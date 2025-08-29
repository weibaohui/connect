package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// LinuxConnector Linux平台的WiFi连接器
type LinuxConnector struct {
	interfaceName string
}

// NewLinuxConnector 创建Linux连接器
func NewLinuxConnector() (*LinuxConnector, error) {
	connector := &LinuxConnector{}
	interfaceName, err := connector.detectInterface()
	if err != nil {
		return nil, err
	}
	connector.interfaceName = interfaceName
	return connector, nil
}

// detectInterface Linux平台的WiFi接口检测
func (l *LinuxConnector) detectInterface() (string, error) {
	// 尝试常见的WiFi接口名称
	commonInterfaces := []string{"wlan0", "wlp2s0", "wlp3s0", "wlo1"}
	for _, iface := range commonInterfaces {
		// 检查接口是否存在
		cmd := exec.Command("ip", "link", "show", iface)
		if err := cmd.Run(); err == nil {
			return iface, nil
		}
	}

	return "", fmt.Errorf("未找到WiFi网络接口")
}

// GetInterface 实现WiFiConnector接口 - 获取WiFi接口名称
func (l *LinuxConnector) GetInterface() (string, error) {
	return l.interfaceName, nil
}

// GetCurrentNetwork 实现WiFiConnector接口 - 获取当前WiFi网络
func (l *LinuxConnector) GetCurrentNetwork() (string, error) {
	// 优先使用iwgetid命令
	cmd := exec.Command("iwgetid", "-r")
	output, err := cmd.Output()
	if err == nil {
		networkName := strings.TrimSpace(string(output))
		if networkName != "" {
			return networkName, nil
		}
	}

	// 备用方案：使用nmcli
	cmd = exec.Command("nmcli", "-t", "-f", "active,ssid", "dev", "wifi")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取当前WiFi失败: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) >= 2 && parts[0] == "yes" {
			return parts[1], nil
		}
	}

	return "", nil // 未连接任何WiFi
}

// Connect 实现WiFiConnector接口 - 连接WiFi网络
func (l *LinuxConnector) Connect(networkName, password string) error {
	var cmd *exec.Cmd
	if password != "" {
		cmd = exec.Command("nmcli", "dev", "wifi", "connect", networkName, "password", password)
	} else {
		cmd = exec.Command("nmcli", "dev", "wifi", "connect", networkName)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("连接WiFi失败: %v", err)
	}

	return nil
}

// IsEnabled 实现WiFiConnector接口 - 检查WiFi是否启用
func (l *LinuxConnector) IsEnabled() bool {
	cmd := exec.Command("ip", "link", "show", l.interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "UP")
}

// Enable 实现WiFiConnector接口 - 启用WiFi
func (l *LinuxConnector) Enable() error {
	cmd := exec.Command("ip", "link", "set", l.interfaceName, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启用WiFi失败: %v", err)
	}
	return nil
}