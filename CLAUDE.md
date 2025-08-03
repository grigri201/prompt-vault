# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Prompt Vault (pv) is a Go CLI application for managing prompts. It uses the Cobra framework for command-line interface and Google Wire for dependency injection.

## Architecture

The codebase follows a clean architecture pattern with clear separation of concerns:

- `cmd/` - CLI commands using Cobra framework
  - `root.go` - Main command entry point
  - `list.go` - List prompts command
- `internal/` - Internal packages
  - `di/` - Dependency injection configuration using Google Wire
  - `infra/` - Infrastructure layer (data storage)
  - `model/` - Domain models
- `main.go` - Application entry point

Key architectural decisions:
- Dependency injection via Google Wire (see `internal/di/wire.go`)
- Interface-based design for storage (`infra.Store`)
- Currently uses in-memory storage (`MemoryStore`)

## Development Commands

### Build
```bash
go build -o pv
```

### Run
```bash
go run main.go
# or after building:
./pv
```

### Generate Wire Dependencies
```bash
go generate ./internal/di
```

### Run Tests
```bash
go test ./...
```

### Format Code
```bash
go fmt ./...
```

### Lint
```bash
go vet ./...
```

## Key Components

- **Store Interface** (`internal/infra/store.go`): Defines the data access contract
- **Prompt Model** (`internal/model/prompt.go`): Core domain model with ID, Name, Author, and GistURL fields
- **Wire Configuration** (`internal/di/wire.go`): Dependency injection setup

## Current Features

- `pv` - Shows greeting message
- `pv list` - Lists all prompts from the store (currently hardcoded in MemoryStore)

## Communication Guidelines

- 用中文交流