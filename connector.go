package main

import (
	"fmt"
	"runtime"
)

// WiFiConnector 定义WiFi连接器接口
type WiFiConnector interface {
	// GetInterface 获取WiFi网卡接口名称
	GetInterface() (string, error)
	// GetCurrentNetwork 获取当前连接的WiFi网络名称
	GetCurrentNetwork() (string, error)
	// Connect 连接到指定的WiFi网络
	Connect(networkName, password string) error
	// IsEnabled 检查WiFi是否已启用
	IsEnabled() bool
	// Enable 启用WiFi
	Enable() error
	// GetIPAddress 获取当前WiFi接口的IP地址
	GetIPAddress() (string, error)
}

// NewWiFiConnector 根据操作系统创建对应的WiFi连接器
func NewWiFiConnector() (WiFiConnector, error) {
	switch runtime.GOOS {
	case "darwin": // macOS
		return NewMacOSConnector()
	case "windows":
		return NewWindowsConnector()
	case "linux":
		return NewLinuxConnector()
	default:
		return nil, fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}