package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
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
func (w *WindowsConnector) detectInterface() (string, error) {
	// 使用 netsh
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	output, err := cmd.Output()
	if err == nil {
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")

		fmt.Printf("netsh wlan show interfaces 输出:\n%s\n", outputStr)

		for _, line := range lines {
			line = strings.TrimSpace(line)
			// 只匹配开头，避免 Description 误判
			if strings.HasPrefix(line, "Name") || strings.HasPrefix(line, "名称") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					ifName := strings.TrimSpace(parts[1])
					if ifName != "" {
						fmt.Printf("找到WiFi接口: %s\n", ifName)
						return ifName, nil
					}
				}
			}
		}
	}

	// fallback: 使用 PowerShell
	cmd2 := exec.Command("powershell", "-Command",
		`Get-NetAdapter | Where-Object {$_.Name -match 'Wi-Fi|无线|WLAN'} | Select-Object -ExpandProperty Name`)
	output2, err2 := cmd2.Output()
	if err2 == nil {
		ifName := strings.TrimSpace(string(output2))
		if ifName != "" {
			fmt.Printf("通过 PowerShell 找到 WiFi 接口: %s\n", ifName)
			return ifName, nil
		}
	}

	return "", fmt.Errorf("未找到 WiFi 网络接口。请确保:\n1. WiFi 适配器已安装并启用\n2. 以管理员权限运行程序\n3. netsh / PowerShell 命令可用")
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
		// 支持多语言环境：查找SSID相关的行
		if (strings.HasPrefix(line, "SSID") ||
			strings.HasPrefix(line, "网络名称") ||
			strings.Contains(strings.ToLower(line), "ssid")) &&
			strings.Contains(line, ":") &&
			!strings.Contains(strings.ToLower(line), "bssid") { // 排除BSSID行

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

	// 等待连接完成并验证连接结果
	for i := 0; i < 10; i++ { // 最多等待10秒
		time.Sleep(1 * time.Second)
		currentNetwork, err := w.GetCurrentNetwork()
		if err != nil {
			continue
		}
		if currentNetwork == networkName {
			return nil // 连接成功
		}
	}

	// 如果10秒后仍未连接到目标网络，返回错误
	return fmt.Errorf("连接超时：无法连接到WiFi网络 '%s'，可能网络不存在或密码错误", networkName)
}

// IsEnabled 实现WiFiConnector接口 - 检查WiFi是否启用
func (w *WindowsConnector) IsEnabled() bool {
	cmd := exec.Command("netsh", "interface", "show", "interface", w.interfaceName)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("检查WiFi状态失败: %v\n", err)
		return false
	}

	outputStr := string(output)
	// 支持多语言环境：检查"Enabled"、"已启用"、"Connected"、"已连接"等状态
	return strings.Contains(outputStr, "Enabled") ||
		strings.Contains(outputStr, "已启用") ||
		strings.Contains(outputStr, "Connected") ||
		strings.Contains(outputStr, "已连接")
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
		// 支持多语言环境：查找IP地址相关的行
		if (strings.Contains(line, "IP Address:") ||
			strings.Contains(line, "IP 地址:") ||
			strings.Contains(line, "IPv4 Address") ||
			strings.Contains(line, "IPv4 地址")) &&
			strings.Contains(line, ":") {

			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ipAddr := strings.TrimSpace(parts[1])
				// 移除可能的额外信息（如子网掩码）
				if strings.Contains(ipAddr, "(") {
					ipAddr = strings.Split(ipAddr, "(")[0]
					ipAddr = strings.TrimSpace(ipAddr)
				}
				if ipAddr != "" && ipAddr != "127.0.0.1" {
					return ipAddr, nil
				}
			}
		}
	}

	return "", fmt.Errorf("未找到IP地址")
}
