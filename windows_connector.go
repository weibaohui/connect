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

// executePowerShellCommand 执行PowerShell命令并返回输出结果
func (w *WindowsConnector) executePowerShellCommand(command string) (string, error) {
	fmt.Printf("[DEBUG] 执行PowerShell命令: %s\n", command)
	// 尝试多种PowerShell调用方式以提高兼容性
	var cmd *exec.Cmd

	// 方式1：使用-NoProfile -ExecutionPolicy Bypass参数
	cmd = exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; "+command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[DEBUG] 方式1失败，尝试方式2\n")
		// 方式2：不使用编码设置
		cmd = exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", command)
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("[DEBUG] 方式2失败，尝试方式3\n")
			// 方式3：使用基本的powershell命令
			cmd = exec.Command("powershell", "-Command", command)
			output, err = cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("[ERROR] 所有PowerShell调用方式都失败: %v\n", err)
				fmt.Printf("[ERROR] 最后一次命令输出: %s\n", string(output))
				return "", err
			}
		}
	}
	result := strings.TrimSpace(string(output))
	fmt.Printf("[DEBUG] PowerShell命令输出: %s\n", result)
	return result, nil
}

// selectBestWiFiInterface 从PowerShell命令输出中解析并选择最佳的WiFi接口
// 参数 output: PowerShell命令的原始输出（可能包含多行）
// 返回值: 最佳的WiFi接口名称，如果没有找到则返回空字符串
func (w *WindowsConnector) selectBestWiFiInterface(output string) string {
	fmt.Printf("[DEBUG] 智能选择WiFi接口，原始输出: %s\n", output)
	// 处理多行输出，提取所有有效接口
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var validInterfaces []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			validInterfaces = append(validInterfaces, line)
		}
	}

	fmt.Printf("[DEBUG] 解析出的接口列表: %v\n", validInterfaces)
	// 如果没有找到任何接口，返回空字符串
	if len(validInterfaces) == 0 {
		fmt.Printf("[DEBUG] 没有找到任何接口\n")
		return ""
	}

	// WiFi接口优先级排序（从高到低）
	wifiPriority := []string{
		"WLAN", "Wi-Fi", "WiFi", "无线", "Wireless",
	}

	// 首先按优先级查找
	for _, priority := range wifiPriority {
		for _, iface := range validInterfaces {
			if strings.Contains(strings.ToLower(iface), strings.ToLower(priority)) {
				fmt.Printf("[DEBUG] 通过优先级匹配找到接口: %s (匹配关键词: %s)\n", iface, priority)
				return iface
			}
		}
	}

	// 如果没有找到优先级匹配的，排除明显的以太网接口
	fmt.Printf("[DEBUG] 未找到优先级匹配，开始排除非WiFi接口\n")
	for _, iface := range validInterfaces {
		lowerIface := strings.ToLower(iface)
		// 排除以太网接口
		if strings.Contains(lowerIface, "以太网") && !strings.Contains(lowerIface, "wi") && !strings.Contains(lowerIface, "wireless") {
			fmt.Printf("[DEBUG] 排除以太网接口: %s\n", iface)
			continue
		}
		if strings.Contains(lowerIface, "ethernet") && !strings.Contains(lowerIface, "wi") && !strings.Contains(lowerIface, "wireless") {
			fmt.Printf("[DEBUG] 排除以太网接口: %s\n", iface)
			continue
		}
		// 排除蓝牙接口
		if strings.Contains(lowerIface, "蓝牙") || strings.Contains(lowerIface, "bluetooth") {
			fmt.Printf("[DEBUG] 排除蓝牙接口: %s\n", iface)
			continue
		}
		fmt.Printf("[DEBUG] 选择接口: %s\n", iface)
		return iface
	}

	// 如果都被排除了，返回第一个（作为最后的备用方案）
	if len(validInterfaces) > 0 {
		fmt.Printf("[DEBUG] 所有接口都被排除，使用第一个作为备用方案: %s\n", validInterfaces[0])
		return validInterfaces[0]
	}
	fmt.Printf("[DEBUG] 没有可用的接口\n")
	return ""
}

// detectInterface 检测WiFi网络接口
func (w *WindowsConnector) detectInterface() (string, error) {
	fmt.Printf("开始检测WiFi接口...\n")

	// 方法1：获取所有网络适配器并显示调试信息
	fmt.Printf("[INFO] 方法1: 获取所有网络适配器信息\n")
	command := `Get-NetAdapter | Format-Table Name, InterfaceDescription, MediaType, Status -AutoSize`
	allAdapters, err := w.executePowerShellCommand(command)
	if err == nil {
		fmt.Printf("所有网络适配器:\n%s\n", allAdapters)
	} else {
		fmt.Printf("[ERROR] 方法1失败: %v\n", err)
	}

	// 方法2：按名称匹配WiFi接口（扩展匹配模式）
	fmt.Printf("[INFO] 方法2: 按名称匹配WiFi接口\n")
	command2 := `Get-NetAdapter | Where-Object {$_.Name -match 'Wi-Fi|无线|WLAN|WiFi|Wireless|以太网|Ethernet.*Wi|Wi.*Fi'} | Select-Object -ExpandProperty Name`
	ifName, err2 := w.executePowerShellCommand(command2)
	if err2 == nil && ifName != "" {
		// 智能选择最佳WiFi接口（内部处理多行输出）
		bestInterface := w.selectBestWiFiInterface(ifName)
		if bestInterface != "" {
			fmt.Printf("通过名称匹配找到 WiFi 接口: %s\n", bestInterface)
			return bestInterface, nil
		}
		fmt.Printf("[WARN] 方法2找到接口但未通过智能选择: %s\n", ifName)
	} else {
		fmt.Printf("[ERROR] 方法2失败: %v\n", err2)
	}

	// 方法3：通过媒体类型查找（不限制状态）
	fmt.Printf("[INFO] 方法3: 通过媒体类型查找WiFi接口\n")
	command3 := `Get-NetAdapter | Where-Object {$_.MediaType -eq 'Native 802.11'} | Select-Object -ExpandProperty Name`
	ifName2, err3 := w.executePowerShellCommand(command3)
	if err3 == nil && ifName2 != "" {
		// 智能选择最佳WiFi接口（内部处理多行输出）
		bestInterface := w.selectBestWiFiInterface(ifName2)
		if bestInterface != "" {
			fmt.Printf("通过媒体类型找到 WiFi 接口: %s\n", bestInterface)
			return bestInterface, nil
		}
		fmt.Printf("[WARN] 方法3找到接口但未通过智能选择: %s\n", ifName2)
	} else {
		fmt.Printf("[ERROR] 方法3失败: %v\n", err3)
	}

	// 方法4：通过接口描述查找
	fmt.Printf("[INFO] 方法4: 通过接口描述查找WiFi接口\n")
	command4 := `Get-NetAdapter | Where-Object {$_.InterfaceDescription -match 'Wireless|Wi-Fi|802.11|WiFi'} | Select-Object -ExpandProperty Name`
	ifName3, err4 := w.executePowerShellCommand(command4)
	if err4 == nil && ifName3 != "" {
		// 智能选择最佳WiFi接口（内部处理多行输出）
		bestInterface := w.selectBestWiFiInterface(ifName3)
		if bestInterface != "" {
			fmt.Printf("通过接口描述找到 WiFi 接口: %s\n", bestInterface)
			return bestInterface, nil
		}
		fmt.Printf("[WARN] 方法4找到接口但未通过智能选择: %s\n", ifName3)
	} else {
		fmt.Printf("[ERROR] 方法4失败: %v\n", err4)
	}

	// 方法5：使用WMI查询（更兼容的方式）
	fmt.Printf("[INFO] 方法5: 使用WMI查询WiFi接口\n")
	command5 := `Get-WmiObject -Class Win32_NetworkAdapter | Where-Object {$_.Name -match 'Wireless|Wi-Fi|无线|WLAN|802.11' -and $_.NetConnectionID -ne $null} | Select-Object -ExpandProperty NetConnectionID`
	ifName4, err5 := w.executePowerShellCommand(command5)
	if err5 == nil && ifName4 != "" {
		// 智能选择最佳WiFi接口（内部处理多行输出）
		bestInterface := w.selectBestWiFiInterface(ifName4)
		if bestInterface != "" {
			fmt.Printf("通过WMI找到 WiFi 接口: %s\n", bestInterface)
			return bestInterface, nil
		}
		fmt.Printf("[WARN] 方法5找到接口但未通过智能选择: %s\n", ifName4)
	} else {
		fmt.Printf("[ERROR] 方法5失败: %v\n", err5)
	}

	// 方法6：获取第一个可用的网络适配器（最后的备用方案）
	fmt.Printf("[INFO] 方法6: 获取可用的网络适配器\n")
	command6 := `Get-NetAdapter | Where-Object {$_.Status -eq 'Up'} | Select-Object -ExpandProperty Name`
	ifName5, err6 := w.executePowerShellCommand(command6)
	if err6 == nil && ifName5 != "" {
		// 智能选择最佳WiFi接口（内部处理多行输出）
		bestInterface := w.selectBestWiFiInterface(ifName5)
		if bestInterface != "" {
			fmt.Printf("使用最佳可用适配器作为 WiFi 接口: %s\n", bestInterface)
			return bestInterface, nil
		}
		fmt.Printf("[WARN] 方法6找到接口但未通过智能选择: %s\n", ifName5)
	} else {
		fmt.Printf("[ERROR] 方法6失败: %v\n", err6)
	}

	// 显示详细的错误信息
	fmt.Printf("所有检测方法都失败了:\n")
	fmt.Printf("- 方法1错误: %v\n", err)
	fmt.Printf("- 方法2错误: %v\n", err2)
	fmt.Printf("- 方法3错误: %v\n", err3)
	fmt.Printf("- 方法4错误: %v\n", err4)
	fmt.Printf("- 方法5错误: %v\n", err5)
	fmt.Printf("- 方法6错误: %v\n", err6)

	return "", fmt.Errorf("未找到 WiFi 网络接口。请确保:\n1. WiFi 适配器已安装并启用\n2. 以管理员权限运行程序\n3. PowerShell 命令可用\n4. 检查上述调试信息确认适配器状态")
}

// GetInterface 实现WiFiConnector接口 - 获取WiFi接口名称
func (w *WindowsConnector) GetInterface() (string, error) {
	return w.interfaceName, nil
}

// GetCurrentNetwork 实现WiFiConnector接口 - 获取当前WiFi网络
func (w *WindowsConnector) GetCurrentNetwork() (string, error) {
	// 使用PowerShell获取当前连接的WiFi网络
	command := `(Get-NetConnectionProfile | Where-Object {$_.InterfaceAlias -eq '` + w.interfaceName + `'}).Name`
	ssid, err := w.executePowerShellCommand(command)
	if err == nil && ssid != "" {
		// 清理SSID名称，去除特殊状态信息
		ssid = strings.TrimSpace(ssid)
		// 如果SSID包含"正在识别"，则认为网络正在切换中
		if strings.Contains(ssid, "正在识别") {
			fmt.Printf("[DEBUG] 网络正在识别中: %s\n", ssid)
			return "正在识别", nil
		}
		// 如果SSID以"正在识别"开头，则认为未连接
		if strings.HasPrefix(ssid, "正在识别") {
			return "正在识别", nil
		}
		fmt.Printf("[DEBUG] 获取到当前网络: %s\n", ssid)
		return ssid, nil
	}

	// 备用方法：通过WiFi配置文件获取
	command2 := `(netsh wlan show interfaces | Select-String 'SSID' | Select-String -NotMatch 'BSSID').ToString().Split(':')[1].Trim()`
	ssid2, err2 := w.executePowerShellCommand(command2)
	if err2 == nil && ssid2 != "" {
		ssid2 = strings.TrimSpace(ssid2)
		if strings.Contains(ssid2, "正在识别") {
			fmt.Printf("[DEBUG] 网络正在识别中（备用方法）: %s\n", ssid2)
			return "正在识别", nil
		}
		fmt.Printf("[DEBUG] 获取到当前网络（备用方法）: %s\n", ssid2)
		return ssid2, nil
	}

	// 最后尝试：使用WMI查询
	command3 := `(Get-WmiObject -Class Win32_NetworkAdapterConfiguration | Where-Object {$_.Description -match 'Wireless|Wi-Fi' -and $_.IPEnabled -eq $true}).Description`
	result, err3 := w.executePowerShellCommand(command3)
	if err3 != nil || result == "" {
		fmt.Printf("[DEBUG] 未连接任何WiFi网络\n")
		return "", nil // 未连接任何WiFi
	}

	fmt.Printf("[DEBUG] 获取到当前网络（WMI方法）: %s\n", result)
	return result, nil
}

// Connect 实现WiFiConnector接口 - 连接WiFi网络
func (w *WindowsConnector) Connect(networkName, password string) error {
	// 使用PowerShell连接WiFi
	var command string
	if password != "" {
		// 有密码的网络
		command = fmt.Sprintf(`$profile = netsh wlan show profiles name="%s" key=clear; if ($profile -match "Key Content") { netsh wlan connect name="%s" } else { $xml = @"
<?xml version="1.0"?>
<WLANProfile xmlns="http://www.microsoft.com/networking/WLAN/profile/v1">
	<name>%s</name>
	<SSIDConfig>
		<SSID>
			<name>%s</name>
		</SSID>
	</SSIDConfig>
	<connectionType>ESS</connectionType>
	<connectionMode>auto</connectionMode>
	<MSM>
		<security>
			<authEncryption>
				<authentication>WPA2PSK</authentication>
				<encryption>AES</encryption>
				<useOneX>false</useOneX>
			</authEncryption>
			<sharedKey>
				<keyType>passPhrase</keyType>
				<protected>false</protected>
				<keyMaterial>%s</keyMaterial>
			</sharedKey>
		</security>
	</MSM>
</WLANProfile>
"@; $xml | Out-File -FilePath "$env:TEMP\wifi_profile.xml" -Encoding UTF8; netsh wlan add profile filename="$env:TEMP\wifi_profile.xml"; netsh wlan connect name="%s" }`, networkName, networkName, networkName, networkName, password, networkName)
	} else {
		// 无密码的网络
		command = fmt.Sprintf(`netsh wlan connect name="%s"`, networkName)
	}

	_, err := w.executePowerShellCommand(command)
	if err != nil {
		return fmt.Errorf("连接WiFi失败: %v", err)
	}

	// 等待连接完成并验证连接结果
	// 增加等待时间并改进验证逻辑
	for i := 0; i < 20; i++ { // 增加到20秒以确保有足够时间完成连接
		time.Sleep(1 * time.Second)
		fmt.Printf("[DEBUG] 第%d次检查连接状态\n", i+1)

		// 使用多种方法检查连接状态
		currentNetwork, err := w.GetCurrentNetwork()
		if err != nil {
			fmt.Printf("[DEBUG] 获取当前网络失败: %v\n", err)
			continue
		}

		fmt.Printf("[DEBUG] 当前网络: '%s', 目标网络: '%s'\n", currentNetwork, networkName)

		// 检查是否已经连接到目标网络
		if currentNetwork == networkName {
			fmt.Printf("[DEBUG] 成功连接到目标网络\n")
			return nil // 连接成功
		}

		// 检查特殊情况：网络正在识别中，继续等待
		if strings.Contains(currentNetwork, "正在识别") {
			fmt.Printf("[DEBUG] 网络正在识别中，继续等待\n")
			continue
		}

		// 检查网络名称是否部分匹配（处理可能的空格或特殊字符差异）
		if strings.Contains(strings.TrimSpace(currentNetwork), strings.TrimSpace(networkName)) ||
			strings.Contains(strings.TrimSpace(networkName), strings.TrimSpace(currentNetwork)) {
			fmt.Printf("[DEBUG] 网络名称部分匹配，认为连接成功\n")
			return nil
		}
	}

	// 如果20秒后仍未连接到目标网络，返回错误
	return fmt.Errorf("连接超时：无法连接到WiFi网络 '%s'，可能网络不存在或密码错误", networkName)
}

// IsEnabled 实现WiFiConnector接口 - 检查WiFi是否启用
func (w *WindowsConnector) IsEnabled() bool {
	// 首先检查WiFi接口是否已连接到网络
	currentNetwork, err := w.GetCurrentNetwork()
	if err == nil && currentNetwork != "" {
		// 如果能获取到当前网络名称，说明WiFi已启用且已连接
		fmt.Printf("WiFi已连接到网络: %s，判断为已启用\n", currentNetwork)
		return true
	}

	// 使用PowerShell检查WiFi适配器状态
	command := fmt.Sprintf(`(Get-NetAdapter -Name "%s").Status`, w.interfaceName)
	status, err := w.executePowerShellCommand(command)
	if err != nil {
		fmt.Printf("检查WiFi状态失败: %v\n", err)
		return false
	}

	fmt.Printf("WiFi接口 %s 状态: %s\n", w.interfaceName, status)

	// 检查状态是否为Up（启用）
	if strings.Contains(strings.ToLower(status), "up") {
		fmt.Printf("WiFi接口状态为Up，判断为已启用\n")
		return true
	}

	// 检查是否为禁用状态
	if strings.Contains(strings.ToLower(status), "disabled") || strings.Contains(strings.ToLower(status), "down") {
		fmt.Printf("WiFi接口状态为禁用\n")
		return false
	}

	// 备用方法：检查是否能获取到WiFi配置文件
	command2 := `netsh wlan show profiles | Select-String "All User Profile"`
	profiles, err2 := w.executePowerShellCommand(command2)
	if err2 == nil && profiles != "" {
		fmt.Printf("能够获取WiFi配置文件，判断为已启用\n")
		return true
	}

	// 默认认为是启用的
	fmt.Printf("无法确定状态，默认判断为已启用\n")
	return true
}

// Enable 实现WiFiConnector接口 - 启用WiFi
func (w *WindowsConnector) Enable() error {
	// 使用PowerShell启用WiFi适配器
	command := fmt.Sprintf(`Enable-NetAdapter -Name "%s" -Confirm:$false`, w.interfaceName)
	_, err := w.executePowerShellCommand(command)
	if err != nil {
		return fmt.Errorf("启用WiFi失败: %v", err)
	}
	return nil
}

// GetIPAddress 实现WiFiConnector接口 - 获取当前WiFi接口的IP地址
func (w *WindowsConnector) GetIPAddress() (string, error) {
	// 使用PowerShell获取IP地址
	command := fmt.Sprintf(`(Get-NetIPAddress -InterfaceAlias "%s" -AddressFamily IPv4).IPAddress`, w.interfaceName)
	ipAddr, err := w.executePowerShellCommand(command)
	if err != nil {
		return "", fmt.Errorf("获取IP地址失败: %v", err)
	}

	// 清理IP地址字符串
	ipAddr = strings.TrimSpace(ipAddr)
	if ipAddr != "" && ipAddr != "127.0.0.1" {
		return ipAddr, nil
	}

	// 备用方法：使用WMI查询
	command2 := fmt.Sprintf(`(Get-WmiObject -Class Win32_NetworkAdapterConfiguration | Where-Object {$_.Description -match '%s' -and $_.IPEnabled -eq $true}).IPAddress[0]`, w.interfaceName)
	ipAddr2, err2 := w.executePowerShellCommand(command2)
	if err2 == nil {
		ipAddr2 = strings.TrimSpace(ipAddr2)
		if ipAddr2 != "" && ipAddr2 != "127.0.0.1" {
			return ipAddr2, nil
		}
	}

	return "", fmt.Errorf("未找到IP地址")
}
