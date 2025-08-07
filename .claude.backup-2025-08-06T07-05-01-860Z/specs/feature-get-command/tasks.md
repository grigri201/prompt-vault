# 任务清单

## 功能名称
feature-get-command

## 功能描述
为 `pv get` 命令实现交互式提示词选择、关键字筛选、URL 直接访问，支持变量替换和剪贴板集成

## 任务依赖
按顺序执行，后续任务依赖前面任务的完成

## 任务列表

### 阶段一：服务层扩展

#### 1. 扩展 PromptService 接口方法
**关联需求**: FR1, TR1
**文件**: `internal/service/prompt_service.go`, `internal/service/prompt_service_impl.go`
- [x] 1.1 重构现有方法：`ListForDeletion` → `ListPrompts`
- [x] 1.2 重构现有方法：`FilterForDeletion` → `FilterPrompts`
- [x] 1.3 新增方法：`GetPromptByURL(gistURL string) (*model.Prompt, error)`
- [x] 1.4 新增方法：`GetPromptContent(prompt *model.Prompt) (string, error)`
- [x] 1.5 更新服务接口文档为重构后的方法名
- [x] 1.6 为新方法添加全面的单元测试覆盖

#### 2. 更新现有命令使用重构方法
**关联需求**: FR1
**文件**: `cmd/delete.go`, `cmd/list.go`
- [x] 2.1 更新 delete 命令使用 `ListPrompts` 替代 `ListForDeletion`
- [x] 2.2 更新 delete 命令使用 `FilterPrompts` 替代 `FilterForDeletion`
- [x] 2.3 更新 list 命令使用重构后的通用方法
- [x] 2.4 验证现有功能正常运行

### 阶段二：新组件实现

#### 3. 变量解析器实现
**关联需求**: FR3, TR2
**文件**: `internal/variable/parser.go`, `internal/variable/parser_test.go`
- [x] 3.1 创建 `Parser` 接口定义
- [x] 3.2 实现 `ExtractVariables(content string) []string` 方法
- [x] 3.3 实现 `ReplaceVariables(content string, values map[string]string) string` 方法
- [x] 3.4 实现 `HasVariables(content string) bool` 方法
- [x] 3.5 添加工厂函数 `NewParser() Parser`
- [x] 3.6 创建全面的单元测试，覆盖各种边缘情况

#### 4. 剪贴板工具实现
**关联需求**: US5, TR2
**文件**: `internal/clipboard/util.go`, `internal/clipboard/util_test.go`
- [x] 4.1 创建 `Util` 接口定义
- [x] 4.2 实现 `Copy(content string) error` 方法
- [x] 4.3 实现 `IsAvailable() bool` 方法
- [x] 4.4 添加工厂函数 `NewUtil() Util`
- [x] 4.5 添加 `github.com/atotto/clipboard` 依赖到 `go.mod`
- [x] 4.6 创建跨平台兼容的单元测试

#### 5. 错误处理类型定义
**关联需求**: FR4, TR2
**文件**: `internal/errors/get_errors.go`, `internal/errors/get_errors_test.go`
- [x] 5.1 创建 `GetError` 错误结构
- [x] 5.2 定义具体错误实例：`ErrNoPromptsFound`, `ErrInvalidGistURL`, `ErrVariableRequired`, `ErrClipboardUnavailable`
- [x] 5.3 实现错误接口 `Error() string`
- [x] 5.4 添加错误处理测试

### 阶段三：TUI 组件扩展

#### 6. 扩展 TUIInterface 接口
**关联需求**: FR2, TR2
**文件**: `internal/tui/interface.go`
- [x] 6.1 添加 `ShowVariableForm(variables []string) (map[string]string, error)` 方法声明

#### 7. 实现 VariableFormModel 组件
**关联需求**: US4, FR2
**文件**: `internal/tui/variable_form.go`, `internal/tui/variable_form_test.go`
- [x] 7.1 创建 `VariableFormModel` 结构体
- [x] 7.2 实现 `NewVariableFormModel(variables []string) VariableFormModel` 工厂函数
- [x] 7.3 实现 `Init()` 方法，初始化 bubbletea 模型
- [x] 7.4 实现 `Update(msg tea.Msg)` 方法，处理用户输入和键盘事件
- [x] 7.5 实现 `View()` 方法，渲染表单界面
- [x] 7.6 添加辅助方法：`IsDone()`, `HasCancelled()`, `GetValues()`, `GetError()`
- [x] 7.7 添加 `github.com/charmbracelet/bubbles/textinput` 依赖
- [x] 7.8 创建 TUI 组件单元测试（使用 Mock）

#### 8. 更新 BubbleTeaTUI 工厂
**关联需求**: FR2
**文件**: `internal/tui/factory.go`
- [x] 8.1 实现 `ShowVariableForm` 方法
- [x] 8.2 集成 VariableFormModel 到现有 TUI 工厂
- [x] 8.3 处理表单运行和状态提取

### 阶段四：命令层实现

#### 9. 创建 get 命令结构
**关联需求**: US1, US2, US3, FR1
**文件**: `cmd/get.go`, `cmd/get_test.go`
- [x] 9.1 创建 `get` 结构体和依赖注入
- [x] 9.2 实现 `execute` 方法，处理三种执行模式
- [x] 9.3 实现 `handleInteractiveMode()` - 交互式提示词选择
- [x] 9.4 实现 `handleFilterMode(keyword)` - 关键字筛选模式
- [x] 9.5 实现 `handleDirectMode(gistURL)` - 直接 URL 模式
- [x] 9.6 实现 URL 检测工具 `isGistURL(arg string) bool`
- [x] 9.7 创建命令层单元测试

#### 10. 实现核心处理流程
**关联需求**: US4, US5, FR3
**文件**: `cmd/get.go`
- [x] 10.1 实现 `processSelectedPrompt(prompt model.Prompt) error` 方法
- [x] 10.2 实现 `showPromptSelection(prompts []model.Prompt) (model.Prompt, error)` 方法
- [x] 10.3 实现 `showVariableForm(variables []string) (map[string]string, error)` 方法
- [x] 10.4 实现 `copyToClipboard(content, promptName string) error` 方法
- [x] 10.5 集成完整的变量处理和剪贴板流程

#### 11. 创建 GetCommand 工厂函数
**关联需求**: FR1, TR1
**文件**: `cmd/get.go`
- [x] 11.1 实现 `NewGetCommand` 工厂函数，集成 Cobra 命令
- [x] 11.2 配置命令属性：`Use`, `Short`, `Long`, `Example`
- [x] 11.3 添加参数验证：`cobra.MaximumNArgs(1)`
- [x] 11.4 完善命令帮助和文档

### 阶段五：依赖注入配置

#### 12. 更新依赖注入提供者
**关联需求**: TR1
**文件**: `internal/di/providers.go`, `internal/di/wire.go`
- [x] 12.1 添加 `ProvideVariableParser() variable.Parser` 提供者
- [x] 12.2 添加 `ProvideClipboardUtil() clipboard.Util` 提供者
- [x] 12.3 更新 `ProvideCommands` 工厂函数，包含 GetCommand
- [x] 12.4 更新 Wire 配置，包含新组件提供者
- [x] 12.5 运行 `go generate ./internal/di` 重新生成依赖注入代码

#### 13. 集成根命令
**关联需求**: FR1
**文件**: `internal/di/providers.go`
- [x] 13.1 在 `ProvideRootCommand` 中添加 get 命令
- [x] 13.2 验证命令正确集成到 CLI 结构中

### 阶段六：测试验证

#### 14. 服务层测试
**关联需求**: TR3
**文件**: `internal/service/prompt_service_test.go`
- [x] 14.1 验证现有测试通过重构
- [x] 14.2 测试 `GetPromptByURL` 方法覆盖
- [x] 14.3 测试 `GetPromptContent` 方法覆盖
- [x] 14.4 验证测试覆盖率达标（>80%）

#### 15. 命令层测试
**关联需求**: US1, US2, US3, TR3
**文件**: `cmd/get_test.go`
- [x] 15.1 创建交互模式测试（使用 Mock TUI）
- [x] 15.2 创建关键字筛选模式测试
- [x] 15.3 创建直接 URL 模式测试
- [x] 15.4 测试错误处理场景
- [x] 15.5 测试工厂函数覆盖

#### 16. TUI 集成测试
**关联需求**: US4, TR3
**文件**: `internal/tui/integration_test.go`
- [x] 16.1 创建变量表单集成测试（使用 go-expect）
- [x] 16.2 测试键盘导航和输入处理
- [x] 16.3 测试表单验证和错误处理
- [x] 16.4 验证 TUI 测试在 CI 环境中正确跳过

### 阶段七：文档和完善

#### 17. 更新项目文档
**关联需求**: 所有需求
**文件**: `CLAUDE.md`, `README.md`
- [x] 17.1 更新 CLAUDE.md 中的功能描述
- [x] 17.2 添加 get 命令使用说明
- [x] 17.3 更新依赖列表和开发命令
- [x] 17.4 补充测试运行指南

#### 18. 性能验证
**关联需求**: NFR1, NFR2, NFR3
- [x] 18.1 运行完整测试套件，验证无回归
- [x] 18.2 验证性能指标：列表加载 < 2s，表单响应 < 100ms
- [x] 18.3 测试剪贴板集成：错误处理和回退机制
- [x] 18.4 跨平台测试：Windows, macOS, Linux 兼容性
- [x] 18.5 最终构建验证和部署准备

## 完成标准

### 功能性
- 三种 get 命令模式按规范工作
- 变量替换正确处理复杂情况
- 剪贴板集成在各平台正常工作

### 代码质量
- 遵循现有代码规范和风格
- 通过 `go fmt` 和 `go vet` 检查
- 所有 TUI 交互保持一致性

### 测试覆盖
- 单元测试覆盖率 >80%
- 所有主要错误场景有测试
- TUI 测试使用 Mock 而非真实终端
- 集成测试覆盖端到端场景

### 性能需求
- 列表加载在正常网络下 < 2s
- 表单 TUI 响应时间 < 100ms
- 剪贴板操作兼容性

## 备注
- [ ] 所有任务已经完成并集成
- [ ] 功能通过完整验证
- [ ] TUI 界面与现有命令一致
- [ ] 剪贴板在各平台正常工作
- [ ] 所有测试通过并达标
- [ ] 性能满足需求标准