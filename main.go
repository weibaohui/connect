package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
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
}

// MacOSConnector macOS平台的WiFi连接器
type MacOSConnector struct {
	interfaceName string
}

// WindowsConnector Windows平台的WiFi连接器
type WindowsConnector struct {
	interfaceName string
}

// LinuxConnector Linux平台的WiFi连接器
type LinuxConnector struct {
	interfaceName string
}

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

// NewWiFiConnector 创建WiFi连接器工厂函数
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

	return nil
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
