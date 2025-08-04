# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Prompt Vault (pv) is a Go CLI application for managing prompts. It uses the Cobra framework for command-line interface and Google Wire for dependency injection.

## Architecture

The codebase follows a clean architecture pattern with clear separation of concerns:

- `cmd/` - CLI commands using Cobra framework
  - `root.go` - Main command entry point
  - `list.go` - List prompts command
  - `add.go` - Add prompts command
  - `delete.go` - Delete prompts command (with del alias)
  - `auth.go` - Authentication commands
- `internal/` - Internal packages
  - `di/` - Dependency injection configuration using Google Wire
  - `infra/` - Infrastructure layer (data storage)
  - `model/` - Domain models
  - `service/` - Business logic layer
  - `tui/` - Terminal User Interface components (bubbletea)
  - `auth/` - Authentication services
  - `config/` - Configuration management
  - `errors/` - Application-specific error handling
  - `validator/` - Input validation
- `main.go` - Application entry point

Key architectural decisions:
- Dependency injection via Google Wire (see `internal/di/wire.go`)
- Interface-based design for storage (`infra.Store`)
- GitHub Gists as the primary storage backend (`GitHubStore`)
- Bubbletea TUI framework for interactive user interfaces
- OAuth2 authentication with GitHub
- Clean error handling with structured error types

### Bubbletea TUI 架构

The Terminal User Interface (TUI) system is built using the bubbletea framework and follows the Model-View-Update (MVU) pattern:

#### Core TUI Components:
- **TUIInterface** (`internal/tui/interface.go`): 定义 TUI 交互契约
- **PromptListModel** (`internal/tui/prompt_list.go`): 提示词列表选择界面
- **ConfirmModel** (`internal/tui/confirm.go`): 确认操作界面
- **ErrorModel** (`internal/tui/error.go`): 错误信息显示界面
- **BubbleTeaTUI** (`internal/tui/factory.go`): TUI 工厂和集成器

#### TUI 交互流程:
1. **列表展示**: 使用 PromptListModel 显示提示词列表，支持键盘导航
2. **用户选择**: 通过方向键和回车键进行选择操作
3. **确认界面**: 显示详细信息并要求用户确认危险操作
4. **错误处理**: 统一的错误界面显示友好的错误信息

#### 测试策略:
- **Mock 接口**: 使用 MockTUI 进行单元测试，避免 TTY 依赖
- **集成测试**: 使用 go-expect 和 vt10x 进行终端交互测试
- **状态验证**: 验证 TUI 状态转换和用户交互逻辑

## Development Commands

### Build
```bash
go build -o pv
```

### Run
```bash
go run main.go
# or after building:
./pv
```

### Generate Wire Dependencies
```bash
go generate ./internal/di
```

### Run Tests
```bash
# 运行所有测试
go test ./...

# 运行带覆盖率的测试
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 运行删除功能专项测试
go test ./cmd -run TestDelete
go test ./internal/service -run TestPromptService.*Delete
go test ./internal/tui -run Test

# 运行集成测试 (需要 TTY 环境)
go test ./integration -run TestDelete

# 运行 TUI 集成测试
go test ./internal/tui -run TestIntegration
```

### 删除功能测试指南

删除功能使用 bubbletea TUI，需要特殊的测试方法：

#### 单元测试
```bash
# 测试服务层删除逻辑
go test ./internal/service -run Delete -v

# 测试 TUI 组件 (使用 Mock)
go test ./internal/tui -v

# 测试命令层参数解析
go test ./cmd -run TestDeleteCommand -v
```

#### 集成测试
```bash
# 设置 TTY 测试环境
export TERM=xterm-256color

# 运行完整的删除工作流测试
go test ./integration -run TestDeleteWorkflow -v

# 测试 TUI 交互 (需要虚拟终端)
go test ./internal/tui -run TestIntegration -v
```

#### 手动测试
```bash
# 测试交互式删除
./pv delete

# 测试关键字筛选删除
./pv delete "test"

# 测试 URL 直接删除
./pv delete https://gist.github.com/user/abc123

# 测试别名功能
./pv del
```

### Format Code
```bash
go fmt ./...
```

### Lint
```bash
go vet ./...
```

## Key Components

### Core Components
- **Store Interface** (`internal/infra/store.go`): Defines the data access contract
- **GitHubStore** (`internal/infra/github_store.go`): GitHub Gists storage implementation
- **Prompt Model** (`internal/model/prompt.go`): Core domain model with ID, Name, Author, and GistURL fields
- **Index Model** (`internal/model/index.go`): Local index for prompt metadata
- **Wire Configuration** (`internal/di/wire.go`): Dependency injection setup

### Service Layer
- **PromptService** (`internal/service/prompt_service.go`): Business logic for prompt operations
- **AuthService** (`internal/service/auth_service.go`): Authentication and authorization logic

### TUI Components
- **TUIInterface** (`internal/tui/interface.go`): TUI interaction contract
- **PromptListModel** (`internal/tui/prompt_list.go`): Interactive prompt list selection
- **ConfirmModel** (`internal/tui/confirm.go`): Confirmation dialogs for dangerous operations
- **ErrorModel** (`internal/tui/error.go`): User-friendly error display

### Authentication
- **GitHubClient** (`internal/auth/github_client.go`): GitHub API client with OAuth2
- **TokenValidator** (`internal/auth/token_validator.go`): OAuth token validation

### Configuration & Validation
- **Config** (`internal/config/config.go`): Application configuration management
- **YAMLValidator** (`internal/validator/yaml_validator.go`): YAML prompt file validation

### Error Handling
- **AppError** (`internal/errors/errors.go`): Structured application error types
- **PromptErrors** (`internal/errors/prompt_errors.go`): Prompt-specific error definitions
- **DeleteErrors** (`internal/errors/delete_errors.go`): Delete operation error definitions
- **AuthErrors** (`internal/errors/auth_errors.go`): Authentication error definitions

## Current Features

- `pv` - Shows greeting message and basic usage information
- `pv list` - Lists all prompts from GitHub Gists with local index
- `pv add <file>` - Adds a new prompt from YAML file to GitHub Gists
- `pv delete [keyword/url]` / `pv del` - Deletes prompts with three modes:
  - Interactive mode (no arguments): Shows TUI list for selection
  - Keyword filtering: Filters prompts by keyword and shows selection TUI
  - Direct URL: Deletes specific prompt by GitHub Gist URL
- `pv auth login` - Authenticate with GitHub OAuth2
- `pv auth logout` - Sign out and clear authentication tokens
- `pv auth status` - Show current authentication status

### Delete Function Modes

1. **Interactive Mode**: `pv delete`
   - Shows all prompts in a navigable TUI list
   - Use arrow keys to navigate, Enter to select, q to quit
   - Displays confirmation dialog before deletion

2. **Keyword Filtering**: `pv delete <keyword>`
   - Filters prompts by name, author, or description
   - Shows filtered results in TUI selection interface
   - Supports partial matching and Chinese keywords

3. **Direct URL**: `pv delete <gist_url>`
   - Directly targets a specific GitHub Gist URL
   - Bypasses list selection, goes straight to confirmation
   - Validates URL format and prompt existence

## Dependencies

The project uses the following key dependencies:

### Core Dependencies
- **github.com/spf13/cobra** - CLI framework for command structure
- **github.com/google/wire** - Dependency injection framework
- **github.com/google/go-github/v74** - GitHub API client library
- **golang.org/x/oauth2** - OAuth2 authentication
- **gopkg.in/yaml.v3** - YAML parsing for prompt files

### TUI Dependencies (New in Delete Feature)
- **github.com/charmbracelet/bubbletea** - Terminal User Interface framework
- **github.com/charmbracelet/lipgloss** - Styling and layout for TUI components

### Testing Dependencies
- **Built-in testing** - Go standard testing package
- **github.com/Netflix/go-expect** - Terminal interaction testing (for TUI)
- **github.com/hinshun/vt10x** - Virtual terminal for testing (for TUI)

### Development Commands for Dependencies

```bash
# Add new dependencies
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest

# Add testing dependencies (dev only)
go get github.com/Netflix/go-expect@latest
go get github.com/hinshun/vt10x@latest

# Update all dependencies
go get -u ./...

# Clean up dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Communication Guidelines

- 用中文交流