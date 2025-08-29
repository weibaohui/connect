# WiFi 自动连接程序

这是一个用Go语言编写的跨平台WiFi自动连接程序，支持Windows、macOS、Linux系统，可以周期性检测WiFi连接状态，并自动连接到指定的WiFi网络。

## 功能特性

- 自动检测WiFi网卡接口（支持多网卡系统）
- 自动检测当前WiFi连接状态
- 如果未连接任何WiFi，自动连接到目标网络
- 如果连接到其他WiFi，自动切换到目标网络
- 周期性检查，确保始终连接到目标网络
- 自动启用WiFi（如果被禁用）
- 详细的日志记录

## 系统要求

### 通用要求
- 管理员权限（用于修改网络设置）
- Go 1.16+ （仅编译时需要）

### 平台特定要求

**macOS**：
- macOS 10.12+
- 需要`networksetup`命令（系统自带）

**Windows**：
- Windows 7+
- 需要`netsh`命令（系统自带）

**Linux**：
- 需要`nmcli`或`iwgetid`命令
- 需要`ip`命令（iproute2包）

## 安装和使用

### 方式一：使用预编译二进制文件

1. 从[Releases页面](https://github.com/weibaohui/connect/releases)下载对应平台的二进制文件
2. 解压并运行

**Windows**：
```cmd
# 下载 connect-windows-amd64.exe
connect-windows-amd64.exe -w "MyWiFi" -p "password"
```

**macOS**：
```bash
# 下载 connect-darwin-amd64 (Intel) 或 connect-darwin-arm64 (Apple Silicon)
chmod +x connect-darwin-arm64
./connect-darwin-arm64 -w "MyWiFi" -p "password"
```

**Linux**：
```bash
# 下载 connect-linux-amd64
chmod +x connect-linux-amd64
./connect-linux-amd64 -w "MyWiFi" -p "password"
```

### 方式二：从源码编译

#### 1. 克隆项目
```bash
git clone https://github.com/weibaohui/connect.git
cd connect
```

#### 2. 单平台编译
```bash
go build -o connect main.go
```

#### 3. 跨平台编译

**使用编译脚本（推荐）**：

在macOS/Linux上：
```bash
./build.sh
```

在Windows上：
```cmd
build.bat
```

**手动编译**：
```bash
# Windows 64位
GOOS=windows GOARCH=amd64 go build -o connect-windows-amd64.exe main.go

# macOS 64位 (Intel)
GOOS=darwin GOARCH=amd64 go build -o connect-darwin-amd64 main.go

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o connect-darwin-arm64 main.go

# Linux 64位
GOOS=linux GOARCH=amd64 go build -o connect-linux-amd64 main.go
```

### 4. 运行程序

程序支持以下命令行参数：

- `-wifi` / `-w`: 目标WiFi网络名称（默认："qqqq"）
- `-password` / `-p`: WiFi密码（可选，如果为空则使用系统保存的密码）
- `-interval` / `-i`: 检查间隔时间，单位秒（默认：10）

#### 基本使用
```bash
# 使用默认设置（连接到"qqqq"网络）
sudo ./connect

# 指定WiFi名称（完整参数）
sudo ./connect -wifi "你的WiFi名称"

# 指定WiFi名称（简写参数）
sudo ./connect -w "你的WiFi名称"

# 指定WiFi名称和密码
sudo ./connect -wifi "你的WiFi名称" -password "你的密码"

# 使用简写参数
sudo ./connect -w "你的WiFi名称" -p "你的密码"

# 指定检查间隔为30秒
sudo ./connect -w "你的WiFi名称" -i 30

# 组合使用所有参数（简写形式）
sudo ./connect -w "你的WiFi名称" -p "你的密码" -i 30
```

#### 查看帮助
```bash
./connect -h
```

## 命令行参数

- `-wifi` / `-w`: 目标WiFi网络名称（默认："qqqq"）
- `-password` / `-p`: WiFi密码（可选，如果为空则使用系统保存的密码）
- `-interval` / `-i`: 检查间隔时间，单位秒（默认：10秒）

## 工作原理

1. **平台检测**：程序启动时自动检测运行平台（Windows/macOS/Linux）
2. **接口检测**：根据平台自动检测WiFi网卡接口
   - macOS：检测en0、en1、en2等接口
   - Windows：使用netsh命令检测WiFi接口
   - Linux：检测wlan0、wlp2s0等常见接口
3. **状态检查**：立即检查当前WiFi状态
4. **自动启用**：如果WiFi未启用，自动启用WiFi
5. **智能连接**：检查当前连接的WiFi网络
   - 如果未连接任何网络，连接到目标网络
   - 如果连接到其他网络，切换到目标网络
   - 如果已连接到目标网络，保持连接
6. **周期检查**：每隔指定时间重复检查

## 注意事项

### 通用注意事项
- 程序需要管理员权限才能修改网络设置
- 目标WiFi网络建议先手动连接一次，确保密码已保存
- 程序会自动检测WiFi网卡接口，支持多网卡系统
- 如果系统有多个WiFi接口，程序会自动选择第一个可用的接口

### 平台特定注意事项

**macOS**：
- 程序使用`networksetup`命令管理WiFi
- 密码会保存在系统钥匙串中
- 可能需要在"系统偏好设置 > 安全性与隐私"中授权

**Windows**：
- 程序使用`netsh`命令管理WiFi
- 需要以管理员身份运行命令提示符或PowerShell
- 密码会保存在Windows的WiFi配置文件中

**Linux**：
- 程序优先使用`nmcli`命令（NetworkManager）
- 如果没有NetworkManager，会尝试使用`iwgetid`和`ip`命令
- 可能需要sudo权限执行网络管理命令

## 停止程序

使用 `Ctrl+C` 停止程序运行。

## 故障排除

如果程序无法正常工作，请检查：

1. 是否有管理员权限
2. 目标WiFi网络名称是否正确
3. 目标网络是否在系统的已知网络列表中
4. 网络密码是否已保存在钥匙串中

## 许可证

MIT License