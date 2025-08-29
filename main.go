package main

import (
	"flag"
	"log"
	"os"
	"time"
)

// 全局变量存储命令行参数
var (
	// 目标WiFi网络名称
	targetWiFi string
	// WiFi密码
	wifiPassword string
	// 检查间隔时间（秒）
	checkInterval int
	// 只运行一次
	runOnce bool
	// 飞书webhook URL
	feishuWebhook string
	// 飞书机器人签名密钥
	feishuSecret string
	// 是否启用通知功能
	enableNotification bool
	// 总是发送通知（即使IP未变化）
	alwaysNotify bool
	// WiFi连接器实例
	connector WiFiConnector
	// IP变化检测器
	ipDetector *IPChangeDetector
	// 飞书通知器
	feishuNotifier *FeishuNotifier
	// 程序版本
	version string = "1.0.0"
)

// checkAndConnect 检查并连接WiFi的主要逻辑
func checkAndConnect() {
	// 自动检测WiFi网卡接口
	interfaceName, err := connector.GetInterface()
	if err != nil {
		log.Printf("获取WiFi接口失败: %v", err)
		return
	}
	log.Printf("检测到WiFi接口: %s", interfaceName)

	// 检查WiFi是否启用
	if !connector.IsEnabled() {
		log.Println("WiFi未启用，正在启用...")
		if err := connector.Enable(); err != nil {
			log.Printf("启用WiFi失败: %v", err)
			return
		}
		log.Println("WiFi已启用")
		// 等待WiFi启用完成
		time.Sleep(3 * time.Second)
	}

	// 获取当前连接的WiFi
	currentWiFi, err := connector.GetCurrentNetwork()
	if err != nil {
		log.Printf("获取当前WiFi失败: %v", err)
		return
	}

	if currentWiFi == "" {
		log.Println("当前未连接任何WiFi网络")
	} else {
		log.Printf("当前连接的WiFi: %s", currentWiFi)
	}

	// 如果当前WiFi不是目标WiFi，则尝试连接
	if currentWiFi != targetWiFi {
		log.Printf("尝试连接到WiFi: %s", targetWiFi)
		if err := connector.Connect(targetWiFi, wifiPassword); err != nil {
			log.Printf("连接WiFi失败: %v", err)
		} else {
			log.Printf("成功连接到WiFi: %s", targetWiFi)
			// 等待网络配置完成
			time.Sleep(2 * time.Second)
			// 获取并显示IP地址
			if ipAddr, err := connector.GetIPAddress(); err != nil {
				log.Printf("获取IP地址失败: %v", err)
			} else {
				log.Printf("分配到的IP地址: %s", ipAddr)
				// 检测IP地址是否变化并发送通知
				if enableNotification && ipDetector != nil && feishuNotifier != nil {
					log.Printf("检测IP地址变化: %s", ipAddr)
					ipChanged := ipDetector.CheckIPChange(ipAddr)
					if ipChanged || alwaysNotify {
						oldIP := ipDetector.GetPreviousIP()
						if ipChanged {
							log.Printf("发送飞书通知(因IP变化): %s -> %s", oldIP, ipAddr)
						} else {
							log.Printf("发送飞书通知(强制发送): %s", ipAddr)
						}
						feishuNotifier.SendIPChangeNotificationAsync(oldIP, ipAddr, targetWiFi)
					} else {
						log.Printf("IP地址未变化，不发送通知: %s", ipAddr)
					}
				} else {
					if !enableNotification {
						log.Printf("通知功能未启用")
					} else if ipDetector == nil {
						log.Printf("IP检测器未初始化")
					} else if feishuNotifier == nil {
						log.Printf("飞书通知器未初始化")
					}
				}
			}
		}
	} else {
		log.Printf("已连接到目标WiFi: %s", targetWiFi)
		// 显示当前IP地址
		if ipAddr, err := connector.GetIPAddress(); err != nil {
			log.Printf("获取IP地址失败: %v", err)
		} else {
			log.Printf("当前IP地址: %s", ipAddr)
			// 检测IP地址是否变化并发送通知
			if enableNotification && ipDetector != nil && feishuNotifier != nil {
				log.Printf("检测IP地址变化: %s", ipAddr)
				ipChanged := ipDetector.CheckIPChange(ipAddr)
				if ipChanged || alwaysNotify {
					oldIP := ipDetector.GetPreviousIP()
					if ipChanged {
						log.Printf("发送飞书通知(因IP变化): %s -> %s", oldIP, ipAddr)
					} else {
						log.Printf("发送飞书通知(强制发送): %s", ipAddr)
					}
					feishuNotifier.SendIPChangeNotificationAsync(oldIP, ipAddr, targetWiFi)
				} else {
					log.Printf("IP地址未变化，不发送通知: %s", ipAddr)
				}
			} else {
				if !enableNotification {
					log.Printf("通知功能未启用")
				} else if ipDetector == nil {
					log.Printf("IP检测器未初始化")
				} else if feishuNotifier == nil {
					log.Printf("飞书通知器未初始化")
				}
			}
		}
	}
}

func main() {
	// 解析命令行参数
	flag.StringVar(&targetWiFi, "w", "", "目标WiFi网络名称")
	flag.StringVar(&wifiPassword, "p", "", "WiFi密码")
	flag.IntVar(&checkInterval, "i", 10, "检查间隔（秒）")
	flag.BoolVar(&runOnce, "once", false, "只运行一次")
	flag.StringVar(&feishuWebhook, "feishu-webhook", "", "飞书webhook URL")
	flag.StringVar(&feishuSecret, "feishu-secret", "", "飞书机器人签名密钥")
	flag.BoolVar(&enableNotification, "enable-notification", false, "是否启用通知功能")
	flag.BoolVar(&alwaysNotify, "always-notify", false, "总是发送通知（即使IP未变化）")
	flag.Parse()

	// 检查必需参数
	if targetWiFi == "" {
		log.Fatal("请指定目标WiFi网络名称，使用 -w 参数")
	}

	// 初始化通知相关组件
	if enableNotification {
		// 优先从环境变量读取配置
		envWebhook := os.Getenv("FEISHU_WEBHOOK_URL")
		envSecret := os.Getenv("FEISHU_SECRET")

		// 如果环境变量不为空，优先使用环境变量
		if envWebhook != "" {
			feishuWebhook = envWebhook
		}
		if envSecret != "" {
			feishuSecret = envSecret
		}

		// 检查配置是否完整
		if feishuWebhook == "" || feishuSecret == "" {
			log.Println("警告: 启用了通知功能但缺少飞书配置，将禁用通知功能")
			log.Println("可以通过设置环境变量 FEISHU_WEBHOOK_URL 和 FEISHU_SECRET 或使用命令行参数 -feishu-webhook 和 -feishu-secret")
			enableNotification = false
		} else {
			ipDetector = NewIPChangeDetector()
			feishuNotifier = NewFeishuNotifier(feishuWebhook, feishuSecret)
			log.Printf("通知功能已启用: %s", feishuWebhook)
		}
	}

	// 创建WiFi连接器
	var err error
	connector, err = NewWiFiConnector()
	if err != nil {
		log.Fatalf("创建WiFi连接器失败: %v", err)
	}

	log.Println("WiFi自动连接程序启动")
	interfaceName, _ := connector.GetInterface()
	log.Printf("检测到WiFi接口: %s", interfaceName)
	log.Printf("目标WiFi网络: %s", targetWiFi)
	if wifiPassword != "" {
		log.Println("已设置WiFi密码")
	} else {
		log.Println("未设置WiFi密码，将尝试使用已保存的密码")
	}
	log.Printf("检查间隔: %d秒", checkInterval)

	// 执行检查和连接
	checkAndConnect()

	// 如果设置为只运行一次，则退出
	if runOnce {
		log.Println("程序执行完毕")
		return
	}

	// 定期检查
	ticker := time.NewTicker(time.Duration(checkInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		checkAndConnect()
	}
}
