# MVP 重构方案

目标：仅保留 `auth/auth_login/auth_logout/auth_status`、`add`、`delete`、`get`、`list` 五个功能，尽量缩减代码体积与依赖。

## 命令层调整
- 删除 `cmd/share.go` 及相关别名；`cmd/root.go` 只注册上述五类命令（`del` 保留为 `delete` 的别名可选）。
- 精简 `cmd/get.go`、`cmd/delete.go` 展示文案，复用 URL/keyword 解析逻辑；可考虑去掉冗长提示。
- 若决定放弃剪贴板支持，`get` 直接打印内容（保留或移除 `internal/clipboard` 视选项）。

## DI 与入口
- `internal/di/providers.go` 仅提供需要的命令、服务、工具；移除 share/TUI 相关的注入项。
- 运行 `wire` 更新生成文件以匹配新的 provider 集合。

## 服务与接口
- `PromptService` 保留：`AddFromFile`、`AddFromURL`、`DeleteByKeyword`、`DeleteByURL`、`ListPrompts`、`FilterPrompts`、`GetPromptByURL`、`GetPromptContent`。删除分享和私有/公开相关接口。
- 对应实现剔除 Share/Export 逻辑、变量解析等不再需要的路径。

## 数据模型与存储
- `internal/model/index.go` 删除 `Exports`/`Parent` 等分享字段。
- `internal/infra/store.go` 接口收缩到列表/增删改查基础方法；去掉导出与公开 Gist 管理方法。
- `internal/infra/github_store.go` 移除 CreatePublicGist/UpdateGist/GetExports/AddExport/UpdateExport/GetGistInfo 等分享相关实现；保留最小可用的索引、Gist CRUD，并确保 add/upload 始终创建公开 Gist。

## TUI/变量/剪贴板
- 若保留交互选择，保留 `tui` 的提示列表和表单；否则改为非交互模式并移除 `internal/tui`/`internal/variable` 依赖以进一步瘦身。
- 根据剪贴板支持决策保留或删除 `internal/clipboard` 并调整 `get` 命令逻辑。

## 文档与测试
- 更新 README 命令矩阵与功能描述，仅涵盖保留的功能。
- 删除 share/exports/同步相关测试与 fixtures；调整剩余测试匹配精简后的接口。
- 最终执行 `make fmt`、`make test-quick`（或 `go test ./...`）验证。
