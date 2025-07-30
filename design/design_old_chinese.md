# Prompt Vault - 设计文档

## 概述

Prompt Vault 是一个通过 GitHub Gists 管理和重用提示模板的命令行应用程序。它提供了一个集中式的方式来存储、组织、搜索和检索具有变量替换功能的提示模板。

## 架构

该应用程序遵循**整洁架构**原则，具有清晰的关注点分离和依赖注入模式。

### 核心原则
- **接口隔离**：组件通过定义良好的接口进行通信
- **依赖注入**：容器模式管理所有依赖关系
- **整洁架构**：领域层、应用层和基础设施层之间的清晰分离
- **错误处理**：具有自动分类的标准化错误类型
- **可测试性**：基于接口的设计支持全面测试

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI层                               │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────┐ │
│  │  sync   │ │ upload  │ │  list   │ │  get    │ │ delete │ │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    容器层                                   │
│                依赖注入和生命周期管理                        │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   应用层                                    │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐ │
│  │     同步     │ │    上传      │ │     导入/分享        │ │
│  │    服务      │ │  重复检测器   │ │      管理器          │ │
│  │              │ │              │ │                      │ │
│  └──────────────┘ └──────────────┘ └──────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    领域层                                   │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐ │
│  │    提示      │ │    索引      │ │     PromptMeta       │ │
│  │    模型      │ │    模型      │ │       模型           │ │
│  └──────────────┘ └──────────────┘ └──────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                基础设施层                                   │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐ │
│  │    GitHub    │ │    本地      │ │     文件系统         │ │
│  │     Gist     │ │    缓存      │ │      和路径          │ │
│  │    客户端    │ │   管理器     │ │                      │ │
│  └──────────────┘ └──────────────┘ └──────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. CLI层 (`internal/cli/`)

**目的**：使用 Cobra 框架作为所有用户交互的入口点

**关键文件**：
- `root.go` - 主命令定义和子命令注册
- `sync.go` - 本地缓存与 GitHub Gists 之间的同步
- `upload.go` - 上传提示模板到 GitHub Gists
- `list.go` - 列出和搜索缓存的提示模板
- `get.go` - 检索和显示带变量替换的提示模板
- `delete.go` - 从 GitHub 和本地缓存删除提示模板
- `login.go` - GitHub 身份验证
- `config.go` - 配置管理命令
- `share.go` - 在用户之间分享提示模板
- `import.go` - 导入分享的提示模板

**模式**：
- 每个命令都有专门的文件和函数
- 通过命令上下文进行依赖注入
- 使用 Cobra 框架构建 CLI 结构
- 使用 Bubble Tea 框架构建交互式 UI 组件

### 2. 容器层 (`internal/container/`)

**目的**：管理所有应用依赖的依赖注入容器

**关键特性**：
- 使用 `Initialize()` 和 `Cleanup()` 方法进行生命周期管理
- 生产和测试容器变体
- 基于接口的依赖管理
- 基于令牌的 Gist 客户端初始化

**管理的依赖**：
- PathManager（具体类型）
- CacheManager（接口）
- ConfigManager（具体类型 - 避免循环依赖）
- AuthManager（接口）
- GistClient（具体类型）

### 3. 领域模型 (`internal/models/`)

**核心模型**：

#### PromptMeta
```go
type PromptMeta struct {
    Name        string   `yaml:"name"`
    Author      string   `yaml:"author"`
    Category    string   `yaml:"category"`
    Tags        []string `yaml:"tags"`
    Version     string   `yaml:"version,omitempty"`
    Description string   `yaml:"description,omitempty"`
    Parent      string   `yaml:"parent,omitempty"`
    ID          string   `yaml:"id,omitempty"`
}
```

#### Prompt
```go
type Prompt struct {
    PromptMeta
    GistID    string    `json:"gist_id"`
    GistURL   string    `json:"gist_url"`
    UpdatedAt time.Time `json:"updated_at"`
    Content   string    `json:"-"` // 不序列化到索引
}
```

#### Index
```go
type Index struct {
    Username        string       `json:"username"`
    Entries         []IndexEntry `json:"entries"`
    ImportedEntries []IndexEntry `json:"imported_entries"`
    UpdatedAt       time.Time    `json:"updated_at"`
}
```

### 4. 存储层

#### GitHub Gist API (`internal/gist/`)
- **客户端**：经过身份验证的 GitHub API 客户端包装器
- **操作**：具有重试逻辑和错误处理的高级 Gist 操作
- **特性**：404 处理、速率限制管理、原子操作

#### 本地缓存 (`internal/cache/`)
- **目的**：下载提示的本地文件系统缓存
- **位置**：`~/.cache/prompt-vault/prompts/`
- **结构**：
  - `index.json` - 所有提示的元数据索引
  - `<gist-id>.yaml` - 单个缓存的提示文件

#### 配置 (`internal/config/`)
- **目的**：应用程序配置和身份验证令牌
- **位置**：`~/.config/prompt-vault/config.yaml`
- **内容**：GitHub 令牌、用户首选项、设置

#### 身份验证 (`internal/auth/`)
- **目的**：GitHub 令牌管理和验证
- **特性**：令牌验证、安全存储、用户身份验证

### 5. 统一组件

#### YAML 解析器 (`internal/parser/`)
- **目的**：提示文件的可配置 YAML 解析
- **模式**：严格和宽松解析模式
- **特性**：前置内容提取、内容分离、验证

#### 错误处理 (`internal/errors/`)
- **类型**：Auth、Network、FileSystem、Parsing、Validation 错误
- **特性**：自动错误分类、标准化创建
- **用法**：不直接使用 `fmt.Errorf` - 所有错误使用 AppError 类型

#### 路径管理 (`internal/paths/`)
- **目的**：所有文件操作的集中路径处理
- **特性**：原子写入、安全文件权限、跨平台支持

### 6. 接口设计 (`internal/interfaces/`)

#### 管理器接口
```go
type Manager interface {
    Initialize(ctx context.Context) error
    Cleanup() error
}
```

#### 缓存接口
```go
type CacheReader interface {
    GetPrompt(gistID string) (*models.Prompt, error)
    GetIndex() (*models.Index, error)
    GetCacheDir() string
}

type CacheWriter interface {
    SavePrompt(prompt *models.Prompt) error
    SaveIndex(index *models.Index) error
    ClearCache() error
}

type CacheManager interface {
    CacheReader
    CacheWriter
    Manager
}
```

#### 身份验证接口
```go
type AuthReader interface {
    GetToken() (string, error)
    IsAuthenticated() bool
}

type AuthWriter interface {
    SaveToken(token string) error
    ClearToken() error
}

type AuthManager interface {
    AuthReader
    AuthWriter
    Manager
}
```

## 关键工作流程

### 1. 身份验证流程
```
用户运行 'pv login'
├─ 提示输入 GitHub 个人访问令牌
├─ 对 GitHub API 验证令牌
├─ 将令牌安全存储在配置文件中
└─ 为将来的操作初始化 Gist 客户端
```

### 2. 同步工作流程
```
用户运行 'pv sync'
├─ 使用存储的令牌连接到 GitHub
├─ 获取用户的 gists 并过滤提示文件
├─ 下载所有提示模板
├─ 更新本地缓存和索引
├─ 处理已删除的 gists（404 错误）
└─ 显示同步摘要
```

### 3. 上传工作流程
```
用户运行 'pv upload <file>'
├─ 解析 YAML 前置内容和内容
├─ 验证提示元数据
├─ 检查重复项（用户确认）
├─ 创建或更新 GitHub Gist
├─ 更新本地缓存和索引
└─ 可选择在上传后自动同步
```

### 4. 获取工作流程
```
用户运行 'pv get <search>'
├─ 在本地索引中搜索匹配的提示
├─ 如有多个匹配项，显示交互式选择器
├─ 从缓存加载提示内容
├─ 交互式变量替换
├─ 将最终提示复制到剪贴板
└─ 显示格式化输出
```

## 文件格式

### 提示模板结构
```yaml
---
name: "提示名称"
author: "用户名"
category: "类别"
tags: ["标签1", "标签2"]
version: "1.0"
description: "描述"
parent: "parent-gist-id"  # 可选，用于共享提示
id: "custom-id"           # 可选，自定义标识符
---
包含 {variables} 填充的提示内容。

您可以使用多个 {variable_name} 占位符。
系统将在 'get' 命令期间提示输入每个唯一变量。
```

### 索引文件结构
```json
{
  "username": "githubuser",
  "entries": [
    {
      "gist_id": "abc123",
      "gist_url": "https://gist.github.com/user/abc123",
      "name": "提示名称",
      "author": "用户名",
      "category": "类别",
      "tags": ["标签1", "标签2"],
      "version": "1.0",
      "description": "描述",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "imported_entries": [...],
  "updated_at": "2024-01-01T12:00:00Z"
}
```

## 目录结构

### 应用程序目录
```
~/.cache/prompt-vault/prompts/    # 本地缓存
├── index.json                   # 提示元数据索引
└── <gist-id>.yaml              # 单个缓存的提示

~/.config/prompt-vault/          # 配置
└── config.yaml                 # 应用配置和身份验证令牌
```

### 源代码结构
```
cmd/pv/                     # 主应用程序入口点
├── main.go                 # 应用程序引导
├── wire.go                 # 依赖注入配置
└── wire_gen.go            # 生成的 wire 代码

internal/
├── auth/                   # 身份验证管理
├── cache/                  # 本地文件缓存
├── cli/                    # CLI 命令和 UI
├── clipboard/              # 系统剪贴板集成
├── config/                 # 配置管理
├── container/              # 依赖注入容器
├── errors/                 # 标准化错误处理
├── gist/                   # GitHub Gist API 客户端
├── imports/                # 导入工作流程管理
├── interfaces/             # 接口定义
├── managers/               # 基础管理器实现
├── models/                 # 领域模型
├── parser/                 # YAML 解析工具
├── paths/                  # 路径管理工具
├── search/                 # 搜索功能
├── share/                  # 分享工作流程管理
├── sync/                   # 同步服务
├── ui/                     # 交互式 UI 组件
├── upload/                 # 上传工作流程和重复检测
├── utils/                  # 工具函数
└── wire/                   # Wire 依赖提供器
```

## 技术栈

### 核心技术
- **语言**：Go 1.21+
- **CLI 框架**：Cobra
- **交互式 UI**：Bubble Tea
- **GitHub API**：go-github/v73
- **YAML 处理**：gopkg.in/yaml.v3
- **依赖注入**：Google Wire

### 外部依赖
- 用于 Gist 存储的 GitHub API
- 用于提示复制的系统剪贴板
- 用于本地缓存和配置的文件系统

## 错误处理策略

### 错误类别
1. **AuthError**：身份验证和授权失败
2. **NetworkError**：网络连接和 API 速率限制
3. **FileSystemError**：文件 I/O 操作
4. **ParsingError**：YAML 解析和验证
5. **ValidationError**：数据验证失败

### 错误创建模式
```go
// 标准错误创建
return errors.NewValidationErrorMsg("function", "validation failed")

// 包装现有错误
return errors.WrapWithMessage(err, "operation failed")
```

## 安全考虑

### 令牌管理
- GitHub 令牌安全存储在配置文件中
- 应用程序启动时进行令牌验证
- 不在环境变量或日志中存储令牌
- 配置文件具有安全文件权限

### 数据隐私
- 所有数据存储在本地或用户的 GitHub Gists 中
- 无第三方数据传输
- 用户通过 GitHub 账户控制所有数据

## 性能特征

### 缓存策略
- 所有下载提示的本地缓存
- 基于索引的元数据搜索
- 提示内容的懒加载
- 具有变更检测的高效同步

### 网络优化
- 尽可能批量操作
- 临时失败的重试逻辑
- 速率限制感知
- 删除 gists 的 404 处理

## 扩展点

### 添加新命令
1. 在 `internal/cli/` 中创建命令文件
2. 使用依赖注入实现命令逻辑
3. 在 `root.go` 中注册命令
4. 添加命令上下文支持

### 添加新存储后端
1. 实现存储接口
2. 更新新依赖的容器
3. 添加配置选项
4. 保持向后兼容性

### 添加新 UI 组件
1. 在 `internal/ui/` 中创建 Bubble Tea 模型
2. 遵循表单和选择器的现有模式
3. 与命令工作流程集成
4. 确保可访问性和可用性

## 构建和开发

### 构建命令
```bash
# 构建应用程序
go build -o pv cmd/pv/main.go

# 运行所有测试
go test ./...

# 生成 wire 依赖
go generate ./...

# 格式化和检查
go fmt ./...
go vet ./...
```

### 测试策略
- 所有组件的单元测试
- 工作流程的集成测试
- 外部依赖的模拟接口
- 常见操作的测试助手

## 未来增强

### 计划功能
- 模板验证和检查
- 批量操作（导入/导出）
- 模板版本控制和历史
- 高级搜索和过滤
- 自定义模板类别
- 模板分享工作流程
- 离线模式改进

### 架构演进
- 扩展的插件系统
- 多存储后端支持
- 全屏模式的增强 UI
- 模板协作功能
- 与其他提示管理工具的集成