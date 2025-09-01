package main

import (
	"fmt"
	"log"
)

// testWindowsConnector 测试Windows连接器的功能
func testWindowsConnector() {
	fmt.Println("=== 测试Windows WiFi连接器 ===")
	
	// 创建Windows连接器
	connector, err := NewWindowsConnector()
	if err != nil {
		log.Printf("创建Windows连接器失败: %v", err)
		return
	}
	
	fmt.Println("✓ Windows连接器创建成功")
	
	// 获取接口名称
	interfaceName, err := connector.GetInterface()
	if err != nil {
		log.Printf("获取接口名称失败: %v", err)
		return
	}
	fmt.Printf("✓ WiFi接口名称: %s\n", interfaceName)
	
	// 检查WiFi是否启用
	if connector.IsEnabled() {
		fmt.Println("✓ WiFi已启用")
	} else {
		fmt.Println("⚠ WiFi未启用，尝试启用...")
		if err := connector.Enable(); err != nil {
			log.Printf("启用WiFi失败: %v", err)
		} else {
			fmt.Println("✓ WiFi已启用")
		}
	}
	
	// 获取当前网络
	currentNetwork, err := connector.GetCurrentNetwork()
	if err != nil {
		log.Printf("获取当前网络失败: %v", err)
	} else if currentNetwork != "" {
		fmt.Printf("✓ 当前连接的WiFi: %s\n", currentNetwork)
		
		// 获取IP地址
		ipAddr, err := connector.GetIPAddress()
		if err != nil {
			log.Printf("获取IP地址失败: %v", err)
		} else {
			fmt.Printf("✓ 当前IP地址: %s\n", ipAddr)
		}
	} else {
		fmt.Println("⚠ 当前未连接任何WiFi网络")
	}
	
	fmt.Println("=== 测试完成 ===")
}

// 运行测试：go run test_windows.go windows_connector.go
func runTest() {
	testWindowsConnector()
}