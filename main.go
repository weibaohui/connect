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

// 全局变量存储命令行参数
var (
	// 目标WiFi网络名称
	targetWiFi string
	// WiFi密码
	wifiPassword string
	// 检查间隔时间（秒）
	checkInterval int
	// WiFi网卡接口名称
	wifiInterface string
	// 程序版本
	version string = "1.0.0"
)

// getWiFiInterface 自动检测WiFi网卡接口名称（跨平台）
func getWiFiInterface() (string, error) {
	switch runtime.GOOS {
	case "darwin": // macOS
		return getWiFiInterfaceMacOS()
	case "windows":
		return getWiFiInterfaceWindows()
	case "linux":
		return getWiFiInterfaceLinux()
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// getWiFiInterfaceMacOS macOS平台的WiFi接口检测
func getWiFiInterfaceMacOS() (string, error) {
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

// getWiFiInterfaceWindows Windows平台的WiFi接口检测
func getWiFiInterfaceWindows() (string, error) {
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

// getWiFiInterfaceLinux Linux平台的WiFi接口检测
func getWiFiInterfaceLinux() (string, error) {
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

// getCurrentWiFi 获取当前连接的WiFi网络名称（跨平台）
func getCurrentWiFi() (string, error) {
	switch runtime.GOOS {
	case "darwin": // macOS
		return getCurrentWiFiMacOS()
	case "windows":
		return getCurrentWiFiWindows()
	case "linux":
		return getCurrentWiFiLinux()
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// getCurrentWiFiMacOS macOS平台获取当前WiFi
func getCurrentWiFiMacOS() (string, error) {
	cmd := exec.Command("networksetup", "-getairportnetwork", wifiInterface)
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

// getCurrentWiFiWindows Windows平台获取当前WiFi
func getCurrentWiFiWindows() (string, error) {
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

// getCurrentWiFiLinux Linux平台获取当前WiFi
func getCurrentWiFiLinux() (string, error) {
	// 尝试使用iwgetid命令
	cmd := exec.Command("iwgetid", "-r")
	output, err := cmd.Output()
	if err == nil {
		ssid := strings.TrimSpace(string(output))
		if ssid != "" {
			return ssid, nil
		}
	}

	// 如果iwgetid失败，尝试使用nmcli
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

// connectToWiFi 连接到指定的WiFi网络（跨平台）
func connectToWiFi(networkName, password string) error {
	log.Printf("正在连接到WiFi网络: %s", networkName)
	switch runtime.GOOS {
	case "darwin": // macOS
		return connectToWiFiMacOS(networkName, password)
	case "windows":
		return connectToWiFiWindows(networkName, password)
	case "linux":
		return connectToWiFiLinux(networkName, password)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// connectToWiFiMacOS macOS平台连接WiFi
func connectToWiFiMacOS(networkName, password string) error {
	var cmd *exec.Cmd
	if password != "" {
		// 如果提供了密码，使用密码连接
		cmd = exec.Command("networksetup", "-setairportnetwork", wifiInterface, networkName, password)
	} else {
		// 如果没有提供密码，尝试使用已保存的密码连接
		cmd = exec.Command("networksetup", "-setairportnetwork", wifiInterface, networkName)
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("连接WiFi失败: %v", err)
	}
	log.Printf("成功连接到WiFi网络: %s", networkName)
	return nil
}

// connectToWiFiWindows Windows平台连接WiFi
func connectToWiFiWindows(networkName, password string) error {
	var cmd *exec.Cmd
	if password != "" {
		// 如果提供了密码，使用密码连接
		cmd = exec.Command("netsh", "wlan", "connect", "name="+networkName, "key="+password)
	} else {
		// 如果没有提供密码，尝试使用已保存的密码连接
		cmd = exec.Command("netsh", "wlan", "connect", "name="+networkName)
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("连接WiFi失败: %v", err)
	}
	log.Printf("成功连接到WiFi网络: %s", networkName)
	return nil
}

// connectToWiFiLinux Linux平台连接WiFi
func connectToWiFiLinux(networkName, password string) error {
	// 尝试使用nmcli连接
	var cmd *exec.Cmd
	if password != "" {
		// 如果提供了密码，使用密码连接
		cmd = exec.Command("nmcli", "dev", "wifi", "connect", networkName, "password", password)
	} else {
		// 如果没有提供密码，尝试使用已保存的密码连接
		cmd = exec.Command("nmcli", "dev", "wifi", "connect", networkName)
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("连接WiFi失败: %v", err)
	}
	log.Printf("成功连接到WiFi网络: %s", networkName)
	return nil
}

// isWiFiEnabled 检查WiFi是否已启用（跨平台）
func isWiFiEnabled() bool {
	switch runtime.GOOS {
	case "darwin": // macOS
		return isWiFiEnabledMacOS()
	case "windows":
		return isWiFiEnabledWindows()
	case "linux":
		return isWiFiEnabledLinux()
	default:
		log.Printf("不支持的操作系统: %s", runtime.GOOS)
		return false
	}
}

// isWiFiEnabledMacOS macOS平台检查WiFi状态
func isWiFiEnabledMacOS() bool {
	cmd := exec.Command("networksetup", "-getairportpower", wifiInterface)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("检查WiFi状态失败: %v", err)
		return false
	}

	result := strings.TrimSpace(string(output))
	return strings.Contains(result, "On")
}

// isWiFiEnabledWindows Windows平台检查WiFi状态
func isWiFiEnabledWindows() bool {
	cmd := exec.Command("netsh", "interface", "show", "interface", wifiInterface)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("检查WiFi状态失败: %v", err)
		return false
	}

	result := strings.TrimSpace(string(output))
	return strings.Contains(result, "Connected")
}

// isWiFiEnabledLinux Linux平台检查WiFi状态
func isWiFiEnabledLinux() bool {
	// 检查网络接口是否启用
	cmd := exec.Command("ip", "link", "show", wifiInterface)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("检查WiFi状态失败: %v", err)
		return false
	}

	result := strings.TrimSpace(string(output))
	return strings.Contains(result, "UP")
}

// enableWiFi 启用WiFi（跨平台）
func enableWiFi() error {
	log.Println("正在启用WiFi...")
	switch runtime.GOOS {
	case "darwin": // macOS
		return enableWiFiMacOS()
	case "windows":
		return enableWiFiWindows()
	case "linux":
		return enableWiFiLinux()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// enableWiFiMacOS macOS平台启用WiFi
func enableWiFiMacOS() error {
	cmd := exec.Command("networksetup", "-setairportpower", wifiInterface, "on")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("启用WiFi失败: %v", err)
	}
	log.Println("WiFi已启用")
	// 等待WiFi启动
	time.Sleep(3 * time.Second)
	return nil
}

// enableWiFiWindows Windows平台启用WiFi
func enableWiFiWindows() error {
	cmd := exec.Command("netsh", "interface", "set", "interface", wifiInterface, "enable")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("启用WiFi失败: %v", err)
	}
	log.Println("WiFi已启用")
	// 等待WiFi启动
	time.Sleep(3 * time.Second)
	return nil
}

// enableWiFiLinux Linux平台启用WiFi
func enableWiFiLinux() error {
	// 尝试启用网络接口
	cmd := exec.Command("ip", "link", "set", wifiInterface, "up")
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
	flag.StringVar(&wifiPassword, "p", "", "WiFi密码")
	flag.IntVar(&checkInterval, "i", 10, "检查间隔时间（秒）")
	flag.Parse()

	// 验证参数
	if targetWiFi == "" {
		log.Fatal("WiFi网络名称不能为空")
	}
	if checkInterval <= 0 {
		log.Fatal("检查间隔必须大于0")
	}

	// 自动检测WiFi网卡接口
	var err error
	wifiInterface, err = getWiFiInterface()
	if err != nil {
		log.Fatalf("检测WiFi网卡接口失败: %v", err)
	}

	log.Println("WiFi自动连接程序启动")
	log.Printf("检测到WiFi网卡接口: %s", wifiInterface)
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
