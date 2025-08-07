# Share 命令任务分解

## 任务概述
将 share 命令功能分解为原子化的实现任务，每个任务专注于单一功能点，可独立实现和测试。

## 任务分类

### 阶段 1: 数据模型增强
- [ ] **Task 1.1**: 增强 Prompt 模型添加 Parent 字段 (TR1)
  - 文件: `internal/model/prompt.go`
  - 添加 `Parent *string` 字段到 Prompt 结构体
  - 更新 JSON 标签为 `json:"parent,omitempty"`
  - 需求: TR1 - Prompt 模型增强

- [ ] **Task 1.2**: 增强 IndexedPrompt 模型添加 Parent 字段 (TR2) 
  - 文件: `internal/model/index.go`
  - 添加 `Parent *string` 字段到 IndexedPrompt 结构体
  - 更新 JSON 标签为 `json:"parent,omitempty"`
  - 需求: TR2 - Index 模型增强

- [ ] **Task 1.3**: 增强 Index 模型添加 Exports 字段 (TR2)
  - 文件: `internal/model/index.go`
  - 添加 `Exports []IndexedPrompt` 字段到 Index 结构体
  - 更新 JSON 标签为 `json:"exports"`
  - 需求: TR2 - Index 模型增强

### 阶段 2: 错误处理扩展
- [ ] **Task 2.1**: 创建 share 相关错误定义
  - 文件: `internal/errors/share_errors.go`
  - 定义 share 操作相关的错误类型和构造函数
  - 包含: ErrGistAlreadyPublic, ErrGistAccessDenied, ErrInvalidGistURL 等
  - 需求: NFR3 - 中文错误消息

- [ ] **Task 2.2**: 扩展现有错误处理支持 add URL 功能
  - 文件: `internal/errors/prompt_errors.go`
  - 添加 URL 导入相关错误: ErrGistNotPublic, ErrInvalidPromptFormat
  - 需求: US5 - Add 命令增强

### 阶段 3: Store 接口扩展
- [ ] **Task 3.1**: 扩展 Store 接口添加 gist 管理方法
  - 文件: `internal/infra/store.go`
  - 添加方法: CreatePublicGist, UpdateGist, GetGistInfo
  - 需求: TR4 - Gist 管理

- [ ] **Task 3.2**: 扩展 Store 接口添加 export 管理方法
  - 文件: `internal/infra/store.go`
  - 添加方法: AddExport, UpdateExport, GetExports
  - 需求: TR2 - Index 模型增强

- [ ] **Task 3.3**: 实现 GitHubStore 的 CreatePublicGist 方法
  - 文件: `internal/infra/github_store.go`
  - 实现创建公开 gist 的 GitHub API 调用
  - 处理认证和错误情况
  - 需求: TR4 - Gist 管理

- [ ] **Task 3.4**: 实现 GitHubStore 的 UpdateGist 方法
  - 文件: `internal/infra/github_store.go`
  - 实现更新现有 gist 内容的 GitHub API 调用
  - 需求: TR4 - Gist 管理

- [ ] **Task 3.5**: 实现 GitHubStore 的 GetGistInfo 方法
  - 文件: `internal/infra/github_store.go`
  - 实现获取 gist 信息（可见性、权限等）的 GitHub API 调用
  - 需求: BR1 - 访问控制

- [ ] **Task 3.6**: 实现 GitHubStore 的 export 管理方法
  - 文件: `internal/infra/github_store.go`
  - 实现 AddExport, UpdateExport, GetExports 方法
  - 操作 index.json 中的 exports 字段
  - 需求: BR2 - 导出管理

### 阶段 4: Service 层扩展
- [ ] **Task 4.1**: 扩展 PromptService 接口添加 share 方法
  - 文件: `internal/service/prompt_service.go`
  - 添加 SharePrompt, ValidateGistAccess, ListPrivatePrompts 等方法
  - 定义 GistInfo 结构体
  - 需求: TR3 - Share 命令模式

- [ ] **Task 4.2**: 扩展 PromptService 接口添加 URL 导入方法
  - 文件: `internal/service/prompt_service.go`  
  - 添加 AddFromURL, FilterPrivatePrompts 方法
  - 需求: US5 - Add 命令增强

- [ ] **Task 4.3**: 实现 PromptService 的 SharePrompt 方法
  - 文件: `internal/service/prompt_service_impl.go`
  - 实现完整的分享流程逻辑（验证权限、检查导出、创建/更新 gist）
  - 需求: US4 - 分享流程执行

- [ ] **Task 4.4**: 实现 PromptService 的 AddFromURL 方法
  - 文件: `internal/service/prompt_service_impl.go`
  - 实现从公开 gist URL 导入 prompt 的逻辑
  - 验证 gist 可见性和格式
  - 需求: US5 - Add 命令增强

- [ ] **Task 4.5**: 实现 PromptService 的私有 prompt 筛选方法
  - 文件: `internal/service/prompt_service_impl.go`
  - 实现 ListPrivatePrompts, FilterPrivatePrompts 方法
  - 需求: US1, US2 - 交互式和筛选分享

- [ ] **Task 4.6**: 实现 PromptService 的 ValidateGistAccess 方法
  - 文件: `internal/service/prompt_service_impl.go`
  - 验证用户对 gist 的访问权限和可见性
  - 需求: BR1 - 访问控制

### 阶段 5: Share 命令实现
- [ ] **Task 5.1**: 创建 share 命令基础结构
  - 文件: `cmd/share.go`
  - 创建 share 命令结构体和基本命令定义
  - 实现参数验证和模式判断逻辑
  - 需求: TR3 - Share 命令模式

- [ ] **Task 5.2**: 实现 share 命令的交互式模式
  - 文件: `cmd/share.go`
  - 实现 handleInteractiveMode 方法
  - 调用 TUI 显示私有 prompts 列表
  - 需求: US1 - 交互式分享选择

- [ ] **Task 5.3**: 实现 share 命令的关键字筛选模式
  - 文件: `cmd/share.go`
  - 实现 handleFilterMode 方法
  - 筛选并显示匹配的私有 prompts
  - 需求: US2 - 关键字过滤分享

- [ ] **Task 5.4**: 实现 share 命令的直接 URL 模式
  - 文件: `cmd/share.go`
  - 实现 handleDirectMode 方法
  - 验证 URL 并直接执行分享流程
  - 需求: US3 - 直接 URL 分享

- [ ] **Task 5.5**: 实现 share 命令的 URL 验证辅助方法
  - 文件: `cmd/share.go`
  - 重用 delete 命令的 URL 验证逻辑
  - 实现 isGistURL, looksLikeURL 等方法
  - 需求: BR3 - 数据完整性

- [ ] **Task 5.6**: 添加 share 命令到根命令
  - 文件: `cmd/root.go`
  - 将 ShareCmd 注册到根命令
  - 更新依赖注入配置
  - 需求: C3 - CLI 一致性

### 阶段 6: Add 命令增强
- [ ] **Task 6.1**: 增强 add 命令支持 URL 参数
  - 文件: `cmd/add.go`
  - 修改参数验证逻辑，支持 URL 格式检测
  - 区分文件路径和 gist URL 参数
  - 需求: US5 - Add 命令增强

- [ ] **Task 6.2**: 实现 add 命令的 URL 处理流程
  - 文件: `cmd/add.go`
  - 添加 handleURLMode 方法处理 gist URL 导入
  - 调用 AddFromURL 服务方法
  - 需求: US5 - Add 命令增强

### 阶段 7: 依赖注入配置
- [ ] **Task 7.1**: 更新 Wire 配置添加新依赖
  - 文件: `internal/di/wire.go`
  - 添加 share 命令的依赖注入配置
  - 确保所有新服务方法可用
  - 需求: 架构一致性

- [ ] **Task 7.2**: 更新 providers 提供新的构造函数
  - 文件: `internal/di/providers.go`
  - 添加 NewShareCommand 构造函数
  - 更新现有构造函数支持新接口
  - 需求: 架构一致性

- [ ] **Task 7.3**: 重新生成 Wire 代码
  - 文件: `internal/di/wire_gen.go`
  - 运行 `go generate ./internal/di` 生成新的依赖注入代码
  - 需求: 构建流程

### 阶段 8: 测试实现
- [ ] **Task 8.1**: 添加 share 相关 service 层单元测试
  - 文件: `internal/service/prompt_service_test.go`
  - 测试 SharePrompt, AddFromURL 等方法的各种场景
  - Mock GitHub API 调用
  - 需求: 质量保证

- [ ] **Task 8.2**: 添加 share 相关 store 层单元测试
  - 文件: `internal/infra/github_store_test.go`
  - 测试 gist 创建、更新、信息获取等功能
  - 需求: 质量保证

- [ ] **Task 8.3**: 添加 share 命令单元测试
  - 文件: `cmd/share_test.go`
  - 测试三种模式的参数解析和执行逻辑
  - 需求: 质量保证

- [ ] **Task 8.4**: 添加数据模型测试
  - 文件: `internal/model/prompt_test.go`, `internal/model/index_test.go`
  - 测试新字段的序列化/反序列化
  - 需求: BR3 - 数据完整性

### 阶段 9: 集成测试和文档
- [ ] **Task 9.1**: 添加 share 命令集成测试
  - 文件: `integration/share_test.go`
  - 测试完整的分享工作流程
  - 使用 mock GitHub API
  - 需求: 质量保证

- [ ] **Task 9.2**: 更新 CLAUDE.md 文档
  - 文件: `CLAUDE.md`
  - 添加 share 命令的使用说明
  - 更新功能特性列表
  - 需求: 文档完整性

- [ ] **Task 9.3**: 添加命令行帮助文本
  - 文件: `cmd/share.go`, `cmd/add.go`
  - 完善命令描述、用法示例和帮助文本
  - 确保中文本地化
  - 需求: NFR3 - 可用性

## 任务依赖关系

### 关键路径
1. **数据模型** (Tasks 1.1-1.3) → **错误处理** (Tasks 2.1-2.2) → **Store 接口** (Tasks 3.1-3.6)
2. **Store 实现** → **Service 接口** (Tasks 4.1-4.2) → **Service 实现** (Tasks 4.3-4.6)
3. **Service 层** → **命令实现** (Tasks 5.1-5.6, 6.1-6.2) → **依赖注入** (Tasks 7.1-7.3)

### 并行执行机会
- **阶段 2 (错误处理)** 可以与 **阶段 3 (Store 接口)** 并行
- **阶段 8 (测试)** 可以在相应功能完成后立即开始
- **阶段 9 (文档)** 可以在功能稳定后开始

## 验收标准检查清单

### 功能验收
- [ ] `pv share` 显示交互式私有 prompts 列表
- [ ] `pv share <keyword>` 根据关键字筛选私有 prompts
- [ ] `pv share <gist_url>` 直接分享指定 URL 的 prompt
- [ ] 首次分享创建公开 gist，再次分享更新内容
- [ ] `pv add <gist_url>` 从公开 gist URL 导入 prompt
- [ ] 所有错误消息使用中文显示

### 技术验收
- [ ] 数据模型正确支持 parent 关系和 exports
- [ ] GitHub API 集成正确处理认证和错误
- [ ] TUI 组件重用现有界面风格
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试覆盖主要工作流程

### 非功能验收
- [ ] 分享操作在 10 秒内完成
- [ ] TUI 响应时间 < 100ms
- [ ] 现有命令功能不受影响
- [ ] 向后兼容现有数据格式

## 实现优先级

### P0 (核心功能)
- Tasks 1.1-1.3 (数据模型)
- Tasks 3.1-3.6 (Store 层)
- Tasks 4.1, 4.3 (Service 核心方法)
- Tasks 5.1-5.4 (Share 命令核心)

### P1 (重要功能)
- Tasks 2.1-2.2 (错误处理)
- Tasks 4.4-4.6 (Service 辅助方法)
- Tasks 6.1-6.2 (Add 命令增强)
- Tasks 7.1-7.3 (依赖注入)

### P2 (测试和完善)
- Tasks 8.1-8.4 (单元测试)
- Tasks 9.1-9.3 (集成测试和文档)

每个任务都被设计为原子化，可以独立实现和测试，同时明确了依赖关系和验收标准。