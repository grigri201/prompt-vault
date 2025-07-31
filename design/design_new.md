# Prompt Vault - 设计文档

## 概述

Prompt Vault 是一个通过 GitHub Gists 管理和重用提示模板的命令行应用程序。它提供了一个集中式的方式来存储、组织、搜索和检索具有变量替换功能的提示模板。

## 架构

该应用程序遵循**整洁架构**原则，具有清晰的关注点分离和依赖注入模式。

### 核心原则
- **简单架构**：清晰的分层结构，避免过度设计
- **依赖注入**：统一的组件管理
- **标准化错误**：一致的错误处理方式

## 系统架构

```
CLI 命令层 (login, add, get, share, del, sync, config)
    ↓
业务逻辑层 (添加、获取、分享、删除、同步服务)
    ↓  
数据模型层 (Prompt, Index, PromptMeta)
    ↓
存储层 (GitHub Gist, 本地缓存, 配置文件)
```

## 核心组件

### 1. CLI层 (`internal/cli/`)
使用 Cobra 框架的命令行接口，包含 login、add、get、share、del、sync、config 等命令

### 2. 数据模型 (`internal/models/`)

#### PromptMeta
提示元数据结构：
```go
type PromptMeta struct {
    Name        string   `yaml:"name"`
    Author      string   `yaml:"author"`
    Tags        []string `yaml:"tags"`
    Version     string   `yaml:"version,omitempty"`
    Description string   `yaml:"description,omitempty"`
    Parent      string   `yaml:"parent,omitempty"`
    ID          string   `yaml:"id,omitempty"`
}
```

#### Prompt
完整的提示模板结构：
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
本地提示索引结构，entries 和 imported_entries 使用相同的数据结构：
```go
type IndexEntry struct {
    GistID      string    `json:"gist_id"`
    GistURL     string    `json:"gist_url"`
    Name        string    `json:"name"`
    Author      string    `json:"author"`
    Tags        []string  `json:"tags"`
    Version     string    `json:"version,omitempty"`
    Description string    `json:"description,omitempty"`
    Parent      string    `json:"parent,omitempty"`
    ID          string    `json:"id,omitempty"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Index struct {
    Username        string       `json:"username"`
    Entries         []IndexEntry `json:"entries"`
    ImportedEntries []IndexEntry `json:"imported_entries"`
    UpdatedAt       time.Time    `json:"updated_at"`
}
```

### 3. 存储层
- **GitHub Gist API** (`internal/gist/`)：与 GitHub 交互的客户端
- **本地缓存** (`internal/cache/`)：本地文件系统缓存管理
- **配置管理** (`internal/config/`)：应用配置和令牌存储

### 4. 业务逻辑
- **添加服务** (`internal/upload/`)：处理文件上传和重复检测
- **同步服务** (`internal/sync/`)：本地与远程数据自动同步
- **搜索功能** (`internal/search/`)：提示模板搜索
- **分享管理** (`internal/share/`)：提示分享功能

## 自动同步机制

### 同步策略

系统采用基于时间戳的双向同步策略，确保本地和远程数据的一致性：

1. **时间戳比较**：比较本地 `index.json` 和 GitHub 上 `index.json` 的 `updated_at` 字段
2. **覆盖策略**：总是用更新的版本覆盖旧版本
3. **冲突解决**：基于时间戳自动解决冲突，无需人工干预
4. **增量同步**：只同步发生变更的提示文件

### 同步流程

```
1. 获取本地 index.json 的 updated_at
2. 获取远程 index.json 的 updated_at  
3. 比较时间戳：
   ├─ 如果远程更新 → 下载远程 index 和相关提示文件
   ├─ 如果本地更新 → 上传本地 index 到远程
   └─ 如果时间相同 → 无需同步
4. 更新本地缓存和索引
```

### 统一同步方法

系统提供一个统一的同步方法，被以下场景调用：

#### 1. 手动同步
- **触发**：用户执行 `pv sync` 命令
- **目的**：主动同步本地和远程数据

#### 2. 前置自动同步
- **触发**：在 `pv add`、`pv get`、`pv share`、`pv del` 命令执行前
- **目的**：确保操作基于最新数据

#### 3. 后置自动同步
- **触发**：在 `pv login`、`pv add`、`pv del` 命令成功后
- **目的**：将本地变更同步到远程

#### 4. 统一同步策略
所有同步调用都使用相同的逻辑：
1. 比较本地和远程 `index.json` 的 `updated_at` 时间戳
2. 以较新的时间戳版本为准
3. 执行必要的数据传输（上传或下载）
4. 更新本地缓存和索引文件

## CLI 命令详细说明

### 1. login - 身份验证命令

**目的**：配置 GitHub 个人访问令牌以访问 Gists API

**基本用法**：
```bash
pv login
```

**功能**：
- 交互式提示用户输入 GitHub 个人访问令牌
- 验证令牌的有效性和必要权限（gist 权限）
- 将令牌安全存储在配置文件中
- 测试与 GitHub API 的连接

**选项**：
- `--token <token>` - 直接提供令牌而不使用交互式输入
- `--verify` - 验证当前存储的令牌是否仍然有效

**示例**：
```bash
# 交互式登录
pv login

# 使用令牌参数登录
pv login --token ghp_xxxxxxxxxxxxxxxxxxxx

# 验证当前令牌
pv login --verify
```

**输出**：
- 成功：显示验证成功和用户信息
- 失败：显示错误信息和解决建议

### 2. add - 添加提示模板命令

**目的**：向 GitHub Gists 添加新的提示模板（合并原 upload 和 import 功能）

**基本用法**：
```bash
# 上传本地文件
pv add <file>

# 导入 GitHub Gist
pv add <gist-url>
```

**选项**：
- `-f --force` - 强制覆盖重复的提示

**示例**：
```bash
# 上传本地 YAML 文件
pv add my-prompt.yaml

# 从 URL 导入提示
pv add https://gist.github.com/user/abc123
```

**功能**：
- 解析 YAML 前置元数据和内容
- 验证提示格式和必填字段
- 检测重复提示并提供选择
- 创建 GitHub Gist 并更新本地缓存

### 3. get - 获取提示模板命令

**目的**：搜索、选择和使用提示模板，支持变量替换

**基本用法**：
```bash
pv get <search-term>
```

**选项**：
- `-o --output <file>` - 输出到文件而不是显示

**示例**：
```bash
# 基本搜索
pv get "code review"

# 输出到文件
pv get "deployment" --output deploy-script.md
```

**功能**：
- 在本地索引中搜索匹配的提示
- 交互式提示选择器（多个匹配时）
- 智能变量检测和替换
- 自动复制最终结果到剪贴板
- 支持输出格式化

### 4. config - 配置管理命令

**目的**：管理应用程序配置和用户偏好

**基本用法**：
```bash
pv config
```

**示例**：
```bash
# 显示所有配置
pv config
```

### 5. share - 分享提示模板命令

**目的**：将指定的提示模板创建为 public gist 然后获取 gist url 作为分享链接

**基本用法**：
```bash
pv share <search-term>
```

**选项**：
- `--copy` - 复制链接到剪贴板（默认启用）

**示例**：
```bash
# 分享提示模板
pv share "code review template"
```

**功能**：
- 搜索并选择要分享的提示
- 检查 Gist 可见性设置
- 生成可访问的分享 URL
- 支持多种输出格式
- 自动复制到剪贴板

### 6. del - 删除提示模板命令

**目的**：删除本地缓存和 GitHub Gist 中的提示模板

**基本用法**：
```bash
pv del <search-term>
```

**选项**：
- `-f --force` - 跳过确认提示

**示例**：
```bash
# 删除提示（交互式确认）
pv del "old template"

# 强制删除不确认
pv del "deprecated" --force
```

**安全特性**：
- 默认要求用户确认
- 显示将要删除的提示详情

### 7. sync - 强制同步命令

**目的**：手动触发本地与远程数据的完整同步

**基本用法**：
```bash
pv sync
```

**示例**：
```bash
# 标准同步
pv sync
```

**功能**：
- 比较本地和远程的 index.json 时间戳
- 自动以 updated_at 更新的版本为准
- 同步所有相关的提示文件
- 显示同步进度和结果统计

**使用场景**：
- 多设备使用时的数据同步
- 网络中断后的数据恢复
- 定期数据备份同步
- 手动触发完整同步

### 全局选项

所有命令都支持以下全局选项：

- `--help, -h` - 显示帮助信息
- `--verbose, -v` - 详细输出模式

## 关键工作流程

### 1. 身份验证流程
```
用户运行 'pv login'
├─ 提示输入 GitHub 个人访问令牌
├─ 对 GitHub API 验证令牌
├─ 将令牌安全存储在配置文件中
├─ 初始化 Gist 客户端
└─ 执行登录后自动同步（获取最新提示列表）
```

**同步时机**：登录验证成功后执行后置同步

### 2. 添加工作流程（合并 upload 和 import）
```
用户运行 'pv add <file>' 或 'pv add <gist-url>'
├─ 执行前置自动同步（确保重复检测基于最新数据）
├─ 自动检测输入类型（本地文件 vs Gist URL）
├─ 如果是本地文件：
│   ├─ 解析 YAML 前置元数据和内容
│   ├─ 验证提示格式和必填字段
│   ├─ 基于最新数据检测重复提示并提供用户确认选项
│   ├─ 创建新的 GitHub Gist
│   └─ 更新本地缓存和索引文件
├─ 如果是 Gist URL：
│   ├─ 提取并验证 Gist ID
│   ├─ 检查 index.importedEntries 中是否已存在
│   ├─ 验证 Gist 是否包含有效的提示模板
│   ├─ 将 Gist 信息添加到 importedEntries
│   └─ 更新本地索引文件
├─ 执行后置自动同步（将新增内容同步到远程）
└─ 显示操作结果和新增提示信息
```

**同步时机**：前置同步 + 后置同步（调用统一同步方法）

### 3. 获取工作流程（内置同步）
```
用户运行 'pv get <search>'
├─ 执行前置自动同步（确保搜索结果包含最新提示）
├─ 在本地索引中搜索匹配的提示
├─ 如有多个匹配项，显示交互式选择器
├─ 从本地缓存或 GitHub 动态加载提示内容
├─ 智能检测变量并进行交互式替换
├─ 将处理后的最终提示复制到剪贴板
└─ 显示格式化输出结果
```

**同步时机**：前置同步（调用统一同步方法）

### 4. 分享工作流程（内置同步）
```
用户运行 'pv share <search>'
├─ 执行前置自动同步（确保分享的是最新版本）
├─ 搜索并选择要分享的提示
├─ 检查目标 Gist 的可见性设置
├─ 生成公共可访问的分享链接
├─ 自动将链接复制到系统剪贴板
└─ 显示分享链接和相关信息
```

**同步时机**：前置同步（调用统一同步方法）

### 5. 删除工作流程（内置同步）
```
用户运行 'pv del <search>'
├─ 执行前置自动同步（确保删除目标准确性）
├─ 搜索并选择要删除的提示
├─ 显示提示详情并要求用户确认
├─ 从 GitHub Gist 中删除（处理404错误）
├─ 从本地缓存中移除相应文件
├─ 更新本地索引文件
└─ 执行后置自动同步（将删除操作同步到远程）
```

**同步时机**：前置同步 + 后置同步（调用统一同步方法）

### 6. 强制同步工作流程
```
用户运行 'pv sync'
├─ 获取本地 index.json 的 updated_at 时间戳
├─ 获取远程 index.json 的 updated_at 时间戳
├─ 比较时间戳并决定同步方向：
│   ├─ 远程更新 → 下载远程索引和提示文件
│   ├─ 本地更新 → 上传本地索引到远程
│   └─ 时间相同 → 显示"已是最新"消息
├─ 执行增量同步（只同步变更的文件）
├─ 更新本地缓存和索引时间戳
└─ 显示同步结果统计信息
```

**同步时机**：用户手动触发（调用统一同步方法）

## 文件格式

### 提示模板结构
```yaml
---
name: "提示名称"
author: "用户名"
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
      "tags": ["标签1", "标签2"],
      "version": "1.0",
      "description": "描述",
      "parent": "parent-gist-id",
      "id": "custom-id",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "imported_entries": [
    {
      "gist_id": "def456",
      "gist_url": "https://gist.github.com/otheruser/def456",
      "name": "导入的提示",
      "author": "其他用户",
      "tags": ["标签3", "标签4"],
      "version": "2.0",
      "description": "从其他用户导入的提示",
      "parent": "original-gist-id",
      "id": "imported-id",
      "updated_at": "2024-01-02T10:30:00Z"
    }
  ],
  "updated_at": "2024-01-01T12:00:00Z"
}
```

**说明**：
- `entries` 和 `imported_entries` 使用完全相同的数据结构
- 两者都包含所有 PromptMeta 字段（除了已移除的 category）
- `entries`：用户自己创建的提示
- `imported_entries`：从其他用户导入的提示

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
├── wire.go                 # Wire 依赖注入配置（手动维护）
└── wire_gen.go            # Wire 生成的依赖注入代码（自动生成）

internal/
├── auth/                   # 身份验证管理
├── cache/                  # 本地文件缓存
├── cli/                    # CLI 命令和 UI（包含add.go合并upload/import逻辑）
├── clipboard/              # 系统剪贴板集成
├── config/                 # 配置管理
├── container/              # 依赖注入容器
├── errors/                 # 标准化错误处理
├── gist/                   # GitHub Gist API 客户端
├── interfaces/             # 接口定义
├── managers/               # 基础管理器实现
├── models/                 # 领域模型
├── parser/                 # YAML 解析工具
├── paths/                  # 路径管理工具
├── search/                 # 搜索功能
├── share/                  # 分享工作流程管理
├── sync/                   # 同步服务（被各命令内置调用）
├── ui/                     # 交互式 UI 组件
├── upload/                 # 上传和重复检测服务
├── utils/                  # 工具函数
└── wire/                   # Wire 依赖提供器（手动编写的 Provider 函数）
```

**说明**：
- `wire.go` 包含手动编写的依赖配置和 Provider 函数
- `wire_gen.go` 由 `wire` 命令自动生成，包含实际的依赖注入代码
- `internal/wire/` 目录包含各个模块的 Provider 函数定义

## 技术栈

### 核心依赖

- **Go 1.24.5** - 主要开发语言（发布日期：2025年7月8日）
- **Cobra v1.9.1** - CLI 框架（`github.com/spf13/cobra`）
- **Bubble Tea v2.0.0-beta.3** - 交互式 UI 框架（`github.com/charmbracelet/bubbletea/v2`）
- **go-github v72.0.0** - GitHub API 客户端（`github.com/google/go-github/v72/github`）
- **yaml.v3 v3.0.1** - YAML 处理库（`gopkg.in/yaml.v3`）
- **Google Wire v0.6.0** - 依赖注入代码生成（`github.com/google/wire`）

### 系统依赖

- **GitHub Gists API** - 远程数据存储服务
- **系统剪贴板** - 提示内容复制功能
- **本地文件系统** - 缓存和配置存储

## 错误处理

使用统一的 AppError 类型处理所有错误，支持自动分类：
- Auth（身份验证）
- Network（网络）  
- FileSystem（文件）
- Parsing（解析）
- Validation（验证）

## 安全设计

- GitHub 令牌安全存储在本地配置文件
- 所有数据仅在用户本地和 GitHub 之间传输
- 用户完全控制自己的数据

## 性能优化

- 本地缓存所有提示数据
- 基于索引的快速搜索
- 懒加载提示内容
- 网络请求重试机制

## 扩展设计

系统支持添加新命令、存储后端和 UI 组件，通过统一的接口和依赖注入模式保证可扩展性。

## 构建和开发

### 依赖注入 (Wire)

项目使用 Google Wire 进行依赖注入代码生成：

```bash
# 安装 Wire 工具
go install github.com/google/wire/cmd/wire@latest

# 生成依赖注入代码
go generate ./...
# 或者直接运行
wire ./cmd/pv/
```

**重要说明**：
- `wire_gen.go` 文件是由 Wire 工具自动生成的，**不应手动编辑**
- 依赖配置定义在 `cmd/pv/wire.go` 文件中
- 每次修改依赖关系后，必须重新运行 `wire` 命令生成新的代码
- `wire_gen.go` 应该提交到版本控制系统，但始终通过命令生成

### 构建命令

```bash
# 生成 Wire 依赖注入代码
go generate ./...

# 构建应用程序
go build -o pv cmd/pv/main.go

# 运行测试
go test ./...

# 代码格式化
go fmt ./...

# 检查代码质量
go vet ./...
```

### Wire 文件结构

```
cmd/pv/
├── main.go        # 主程序入口
├── wire.go        # Wire 依赖配置（手动维护）
└── wire_gen.go    # 生成的依赖注入代码（自动生成，不可手动编辑）
```

### go.mod 依赖配置示例

```go
module github.com/username/prompt-vault

go 1.24.5

require (
    github.com/spf13/cobra v1.9.1
    github.com/charmbracelet/bubbletea/v2 v2.0.0-beta.3
    github.com/google/go-github/v72 v72.0.0
    gopkg.in/yaml.v3 v3.0.1
    github.com/google/wire v0.6.0
)
```

**版本说明**：
- 以上版本为 2025年7月的最新稳定版本
- Bubble Tea v2 当前为 beta 版本，提供更好的键盘处理和 UI 组件
- 建议定期检查并更新到最新版本以获得安全修复和新功能

## 未来增强

计划功能包括：模板版本控制、批量操作、高级搜索、插件系统、多存储后端支持等。