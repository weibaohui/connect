# WiFi 自动连接程序

这是一个用Go语言编写的跨平台WiFi自动连接程序，支持Windows、macOS、Linux系统，可以周期性检测WiFi连接状态，并自动连接到指定的WiFi网络。

## 使用场景

### 远程办公设备管理

**问题场景**：如果您有多台笔记本电脑需要管理，经常需要带着设备外出办公，回到办公室或家中时可能遇到以下问题：
- 设备没有自动连接到期望的WiFi网络
- 设备连接了其他WiFi网络（如客人网络、邻居网络等）
- 不知道设备当前的IP地址，无法进行远程连接
- 需要物理接触设备才能检查网络状态和重新连接WiFi

**解决方案**：使用本程序可以：
1. **自动网络管理**：确保设备始终连接到指定的WiFi网络，避免连接错误网络
2. **实时IP通知**：当IP地址发生变化时，自动通过飞书发送通知，随时了解设备的最新网络状态
3. **无人值守运行**：程序可以作为后台服务运行，无需人工干预
4. **远程管理便利**：通过飞书通知获得最新IP地址，可以直接进行SSH、RDP等远程连接

**典型工作流程**：
```
设备开机/回到网络环境 → 自动检测WiFi状态 → 自动连接指定网络 → 获取IP地址 → 飞书通知 → 远程连接设备
```

这样，您就不需要物理接触每台设备，通过飞书消息就能知道所有设备的网络状态和IP地址，大大提升了远程办公的效率。

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

# 启用飞书通知功能
FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id \
FEISHU_SECRET=your-secret-key \
sudo ./connect -w "你的WiFi名称" -p "你的密码" --enable-notification
```

#### 查看帮助
```bash
./connect -h
```

## 命令行参数

- `-wifi` / `-w`: 目标WiFi网络名称（必需）
- `-password` / `-p`: WiFi密码（可选，如果为空则使用系统保存的密码）
- `-interval` / `-i`: 检查间隔时间，单位秒（默认：10秒）
- `--enable-notification`: 启用飞书通知功能（可选）

## 飞书通知功能

程序支持在IP地址发生变化时向飞书群发送通知消息。当启用通知功能后，系统会监控IP地址变化并自动发送包含网络信息和IP变化详情的通知。

### 配置方式

#### 方式一：使用环境变量（推荐）

设置以下环境变量：
- `FEISHU_WEBHOOK_URL`: 飞书机器人的Webhook地址
- `FEISHU_SECRET`: 飞书机器人的签名密钥

```bash
# 设置环境变量
export FEISHU_WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-url"
export FEISHU_SECRET="your-secret-key"

# 运行程序并启用通知功能
./connect -w "CMCC-qqqq-5G" -p "cmcccmcc" --enable-notification
```

#### 方式二：运行时设置环境变量

```bash
# 在命令行中直接设置环境变量并运行程序
FEISHU_WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id" \
FEISHU_SECRET="your-secret-key" \
./connect -w "CMCC-qqqq-5G" -p "cmcccmcc" --enable-notification
```

### 使用示例

```bash
# 基础使用：连接WiFi并启用通知
FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id \
FEISHU_SECRET=your-secret-key \
./connect -w CMCC-qqqq-5G -p cmcccmcc --enable-notification

# 使用自定义检查间隔
FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id \
FEISHU_SECRET=your-secret-key \
./connect -w CMCC-qqqq-5G -p cmcccmcc --enable-notification -i 30
```

### 通知内容

当IP地址发生变化时，系统会发送包含以下信息的通知：
- 网络名称
- IP地址变化情况（从旧IP到新IP）
- 通知时间

**首次连接通知示例**：
```
🌐 WiFi连接状态通知
网络：CMCC-qqqq-5G
✅ 已连接，IP地址：192.168.1.100
时间：2024-01-15 14:30:25
```

**IP变化通知示例**：
```
🌐 WiFi连接状态更新
网络：CMCC-qqqq-5G
🔄 IP地址变化：192.168.1.100 → 192.168.1.105
时间：2024-01-15 14:35:30
```

### 注意事项

1. **安全性**：请妥善保管飞书机器人的Webhook URL和签名密钥，避免泄露
2. **网络要求**：发送通知需要网络连接，请确保设备能够访问飞书服务
3. **通知条件**：只有在IP地址真正发生变化时才会发送通知，相同IP不会重复通知
4. **重试机制**：内置3次重试机制，如果发送失败会自动重试
5. **异步发送**：通知发送为异步操作，不会阻塞主程序运行

### 获取飞书机器人配置

1. 在飞书群中添加自定义机器人
2. 获取Webhook URL和签名密钥
3. 确保机器人有发送消息的权限

详细配置步骤请参考[飞书开放平台文档](https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN)

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

### 基本问题
1. 是否有管理员权限
2. 目标WiFi网络名称是否正确
3. 目标网络是否在系统的已知网络列表中
4. 网络密码是否已保存在钥匙串中

### 飞书通知问题
1. **通知未发送**：
   - 检查是否使用了 `--enable-notification` 参数
   - 确认 `FEISHU_WEBHOOK_URL` 和 `FEISHU_SECRET` 环境变量是否正确设置
   - 检查网络连接是否正常

2. **通知发送失败**：
   - 验证飞书Webhook URL是否正确
   - 检查飞书机器人签名密钥是否匹配
   - 确认飞书机器人有发送消息的权限

3. **IP地址未变化但想发送通知**：
   - 默认情况下只有IP地址发生变化时才会发送通知
   - 这是正常行为，可以避免重复通知干扰

## 许可证

MIT License