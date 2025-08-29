# WiFi 自动连接程序

这是一个用Go语言编写的WiFi自动连接程序，可以周期性检测WiFi连接状态，并自动连接到指定的WiFi网络。

## 功能特性

- 自动检测当前WiFi连接状态
- 如果未连接任何WiFi，自动连接到目标网络
- 如果连接到其他WiFi，自动切换到目标网络
- 周期性检查，确保始终连接到目标网络
- 自动启用WiFi（如果被禁用）
- 详细的日志记录

## 系统要求

- macOS 系统
- Go 1.16 或更高版本
- 管理员权限（用于修改网络设置）

## 安装和使用

### 1. 克隆项目
```bash
git clone https://github.com/weibaohui/connect.git
cd connect
```

### 2. 编译程序
```bash
go build -o connect main.go
```

### 3. 运行程序

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

1. 程序启动后立即检查当前WiFi状态
2. 如果WiFi未启用，自动启用WiFi
3. 检查当前连接的WiFi网络：
   - 如果未连接任何网络，连接到目标网络
   - 如果连接到其他网络，切换到目标网络
   - 如果已连接到目标网络，保持连接
4. 每隔指定时间重复检查

## 注意事项

- 程序需要管理员权限才能修改网络设置
- 目标WiFi网络必须在系统的已知网络列表中
- 如果目标网络需要密码，请确保密码已保存在系统钥匙串中
- 程序使用 `networksetup` 命令，仅适用于macOS系统

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