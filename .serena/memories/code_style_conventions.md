# Code Style and Conventions

## Go Style Guidelines
- Follow standard Go formatting (`go fmt`)
- Use Go naming conventions:
  - PascalCase for exported types/functions
  - camelCase for unexported types/functions
  - ALL_CAPS for constants

## Project Conventions

### Package Structure
- `cmd/` - CLI command implementations using Cobra
- `internal/` - Private packages not intended for external use
- Clear separation of concerns with dedicated packages for each layer

### Naming Patterns
- Command files: `<command_name>.go` (e.g., `list.go`, `auth_login.go`)
- Interface implementations: `<interface_name>_impl.go` 
- Mock implementations: `<interface_name>_mock.go`
- Test files: `<file_name>_test.go`

### Architecture Patterns
- Interface-based design for abstracting external dependencies
- Dependency injection using Google Wire
- Clean architecture with domain models separate from infrastructure
- Error handling using custom error types in `internal/errors/`

### Code Organization
- Group related functionality in packages
- Use struct receivers for methods
- Implement interfaces in separate files when appropriate
- Keep domain models pure (no external dependencies)

## Testing Conventions
- Unit tests in `*_test.go` files
- Mock implementations for testing external dependencies
- Test coverage for service layer implementations