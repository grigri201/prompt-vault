# 离线缓存功能实现计划

## 任务概述
实现离线缓存功能，通过装饰器模式为现有的 GitHubStore 添加本地缓存层。采用"远程优先，缓存回退"的策略，新增 CacheManager 进行文件管理，并扩展现有命令支持离线操作。实现严格遵循单一职责和最小修改原则。

## 指导文档合规性
- **structure.md**: 所有新文件位于 `internal/infra/` 和 `cmd/` 目录，遵循现有命名约定
- **tech.md**: 使用装饰器模式保持清洁架构，复用现有错误处理和文件权限工具
- **依赖注入**: 通过 Wire 管理新增组件，保持接口驱动设计

## 原子化任务要求
每个任务满足以下标准：
- **文件范围**: 最多涉及 1-3 个相关文件
- **时间限制**: 15-30 分钟可完成
- **单一目的**: 一个可测试的结果
- **具体文件**: 明确指定要创建/修改的文件路径
- **代理友好**: 清晰的输入输出，最少的上下文切换

## 任务

### 阶段 1：核心缓存管理器

- [x] 1. 创建缓存目录工具函数在 internal/config/cache.go
  - 文件: internal/config/cache.go (新建)
  - 实现 getCacheDir() 函数，模仿 getConfigDir() 但返回 ~/.cache/pv
  - 支持 Windows (LOCALAPPDATA) 和 Unix (XDG_CACHE_HOME) 规范
  - 支持 PV_CACHE_DIR 环境变量覆盖
  - _需求: 1.1, 1.4_
  - _复用: internal/config/store.go 的 getConfigDir() 模式_

- [x] 2. 创建 CacheManager 结构体在 internal/infra/cache_manager.go  
  - 文件: internal/infra/cache_manager.go (新建)
  - 定义 CacheManager 结构体和基础构造函数
  - 实现 EnsureCacheDir() 方法创建缓存目录结构
  - 设置正确的目录权限 (0700)
  - _需求: 1.1, 1.2, 1.3_
  - _复用: internal/config/cache.go, internal/config/store.go 的 writeFileWithPermissions_

- [x] 3. 实现 CacheManager 索引操作方法在 cache_manager.go
  - 文件: internal/infra/cache_manager.go (继续任务2)
  - 添加 LoadIndex() 方法从 index.json 读取缓存索引
  - 添加 SaveIndex() 方法保存 model.Index 到 index.json
  - 使用 JSON 序列化，复用现有权限设置
  - _需求: 1.1, 2.2_
  - _复用: internal/model/index.go, internal/config/store.go_

- [x] 4. 实现 CacheManager 内容操作方法在 cache_manager.go
  - 文件: internal/infra/cache_manager.go (继续任务3)
  - 添加 LoadContent() 方法从 prompts/{gist_id}.yaml 读取原始内容
  - 添加 SaveContent() 方法保存原始 YAML 内容到文件
  - 确保文件扩展名为 .yaml，内容与 Gist 完全一致
  - _需求: 2.1, 2.2_
  - _复用: internal/config/store.go 的文件操作模式_

- [x] 5. 添加 CacheManager 辅助方法在 cache_manager.go
  - 文件: internal/infra/cache_manager.go (继续任务4)
  - 添加 GetCacheInfo() 方法返回缓存统计信息
  - 实现目录大小计算和文件计数
  - 添加错误处理，优雅处理文件系统错误
  - _需求: 性能要求，可用性要求_

### 阶段 2：缓存装饰器实现

- [x] 6. 创建 CachedStore 装饰器结构体在 internal/infra/cached_store.go
  - 文件: internal/infra/cached_store.go (新建)
  - 定义 CachedStore 结构体包装 GitHubStore 和 CacheManager
  - 实现构造函数 NewCachedStore()
  - 添加 forceRemote 标志支持
  - _需求: 2.1, 5.2_
  - _复用: internal/infra/store.go, internal/infra/github_store.go_

- [x] 7. 实现 CachedStore 的 List() 方法在 cached_store.go
  - 文件: internal/infra/cached_store.go (继续任务6)
  - 实现远程优先的 List() 逻辑
  - 支持 --remote 强制远程获取模式
  - 远程失败时回退到本地缓存
  - 成功时更新缓存索引
  - _需求: 2.1, 2.2, 2.3, 5.1, 5.2_
  - _复用: internal/infra/github_store.go 的 List() 方法_

- [x] 8. 实现 CachedStore 的 GetContent() 方法在 cached_store.go
  - 文件: internal/infra/cached_store.go (继续任务7)
  - 实现远程优先的 GetContent() 逻辑
  - 远程成功时更新缓存内容
  - 远程失败时从缓存读取
  - 保持内容完全一致性
  - _需求: 2.1, 2.2, 2.3, 4.3_
  - _复用: internal/infra/github_store.go 的 GetContent() 方法_

- [x] 9. 实现 CachedStore 其他 Store 接口方法在 cached_store.go
  - 文件: internal/infra/cached_store.go (继续任务8)
  - 实现 Add(), Delete(), Update(), Get() 方法
  - 直接委托给 GitHubStore，成功后清理相关缓存
  - 确保缓存与远程数据同步
  - _需求: 2.1, 2.2_
  - _复用: internal/infra/github_store.go 的相应方法_

### 阶段 3：命令扩展

- [x] 10. 创建 sync 命令在 cmd/sync.go
  - 文件: cmd/sync.go (新建)
  - 创建 sync 命令结构体和构造函数
  - 实现基本的命令框架，依赖 PromptService
  - 添加进度显示和错误统计功能
  - _需求: 3.1, 3.4_
  - _复用: cmd/list.go 的命令模式, cmd/add.go 的依赖注入_

- [x] 11. 实现 sync 命令的执行逻辑在 sync.go
  - 文件: cmd/sync.go (继续任务10)
  - 实现完整的同步流程：获取索引 -> 下载内容
  - 添加串行下载，显示 "正在下载 X/Y" 进度
  - 单个提示词失败时继续处理其他提示词
  - 显示最终同步统计信息
  - _需求: 3.1, 3.2, 3.3, 3.4_
  - _复用: internal/service/prompt_service.go_

- [x] 12. 修改 list 命令添加 --remote 选项在 cmd/list.go
  - 文件: cmd/list.go (修改现有)
  - 添加 --remote 布尔标志到命令定义
  - 修改执行逻辑传递 forceRemote 参数给 Store
  - 添加缓存数据显示最后更新时间
  - _需求: 5.1, 5.2, 5.3, 5.5_
  - _复用: 现有 cmd/list.go 结构_

- [x] 13. 修改 get 命令集成快速同步在 cmd/get.go  
  - 文件: cmd/get.go (修改现有)
  - 在 get 操作前添加快速同步尝试
  - 同步失败时显示警告并继续使用缓存
  - 标记缓存数据来源在输出中
  - _需求: 4.1, 4.2, 4.3, 4.4_
  - _复用: 现有 cmd/get.go 结构_

### 阶段 4：依赖注入和集成

- [x] 14. 更新 Wire 配置集成缓存组件在 internal/di/wire.go
  - 文件: internal/di/wire.go (修改现有)
  - 添加 CacheManager 和 CachedStore 的 Provider 函数
  - 修改现有 Store 注入使用 CachedStore 替代 GitHubStore
  - 保持向后兼容，不影响现有服务
  - _需求: 架构集成_
  - _复用: 现有 Wire 配置模式_

- [x] 15. 添加 sync 命令到根命令在 cmd/root.go
  - 文件: cmd/root.go (修改现有)
  - 导入 sync 命令并添加到根命令
  - 确保命令在帮助中正确显示
  - _需求: 3.1_
  - _复用: 现有命令注册模式_

### 阶段 5：数据模型扩展

- [x] 16. 创建缓存信息模型在 internal/model/cache.go
  - 文件: internal/model/cache.go (新建)  
  - 定义 CacheInfo 结构体用于缓存统计
  - 添加 JSON 标签和合理的默认值
  - _需求: 可用性要求_

### 阶段 6：测试

- [x] 17. 创建 CacheManager 单元测试在 internal/infra/cache_manager_test.go
  - 文件: internal/infra/cache_manager_test.go (新建)
  - 测试目录创建、权限设置和文件读写操作
  - 测试边界情况：损坏文件、磁盘空间不足
  - 使用临时目录避免影响真实缓存
  - _需求: 可靠性要求_
  - _复用: 现有测试模式和工具_

- [x] 18. 创建 CachedStore 单元测试在 internal/infra/cached_store_test.go
  - 文件: internal/infra/cached_store_test.go (新建)
  - 测试装饰器逻辑：远程优先、缓存回退
  - 模拟网络失败场景和缓存命中/未命中
  - 验证数据一致性和错误处理
  - _需求: 2.1, 2.2, 2.3_
  - _复用: 现有 Store 测试模式_

- [x] 19. 创建 sync 命令测试在 cmd/sync_test.go
  - 文件: cmd/sync_test.go (新建)
  - 测试同步流程的各种场景
  - 验证进度显示和错误统计
  - 测试认证失败和部分同步失败
  - _需求: 3.1, 3.2, 3.3, 3.4, 3.5_
  - _复用: 现有命令测试模式_

- [x] 20. 更新现有命令测试适配缓存功能
  - 文件: cmd/list_test.go, cmd/get_test.go (修改现有)
  - 更新测试以适配新的缓存行为
  - 验证 --remote 选项和离线模式
  - 确保向后兼容性
  - _需求: 4.1-4.5, 5.1-5.5_
  - _复用: 现有测试结构_