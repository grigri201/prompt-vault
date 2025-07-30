# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
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
```

## Architecture Overview

This is a CLI application for managing prompt templates using GitHub Gists as storage backend. The architecture follows clean architecture principles with clear separation of concerns.

### Core Components

1. **CLI Layer** (`internal/cli/`)
   - Entry point for all commands using Cobra framework
   - Each command has its own file (e.g., `sync.go`, `upload.go`, `delete.go`)
   - Commands use dependency injection pattern for testability
   - Interactive UI components use Bubble Tea framework

2. **Domain Models** (`internal/models/`)
   - `Prompt`: Core model representing a prompt template with YAML frontmatter
   - `Index`: Maintains list of all prompts with metadata
   - `PromptMeta`: Metadata structure for prompts (name, author, category, tags, etc.)

3. **Storage Layers**
   - **GitHub Gist API** (`internal/gist/`): Handles all GitHub Gist operations
   - **Local Cache** (`internal/cache/`): Manages local file cache in `~/.cache/prompt-vault/prompts/`
   - **Config** (`internal/config/`): Manages application configuration in `~/.config/prompt-vault/`
   - **Auth** (`internal/auth/`): Handles GitHub token authentication

4. **Unified Components**
   - **YAMLParser** (`internal/parser/yaml_parser.go`): Configurable YAML parsing with strict/lenient modes
   - **GistOperations** (`internal/gist/operations.go`): Wrapper for Gist operations with retry logic and 404 handling
   - **Standardized Error Handling** (`internal/errors/`): AppError types with automatic error categorization

5. **Managers Pattern & Interfaces**
   - Components implement Manager pattern with `Initialize()` and `Cleanup()` methods
   - Base manager in `internal/managers/` provides common functionality
   - **Interface Segregation** (`internal/interfaces/`):
     - `CacheManager`, `CacheReader`, `CacheWriter`: Cache operations
     - `AuthManager`, `AuthReader`, `AuthWriter`: Authentication operations
     - `Manager`: Base interface for lifecycle management
   - Managers communicate through interfaces, not direct dependencies

6. **Container Pattern** (`internal/container/`)
   - Dependency injection container that initializes all managers
   - Provides both production and test containers
   - Uses interfaces for managers to enable testing and loose coupling
   - Exception: ConfigManager uses concrete type to avoid circular dependencies

7. **Path Management** (`internal/paths/`)
   - Centralized path handling for all file operations
   - Supports atomic writes and secure file permissions

## Key Workflows

### Authentication Flow
1. User runs `pv login`
2. Prompted for GitHub personal access token
3. Token validated against GitHub API
4. Token stored securely in config file

### Sync Workflow
1. Connects to GitHub using stored token
2. Fetches all gists for authenticated user
3. Filters for prompt files (YAML frontmatter)
4. Downloads and caches prompts locally
5. Updates index.json with metadata

### Prompt File Format
```yaml
---
name: "Prompt Name"
author: "username"
category: "category"
tags: ["tag1", "tag2"]
version: "1.0"
description: "Description"
parent: "parent-gist-id"  # Optional, for shared prompts
---
Prompt content with {variables} to fill
```

## Testing Approach

The project follows TDD methodology:
- Unit tests for all components
- Integration tests for workflows
- Mock interfaces for external dependencies
- Test helpers in `internal/testhelpers/`

## Important Implementation Details

1. **Gist Storage**: Each prompt is stored as a single file in a GitHub Gist
2. **Index Management**: A special `index.json` file tracks all prompts
3. **Variable Handling**: Variables in prompts use `{variable}` syntax
4. **Error Handling**: 
   - Custom AppError types with automatic categorization
   - Error types: Auth, Network, FileSystem, Parsing, Validation
   - Standardized error creation with `NewError` and `WrapError`
   - No direct `fmt.Errorf` usage - all errors use AppError
5. **UI Components**: Custom Bubble Tea models for forms, selectors, and progress indicators
6. **Code Organization**:
   - No code duplication - unified components for common operations
   - Interface-based design for testability and flexibility
   - Clear separation of concerns between layers

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