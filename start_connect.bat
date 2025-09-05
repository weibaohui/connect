@echo off
chcp 65001 >nul

REM WiFi自动连接程序启动脚本
REM 适用于Windows平台

REM ================================
REM 配置区域 - 请根据实际情况修改以下变量
REM ================================

REM 目标WiFi网络名称
set WIFI_NAME=qqqq

REM WiFi密码
set WIFI_PASSWORD=cmcccmcc

REM 检查间隔（秒）
set CHECK_INTERVAL=10

REM 是否启用飞书通知功能（true/false）
set ENABLE_NOTIFICATION=true

REM 飞书Webhook URL（启用通知功能时必需）
set FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id

REM 飞书机器人签名密钥（启用通知功能时必需）
set FEISHU_SECRET=your-secret-key

REM ================================
REM 程序执行区域 - 一般不需要修改以下内容
REM ================================

REM 检查程序文件是否存在
if not exist "connect-windows-amd64.exe" (
    echo 错误: 未找到 connect-windows-amd64.exe 文件
    echo 请确保程序文件与脚本在同一目录下
    pause
    exit /b 1
)

REM 显示配置信息
echo ================================
echo WiFi自动连接程序启动配置
echo ================================
echo 目标WiFi网络: %WIFI_NAME%
echo 检查间隔: %CHECK_INTERVAL% 秒
echo 通知功能: %ENABLE_NOTIFICATION%
if "%ENABLE_NOTIFICATION%"=="true" (
    echo 飞书Webhook: %FEISHU_WEBHOOK_URL%
)
echo ================================

REM 设置环境变量（如果启用了通知功能）
if "%ENABLE_NOTIFICATION%"=="true" (
    set "FEISHU_WEBHOOK_URL=%FEISHU_WEBHOOK_URL%"
    set "FEISHU_SECRET=%FEISHU_SECRET%"
    echo 已设置飞书通知环境变量
)

REM 构建命令行参数
set "CMD_ARGS=-w %WIFI_NAME% -p %WIFI_PASSWORD% -i %CHECK_INTERVAL%"

REM 如果启用了通知功能，添加相应参数
if "%ENABLE_NOTIFICATION%"=="true" (
    set "CMD_ARGS=%CMD_ARGS% --enable-notification"
)

REM 执行程序
echo 启动WiFi自动连接程序...
echo 执行命令: connect-windows-amd64.exe %CMD_ARGS%
connect-windows-amd64.exe %CMD_ARGS%

REM 如果程序退出，显示提示信息
echo.
echo 程序已退出
pause