# Prompt Vault

A command-line tool for managing and using prompt templates with GitHub Gist as storage backend.

## Features

- Store prompt templates in private GitHub Gists
- Interactive variable filling
- Search and filter prompts by name, category, tags, author, and description
- Local caching for offline access
- Cross-platform clipboard support

## Installation

```bash
go install github.com/[your-username]/prompt-vault/cmd/pv@latest
```

## Usage

### First Time Setup

```bash
# Authenticate with GitHub
pv login
```

### Managing Prompts

```bash
# Upload a prompt template
pv upload template.yaml

# List all prompts
pv list

# Search prompts
pv get [keyword]

# Delete a prompt
pv delete <prompt-name>

# Sync local cache
pv sync
```

## Prompt File Format

```yaml
---
name: "API Documentation"
author: "john"
category: "docs"
tags: ["api", "swagger"]
version: "1.0"
description: "Generate API documentation"
---
Generate {format} documentation for {endpoint}
```

## Development

This project follows Test-Driven Development (TDD) methodology. Run tests with:

```bash
go test ./...
```

## License

MIT