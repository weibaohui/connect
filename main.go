package main

import (
	"flag"
	"log"
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
	// WiFi连接器实例
	connector WiFiConnector
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
		}
	} else {
		log.Printf("已连接到目标WiFi: %s", targetWiFi)
	}
}

func main() {
	// 解析命令行参数
	flag.StringVar(&targetWiFi, "wifi", "", "目标WiFi网络名称")
	flag.StringVar(&wifiPassword, "password", "", "WiFi密码")
	flag.IntVar(&checkInterval, "interval", 30, "检查间隔（秒）")
	flag.BoolVar(&runOnce, "once", false, "只运行一次")
	flag.Parse()

	// 检查必需参数
	if targetWiFi == "" {
		log.Fatal("请指定目标WiFi网络名称，使用 -wifi 参数")
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

	for {
		select {
		case <-ticker.C:
			checkAndConnect()
		}
	}
}
