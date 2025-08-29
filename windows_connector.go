package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// WindowsConnector Windows平台的WiFi连接器
type WindowsConnector struct {
	interfaceName string
}

// NewWindowsConnector 创建Windows连接器
func NewWindowsConnector() (*WindowsConnector, error) {
	connector := &WindowsConnector{}
	interfaceName, err := connector.detectInterface()
	if err != nil {
		return nil, err
	}
	connector.interfaceName = interfaceName
	return connector, nil
}

// detectInterface Windows平台的WiFi接口检测
func (w *WindowsConnector) detectInterface() (string, error) {
	// 使用netsh命令获取WiFi接口
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取WiFi接口失败: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Name") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				interfaceName := strings.TrimSpace(parts[1])
				if interfaceName != "" {
					return interfaceName, nil
				}
			}
		}
	}

	return "", fmt.Errorf("未找到WiFi网络接口")
}

// GetInterface 实现WiFiConnector接口 - 获取WiFi接口名称
func (w *WindowsConnector) GetInterface() (string, error) {
	return w.interfaceName, nil
}

// GetCurrentNetwork 实现WiFiConnector接口 - 获取当前WiFi网络
func (w *WindowsConnector) GetCurrentNetwork() (string, error) {
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取当前WiFi失败: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SSID") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ssid := strings.TrimSpace(parts[1])
				if ssid != "" {
					return ssid, nil
				}
			}
		}
	}

	return "", nil // 未连接任何WiFi
}

// Connect 实现WiFiConnector接口 - 连接WiFi网络
func (w *WindowsConnector) Connect(networkName, password string) error {
	var cmd *exec.Cmd
	if password != "" {
		cmd = exec.Command("netsh", "wlan", "connect", "name="+networkName, "key="+password)
	} else {
		cmd = exec.Command("netsh", "wlan", "connect", "name="+networkName)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("连接WiFi失败: %v", err)
	}

	return nil
}

// IsEnabled 实现WiFiConnector接口 - 检查WiFi是否启用
func (w *WindowsConnector) IsEnabled() bool {
	cmd := exec.Command("netsh", "interface", "show", "interface", w.interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "Enabled")
}

// Enable 实现WiFiConnector接口 - 启用WiFi
func (w *WindowsConnector) Enable() error {
	cmd := exec.Command("netsh", "interface", "set", "interface", w.interfaceName, "enable")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启用WiFi失败: %v", err)
	}
	return nil
}

// GetIPAddress 实现WiFiConnector接口 - 获取当前WiFi接口的IP地址
func (w *WindowsConnector) GetIPAddress() (string, error) {
	cmd := exec.Command("netsh", "interface", "ip", "show", "address", w.interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取IP地址失败: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "IP Address:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ipAddr := strings.TrimSpace(parts[1])
				if ipAddr != "" && ipAddr != "127.0.0.1" {
					return ipAddr, nil
				}
			}
		}
	}

	return "", fmt.Errorf("未找到IP地址")
}