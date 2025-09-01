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

// detectInterface Windows平台的WiFi接口检测
func (w *WindowsConnector) detectInterface() (string, error) {
	// 使用netsh命令获取WiFi接口
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取WiFi接口失败: %v", err)
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	
	// 调试信息：打印原始输出
	fmt.Printf("netsh wlan show interfaces 输出:\n%s\n", outputStr)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 支持多语言环境：查找包含"Name"、"名称"或"接口名称"的行
		if (strings.HasPrefix(line, "Name") || 
			strings.HasPrefix(line, "名称") || 
			strings.Contains(strings.ToLower(line), "name")) && 
			strings.Contains(line, ":") {
			
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				interfaceName := strings.TrimSpace(parts[1])
				if interfaceName != "" {
					fmt.Printf("找到WiFi接口: %s\n", interfaceName)
					return interfaceName, nil
				}
			}
		}
	}

	// 如果上述方法失败，尝试使用另一种方法：通过wmic命令获取网络适配器
	cmd2 := exec.Command("wmic", "path", "win32_networkadapter", "where", "NetConnectionID like '%Wi-Fi%' or NetConnectionID like '%无线%' or NetConnectionID like '%WLAN%'", "get", "NetConnectionID", "/format:list")
	output2, err2 := cmd2.Output()
	if err2 == nil {
		lines2 := strings.Split(string(output2), "\n")
		for _, line := range lines2 {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "NetConnectionID=") {
				interfaceName := strings.TrimPrefix(line, "NetConnectionID=")
				interfaceName = strings.TrimSpace(interfaceName)
				if interfaceName != "" {
					fmt.Printf("通过wmic找到WiFi接口: %s\n", interfaceName)
					return interfaceName, nil
				}
			}
		}
	}

	return "", fmt.Errorf("未找到WiFi网络接口。请确保:\n1. WiFi适配器已安装并启用\n2. 以管理员权限运行程序\n3. netsh命令可用")
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