# Technology Steering Document

## Technology Stack
### Core Technologies
- **Language**: Go 1.24.5
- **CLI Framework**: Cobra v1.9.1
- **Dependency Injection**: Google Wire v0.6.0
- **Storage Backend**: GitHub Gist API

## Architecture Decisions
### Clean Architecture
- 明确的层级分离
- 依赖倒置原则
- 接口驱动设计

### Storage Strategy
- **Primary**: GitHub Gist（在线存储）
- **Secondary**: Local file cache（离线支持）
- **Interface**: `Store` 接口支持多种实现

## Authentication
### 认证策略
- **认证方法**: GitHub Personal Access Token (PAT)
- **所需权限**: `gist` scope（读写 Gists）
- **存储位置**: `~/.pv/config.json`（用户主目录）
- **文件权限**: 0600（仅用户可读写）
- **环境变量**: 支持 `PV_GITHUB_TOKEN` 覆盖本地配置

### 认证流程
1. **登录流程** (`pv auth login`)
   - 提示用户输入 PAT
   - 验证 token 有效性和权限
   - 安全存储到本地配置文件
   - 显示认证成功的 GitHub 用户名

2. **状态检查** (`pv auth status`)
   - 读取本地存储的 token
   - 使用 GitHub API 验证 token 当前状态
   - 显示用户名或未认证状态

3. **登出流程** (`pv auth logout`)
   - 删除本地存储的 token
   - 清理相关缓存数据

## Technical Constraints
- CLI-only，不提供 Web UI 或 API
- 必须兼容 GitHub API 限制
- 需要处理网络不稳定的情况

## Testing Strategy
- **Unit Tests**: 测试独立组件和业务逻辑
- **Integration Tests**: 测试与 GitHub API 的集成
- **Mock Strategy**: 使用接口便于测试时替换实现

## Error Handling
### 错误处理策略
采用**自定义错误类型 + 错误处理中间件**的组合方案：

1. **自定义错误类型** (`internal/errors/errors.go`)
   - 定义 `AppError` 结构体包含错误类型、消息和原始错误
   - 错误类型枚举：`ErrNetwork`, `ErrAuth`, `ErrNotFound`, `ErrValidation`
   - 支持错误包装和错误链

2. **错误处理中间件** (`internal/errors/handler.go`)
   - 集中式错误处理器 `ErrorHandler`
   - 根据错误类型提供用户友好的错误信息
   - 支持 `--verbose` 标志显示详细错误信息
   - 统一的退出码管理

3. **错误处理原则**
   - 在服务层返回结构化错误
   - 在命令层通过错误处理器统一处理
   - 区分致命错误和警告（如本地缓存失败）
   - 提供清晰的错误恢复指导

### 示例错误信息
- 认证失败：`"Authentication failed. Please run 'pv config set-token' to update your GitHub token"`
- 网络错误：`"Network error. Please check your internet connection"`
- 资源未找到：`"Prompt not found: <id>"`

## Performance Considerations
- 本地缓存减少 API 调用
- 批量操作优化
- 异步同步机制（如需要）