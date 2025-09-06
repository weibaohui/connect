# WiFi自动连接程序 Makefile
# 支持Windows、macOS、Linux多种架构

# 程序名称
APP_NAME := connect

# 版本信息
VERSION := 1.0.0

# 输出目录
OUTPUT_DIR := bin

# 默认目标 - 编译当前平台
.PHONY: all build clean build-all build-windows build-macos build-linux

# 默认行为：编译当前平台
all: current-platform

# 编译当前平台
current-platform:
	@echo "编译当前平台..."
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUTPUT_DIR)/$(APP_NAME)" .

# 构建所有平台（make build）
build: build-all

# 构建所有平台
build-all: clean build-windows build-macos build-linux

# 清理输出目录
clean:
	@echo "清理输出目录..."
	@rm -rf $(OUTPUT_DIR)
	@mkdir -p $(OUTPUT_DIR)

# Windows 平台编译
build-windows:
	@echo "=== Windows 平台 ==="
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUTPUT_DIR)/$(APP_NAME)-windows-amd64.exe" .
	GOOS=windows GOARCH=386 go build -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUTPUT_DIR)/$(APP_NAME)-windows-386.exe" .
	GOOS=windows GOARCH=arm64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUTPUT_DIR)/$(APP_NAME)-windows-arm64.exe" .

# macOS 平台编译
build-macos:
	@echo "=== macOS 平台 ==="
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUTPUT_DIR)/$(APP_NAME)-darwin-amd64" .
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUTPUT_DIR)/$(APP_NAME)-darwin-arm64" .

# Linux 平台编译
build-linux:
	@echo "=== Linux 平台 ==="
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUTPUT_DIR)/$(APP_NAME)-linux-amd64" .
	GOOS=linux GOARCH=386 go build -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUTPUT_DIR)/$(APP_NAME)-linux-386" .
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUTPUT_DIR)/$(APP_NAME)-linux-arm64" .
	GOOS=linux GOARCH=arm go build -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUTPUT_DIR)/$(APP_NAME)-linux-arm" .

# 显示帮助信息
help:
	@echo "可用的命令:"
	@echo "  make             - 编译当前平台的二进制文件"
	@echo "  make build       - 构建所有平台的二进制文件"
	@echo "  make build-all   - 构建所有平台的二进制文件"
	@echo "  make clean       - 清理输出目录"
	@echo "  make build-windows - 编译Windows平台二进制文件"
	@echo "  make build-macos   - 编译macOS平台二进制文件"
	@echo "  make build-linux   - 编译Linux平台二进制文件"
	@echo "  make help        - 显示帮助信息"