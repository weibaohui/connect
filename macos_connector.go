package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// MacOSConnector macOS平台的WiFi连接器
type MacOSConnector struct {
	interfaceName string
}

// NewMacOSConnector 创建macOS连接器
func NewMacOSConnector() (*MacOSConnector, error) {
	connector := &MacOSConnector{}
	interfaceName, err := connector.detectInterface()
	if err != nil {
		return nil, err
	}
	connector.interfaceName = interfaceName
	return connector, nil
}

// detectInterface macOS平台的WiFi接口检测
func (m *MacOSConnector) detectInterface() (string, error) {
	cmd := exec.Command("networksetup", "-listallhardwareports")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取网络接口列表失败: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if strings.Contains(line, "Wi-Fi") || strings.Contains(line, "AirPort") {
			// 找到WiFi接口，下一行应该是Device
			if i+1 < len(lines) {
				deviceLine := lines[i+1]
				if strings.HasPrefix(deviceLine, "Device: ") {
					device := strings.TrimPrefix(deviceLine, "Device: ")
					return strings.TrimSpace(device), nil
				}
			}
		}
	}

	// 如果没有找到WiFi接口，尝试常见的接口名称
	commonInterfaces := []string{"en0", "en1", "en2"}
	for _, iface := range commonInterfaces {
		cmd := exec.Command("networksetup", "-getairportpower", iface)
		if err := cmd.Run(); err == nil {
			return iface, nil
		}
	}

	return "", fmt.Errorf("未找到WiFi网络接口")
}

// GetInterface 实现WiFiConnector接口 - 获取WiFi接口名称
func (m *MacOSConnector) GetInterface() (string, error) {
	return m.interfaceName, nil
}

// GetCurrentNetwork 实现WiFiConnector接口 - 获取当前WiFi网络
func (m *MacOSConnector) GetCurrentNetwork() (string, error) {
	cmd := exec.Command("networksetup", "-getairportnetwork", m.interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取当前WiFi失败: %v", err)
	}

	result := strings.TrimSpace(string(output))
	if strings.Contains(result, "You are not associated with an AirPort network") {
		return "", nil // 未连接任何WiFi
	}

	// 解析输出格式: "Current Wi-Fi Network: NetworkName"
	if strings.HasPrefix(result, "Current Wi-Fi Network: ") {
		networkName := strings.TrimPrefix(result, "Current Wi-Fi Network: ")
		return strings.TrimSpace(networkName), nil
	}

	return "", fmt.Errorf("无法解析WiFi网络名称: %s", result)
}

// Connect 实现WiFiConnector接口 - 连接WiFi网络
func (m *MacOSConnector) Connect(networkName, password string) error {
	var cmd *exec.Cmd
	if password != "" {
		cmd = exec.Command("networksetup", "-setairportnetwork", m.interfaceName, networkName, password)
	} else {
		cmd = exec.Command("networksetup", "-setairportnetwork", m.interfaceName, networkName)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("连接WiFi失败: %v", err)
	}

	// 等待连接完成并验证连接结果
	for i := 0; i < 10; i++ { // 最多等待10秒
		time.Sleep(1 * time.Second)
		currentNetwork, err := m.GetCurrentNetwork()
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
func (m *MacOSConnector) IsEnabled() bool {
	cmd := exec.Command("networksetup", "-getairportpower", m.interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "On")
}

// Enable 实现WiFiConnector接口 - 启用WiFi
func (m *MacOSConnector) Enable() error {
	cmd := exec.Command("networksetup", "-setairportpower", m.interfaceName, "on")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启用WiFi失败: %v", err)
	}
	return nil
}

// GetIPAddress 实现WiFiConnector接口 - 获取当前WiFi接口的IP地址
func (m *MacOSConnector) GetIPAddress() (string, error) {
	cmd := exec.Command("ifconfig", m.interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取IP地址失败: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") && !strings.Contains(line, "127.0.0.1") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("未找到IP地址")
}
