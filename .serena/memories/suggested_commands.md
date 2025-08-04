# Suggested Commands for Development

## Build Commands
```bash
# Build the application
go build -o pv

# Run without building
go run main.go
```

## Development Commands
```bash
# Generate Wire dependencies (required after changing DI configuration)
go generate ./internal/di

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint code
go vet ./...
```

## Application Usage
```bash
# Show help/greeting
./pv

# List all prompts
./pv list

# Authentication commands
./pv auth login
./pv auth logout
./pv auth status
```

## System Commands (Linux)
```bash
# Basic file operations
ls          # List files
cd <dir>    # Change directory
grep        # Search text
find        # Find files
git         # Git operations
```

## Important Notes
- Always run `go generate ./internal/di` after modifying dependency injection configuration
- Use `go fmt ./...` before committing code changes
- Run `go test ./...` to ensure all tests pass