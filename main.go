package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
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
)

// getCurrentWiFi 获取当前连接的WiFi网络名称
func getCurrentWiFi() (string, error) {
	cmd := exec.Command("networksetup", "-getairportnetwork", "en0")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取当前WiFi失败: %v", err)
	}

	result := strings.TrimSpace(string(output))
	if strings.Contains(result, "You are not associated with an AirPort network") {
		return "", nil // 未连接任何WiFi
	}

	// 解析输出格式: "Current Wi-Fi Network: NetworkName"
	parts := strings.Split(result, ": ")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1]), nil
	}

	return "", fmt.Errorf("无法解析WiFi网络名称: %s", result)
}

// connectToWiFi 连接到指定的WiFi网络
func connectToWiFi(networkName, password string) error {
	log.Printf("正在连接到WiFi网络: %s", networkName)
	var cmd *exec.Cmd
	if password != "" {
		// 如果提供了密码，使用密码连接
		cmd = exec.Command("networksetup", "-setairportnetwork", "en0", networkName, password)
	} else {
		// 如果没有提供密码，尝试使用已保存的密码连接
		cmd = exec.Command("networksetup", "-setairportnetwork", "en0", networkName)
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("连接WiFi失败: %v", err)
	}
	log.Printf("成功连接到WiFi网络: %s", networkName)
	return nil
}

// isWiFiEnabled 检查WiFi是否已启用
func isWiFiEnabled() bool {
	cmd := exec.Command("networksetup", "-getairportpower", "en0")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("检查WiFi状态失败: %v", err)
		return false
	}

	result := strings.TrimSpace(string(output))
	return strings.Contains(result, "On")
}

// enableWiFi 启用WiFi
func enableWiFi() error {
	log.Println("正在启用WiFi...")
	cmd := exec.Command("networksetup", "-setairportpower", "en0", "on")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("启用WiFi失败: %v", err)
	}
	log.Println("WiFi已启用")
	// 等待WiFi启动
	time.Sleep(3 * time.Second)
	return nil
}

// checkAndConnect 检查WiFi连接状态并根据需要进行连接
func checkAndConnect() {
	// 检查WiFi是否启用
	if !isWiFiEnabled() {
		log.Println("WiFi未启用，正在启用...")
		if err := enableWiFi(); err != nil {
			log.Printf("启用WiFi失败: %v", err)
			return
		}
	}

	// 获取当前连接的WiFi
	currentWiFi, err := getCurrentWiFi()
	if err != nil {
		log.Printf("获取当前WiFi状态失败: %v", err)
		return
	}

	if currentWiFi == "" {
		// 未连接任何WiFi，连接到目标网络
		log.Println("未连接任何WiFi网络")
		if err := connectToWiFi(targetWiFi, wifiPassword); err != nil {
			log.Printf("连接到目标WiFi失败: %v", err)
		}
	} else if currentWiFi != targetWiFi {
		// 连接到了其他WiFi，切换到目标网络
		log.Printf("当前连接到: %s，切换到目标网络: %s", currentWiFi, targetWiFi)
		if err := connectToWiFi(targetWiFi, wifiPassword); err != nil {
			log.Printf("切换到目标WiFi失败: %v", err)
		}
	} else {
		// 已经连接到目标WiFi
		log.Printf("已连接到目标WiFi: %s", currentWiFi)
	}
}

func main() {
	// 解析命令行参数
	flag.StringVar(&targetWiFi, "w", "qqqq", "目标WiFi网络名称")
	flag.StringVar(&wifiPassword, "p", "", "WiFi密码（可选，如果为空则使用系统保存的密码）")
	flag.IntVar(&checkInterval, "i", 10, "检查间隔时间（秒）")
	flag.Parse()

	// 验证参数
	if targetWiFi == "" {
		log.Fatal("WiFi网络名称不能为空")
	}
	if checkInterval <= 0 {
		log.Fatal("检查间隔必须大于0")
	}

	log.Println("WiFi自动连接程序启动")
	log.Printf("目标WiFi网络: %s", targetWiFi)
	if wifiPassword != "" {
		log.Println("使用提供的密码")
	} else {
		log.Println("使用系统保存的密码")
	}
	log.Printf("检查间隔: %d秒", checkInterval)

	// 立即执行一次检查
	checkAndConnect()

	// 创建定时器，周期性检查
	ticker := time.NewTicker(time.Duration(checkInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkAndConnect()
		}
	}
}
