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
# Add a prompt template (local file or Gist URL)
pv add template.yaml
pv add https://gist.github.com/user/gist-id

# Search and use prompts
pv get [keyword]

# Delete a prompt
pv del <prompt-name>

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