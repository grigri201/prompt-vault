# Prompt Vault Requirements Document

## Introduction

Prompt Vault is a command-line tool written in Go that enables users to manage, store, and reuse prompt templates through GitHub Gists. The tool allows users to create prompt templates with variables, store them privately in GitHub Gists, search and retrieve templates, fill in variables interactively, and copy the final prompt to the clipboard.

## Requirements

### 1. Authentication and Configuration

**User Story**: As a user, I want to authenticate with GitHub using a Personal Access Token, so that I can securely store and retrieve my private prompts.

**Acceptance Criteria**:
1. The system SHALL provide a `pv login` command that prompts the user to enter their GitHub Personal Access Token
2. The system SHALL display instructions on how to create a GitHub Personal Access Token (Settings > Developer settings > Personal access tokens)
3. The system SHALL provide a secure input prompt for users to paste their token (hidden input)
4. The system SHALL validate the provided token by making a test API call to GitHub
5. The system SHALL store the token securely in `~/.config/prompt-vault/config.yaml`
6. The system SHALL create the configuration directory if it does not exist
7. The system SHALL display an error message if the token is invalid
8. The system SHALL allow users to re-run `pv login` to update their token

### 2. Prompt Template Upload

**User Story**: As a prompt author, I want to upload my prompt templates to GitHub Gists, so that I can store and version them centrally.

**Acceptance Criteria**:
1. The system SHALL provide a `pv upload [file_name]` command to upload prompt files
2. The system SHALL parse the YAML front matter from the prompt file to extract metadata
3. The system SHALL validate that the prompt file contains required fields: name, author, category, tags
4. The system SHALL support optional description field in the YAML front matter
5. The system SHALL automatically set the version field to current timestamp (Unix milliseconds) if not provided
6. The system SHALL create a new private Gist named `{author}-{name}` for the prompt
7. The system SHALL set the Gist description to the prompt's description if provided
8. The system SHALL replace existing Gist content if a prompt with the same author and name already exists
9. The system SHALL update the user's index Gist (`{username}-promptvault-index`) with the new prompt metadata
10. The system SHALL store each prompt entry in the index as a JSON object containing: gist_id, gist_url, name, author, category, tags, version, description, updated_at
11. The system SHALL format the index file as a JSON array of prompt entries
12. The system SHALL display success message with the Gist URL after successful upload

### 3. Prompt Discovery and Listing

**User Story**: As a user, I want to list and search through available prompts, so that I can find the right template for my needs.

**Acceptance Criteria**:
1. The system SHALL provide a `pv list` command to display all available prompts
2. The system SHALL display prompts in a paginated list with 20 items per page
3. The system SHALL support left/right arrow keys for pagination navigation
4. The system SHALL display current page number and total pages (e.g., "Page 1/5")
5. The system SHALL display prompts with numeric indices and show: name, author, category, version
6. The system SHALL provide a `pv get [keyword]` command to search prompts
7. The system SHALL search through prompt names, categories, tags, authors, and descriptions
8. The system SHALL display search results in a paginated list with numeric indices (1, 2, 3, etc.)
9. The system SHALL cache prompt metadata locally in `~/.cache/prompt-vault/prompts/`
10. The system SHALL use cached data when available to improve performance

### 4. Prompt Template Usage

**User Story**: As a user, I want to use a prompt template by filling in variables and copying the result to my clipboard, so that I can quickly generate customized prompts.

**Acceptance Criteria**:
1. The system SHALL allow users to select a prompt by entering its numeric index from search results
2. The system SHALL download and cache the selected prompt template if not already cached
3. The system SHALL parse the template to identify all variables in the format `{variable_name}`
4. The system SHALL display an interactive form for variable input with all variables listed
5. The system SHALL show each variable name with an input field
6. The system SHALL support up/down arrow keys to navigate between input fields
7. The system SHALL highlight the currently active input field
8. The system SHALL allow users to submit the form with Enter key when all variables are filled
9. The system SHALL replace all variable placeholders with user-provided values
10. The system SHALL copy the final prompt text to the system clipboard
11. The system SHALL display a confirmation message after copying to clipboard

### 5. Prompt Deletion

**User Story**: As a prompt author, I want to delete prompts I've uploaded, so that I can remove outdated or incorrect templates.

**Acceptance Criteria**:
1. The system SHALL provide a `pv delete <name>` command to delete prompts
2. The system SHALL require the complete prompt name as input
3. The system SHALL only allow users to delete prompts where they are the author
4. The system SHALL display a warning message stating "This operation cannot be undone. Are you sure? (y/N)"
5. The system SHALL require explicit confirmation (y/yes) to proceed with deletion
6. The system SHALL abort deletion if user enters anything other than y/yes
7. The system SHALL delete the corresponding Gist from GitHub upon confirmation
8. The system SHALL update the user's index Gist to remove the deleted prompt
9. The system SHALL remove the prompt from local cache if present
10. The system SHALL display an error if the prompt is not found or user lacks permission

### 6. Cache Synchronization

**User Story**: As a user, I want to synchronize my local cache with GitHub, so that I have access to the latest prompts.

**Acceptance Criteria**:
1. The system SHALL provide a `pv sync` command to synchronize local cache
2. The system SHALL fetch the user's index Gist to get the list of available prompts
3. The system SHALL display progress indicator showing "Syncing... (X/Y)"
4. The system SHALL show current item being synced with its name
5. The system SHALL update local cache with new or modified prompts
6. The system SHALL remove cached prompts that no longer exist in the index
7. The system SHALL display summary after sync completion (e.g., "Sync completed: X added, Y updated, Z deleted")
8. The system SHALL handle network errors gracefully with retry logic
9. The system SHALL update the last sync timestamp in the cache

### 7. Error Handling and User Experience

**User Story**: As a user, I want clear error messages and helpful feedback, so that I can understand and resolve issues quickly.

**Acceptance Criteria**:
1. The system SHALL provide clear error messages for common issues (network errors, authentication failures, file not found)
2. The system SHALL validate user inputs before making API calls
3. The system SHALL provide helpful usage instructions when commands are used incorrectly
4. The system SHALL handle GitHub API rate limits gracefully
5. The system SHALL provide verbose mode (-v flag) for debugging
6. The system SHALL ensure all user-facing messages are in English

### 8. Prompt File Format

**User Story**: As a prompt author, I want a clear file format for my prompts, so that I can create valid templates easily.

**Acceptance Criteria**:
1. The system SHALL support YAML front matter followed by prompt content
2. The system SHALL require the following fields in front matter: name, author, category, tags
3. The system SHALL support optional version field (any string format)
4. The system SHALL support optional description field for prompt summary
5. The system SHALL support tags as a YAML array
6. The system SHALL treat everything after the front matter separator (---) as prompt content
7. The system SHALL preserve formatting and whitespace in prompt content
8. The system SHALL support multiple variables with the same name in templates

Example prompt format:
```yaml
---
name: "API 文档生成"
author: "zhangsan"
category: "文档"
tags: ["api", "documentation", "swagger"]
version: "1705678200000"  # 默认为时间戳数值，但支持任意字符串
description: "根据 API 接口信息自动生成各种格式的文档"  # 可选字段
---
请根据以下 API 接口信息生成 {format} 格式的文档：
API 端点：{endpoint}
请求方法：{method}
```

### 9. Index File Format

**User Story**: As a developer, I want a standardized index file format, so that the system can reliably search and download prompts.

**Acceptance Criteria**:
1. The system SHALL store the index in a Gist named `{username}-promptvault-index`
2. The system SHALL format the index as a JSON file containing an array of prompt entries
3. The system SHALL include the following fields for each prompt entry:
   - gist_id: The GitHub Gist ID for direct API access
   - gist_url: The full URL to the Gist
   - name: The prompt name
   - author: The prompt author
   - category: The prompt category
   - tags: Array of tag strings
   - version: The version string
   - description: Optional prompt description (empty string if not provided)
   - updated_at: ISO 8601 timestamp of last update
4. The system SHALL maintain the index in valid JSON format at all times
5. The system SHALL update the index atomically to prevent corruption
6. The system SHALL use the gist_id from the index to fetch prompt content

Example index format:
```json
[
  {
    "gist_id": "abc123def456",
    "gist_url": "https://gist.github.com/username/abc123def456",
    "name": "API 文档生成",
    "author": "zhangsan",
    "category": "文档",
    "tags": ["api", "documentation", "swagger"],
    "version": "1705678200000",
    "description": "根据 API 接口信息自动生成各种格式的文档",
    "updated_at": "2024-01-19T10:30:00Z"
  }
]
```

---

Do the requirements look good? If so, we can move on to the design.