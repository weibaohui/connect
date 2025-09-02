# WiFi接口检测测试脚本
# 请在Windows PowerShell中以管理员权限运行此脚本

Write-Host "=== WiFi接口检测测试 ===" -ForegroundColor Green
Write-Host ""

# 测试1：显示所有网络适配器
Write-Host "1. 所有网络适配器:" -ForegroundColor Yellow
try {
    Get-NetAdapter | Format-Table Name, InterfaceDescription, MediaType, Status -AutoSize
} catch {
    Write-Host "错误: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试2：按名称匹配
Write-Host "2. 按名称匹配WiFi接口:" -ForegroundColor Yellow
try {
    $result = Get-NetAdapter | Where-Object {$_.Name -match 'Wi-Fi|无线|WLAN|WiFi|Wireless|以太网|Ethernet.*Wi|Wi.*Fi'} | Select-Object -ExpandProperty Name
    if ($result) {
        Write-Host "找到: $result" -ForegroundColor Green
    } else {
        Write-Host "未找到匹配的接口" -ForegroundColor Red
    }
} catch {
    Write-Host "错误: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试3：按媒体类型匹配
Write-Host "3. 按媒体类型匹配(Native 802.11):" -ForegroundColor Yellow
try {
    $result = Get-NetAdapter | Where-Object {$_.MediaType -eq 'Native 802.11'} | Select-Object -ExpandProperty Name
    if ($result) {
        Write-Host "找到: $result" -ForegroundColor Green
    } else {
        Write-Host "未找到802.11接口" -ForegroundColor Red
    }
} catch {
    Write-Host "错误: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试4：按接口描述匹配
Write-Host "4. 按接口描述匹配:" -ForegroundColor Yellow
try {
    $result = Get-NetAdapter | Where-Object {$_.InterfaceDescription -match 'Wireless|Wi-Fi|802.11|WiFi'} | Select-Object -ExpandProperty Name
    if ($result) {
        Write-Host "找到: $result" -ForegroundColor Green
    } else {
        Write-Host "未找到无线接口" -ForegroundColor Red
    }
} catch {
    Write-Host "错误: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试5：WMI查询
Write-Host "5. WMI查询:" -ForegroundColor Yellow
try {
    $result = Get-WmiObject -Class Win32_NetworkAdapter | Where-Object {$_.Name -match 'Wireless|Wi-Fi|无线|WLAN|802.11' -and $_.NetConnectionID -ne $null} | Select-Object -ExpandProperty NetConnectionID
    if ($result) {
        Write-Host "找到: $result" -ForegroundColor Green
    } else {
        Write-Host "WMI未找到无线接口" -ForegroundColor Red
    }
} catch {
    Write-Host "错误: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试6：显示PowerShell版本
Write-Host "6. PowerShell版本信息:" -ForegroundColor Yellow
$PSVersionTable.PSVersion
Write-Host ""

# 测试7：检查执行策略
Write-Host "7. PowerShell执行策略:" -ForegroundColor Yellow
Get-ExecutionPolicy
Write-Host ""

Write-Host "=== 测试完成 ===" -ForegroundColor Green
Write-Host "请将上述输出结果发送给开发者以便进一步诊断问题。" -ForegroundColor Cyan