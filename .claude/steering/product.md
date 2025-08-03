# Product Steering Document

## Product Vision
Prompt Vault (pv) 是一个命令行工具，通过使用 GitHub Gist 作为后端存储，帮助用户在多个平台之间同步和管理他们的 prompts。

## Target Users
- 有 GitHub 使用经验的用户
- 需要在多个设备/平台之间同步 prompts 的用户
- 重视数据所有权和版本控制的用户

## Core Features
### Current
- `pv` - 显示欢迎信息
- `pv list` - 列出所有 prompts

### Planned
#### Authentication Commands
- `pv auth login` - 使用 GitHub Personal Access Token 登录
- `pv auth status` - 查看当前认证状态和已登录的 GitHub 账户
- `pv auth logout` - 登出并清除存储的认证信息

#### Prompt Management Commands
- `pv add` - 添加新的 prompt
- `pv delete` - 删除指定的 prompt
- `pv get` - 获取特定的 prompt 详情
- `pv share` - 分享 prompt（生成可分享的链接）
- `pv sync` - 同步本地和 GitHub Gist 上的 prompts

## Key Benefits
1. **跨平台同步** - 使用 GitHub Gist 实现数据在不同设备间的同步
2. **版本控制** - 利用 Git 的版本控制能力追踪 prompt 的变化
3. **离线支持** - 本地缓存支持离线访问和使用
4. **数据所有权** - 数据存储在用户自己的 GitHub 账户中

## Success Metrics
- 命令执行的可靠性和速度
- 同步操作的成功率
- 离线模式下的功能完整性
- 用户数据的安全性和隐私保护