package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// IPChangeDetector IP地址变化检测器
type IPChangeDetector struct {
	previousIP string
	currentIP  string
	mutex      sync.RWMutex
}

// NewIPChangeDetector 创建新的IP变化检测器
func NewIPChangeDetector() *IPChangeDetector {
	return &IPChangeDetector{}
}

// CheckIPChange 检测IP地址是否发生变化
func (d *IPChangeDetector) CheckIPChange(newIP string) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// 如果当前IP为空，说明是第一次设置
	if d.currentIP == "" {
		d.currentIP = newIP
		return true
	}

	// 如果IP地址发生变化
	if d.currentIP != newIP {
		d.previousIP = d.currentIP
		d.currentIP = newIP
		return true
	}

	// IP地址未变化
	return false
}

// GetPreviousIP 获取之前的IP地址
func (d *IPChangeDetector) GetPreviousIP() string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.previousIP
}

// GetCurrentIP 获取当前的IP地址
func (d *IPChangeDetector) GetCurrentIP() string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.currentIP
}

// FeishuMessage 飞书消息结构
type FeishuMessage struct {
	MsgType   string         `json:"msg_type"`
	Content   MessageContent `json:"content"`
	Timestamp int64          `json:"timestamp"`
	Sign      string         `json:"sign"`
}

// MessageContent 消息内容
type MessageContent struct {
	Text string `json:"text"`
}

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
	webhookURL string
	secret     string
	httpClient *http.Client
}

// NewFeishuNotifier 创建新的飞书通知器
func NewFeishuNotifier(webhookURL, secret string) *FeishuNotifier {
	// 如果没有传入参数，则从环境变量读取
	if webhookURL == "" {
		webhookURL = os.Getenv("FEISHU_WEBHOOK_URL")
	}
	if secret == "" {
		secret = os.Getenv("FEISHU_SECRET")
	}

	return &FeishuNotifier{
		webhookURL: webhookURL,
		secret:     secret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewFeishuNotifierFromEnv 从环境变量创建飞书通知器
func NewFeishuNotifierFromEnv() *FeishuNotifier {
	return NewFeishuNotifier("", "")
}

// generateSignature 生成飞书机器人签名
func (f *FeishuNotifier) generateSignature(timestamp int64) string {
	// 根据飞书文档：使用timestamp + "\n" + secret作为签名字符串
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, f.secret)
	// 使用空字符串作为待签名内容，stringToSign作为key
	h := hmac.New(sha256.New, []byte(stringToSign))
	h.Write([]byte(""))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature
}

// buildMessage 构建飞书消息
func (f *FeishuNotifier) buildMessage(oldIP, newIP, networkName string) *FeishuMessage {
	timestamp := time.Now().Unix()

	var messageText string
	if oldIP == "" {
		// 首次获取IP地址或WiFi重新连接
		messageText = fmt.Sprintf("🌐 WiFi连接状态通知\n网络：%s\n✅ 已连接，IP地址：%s\n时间：%s",
			networkName, newIP, time.Now().Format("2006-01-02 15:04:05"))
	} else {
		// IP地址发生变化
		messageText = fmt.Sprintf("🌐 WiFi连接状态更新\n网络：%s\n🔄 IP地址变化：%s → %s\n时间：%s",
			networkName, oldIP, newIP, time.Now().Format("2006-01-02 15:04:05"))
	}

	return &FeishuMessage{
		MsgType: "text",
		Content: MessageContent{
			Text: messageText,
		},
		Timestamp: timestamp,
		Sign:      f.generateSignature(timestamp),
	}
}

// buildWiFiReconnectMessage 构建WiFi重新连接消息
func (f *FeishuNotifier) buildWiFiReconnectMessage(ip, networkName string) *FeishuMessage {
	timestamp := time.Now().Unix()

	messageText := fmt.Sprintf("🌐 WiFi重新连接通知\n网络：%s\n✅ 已重新连接，IP地址：%s\n时间：%s",
		networkName, ip, time.Now().Format("2006-01-02 15:04:05"))

	return &FeishuMessage{
		MsgType: "text",
		Content: MessageContent{
			Text: messageText,
		},
		Timestamp: timestamp,
		Sign:      f.generateSignature(timestamp),
	}
}

// SendIPChangeNotification 发送IP变化通知
func (f *FeishuNotifier) SendIPChangeNotification(oldIP, newIP, networkName string) error {
	if f.webhookURL == "" || f.secret == "" {
		return fmt.Errorf("飞书通知配置不完整")
	}

	message := f.buildMessage(oldIP, newIP, networkName)

	// 序列化消息
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 发送HTTP请求
	resp, err := f.httpClient.Post(f.webhookURL, "application/json", bytes.NewBuffer(messageData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	log.Printf("飞书通知发送成功: %s", networkName)
	return nil
}

// SendIPChangeNotificationAsync 异步发送IP变化通知
func (f *FeishuNotifier) SendIPChangeNotificationAsync(oldIP, newIP, networkName string) {
	go func() {
		// 实现重试机制
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			err := f.SendIPChangeNotification(oldIP, newIP, networkName)
			if err == nil {
				return
			}

			log.Printf("飞书通知发送失败 (第%d次重试): %v", i+1, err)
			if i < maxRetries-1 {
				// 指数退避重试
				time.Sleep(time.Duration(i+1) * time.Second)
			}
		}
		log.Printf("飞书通知发送最终失败，已重试%d次", maxRetries)
	}()
}

// SendWiFiReconnectNotification 发送WiFi重新连接通知
func (f *FeishuNotifier) SendWiFiReconnectNotification(ip, networkName string) error {
	if f.webhookURL == "" || f.secret == "" {
		return fmt.Errorf("飞书通知配置不完整")
	}

	message := f.buildWiFiReconnectMessage(ip, networkName)

	// 序列化消息
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 发送HTTP请求
	resp, err := f.httpClient.Post(f.webhookURL, "application/json", bytes.NewBuffer(messageData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	log.Printf("飞书WiFi重新连接通知发送成功: %s", networkName)
	return nil
}

// SendWiFiReconnectNotificationAsync 异步发送WiFi重新连接通知
func (f *FeishuNotifier) SendWiFiReconnectNotificationAsync(ip, networkName string) {
	go func() {
		// 实现重试机制
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			err := f.SendWiFiReconnectNotification(ip, networkName)
			if err == nil {
				return
			}

			log.Printf("飞书WiFi重新连接通知发送失败 (第%d次重试): %v", i+1, err)
			if i < maxRetries-1 {
				// 指数退避重试
				time.Sleep(time.Duration(i+1) * time.Second)
			}
		}
		log.Printf("飞书WiFi重新连接通知发送最终失败，已重试%d次", maxRetries)
	}()
}
