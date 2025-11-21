# Prompt Vault (pv) - Makefile
# 为 Go CLI 应用提供构建相关工具

# 变量定义
BINARY_NAME=pv
MAIN_PACKAGE=.
BUILD_DIR=./build

# Go 相关变量
GO_VERSION := $(shell go version 2>/dev/null)
GO_MOD_NAME := $(shell go list -m 2>/dev/null)

# 版本和构建信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 默认目标
.DEFAULT_GOAL := build

# 构建
.PHONY: build
build: ## 构建应用程序
	@echo "构建 $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "构建完成: $(BINARY_NAME)"

# 清理构建文件
.PHONY: clean
clean: ## 清理构建文件和缓存
	@echo "清理文件..."
	go clean
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	@echo "清理完成"
