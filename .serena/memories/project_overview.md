# Project Overview

## Purpose
Prompt Vault (pv) is a Go CLI application for managing prompts. It allows users to store, list, and manage prompts with integration to GitHub Gists for remote storage.

## Tech Stack
- **Language**: Go 1.24.5
- **CLI Framework**: Cobra (github.com/spf13/cobra v1.9.1)
- **Dependency Injection**: Google Wire (github.com/google/wire v0.6.0)
- **GitHub Integration**: go-github/v74 for GitHub API interactions
- **Authentication**: OAuth2 (golang.org/x/oauth2 v0.30.0)

## Architecture Pattern
Clean Architecture with clear separation of concerns:
- Domain models in `internal/model/`
- Infrastructure layer in `internal/infra/`
- Service layer in `internal/service/`
- Authentication in `internal/auth/`
- Configuration in `internal/config/`
- Dependency injection setup in `internal/di/`
- CLI commands in `cmd/`

## Key Design Principles
- Interface-based design for storage (`infra.Store`)
- Dependency injection via Google Wire
- Clean separation between domain, infrastructure, and presentation layers