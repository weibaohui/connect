#!/bin/bash

# WiFi自动连接程序启动脚本
# 适用于macOS平台

# ================================
# 配置区域 - 请根据实际情况修改以下变量
# ================================

# 目标WiFi网络名称
WIFI_NAME="qqqq"

# WiFi密码
WIFI_PASSWORD="cmcccmcc"

# 检查间隔（秒）
CHECK_INTERVAL=10

# 是否启用飞书通知功能（true/false）
ENABLE_NOTIFICATION="true"

# 飞书Webhook URL（启用通知功能时必需）
FEISHU_WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id"

# 飞书机器人签名密钥（启用通知功能时必需）
FEISHU_SECRET="your-secret-key"

# ================================
# 程序执行区域 - 一般不需要修改以下内容
# ================================

# 检查程序文件是否存在
if [ ! -f "./connect-darwin-amd64" ] && [ ! -f "./connect-darwin-arm64" ]; then
    echo "错误: 未找到 connect-darwin-amd64 或 connect-darwin-arm64 文件"
    echo "请确保程序文件与脚本在同一目录下"
    exit 1
fi

# 检测系统架构并选择正确的程序文件
if [ "$(uname -m)" = "arm64" ] || [ "$(uname -m)" = "aarch64" ]; then
    PROGRAM_FILE="./connect-darwin-arm64"
    echo "检测到ARM架构，使用 $PROGRAM_FILE"
else
    PROGRAM_FILE="./connect-darwin-amd64"
    echo "检测到Intel架构，使用 $PROGRAM_FILE"
fi

# 显示配置信息
echo "================================"
echo "WiFi自动连接程序启动配置"
echo "================================"
echo "目标WiFi网络: $WIFI_NAME"
echo "检查间隔: $CHECK_INTERVAL 秒"
echo "通知功能: $ENABLE_NOTIFICATION"
if [ "$ENABLE_NOTIFICATION" = "true" ]; then
    echo "飞书Webhook: $FEISHU_WEBHOOK_URL"
fi
echo "================================"

# 设置环境变量（如果启用了通知功能）
if [ "$ENABLE_NOTIFICATION" = "true" ]; then
    export FEISHU_WEBHOOK_URL="$FEISHU_WEBHOOK_URL"
    export FEISHU_SECRET="$FEISHU_SECRET"
    echo "已设置飞书通知环境变量"
fi

# 构建命令行参数
CMD_ARGS="-w $WIFI_NAME -p $WIFI_PASSWORD -i $CHECK_INTERVAL"

# 如果启用了通知功能，添加相应参数
if [ "$ENABLE_NOTIFICATION" = "true" ]; then
    CMD_ARGS="$CMD_ARGS --enable-notification"
fi

# 执行程序
echo "启动WiFi自动连接程序..."
echo "执行命令: sudo $PROGRAM_FILE $CMD_ARGS"
sudo $PROGRAM_FILE $CMD_ARGS

# 如果程序退出，显示提示信息
echo ""
echo "程序已退出"