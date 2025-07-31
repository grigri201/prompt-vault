# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Generate Wire dependency injection code (required before building)
go generate ./...
# or directly
wire ./cmd/pv/

# Build the application
go build -o pv cmd/pv/main.go

# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests for a specific package
go test ./internal/cli

# Run a specific test
go test -run TestSyncCommand ./internal/cli -v

# Run tests with verbose output
go test -v ./...

# Format code
go fmt ./...

# Check for linting issues
go vet ./...

# Run quality checks (from scripts)
./scripts/quality_checks.sh

# Run functional tests
./scripts/functional_tests.sh

# Cross-platform build
./scripts/cross_platform_build.sh
```

## Architecture Overview

This is a CLI application for managing prompt templates using GitHub Gists as storage backend. The architecture follows clean architecture principles with clear separation of concerns and uses Google Wire for dependency injection.

### Core Components

1. **CLI Layer** (`internal/cli/`)
   - Entry point for all commands using Cobra framework
   - Commands: `add` (merged upload/import), `get`, `del`, `sync`, `share`, `login`, `config`
   - Each command implements sync middleware for automatic synchronization
   - Interactive UI components use Bubble Tea v2 framework

2. **Domain Models** (`internal/models/`)
   - `Prompt`: Core model with YAML frontmatter and content
   - `Index`: Tracks all prompts with `entries` and `imported_entries`
   - `PromptMeta`: Metadata (name, author, tags, version, description, parent, id)
   - Note: `category` field has been removed from the codebase

3. **Storage Layers**
   - **GitHub Gist API** (`internal/gist/`): All GitHub Gist operations with retry logic
   - **Local Cache** (`internal/cache/`): File cache in `~/.cache/prompt-vault/prompts/`
   - **Config** (`internal/config/`): App configuration in `~/.config/prompt-vault/`
   - **Auth** (`internal/auth/`): GitHub token authentication management

4. **Core Services**
   - **Sync Manager** (`internal/sync/`): Unified sync logic used by all commands
   - **Import Manager** (`internal/imports/`): Handles Gist URL imports
   - **Share Manager** (`internal/share/`): Creates public Gists for sharing
   - **Search** (`internal/search/`): Fuzzy search across prompts
   - **Parser** (`internal/parser/`): YAML parsing with strict/lenient modes

5. **Infrastructure**
   - **Container** (`internal/container/`): Dependency injection container
   - **Wire** (`cmd/pv/wire.go`, `internal/wire/`): Wire configuration and providers
   - **Interfaces** (`internal/interfaces/`): Interface definitions for all managers
   - **Errors** (`internal/errors/`): Standardized AppError with automatic categorization
   - **Paths** (`internal/paths/`): Centralized path management with atomic writes

6. **UI Components** (`internal/ui/`)
   - Form: Interactive input forms
   - Selector: List selection with search
   - Paginator: Paginated list display
   - Progress: Progress indicators

## Key Workflows

### Authentication Flow
1. User runs `pv login`
2. Prompted for GitHub personal access token
3. Token validated against GitHub API
4. Token stored securely in config file
5. Post-login sync executed automatically

### Unified Sync Strategy
The app uses timestamp-based bidirectional sync:
- Compares local and remote `index.json` `updated_at` timestamps
- Always uses the newer version (no manual conflict resolution)
- Sync is called:
  - **Pre-command**: Before `add`, `get`, `share`, `del` (ensures latest data)
  - **Post-command**: After `login`, `add`, `del` (pushes changes)
  - **Manual**: Via `pv sync` command

### Add Command Flow (Unified Upload/Import)
1. Pre-sync to ensure latest data
2. Auto-detects input type (file vs Gist URL)
3. For files: Creates Gist, updates local cache
4. For URLs: Adds to `imported_entries` in index
5. Post-sync to push changes

### Prompt File Format
```yaml
---
name: "Prompt Name"
author: "username"
tags: ["tag1", "tag2"]
version: "1.0"
description: "Description"
parent: "parent-gist-id"  # Optional, for shared prompts
id: "custom-id"          # Optional, custom identifier
---
Prompt content with {variables} to fill
```

## Testing Approach

The project follows TDD methodology:
- Unit tests for all components
- Integration tests in `internal/integration/`
- Validation tests in `internal/validation/`
- Mock interfaces for external dependencies
- Test helpers in `internal/testhelpers/`

## Important Implementation Details

1. **Wire Dependency Injection**:
   - `wire_gen.go` is auto-generated - never edit manually
   - Run `go generate ./...` after changing dependencies
   - Wire configuration in `cmd/pv/wire.go`

2. **Sync Middleware**: All data-modifying commands use sync middleware for consistency

3. **Error Handling**:
   - All errors use `AppError` type
   - Automatic categorization: Auth, Network, FileSystem, Parsing, Validation
   - Use `NewError()` and `WrapError()` - never `fmt.Errorf()`

4. **Index Structure**:
   - `entries`: User's own prompts
   - `imported_entries`: Prompts imported from other users
   - Both use identical `IndexEntry` structure

5. **Variable Handling**: Uses `{variable}` syntax with interactive replacement

6. **Atomic Operations**: All file writes use atomic operations via PathManager

## Environment Variables

- `HOME`: Used to determine default paths for cache and config
- GitHub token is stored in config, not environment variables

## Directory Structure

```
~/.cache/prompt-vault/prompts/    # Local cache
    ├── index.json               # Prompt index
    └── <gist-id>.yaml          # Cached prompts

~/.config/prompt-vault/          # Configuration
    └── config.yaml             # App config and auth token
```