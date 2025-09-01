#!/bin/bash

# WiFi自动连接程序跨平台编译脚本
# 支持Windows、macOS、Linux多种架构

set -e

# 程序名称
APP_NAME="connect"

# 版本信息
VERSION="1.0.0"

# 输出目录
OUTPUT_DIR="dist"

# 清理输出目录
echo "清理输出目录..."
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

# 编译函数
build_binary() {
    local os=$1
    local arch=$2
    local ext=$3
    
    echo "编译 ${os}/${arch}..."
    
    local output_name="${APP_NAME}-${os}-${arch}${ext}"
    local output_path="${OUTPUT_DIR}/${output_name}"
    
    GOOS=$os GOARCH=$arch go build -ldflags "-s -w -X main.version=${VERSION}" -o "$output_path" *.go
    
    if [ $? -eq 0 ]; then
        echo "✓ 编译成功: $output_path"
    else
        echo "✗ 编译失败: ${os}/${arch}"
        exit 1
    fi
}

echo "开始跨平台编译..."
echo "版本: $VERSION"
echo ""

# Windows 平台
echo "=== Windows 平台 ==="
build_binary "windows" "amd64" ".exe"
build_binary "windows" "386" ".exe"
build_binary "windows" "arm64" ".exe"

# macOS 平台
echo ""
echo "=== macOS 平台 ==="
build_binary "darwin" "amd64" ""
build_binary "darwin" "arm64" ""

# Linux 平台
echo ""
echo "=== Linux 平台 ==="
build_binary "linux" "amd64" ""
build_binary "linux" "386" ""
build_binary "linux" "arm64" ""
build_binary "linux" "arm" ""

echo ""
echo "编译完成！输出文件:"
ls -la $OUTPUT_DIR

echo ""
echo "文件大小统计:"
du -h $OUTPUT_DIR/*

echo ""
echo "所有平台编译完成！"