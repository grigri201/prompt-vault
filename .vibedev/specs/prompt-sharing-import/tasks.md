# Implementation Tasks: Prompt Sharing and Import Feature

This document contains the test-driven development tasks for implementing the `pv share` and `pv import` commands.

## 1. Data Model Enhancements

- [x] **1.1 Write tests for PromptMeta with Parent field**
  - Create test file: `internal/models/prompt_parent_test.go`
  - Test YAML marshaling/unmarshaling with parent field
  - Test backward compatibility when parent field is empty
  - References: Requirement 3.1, 3.4

- [x] **1.2 Implement Parent field in PromptMeta**
  - Add Parent field to PromptMeta struct in `internal/models/prompt.go`
  - Add YAML tag `parent,omitempty`
  - Ensure field is preserved during read/write operations
  - References: Requirement 3.1

- [x] **1.3 Write tests for Index with ImportedEntries field**
  - Extend `internal/models/prompt_test.go` or create new test file
  - Test JSON marshaling/unmarshaling with imported_entries
  - Test adding/updating entries in ImportedEntries
  - References: Requirement 3.2

- [x] **1.4 Implement ImportedEntries field in Index**
  - Add ImportedEntries field to Index struct in `internal/models/prompt.go`
  - Add JSON tag `imported_entries`
  - Implement helper methods for managing imported entries
  - References: Requirement 3.2, 3.3

## 2. GitHub Client Extensions

- [x] **2.1 Write tests for public gist creation**
  - Add tests to `internal/gist/client_test.go`
  - Mock GitHub API responses for public gist creation
  - Test error scenarios (network failures, auth errors)
  - References: Requirement 1.1, 4.1, 4.2

- [x] **2.2 Implement CreatePublicGist method**
  - Add method to `internal/gist/client.go`
  - Set Public field to true in gist creation
  - Return gist ID and URL
  - References: Requirement 1.1

- [x] **2.3 Write tests for listing user gists**
  - Add tests for ListUserGists method
  - Mock paginated responses
  - Test filtering by description or filename
  - References: Requirement 1.5

- [x] **2.4 Implement ListUserGists method**
  - Add method to fetch all user gists
  - Support pagination for users with many gists
  - Enable filtering capabilities
  - References: Requirement 1.5

- [x] **2.5 Write tests for gist URL parsing**
  - Test various GitHub gist URL formats
  - Test invalid URL handling
  - Test extraction of gist ID from URL
  - References: Requirement 2.1, 5.4

- [x] **2.6 Implement GetGistByURL method**
  - Parse gist URL to extract ID
  - Validate URL format
  - Fetch gist using extracted ID
  - References: Requirement 2.1, 5.4

## 3. Share Manager Implementation

- [x] **3.1 Create ShareManager tests**
  - Create `internal/share/manager_test.go`
  - Test sharing new public gist from private
  - Test updating existing public gist
  - Test error cases (public gist, missing gist)
  - References: Requirement 1.1, 1.2, 1.9

- [x] **3.2 Implement ShareManager structure**
  - Create `internal/share/manager.go`
  - Define Manager struct with gistClient and ui dependencies
  - Implement constructor
  - References: Design ShareManager interface

- [ ] **3.3 Write tests for finding existing public gist**
  - Test searching gists by parent field
  - Test when no public version exists
  - Test multiple gists with same parent
  - References: Requirement 1.5

- [ ] **3.4 Implement findExistingPublicGist method**
  - Search user's gists for matching parent field
  - Parse YAML metadata to check parent field
  - Return gist ID if found
  - References: Requirement 1.5, 1.4

- [ ] **3.5 Write tests for public gist creation with parent**
  - Test parent field is included in metadata
  - Test gist is created as public
  - Test success response handling
  - References: Requirement 1.3, 1.4, 1.8

- [ ] **3.6 Implement createPublicGist method**
  - Read private gist content
  - Add parent field to metadata
  - Create new public gist
  - Return URL and success status
  - References: Requirement 1.1, 1.3, 1.4, 1.8

- [ ] **3.7 Write tests for public gist update**
  - Test user confirmation prompt
  - Test content synchronization
  - Test version update
  - References: Requirement 1.6, 1.7

- [ ] **3.8 Implement updatePublicGist method**
  - Prompt user for update confirmation
  - Sync content from private gist
  - Update version in metadata
  - References: Requirement 1.6, 1.7

- [ ] **3.9 Write integration test for SharePrompt**
  - Test complete share workflow
  - Test both create and update paths
  - Test error handling
  - References: Requirement 1.1-1.10

- [ ] **3.10 Implement SharePrompt method**
  - Validate private gist
  - Check for existing public version
  - Create or update as appropriate
  - Return result with URL
  - References: Requirement 1.1-1.10

## 4. Import Manager Implementation

- [ ] **4.1 Create ImportManager tests**
  - Create `internal/imports/manager_test.go`
  - Test importing new gist
  - Test updating existing import
  - Test validation failures
  - References: Requirement 2.1, 2.2, 2.8

- [ ] **4.2 Implement ImportManager structure**
  - Create `internal/imports/manager.go`
  - Define Manager struct with dependencies
  - Implement constructor
  - References: Design ImportManager interface

- [ ] **4.3 Write tests for gist ID extraction**
  - Test various URL formats
  - Test invalid URLs
  - Test edge cases
  - References: Requirement 2.1, 4.3

- [ ] **4.4 Implement extractGistID method**
  - Parse GitHub gist URLs
  - Extract gist ID component
  - Validate format
  - References: Requirement 2.1

- [ ] **4.5 Write tests for prompt validation**
  - Test valid prompt structure
  - Test missing required fields
  - Test field validation
  - References: Requirement 2.2, 2.3, 4.4

- [ ] **4.6 Implement validatePromptGist method**
  - Parse gist content for YAML metadata
  - Validate required fields
  - Create Prompt model from content
  - References: Requirement 2.2, 2.3

- [ ] **4.7 Write tests for existing import check**
  - Test finding existing entry by gist ID
  - Test version comparison
  - Test when no existing entry
  - References: Requirement 2.6

- [ ] **4.8 Implement checkExistingImport method**
  - Search ImportedEntries by gist ID
  - Compare versions if found
  - Return entry and found status
  - References: Requirement 2.6

- [ ] **4.9 Write tests for version update confirmation**
  - Test user prompt for version conflict
  - Test confirmation acceptance/rejection
  - Test version display format
  - References: Requirement 2.6, 2.7

- [ ] **4.10 Implement confirmVersionUpdate method**
  - Display old and new versions
  - Prompt user for confirmation
  - Return user decision
  - References: Requirement 2.6, 2.7

- [ ] **4.11 Write integration test for ImportPrompt**
  - Test complete import workflow
  - Test index update
  - Test error scenarios
  - References: Requirement 2.1-2.10

- [ ] **4.12 Implement ImportPrompt method**
  - Extract and validate gist
  - Check for existing import
  - Handle version conflicts
  - Update index and persist
  - References: Requirement 2.1-2.10

## 5. CLI Command Implementation

- [ ] **5.1 Write tests for share command**
  - Create `internal/cli/share_test.go`
  - Test command parsing and validation
  - Test manager integration
  - Test output formatting
  - References: Requirement 1.8, 5.1, 5.2

- [ ] **5.2 Implement share command**
  - Create `internal/cli/share.go`
  - Define command structure
  - Parse gist ID argument
  - Call ShareManager and display result
  - References: Requirement 1.1, 1.8, 5.1, 5.2

- [ ] **5.3 Write tests for import command**
  - Create `internal/cli/import_test.go`
  - Test URL parsing
  - Test manager integration
  - Test success messages
  - References: Requirement 2.1, 5.1, 5.2

- [ ] **5.4 Implement import command**
  - Create `internal/cli/import.go`
  - Define command structure
  - Parse gist URL argument
  - Call ImportManager and display result
  - References: Requirement 2.1, 5.1, 5.2

- [ ] **5.5 Register commands with root**
  - Add share and import commands to root command
  - Update `internal/cli/root.go`
  - Ensure commands are available
  - References: Requirement 5.2

## 6. Integration and Error Handling

- [ ] **6.1 Write integration tests for share workflow**
  - Test end-to-end share scenarios
  - Test with mock GitHub API
  - Test error propagation
  - References: Requirement 6.4, 6.7

- [ ] **6.2 Write integration tests for import workflow**
  - Test end-to-end import scenarios
  - Test index persistence
  - Test error handling
  - References: Requirement 6.4, 6.7

- [ ] **6.3 Implement comprehensive error messages**
  - Add error constants to `internal/errors/messages.go`
  - Ensure all error cases have clear messages
  - Test error message formatting
  - References: Requirement 4.1-4.5, 5.1

- [ ] **6.4 Write tests for progress indicators**
  - Test progress display during long operations
  - Test cancellation handling
  - Ensure UI responsiveness
  - References: Requirement 5.3

- [ ] **6.5 Implement progress indicators**
  - Add progress feedback for network operations
  - Use existing UI patterns from codebase
  - Handle operation cancellation
  - References: Requirement 5.3

---

任务列表看起来怎么样？