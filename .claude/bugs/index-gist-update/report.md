# Bug Report

## Bug Summary
share 命令和 add <gist_url> 无法正确更新 GitHub 中的 index gist，导致索引信息不同步

## Bug Details

### Expected Behavior
- 使用 `pv share` 命令分享私有提示词为公开 Gist 时，应该将导出记录添加到 index gist 的 exports 部分
- 使用 `pv add <gist_url>` 从 URL 导入提示词时，应该将新提示词信息添加到 index gist 的 prompts 部分
- index gist 应该保持最新状态，反映所有添加、分享、删除操作

### Actual Behavior  
- share 命令执行后，导出操作成功创建了公开 Gist，但 index gist 中的 exports 部分没有更新
- add <gist_url> 命令执行后，提示词被导入到本地索引，但 index gist 没有正确更新
- 本地操作成功但 GitHub 上的 index gist 状态不一致

### Steps to Reproduce
1. 创建或确认存在私有提示词
2. 运行 `pv share <keyword>` 或 `pv share <gist_url>` 
3. 检查 GitHub 上的 index gist，发现 exports 部分为空或未更新
4. 运行 `pv add <public_gist_url>` 导入公开提示词
5. 检查 GitHub 上的 index gist，发现 prompts 部分未正确添加新条目

### Environment
- **Version**: 当前开发版本
- **Platform**: Linux/macOS/Windows (CLI 应用)
- **Configuration**: 需要 GitHub OAuth 认证和 Gist 访问权限

## Impact Assessment

### Severity
- [x] High - Major functionality broken
- [ ] Critical - System unusable
- [ ] Medium - Feature impaired but workaround exists
- [ ] Low - Minor issue or cosmetic

### Affected Users
所有使用 share 和 add URL 功能的用户，特别是依赖 GitHub index gist 进行跨设备同步的用户

### Affected Features
- share 命令的导出记录功能
- add <gist_url> 命令的索引更新功能
- GitHub index gist 的同步机制
- 跨设备的提示词库同步

## Additional Context

### Error Messages
暂未发现明显的错误消息，操作显示成功但索引更新失败

### Screenshots/Media
需要检查 GitHub Gist 界面中的 index 文件内容

### Related Issues
可能与 CachedStore 的缓存机制有关，或与 GitHubStore 的 saveIndex() 调用时机有关

## Initial Analysis

### Suspected Root Cause
**经过代码分析，发现核心问题在于 CachedStore 装饰器实现存在缺陷:**

1. **CachedStore 缺少 share 相关方法的缓存更新**: 
   - `AddExport()` 和 `UpdateExport()` 方法直接代理到远程存储，但没有更新本地缓存
   - 这导致缓存与 GitHub index gist 状态不一致

2. **index.Exports 字段初始化问题**:
   - 创建空索引时可能未初始化 `Exports` 字段
   - `initializeIndex()` 方法创建的空索引只包含 `Prompts` 字段

3. **AddFromURL 可能使用缓存而非远程更新**:
   - 如果缓存存在，可能优先使用缓存数据而非更新远程索引

### Affected Components
- `internal/infra/cached_store.go` - AddExport(), UpdateExport() 方法缺少缓存同步
- `internal/infra/github_store.go` - initializeIndex() 方法可能未初始化 Exports 字段
- `internal/service/prompt_service_impl.go` - SharePrompt() 和 AddFromURL() 依赖 Store 实现

### 已确认的技术发现
1. **CachedStore 装饰器**: 虽然正确代理了 AddExport/UpdateExport 到远程存储，但没有相应的缓存更新逻辑
2. **GitHubStore 实现**: AddExport/UpdateExport/GetExports 方法实现正确，直接操作 index.Exports
3. **Index 模型**: 正确包含了 `Exports []IndexedPrompt` 字段
4. **装饰器模式**: 其他方法如 Add/Update/Delete 都正确实现了缓存同步

### Analysis Focus Areas
1. CachedStore 中 AddExport/UpdateExport 方法缺少缓存同步逻辑
2. GitHubStore.initializeIndex() 是否正确初始化 Exports 字段
3. share 和 add URL 操作的完整执行流程验证