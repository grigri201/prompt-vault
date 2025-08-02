# Prompt Vault (pv) - Makefile
# 为 Go CLI 应用提供构建、测试和开发工具

# 变量定义
BINARY_NAME=pv
MAIN_PACKAGE=.
BUILD_DIR=./build
SCRIPTS_DIR=./scripts
COVERAGE_DIR=./coverage

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
.DEFAULT_GOAL := help

# 检查 Go 环境
check-go:
ifndef GO_VERSION
	@echo "错误: Go 未安装或不在 PATH 中"
	@exit 1
endif
	@echo "使用 Go 版本: $(GO_VERSION)"

# 安装依赖
.PHONY: deps
deps: check-go ## 安装项目依赖
	@echo "安装依赖..."
	go mod download
	go mod tidy
	@echo "依赖安装完成"

# 安装开发依赖
.PHONY: dev-deps
dev-deps: deps ## 安装开发依赖
	@echo "安装开发依赖..."
	@# 安装测试覆盖率合并工具
	@if ! command -v gocovmerge >/dev/null 2>&1; then \
		echo "安装 gocovmerge..."; \
		go install github.com/wadey/gocovmerge@latest; \
	fi
	@# 安装代码格式化工具
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "安装 goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@# 安装 linting 工具
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	fi
	@echo "开发依赖安装完成"

# 构建
.PHONY: build
build: check-go ## 构建应用程序
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
	rm -rf $(COVERAGE_DIR)
	rm -f coverage*.out coverage*.html
	@echo "清理完成"

# 运行应用
.PHONY: run
run: build ## 构建并运行应用
	./$(BINARY_NAME)

# 基础测试
.PHONY: test
test: check-go ## 运行所有单元测试
	@echo "运行单元测试..."
	@if command -v gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test ./... -v -race -coverprofile=coverage.out; \
	else \
		echo "警告: gcc 不可用，跳过竞态检测..."; \
		go test ./... -v -coverprofile=coverage.out; \
	fi
	@echo "单元测试完成"

# 删除功能测试
.PHONY: test-delete
test-delete: check-go ## 运行删除功能相关测试
	@echo "运行删除功能测试..."
	@echo "1. 运行删除服务测试..."
	@if command -v gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -v -race ./internal/service/... -run=".*Delete.*" -coverprofile=coverage-delete-service.out; \
	else \
		echo "警告: gcc 不可用，跳过竞态检测..."; \
		go test -v ./internal/service/... -run=".*Delete.*" -coverprofile=coverage-delete-service.out; \
	fi
	@echo "2. 运行删除命令测试..."
	@if command -v gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -v -race ./cmd/... -run=".*[Dd]elete.*" -coverprofile=coverage-delete-cmd.out; \
	else \
		echo "警告: gcc 不可用，跳过竞态检测..."; \
		go test -v ./cmd/... -run=".*[Dd]elete.*" -coverprofile=coverage-delete-cmd.out; \
	fi
	@echo "3. 运行删除错误处理测试..."
	@if command -v gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -v -race ./internal/errors/... -run=".*[Dd]elete.*" -coverprofile=coverage-delete-errors.out; \
	else \
		echo "警告: gcc 不可用，跳过竞态检测..."; \
		go test -v ./internal/errors/... -run=".*[Dd]elete.*" -coverprofile=coverage-delete-errors.out; \
	fi
	@echo "4. 运行删除集成测试..."
	@if command -v gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -v -race ./integration/... -run=".*[Dd]elete.*" -coverprofile=coverage-delete-integration.out; \
	else \
		echo "警告: gcc 不可用，跳过竞态检测..."; \
		go test -v ./integration/... -run=".*[Dd]elete.*" -coverprofile=coverage-delete-integration.out; \
	fi
	@echo "删除功能测试完成"

# TUI 测试
.PHONY: test-tui
test-tui: check-go ## 运行 TUI 相关测试
	@echo "运行 TUI 测试..."
	@if [ -f $(SCRIPTS_DIR)/test-tui.sh ]; then \
		$(SCRIPTS_DIR)/test-tui.sh -c; \
	else \
		echo "警告: TUI 测试脚本不存在，运行基础 TUI 测试..."; \
		if command -v gcc >/dev/null 2>&1; then \
			CGO_ENABLED=1 go test -v -race ./internal/tui/... -coverprofile=coverage-tui.out; \
		else \
			echo "警告: gcc 不可用，跳过竞态检测..."; \
			go test -v ./internal/tui/... -coverprofile=coverage-tui.out; \
		fi; \
	fi
	@echo "TUI 测试完成"

# 集成测试
.PHONY: test-integration
test-integration: check-go ## 运行集成测试
	@echo "运行集成测试..."
	CGO_ENABLED=1 go test -v -race ./integration/... -tags=integration -timeout=60s -coverprofile=coverage-integration.out
	@echo "集成测试完成"

# 完整测试套件
.PHONY: test-all
test-all: test test-delete test-tui test-integration ## 运行所有测试
	@echo "运行完整测试套件..."
	@echo "合并覆盖率报告..."
	@if command -v gocovmerge >/dev/null 2>&1; then \
		gocovmerge coverage*.out > coverage-total.out; \
		go tool cover -html=coverage-total.out -o coverage-total.html; \
		coverage_percent=$$(go tool cover -func=coverage-total.out | tail -1 | awk '{print $$3}'); \
		echo "总体测试覆盖率: $$coverage_percent"; \
		echo "覆盖率报告已生成: coverage-total.html"; \
	else \
		echo "警告: gocovmerge 未安装，无法合并覆盖率报告"; \
	fi
	@echo "完整测试套件执行完成"

# 快速测试 (无竞态检测)
.PHONY: test-quick
test-quick: check-go ## 快速运行测试 (无竞态检测)
	@echo "快速测试..."
	go test ./... -short
	@echo "快速测试完成"

# 代码格式化
.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi
	@echo "代码格式化完成"

# 代码检查
.PHONY: lint
lint: ## 运行代码检查
	@echo "运行代码检查..."
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "警告: golangci-lint 未安装，跳过 linting"; \
	fi
	@echo "代码检查完成"

# 安全扫描
.PHONY: security
security: ## 运行安全扫描
	@echo "运行安全扫描..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec 未安装，尝试安装..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi
	@echo "安全扫描完成"

# 性能基准测试
.PHONY: bench
bench: check-go ## 运行性能基准测试
	@echo "运行性能基准测试..."
	CGO_ENABLED=1 go test -bench=. -benchmem ./...
	@echo "性能基准测试完成"

# 生成依赖注入代码
.PHONY: generate
generate: check-go ## 生成 Wire 依赖注入代码
	@echo "生成依赖注入代码..."
	go generate ./internal/di
	@echo "依赖注入代码生成完成"

# 项目初始化
.PHONY: init
init: dev-deps generate ## 初始化项目开发环境
	@echo "初始化项目开发环境..."
	@echo "1. 创建必要目录..."
	@mkdir -p $(BUILD_DIR) $(COVERAGE_DIR)
	@echo "2. 检查项目结构..."
	@if [ ! -f go.mod ]; then echo "警告: go.mod 不存在"; fi
	@if [ ! -f main.go ]; then echo "警告: main.go 不存在"; fi
	@echo "3. 运行快速测试..."
	@$(MAKE) test-quick
	@echo "项目初始化完成"

# Docker 构建
.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@if [ -f Dockerfile ]; then \
		echo "构建 Docker 镜像..."; \
		docker build -t $(BINARY_NAME):$(VERSION) .; \
		echo "Docker 镜像构建完成: $(BINARY_NAME):$(VERSION)"; \
	else \
		echo "错误: Dockerfile 不存在"; \
		exit 1; \
	fi

# 发布构建 (多平台)
.PHONY: release
release: clean ## 构建发布版本 (多平台)
	@echo "构建发布版本..."
	@mkdir -p $(BUILD_DIR)
	@echo "构建 Linux amd64..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	@echo "构建 Linux arm64..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	@echo "构建 macOS amd64..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	@echo "构建 macOS arm64..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	@echo "构建 Windows amd64..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "发布版本构建完成，文件位于 $(BUILD_DIR)/"

# 项目信息
.PHONY: info
info: ## 显示项目信息
	@echo "项目信息:"
	@echo "  名称: $(GO_MOD_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git 提交: $(GIT_COMMIT)"
	@echo "  Go 版本: $(GO_VERSION)"
	@echo "  二进制文件: $(BINARY_NAME)"

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "Prompt Vault (pv) - Makefile 帮助"
	@echo ""
	@echo "可用目标:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "快速开始:"
	@echo "  make init          # 初始化开发环境"
	@echo "  make build         # 构建应用"
	@echo "  make test-all      # 运行所有测试"
	@echo "  make run           # 运行应用"
	@echo ""
	@echo "常用命令组合:"
	@echo "  make clean build test-all  # 完整的构建和测试流程"
	@echo "  make fmt lint test-quick   # 快速的代码检查和测试"