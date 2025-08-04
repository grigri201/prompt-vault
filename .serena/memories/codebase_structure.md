# Codebase Structure

## Directory Layout
```
.
├── cmd/                    # CLI commands using Cobra
│   ├── root.go            # Main command entry point
│   ├── list.go            # List prompts command
│   ├── auth.go            # Auth commands
│   ├── auth_login.go      # Login command
│   ├── auth_logout.go     # Logout command
│   └── auth_status.go     # Auth status command
├── internal/              # Internal packages
│   ├── auth/              # Authentication logic
│   │   ├── github_client.go
│   │   ├── github_client_impl.go
│   │   ├── github_client_mock.go
│   │   └── token_validator.go
│   ├── config/            # Configuration management
│   │   ├── config.go
│   │   ├── store.go
│   │   ├── store_test.go
│   │   └── obfuscate.go
│   ├── di/                # Dependency injection
│   │   ├── wire.go        # Wire configuration
│   │   ├── wire_gen.go    # Generated wire code
│   │   └── providers.go   # Provider functions
│   ├── errors/            # Error definitions
│   │   ├── errors.go
│   │   └── auth_errors.go
│   ├── infra/             # Infrastructure layer
│   │   ├── store.go       # Store interface
│   │   └── github_store.go # GitHub-based storage
│   ├── model/             # Domain models
│   │   ├── prompt.go      # Prompt struct
│   │   └── index.go       # Index model
│   └── service/           # Service layer
│       ├── auth_service.go
│       ├── auth_service_impl.go
│       └── auth_service_test.go
├── main.go                # Application entry point
├── go.mod                 # Go module definition
├── go.sum                 # Go dependencies checksum
└── CLAUDE.md              # Project documentation
```

## Key Components
- **Store Interface** (`internal/infra/store.go`): Defines data access contract with methods: List(), Add(), Delete(), Update(), Get()
- **Prompt Model** (`internal/model/prompt.go`): Core domain model with ID, Name, Author, and GistURL fields
- **Wire Configuration** (`internal/di/wire.go`): Dependency injection setup