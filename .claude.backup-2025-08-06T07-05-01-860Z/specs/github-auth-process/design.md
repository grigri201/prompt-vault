# 设计文档 - github-auth-process

## 架构概览

GitHub 认证功能将遵循现有的清洁架构模式，通过新增认证相关的服务、存储和命令来实现。整体设计保持与现有代码库的一致性，使用依赖注入和接口驱动的方式。

```mermaid
graph TB
    subgraph "CLI Layer"
        AC[auth command]
        AL[auth login]
        AS[auth status]
        ALO[auth logout]
    end
    
    subgraph "Service Layer"
        ASV[AuthService]
    end
    
    subgraph "Infrastructure Layer"
        CS[ConfigStore]
        GC[GitHubClient]
        TV[TokenValidator]
    end
    
    subgraph "External"
        GA[GitHub API]
        FS[File System]
    end
    
    AC --> AL
    AC --> AS
    AC --> ALO
    
    AL --> ASV
    AS --> ASV
    ALO --> ASV
    
    ASV --> CS
    ASV --> GC
    ASV --> TV
    
    CS --> FS
    GC --> GA
    TV --> GA
```

## 组件设计

### 1. CLI 命令组件

#### AuthCmd (cmd/auth.go)
```go
type AuthCmd = *cobra.Command

func NewAuthCommand(loginCmd AuthLoginCmd, statusCmd AuthStatusCmd, logoutCmd AuthLogoutCmd) AuthCmd {
    cmd := &cobra.Command{
        Use:   "auth",
        Short: "Manage GitHub authentication",
    }
    cmd.AddCommand(loginCmd, statusCmd, logoutCmd)
    return cmd
}
```

#### AuthLoginCmd (cmd/auth_login.go)
```go
type AuthLoginCmd = *cobra.Command

type authLogin struct {
    service service.AuthService
}

func NewAuthLoginCommand(service service.AuthService) AuthLoginCmd {
    al := &authLogin{service: service}
    return &cobra.Command{
        Use:   "login",
        Short: "Login with GitHub Personal Access Token",
        Run:   al.execute,
    }
}
```

#### AuthStatusCmd (cmd/auth_status.go)
```go
type AuthStatusCmd = *cobra.Command

type authStatus struct {
    service service.AuthService
}

func NewAuthStatusCommand(service service.AuthService) AuthStatusCmd {
    as := &authStatus{service: service}
    return &cobra.Command{
        Use:   "status",
        Short: "Check authentication status",
        Run:   as.execute,
    }
}
```

#### AuthLogoutCmd (cmd/auth_logout.go)
```go
type AuthLogoutCmd = *cobra.Command

type authLogout struct {
    service service.AuthService
}

func NewAuthLogoutCommand(service service.AuthService) AuthLogoutCmd {
    al := &authLogout{service: service}
    return &cobra.Command{
        Use:   "logout",
        Short: "Logout and clear authentication",
        Run:   al.execute,
    }
}
```

### 2. 服务层组件

#### AuthService (internal/service/auth_service.go)
```go
type AuthService interface {
    Login(token string) error
    GetStatus() (*AuthStatus, error)
    Logout() error
}

type authService struct {
    configStore config.Store
    githubClient auth.GitHubClient
    tokenValidator auth.TokenValidator
}

type AuthStatus struct {
    IsAuthenticated bool
    Username        string
    Email           string
}
```

### 3. 基础设施层组件

#### ConfigStore (internal/config/store.go)
```go
type Store interface {
    SaveToken(token string) error
    GetToken() (string, error)
    DeleteToken() error
    GetConfigPath() string
}

type fileStore struct {
    configPath string
}

type Config struct {
    GitHubToken string `json:"github_token,omitempty"`
}
```

#### GitHubClient (internal/auth/github_client.go)
```go
type GitHubClient interface {
    GetAuthenticatedUser(token string) (*User, error)
    ValidateScopes(token string) ([]string, error)
}

type githubClient struct {
    baseURL string
    client  *http.Client
}

type User struct {
    Login string `json:"login"`
    Email string `json:"email"`
    Name  string `json:"name"`
}
```

#### TokenValidator (internal/auth/token_validator.go)
```go
type TokenValidator interface {
    Validate(token string) (*ValidationResult, error)
}

type tokenValidator struct {
    githubClient GitHubClient
}

type ValidationResult struct {
    IsValid    bool
    HasGistScope bool
    User       *User
    Error      string
}
```

## 数据模型

### 配置文件结构
```json
{
  "github_token": "encrypted_or_obfuscated_token_string"
}
```

### 文件位置
- **Linux/macOS**: `~/.pv/config.json`
- **Windows**: `%USERPROFILE%\.pv\config.json`

## API 设计

### GitHub API 集成

#### 认证用户端点
```
GET https://api.github.com/user
Headers:
  Authorization: token PERSONAL_ACCESS_TOKEN
```

#### 检查权限端点
```
HEAD https://api.github.com/
Headers:
  Authorization: token PERSONAL_ACCESS_TOKEN
Response Headers:
  X-OAuth-Scopes: gist, repo
```

### 内部 API 流程

#### Login 流程
```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant AuthService
    participant TokenValidator
    participant GitHubClient
    participant ConfigStore
    
    User->>CLI: pv auth login
    CLI->>User: Enter GitHub PAT:
    User->>CLI: [token]
    CLI->>AuthService: Login(token)
    AuthService->>TokenValidator: Validate(token)
    TokenValidator->>GitHubClient: ValidateScopes(token)
    GitHubClient->>GitHub API: HEAD /
    GitHub API-->>GitHubClient: X-OAuth-Scopes
    GitHubClient-->>TokenValidator: scopes
    TokenValidator->>GitHubClient: GetAuthenticatedUser(token)
    GitHubClient->>GitHub API: GET /user
    GitHub API-->>GitHubClient: user info
    GitHubClient-->>TokenValidator: user
    TokenValidator-->>AuthService: ValidationResult
    AuthService->>ConfigStore: SaveToken(token)
    ConfigStore->>File System: Write config.json (600)
    AuthService-->>CLI: success
    CLI-->>User: Logged in as [username]
```

## 用户界面

### 命令行交互设计

#### Login 交互
```bash
$ pv auth login
Enter your GitHub Personal Access Token: 
✓ Token validated successfully
✓ Logged in as username
```

#### Status 交互
```bash
$ pv auth status
✓ Authenticated as username (user@example.com)
```

#### Logout 交互
```bash
$ pv auth logout
✓ Successfully logged out
```

### 错误消息设计
- 无效令牌: "Error: Invalid GitHub token. Please check your token and try again."
- 权限不足: "Error: Token lacks 'gist' scope. Please create a new token with gist permissions."
- 网络错误: "Error: Unable to connect to GitHub API. Please check your internet connection."

## 集成点

### 与现有系统的集成

1. **Wire 依赖注入**
   - 在 `internal/di/wire.go` 中添加 AuthService 及其依赖
   - 更新 `BuildCLI` 函数以包含认证命令

2. **命令注册**
   - 在 `NewRootCommand` 中添加 `auth` 命令

3. **未来集成**
   - GistStore 将使用 ConfigStore 获取认证令牌
   - 所有 GitHub API 调用都将通过 GitHubClient

## 安全设计

### 令牌存储安全

1. **文件权限**
   - Unix 系统: 0600 (仅用户可读写)
   - Windows: 使用 ACL 限制为当前用户

2. **令牌混淆**
   - 使用简单的 XOR 混淆避免明文存储
   - 未来可升级为使用系统密钥链

3. **内存安全**
   - 令牌在内存中的时间最小化
   - 使用后立即清零敏感数据

### API 通信安全

1. **HTTPS 强制**
   - 所有 GitHub API 调用使用 HTTPS
   - 验证 SSL 证书

2. **令牌传输**
   - 令牌仅在 Authorization header 中传输
   - 永不在 URL 参数中传递

## 错误处理

### 错误类型定义

```go
// internal/errors/auth_errors.go
const (
    ErrInvalidToken = AppError{
        Type: ErrAuth,
        Message: "Invalid GitHub token",
    }
    
    ErrMissingScope = AppError{
        Type: ErrAuth,
        Message: "Token lacks required 'gist' scope",
    }
    
    ErrTokenNotFound = AppError{
        Type: ErrAuth,
        Message: "No authentication token found. Please run 'pv auth login' first",
    }
)
```

### 错误恢复策略

1. **网络错误**: 提示用户检查网络连接，可重试
2. **认证错误**: 引导用户重新登录
3. **权限错误**: 提示创建具有正确权限的新令牌
4. **文件系统错误**: 提供手动修复说明