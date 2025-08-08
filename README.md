# Prompt Vault (pv) - CLI 提示词管理工具

Prompt Vault 是一个 Go 命令行工具，用于管理 AI 提示词。

## 功能特性

使用 GitHub Gist 管理、分享和导入提示词。同时提供了本地缓存可以在离线的时候使用。

## 安装

### 从源码构建

```bash
git clone https://github.com/grigri/pv.git
cd pv
go build -o pv
# 或使用 Makefile
make build
```

## 快速开始

### 1. 身份验证

首次使用需要登录 GitHub，创建一个具有 `gist` 权限的 Personal Access Token：

1. 访问 https://github.com/settings/tokens
2. 点击 "Generate new token"
3. 选择 `gist` 权限
4. 生成并复制 token

```bash
pv auth login
```

### 2. 添加提示词

#### 从 YAML 文件添加

创建一个 YAML 格式的提示词文件：

```yaml
# my-prompt.yaml
name: "代码审查助手"
author: "grigri"
description: "帮助进行代码审查，发现潜在问题和改进建议"
tags:
  - "代码审查"
  - "开发工具"
version: "1.0"
---
你是一个专业的代码审查助手。请仔细检查以下代码，并提供详细的审查反馈。

请按以下格式提供反馈：
- 代码质量问题
- 潜在的 bug 或安全隐患
- 改进建议
```

然后添加到 Prompt Vault：

```bash
pv add my-prompt.yaml
```

#### 从公开 Gist URL 添加

```bash
pv add https://gist.github.com/username/abc123def456
```

### 3. 浏览提示词

```bash
# 使用本地缓存（快速）
pv list

# 强制从远程获取最新数据
pv list --remote
```

### 4. 获取提示词内容

```bash
# 交互式选择并复制到剪贴板
pv get

# 按关键字筛选获取
pv get "代码审查"

# 直接通过 Gist URL 获取
pv get https://gist.github.com/username/abc123
```

### 5. 同步提示词缓存

```bash
# 同步所有提示词到本地缓存
pv sync

# 显示详细同步过程
pv sync --verbose
```

### 6. 分享提示词

```bash
# 交互式选择私有提示词进行分享
pv share

# 按关键字筛选私有提示词进行分享
pv share "我的提示词"

# 直接分享指定 URL 的私有提示词
pv share https://gist.github.com/username/private_gist_id
```

### 7. 删除提示词

```bash
# 交互式删除 - 显示所有提示词供选择
pv delete

# 按关键字筛选删除
pv delete "代码审查"

# 直接通过 Gist URL 删除
pv delete https://gist.github.com/username/abc123

# 使用简短别名
pv del
```

## 命令参考

| 命令 | 别名 | 描述 | 示例 |
|------|------|------|------|
| `pv` | - | 显示欢迎信息 | `pv` |
| `pv list [--remote]` | - | 列出所有提示词 | `pv list -r` |
| `pv add <file\|url>` | - | 添加提示词 | `pv add prompt.yaml` |
| `pv get [keyword\|url]` | - | 获取提示词到剪贴板 | `pv get "golang"` |
| `pv sync [--verbose]` | - | 同步远程数据到本地 | `pv sync -v` |
| `pv share [keyword\|url]` | - | 分享私有提示词 | `pv share "密码"` |
| `pv delete [keyword\|url]` | `pv del` | 删除提示词 | `pv delete "golang"` |
| `pv auth login` | - | 登录 GitHub 账户 | `pv auth login` |
| `pv auth logout` | - | 登出当前账户 | `pv auth logout` |
| `pv auth status` | - | 查看认证状态 | `pv auth status` |

## 功能详解

### 添加功能

支持两种方式添加提示词：

1. **从 YAML 文件添加** - 本地文件格式详见下文
2. **从公开 Gist URL 添加** - 直接从其他用户的公开 Gist 导入

### 获取功能

提供三种获取模式，自动复制内容到剪贴板：

1. **交互式获取** - 显示所有提示词列表供选择
2. **关键字筛选获取** - 根据关键字筛选提示词
3. **直接 URL 获取** - 通过 Gist URL 直接获取

### 同步功能

完整的缓存同步流程：

- 获取远程提示词索引列表
- 串行下载所有提示词内容到本地缓存
- 显示 "正在下载 X/Y" 进度信息
- 单个提示词失败时继续处理其他提示词
- 显示最终同步统计信息（成功/失败数量）

### 分享功能

将私有 GitHub Gists 转换为公开 Gists：

1. **交互式分享** - 显示所有私有提示词列表供选择
2. **关键字筛选分享** - 根据关键字筛选私有提示词
3. **直接 URL 分享** - 直接分享指定 URL 的私有提示词

### 删除功能

提供三种删除模式，所有删除操作都有确认界面：

1. **交互式删除** - 显示所有提示词的交互式列表
2. **关键字筛选删除** - 通过关键字筛选提示词进行删除
3. **直接 URL 删除** - 通过 GitHub Gist URL 直接删除

## 提示词文件格式

Prompt Vault 使用 YAML 格式存储提示词：

```yaml
name: "提示词名称"           # 必需：提示词的名称
author: "作者名称"          # 必需：创建者
description: "简要描述"      # 可选：提示词用途说明
tags:                      # 可选：分类标签
  - "标签1"
  - "标签2"
version: "1.0"             # 可选：版本号
---
这里是提示词的完整内容。
支持多行文本。
可以包含各种格式和说明。
```

## 配置

配置文件位于 `~/.config/pv/config.yaml`，包含：

- GitHub Personal Access Token（加密存储）
- 用户偏好设置
- 本地索引缓存

## 缓存机制

PV 支持本地缓存机制，提供以下优势：

- **快速访问** - `pv list` 默认使用本地缓存，秒级响应
- **离线使用** - 缓存的提示词可在离线状态下访问
- **网络优化** - 减少 GitHub API 调用，避免速率限制
- **手动控制** - 通过 `pv sync` 手动更新缓存，`pv list --remote` 强制远程获取

## 故障排除

### 认证问题

如果遇到认证相关错误：

```bash
# 检查当前认证状态
pv auth status

# 重新登录
pv auth logout
pv auth login
```

### 网络连接问题

确保你的网络可以访问 GitHub API：

```bash
curl -I https://api.github.com
```

### 权限问题

确保你的 GitHub token 具有创建和管理 Gists 的权限（`gist` scope）。

### 缓存问题

如果缓存数据不一致：

```bash
# 强制同步远程数据
pv sync

# 或使用远程模式列出
pv list --remote
```

## 开发

### 环境要求

- Go 1.24.5 或更高版本
- CGO 支持（用于竞态检测测试）

### 构建项目

```bash
# 使用 Go 命令
go build -o pv

# 使用 Makefile（推荐）
make build

# 构建多平台发布版本
make release
```

### 运行测试

```bash
# 运行所有测试
make test

# 快速测试（无竞态检测）
make test-quick

# 特定功能测试
make test-delete
make test-tui
make test-integration

# 完整测试套件（生成合并覆盖率报告）
make test-all
```

### 生成依赖注入代码

```bash
go generate ./internal/di
# 或
make generate
```

### 开发工具

```bash
# 安装开发依赖
make dev-deps

# 代码格式化
make fmt

# 代码检查
make lint

# 安全扫描
make security

# 性能基准测试
make bench
```

### 项目初始化

首次开发时运行：

```bash
make init
```

## 技术栈

- **Go 1.24.5** - 主要编程语言
- **Cobra** - CLI 框架
- **Bubbletea** - 终端用户界面
- **Lipgloss** - TUI 样式和布局
- **Wire** - 依赖注入
- **GitHub API** - 云存储后端
- **OAuth2** - 身份认证

### 测试依赖

- **go-expect** - 终端交互测试
- **vt10x** - 虚拟终端测试

## 架构

项目采用清晰的分层架构：

- **命令层** (`cmd/`) - CLI 命令实现
- **服务层** (`internal/service/`) - 业务逻辑
- **基础设施层** (`internal/infra/`) - 数据存储
- **领域层** (`internal/model/`) - 核心数据模型
- **TUI 层** (`internal/tui/`) - 用户界面组件
- **认证层** (`internal/auth/`) - GitHub OAuth2 集成
