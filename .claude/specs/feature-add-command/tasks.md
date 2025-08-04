# Tasks - Add Command Feature

## 任务概览

将 add 命令功能分解为 15 个原子性任务，按依赖关系分组执行。每个任务专注于 1-3 个相关文件，可在 15-30 分钟内完成。

## 任务分组和依赖关系

### 第一组：基础模型和错误定义 (并行执行)
这些任务为其他组件提供基础结构，可以并行开发。

### 第二组：验证和服务层 (依赖第一组)
实现核心业务逻辑，依赖于基础模型和错误定义。

### 第三组：命令和集成 (依赖第二组)
实现 CLI 命令和依赖注入，需要服务层组件。

### 第四组：增强和完善 (依赖第三组)
完善现有组件以支持新功能。

## 详细任务列表

### 第一组：基础结构 (并行执行)

#### 任务 1：扩展 Prompt 模型
- [x] **文件**: `internal/model/prompt.go`
- [ ] **需求引用**: FR-002 (文件格式验证), TC-001 (YAML 格式规范)
- [ ] **描述**: 在现有 Prompt 结构体中添加 Description、Tags、Version、Content 字段
- [ ] **验收标准**: 
  - 添加新字段：Description string, Tags []string, Version string, Content string
  - 保持现有字段不变 (ID, Name, Author, GistURL)
  - 添加适当的 JSON 标签用于序列化
- [ ] **现有代码复用**: 扩展现有的 Prompt 结构体定义

#### 任务 2：创建 prompt 相关错误定义
- [x] **文件**: `internal/errors/prompt_errors.go`
- [ ] **需求引用**: FR-004 (错误处理)
- [ ] **描述**: 创建新文件定义 prompt 操作相关的错误类型和变量
- [ ] **验收标准**:
  - 定义 ErrFileNotFound, ErrInvalidYAML, ErrMissingRequired, ErrInvalidMetadata 错误变量
  - 创建 ValidationError 结构体，包含 Field 和 Message 字段
  - 实现 ValidationError 的 Error() 方法
- [ ] **现有代码复用**: 参考 `internal/errors/auth_errors.go` 的错误定义模式

#### 任务 3：创建 YAML 验证器接口
- [x] **文件**: `internal/validator/yaml_validator.go`
- [ ] **需求引用**: FR-002 (文件格式验证), US-002 (YAML 格式验证)
- [ ] **描述**: 定义 YAML 验证相关的接口和数据结构
- [ ] **验收标准**:
  - 定义 YAMLValidator 接口，包含 ValidatePromptFile 和 ValidateRequired 方法
  - 定义 PromptFileContent 结构体，包含 Metadata 和 Content 字段
  - 定义 PromptMetadata 结构体，包含所有 YAML 元数据字段及其标签
- [ ] **现有代码复用**: 新建文件，但遵循项目的接口定义模式

### 第二组：核心业务逻辑 (依赖第一组)

#### 任务 4：实现 YAML 验证器
- [x] **文件**: `internal/validator/yaml_validator_impl.go`
- [ ] **需求引用**: FR-002 (文件格式验证), TC-001 (YAML 格式规范)
- [ ] **描述**: 实现 YAMLValidator 接口，提供实际的 YAML 解析和验证功能
- [ ] **验收标准**:
  - 实现 yamlValidatorImpl 结构体
  - 实现 ValidatePromptFile 方法：解析 YAML 前置数据和内容部分
  - 实现 ValidateRequired 方法：验证必需字段 (name, author)
  - 处理 YAML 解析错误并返回适当的错误类型
- [ ] **现有代码复用**: 新实现，但参考项目中其他 impl 文件的结构模式

#### 任务 5：创建 PromptService 接口
- [x] **文件**: `internal/service/prompt_service.go`
- [x] **需求引用**: US-001 (添加单个 Prompt 文件), FR-001 (命令行接口)
- [x] **描述**: 定义 prompt 业务逻辑的服务接口
- [x] **验收标准**:
  - 定义 PromptService 接口，包含 AddFromFile(filePath string) (*model.Prompt, error) 方法
  - 添加适当的导入语句和包声明
- [x] **现有代码复用**: 参考 `internal/service/auth_service.go` 的接口定义模式

#### 任务 6：实现 PromptService
- [x] **文件**: `internal/service/prompt_service_impl.go`
- [ ] **需求引用**: US-001, US-003 (GitHub Gist 集成), US-004 (索引更新)
- [ ] **描述**: 实现 PromptService 接口，协调文件读取、验证、存储等操作
- [ ] **验收标准**:
  - 实现 promptServiceImpl 结构体，包含 store 和 validator 依赖
  - 实现 AddFromFile 方法：读取文件、验证内容、转换为 Prompt 模型、调用 store.Add
  - 实现 NewPromptService 构造函数
  - 完整的错误处理和类型转换
- [ ] **现有代码复用**: 参考 `internal/service/auth_service_impl.go` 的实现模式

### 第三组：命令层和依赖注入 (依赖第二组)

#### 任务 7：创建 add 命令实现
- [ ] **文件**: `cmd/add.go`
- [ ] **需求引用**: FR-001 (命令行接口), US-001 (添加单个 Prompt 文件)
- [ ] **描述**: 实现 add 命令的 Cobra 命令结构和执行逻辑
- [ ] **验收标准**:
  - 定义 AddCmd 类型别名
  - 实现 add 结构体，包含 promptService 依赖
  - 实现 execute 方法：参数验证、调用服务、错误处理、成功消息显示
  - 实现 NewAddCommand 构造函数，配置 Cobra 命令参数
- [ ] **现有代码复用**: 完全复用 `cmd/list.go` 的结构和模式

#### 任务 8：添加 YAML 验证器的依赖注入配置
- [x] **文件**: `internal/di/providers.go`
- [x] **需求引用**: 架构集成要求
- [x] **描述**: 在现有 providers 文件中添加 YAMLValidator 的 provider 函数
- [x] **验收标准**:
  - 添加 ProvideYAMLValidator() validator.YAMLValidator 函数
  - 函数返回 validator.NewYAMLValidator() 实例
  - 添加必要的导入语句
- [x] **现有代码复用**: 扩展现有文件，遵循现有 provider 函数的模式

#### 任务 9：添加 PromptService 的依赖注入配置
- [x] **文件**: `internal/di/providers.go`
- [x] **需求引用**: 架构集成要求
- [x] **描述**: 在现有 providers 文件中添加 PromptService 的 provider 函数
- [x] **验收标准**:
  - 添加 ProvidePromptService(store infra.Store, validator validator.YAMLValidator) service.PromptService 函数
  - 函数返回 service.NewPromptService(store, validator) 实例
  - 添加必要的导入语句
- [x] **现有代码复用**: 扩展现有文件，遵循现有 provider 函数的模式

#### 任务 10：更新 Commands 结构体和 provider
- [x] **文件**: `internal/di/providers.go`
- [ ] **需求引用**: 架构集成要求
- [ ] **描述**: 修改现有的 Commands 结构体和 ProvideCommands 函数以包含 AddCmd
- [ ] **验收标准**:
  - 在 Commands 结构体中添加 AddCmd cmd.AddCmd 字段
  - 修改 ProvideCommands 函数添加 promptService 参数
  - 在 ProvideCommands 中调用 cmd.NewAddCommand(promptService)
  - 更新返回的 Commands 结构体包含 AddCmd
- [ ] **现有代码复用**: 修改现有的 Commands 结构体和函数

### 第四组：集成和完善 (依赖第三组)

#### 任务 11：更新 root 命令集成 add 命令
- [x] **文件**: `cmd/root.go`
- [x] **需求引用**: FR-001 (命令行接口)
- [x] **描述**: 修改 root 命令以包含新的 add 命令
- [x] **验收标准**:
  - 修改 NewRootCommand 函数签名添加 addCmd AddCmd 参数
  - 在 root.AddCommand 调用中添加 addCmd
  - 确保现有功能保持不变
- [x] **现有代码复用**: 修改现有的 NewRootCommand 函数

#### 任务 12：更新 Wire 配置
- [x] **文件**: `internal/di/wire.go`
- [x] **需求引用**: 架构集成要求
- [x] **描述**: 更新 Wire 依赖注入配置以包含新的 providers
- [x] **验收标准**:
  - 在 BuildCLI 函数的 wire.Build 调用中添加新的 providers (ServiceSet包含所需providers)
  - 添加 ProvideYAMLValidator 和 ProvidePromptService 到 providers 列表 (已在ServiceSet中配置)
  - 确保依赖关系正确配置 (已正确配置)
- [x] **现有代码复用**: 修改现有的 wire.Build 配置

#### 任务 13：增强 GitHubStore.Add 方法支持文件内容
- [x] **文件**: `internal/infra/github_store.go`
- [x] **需求引用**: US-003 (GitHub Gist 集成), FR-003 (GitHub Gist 创建)
- [x] **描述**: 修改现有的 Add 方法以支持上传实际的 prompt 文件内容而非占位符
- [x] **验收标准**:
  - 修改 gist 创建逻辑使用 prompt.Content 而非硬编码占位符
  - 更新 gist 描述包含更多 prompt 元数据 (如 description)
  - 文件名使用 prompt.Name + ".yaml" 格式
  - 保持现有的索引更新逻辑
- [x] **现有代码复用**: 修改现有的 Add 方法实现

### 第五组：测试和验证 (可选，依赖所有其他组)

#### 任务 14：创建 PromptService 单元测试
- [x] **文件**: `internal/service/prompt_service_test.go`
- [x] **需求引用**: 测试策略要求
- [x] **描述**: 为 PromptService 实现创建单元测试
- [x] **验收标准**:
  - 测试 AddFromFile 的成功场景
  - 测试文件不存在的错误场景
  - 测试 YAML 验证失败的错误场景
  - 使用 mock Store 和 validator 进行测试
- [x] **现有代码复用**: 参考 `internal/service/auth_service_test.go` 的测试模式

#### 任务 15：创建 YAML 验证器单元测试
- [x] **文件**: `internal/validator/yaml_validator_test.go`
- [x] **需求引用**: 测试策略要求
- [x] **描述**: 为 YAMLValidator 实现创建单元测试
- [x] **验收标准**:
  - 测试有效 YAML 文件的解析
  - 测试无效 YAML 格式的错误处理
  - 测试缺少必需字段的验证
  - 测试各种边界情况
- [x] **现有代码复用**: 参考现有测试文件的结构和模式

## 执行顺序建议

### 阶段 1 (并行)：基础结构
```
任务 1 → 任务 2 → 任务 3
```

### 阶段 2 (顺序)：核心逻辑  
```
任务 4 → 任务 5 → 任务 6
```

### 阶段 3 (顺序)：命令集成
```
任务 7 → 任务 8 → 任务 9 → 任务 10
```

### 阶段 4 (顺序)：系统集成
```
任务 11 → 任务 12 → 任务 13
```

### 阶段 5 (并行)：测试
```
任务 14 → 任务 15
```

## 完成标准

所有任务完成后，用户应该能够：
- [ ] 执行 `pv add /path/to/prompt.yaml` 命令
- [ ] 看到文件格式验证和错误提示
- [ ] 成功添加有效的 prompt 文件到 GitHub Gist
- [ ] 在 `pv list` 命令中看到新添加的 prompt
- [ ] 获得清晰的错误消息用于各种失败场景

## 风险和依赖

**关键依赖**:
- 任务 4-6 依赖任务 1-3 完成
- 任务 7-10 依赖任务 4-6 完成  
- 任务 11-13 依赖任务 7-10 完成

**潜在风险**:
- GitHub API 集成可能需要额外的认证处理
- YAML 解析库选择可能影响验证实现
- 现有 Store 接口可能需要微调以支持新功能