# Prompt Vault (pv) - CLI 提示词管理工具

Prompt Vault 是一个基于 Go 的命令行工具，用于管理和组织你的 AI 提示词。它将提示词存储在 GitHub Gists 中，提供简洁的界面来添加、浏览和删除提示词。

## 功能特性

- 📝 **添加提示词** - 从 YAML 文件添加新的提示词到 GitHub Gists
- 📋 **浏览提示词** - 列出所有已保存的提示词
- 🗑️ **删除提示词** - 通过交互式界面或关键字删除提示词
- 🔐 **GitHub 集成** - 使用 GitHub OAuth2 安全认证
- 🎨 **交互式界面** - 基于 bubbletea 的直观用户界面
- ☁️ **云存储** - 所有提示词安全存储在 GitHub Gists 中

## 安装

### 从源码构建

```bash
git clone https://github.com/grigri/pv.git
cd pv
go build -o pv
```

### 使用预编译二进制文件

从 [Releases](https://github.com/grigri/pv/releases) 页面下载适合你系统的二进制文件。

## 快速开始

### 1. 身份验证

首次使用需要登录 GitHub：

```bash
pv auth login
```

### 2. 添加提示词

创建一个 YAML 格式的提示词文件：

```yaml
# my-prompt.yaml
name: "代码审查助手"
author: "grigri"
description: "帮助进行代码审查的 AI 提示词"
content: |
  请作为一个资深的代码审查员，仔细检查以下代码：
  
  1. 代码逻辑是否正确
  2. 是否遵循最佳实践
  3. 性能是否可以优化
  4. 安全性是否有问题
  
  请提供具体的改进建议。
```

然后添加到 Prompt Vault：

```bash
pv add my-prompt.yaml
```

### 3. 浏览提示词

```bash
pv list
```

### 4. 删除提示词

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
| `pv list` | - | 列出所有提示词 | `pv list` |
| `pv add <file>` | - | 从文件添加提示词 | `pv add prompt.yaml` |
| `pv delete [keyword/url]` | `pv del` | 删除提示词 | `pv delete "golang"` |
| `pv auth login` | - | 登录 GitHub 账户 | `pv auth login` |
| `pv auth logout` | - | 登出当前账户 | `pv auth logout` |
| `pv auth status` | - | 查看认证状态 | `pv auth status` |

## 删除功能详解

Prompt Vault 提供三种删除模式：

### 1. 交互式删除

运行不带参数的删除命令，会显示所有提示词的交互式列表：

```bash
pv delete
```

使用方向键导航，按 Enter 选择要删除的提示词，按 q 退出。

### 2. 关键字筛选删除

通过关键字筛选提示词，然后从匹配结果中选择删除：

```bash
pv delete "golang"
pv delete "代码审查"
```

### 3. 直接 URL 删除

如果你知道具体的 GitHub Gist URL，可以直接删除：

```bash
pv delete https://gist.github.com/username/abc123def456
```

所有删除操作都会显示确认界面，确保不会误删重要的提示词。

## 提示词文件格式

Prompt Vault 使用 YAML 格式存储提示词：

```yaml
name: "提示词名称"           # 必需：提示词的名称
author: "作者名称"          # 必需：创建者
description: "简要描述"      # 可选：提示词用途说明
content: |                  # 必需：提示词的实际内容
  这里是提示词的完整内容。
  支持多行文本。
  可以包含各种格式和说明。
```

## 配置

配置文件位于 `~/.config/pv/config.yaml`，包含：

- GitHub OAuth token（加密存储）
- 用户偏好设置
- 索引缓存

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

确保你的 GitHub token 具有创建和管理 Gists 的权限。

## 开发

### 构建项目

```bash
go build -o pv
```

### 运行测试

```bash
go test ./...

# 运行带覆盖率的测试
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 生成依赖注入代码

```bash
go generate ./internal/di
```

## 技术栈

- **Go 1.24.5** - 主要编程语言
- **Cobra** - CLI 框架
- **Bubbletea** - 终端用户界面
- **Wire** - 依赖注入
- **GitHub API** - 云存储后端
- **OAuth2** - 身份认证

## 贡献

欢迎贡献代码！请确保：

1. 遵循现有的代码风格
2. 添加相应的测试
3. 更新相关文档

## 许可证

MIT License - 查看 [LICENSE](LICENSE) 文件了解详情。

## 更新日志

### v1.1.0 (即将发布)
- ✨ 新增删除功能，支持三种删除模式
- 🎨 添加基于 bubbletea 的交互式用户界面
- 🔧 改进错误处理和用户反馈
- 📚 完善文档和使用指南

### v1.0.0
- 🎉 初始版本发布
- ✅ 基本的添加和列表功能
- 🔐 GitHub OAuth2 认证
- ☁️ GitHub Gists 存储集成