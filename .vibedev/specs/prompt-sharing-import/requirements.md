# Requirements: Prompt Sharing and Import Feature

## Introduction

This feature introduces two new commands to the prompts-vault CLI tool:

1. **`pv share`** - Creates and maintains public copies of private prompt gists, enabling users to share their prompts while keeping the original versions private
2. **`pv import`** - Imports external public prompt gists into the user's prompt collection, allowing users to build their library from shared community prompts

These commands enhance the collaborative capabilities of prompts-vault by facilitating prompt sharing and collection management.

## Requirements

### 1. Share Command (`pv share`)

**User Story**: As a prompt author, I want to create public versions of my private prompts, so that I can share them with others while maintaining my private working copies.

#### Acceptance Criteria:

1.1. **The system SHALL** create a new public gist as a copy when the user executes `pv share <gist-id>` on a private gist.

1.2. **The system SHALL** preserve the original private gist completely unchanged during the sharing process.

1.3. **The system SHALL** add a `parent` field to the PromptMeta structure containing the original private gist ID.

1.4. **The system SHALL** include the `parent` field in the public gist's YAML metadata to track its origin.

1.5. **The system SHALL** check for an existing public version (by parent field) when sharing a gist for the second time.

1.6. **IF** a public version already exists, **THEN the system SHALL** prompt the user to confirm updating the existing public gist.

1.7. **WHEN** updating an existing public gist, **THEN the system SHALL** synchronize the content and version from the private gist.

1.8. **The system SHALL** display the public gist URL after successful creation or update.

1.9. **IF** the user attempts to share an already public gist, **THEN the system SHALL** display an appropriate error message.

1.10. **The system SHALL** handle authentication errors and network failures gracefully with clear error messages.

### 2. Import Command (`pv import`)

**User Story**: As a prompt collector, I want to import public prompt gists into my collection, so that I can build a curated library of useful prompts from various sources.

#### Acceptance Criteria:

2.1. **The system SHALL** accept a gist URL as input: `pv import <gist-url>`.

2.2. **The system SHALL** fetch and validate the gist content to ensure it's a valid prompt gist (contains required PromptMeta fields).

2.3. **The system SHALL** verify that the gist contains all required fields: name, author, category, and at least one tag.

2.4. **The system SHALL** preserve all original metadata including the original author information.

2.5. **The system SHALL** add a new `imported_entries` field to the Index structure to store imported prompts separately.

2.6. **IF** the gist already exists in imported_entries, **THEN the system SHALL** compare versions and prompt the user to confirm updates when versions differ.

2.7. **WHEN** the user confirms an update, **THEN the system SHALL** replace the existing entry with the newer version.

2.8. **The system SHALL** reject private gists with an appropriate error message explaining they cannot be accessed.

2.9. **The system SHALL** support importing gists from any GitHub user, not just the authenticated user.

2.10. **The system SHALL** update the index gist after successful import to persist the changes.

### 3. Data Model Updates

**User Story**: As a developer, I want the data models to support the new sharing and import features, so that the functionality can be implemented reliably.

#### Acceptance Criteria:

3.1. **The PromptMeta structure SHALL** include a new optional `Parent` field of type string with YAML tag `parent,omitempty`.

3.2. **The Index structure SHALL** include a new `ImportedEntries` field of type `[]IndexEntry` with JSON tag `imported_entries`.

3.3. **The system SHALL** maintain backward compatibility with existing prompt and index files.

3.4. **The system SHALL** preserve the parent field when reading and writing prompt YAML files.

### 4. Error Handling

**User Story**: As a user, I want clear error messages and graceful handling of edge cases, so that I understand what went wrong and how to fix it.

#### Acceptance Criteria:

4.1. **IF** a network error occurs, **THEN the system SHALL** display a clear message about connectivity issues.

4.2. **IF** authentication fails, **THEN the system SHALL** guide the user to check their GitHub token.

4.3. **IF** a gist is not found, **THEN the system SHALL** display a message indicating the gist ID or URL is invalid.

4.4. **IF** a gist lacks required fields, **THEN the system SHALL** list the missing fields in the error message.

4.5. **The system SHALL** use the existing error handling patterns established in the codebase.

### 5. User Experience

**User Story**: As a user, I want intuitive command interfaces and helpful feedback, so that I can use the new features effectively.

#### Acceptance Criteria:

5.1. **The system SHALL** provide clear success messages showing the result of each operation.

5.2. **The system SHALL** use consistent command patterns matching existing pv commands.

5.3. **The system SHALL** provide progress indicators for long-running operations (network requests).

5.4. **The system SHALL** support both gist IDs and full GitHub URLs where applicable.

5.5. **WHEN** displaying URLs, **THEN the system SHALL** show them in a format that's easy to copy and share.

### 6. Test-Driven Development Requirements

**User Story**: As a developer, I want to follow TDD practices, so that the code is well-tested and maintainable.

#### Acceptance Criteria:

6.1. **The development SHALL** start by writing failing unit tests for each new component before implementation.

6.2. **Each test SHALL** be written to verify one specific acceptance criterion from the requirements.

6.3. **The test suite SHALL** include unit tests for:
   - PromptMeta with Parent field validation
   - Index with ImportedEntries field operations
   - Share command logic (creating, updating, error handling)
   - Import command logic (validation, version comparison, error handling)

6.4. **The test suite SHALL** include integration tests for:
   - Full share command workflow with GitHub API
   - Full import command workflow with GitHub API
   - Index file updates and persistence

6.5. **The tests SHALL** use mocks for external dependencies (GitHub API) to ensure reliable and fast execution.

6.6. **The tests SHALL** follow the existing testing patterns in the codebase (using the same test framework and conventions).

6.7. **Each feature SHALL** be implemented incrementally following the Red-Green-Refactor cycle:
   - Red: Write a failing test
   - Green: Write minimal code to make the test pass
   - Refactor: Improve the code while keeping tests green

6.8. **The test coverage SHALL** be maintained at the same level or higher than the existing codebase.

## Success Criteria

- Users can successfully share private prompts as public gists
- Users can import any valid public prompt gist into their collection
- The parent-child relationship between private and public gists is maintained
- Imported prompts are tracked separately in the index
- All operations handle errors gracefully and provide helpful feedback
- All code is developed using TDD methodology with comprehensive test coverage
- Tests are written before implementation following Red-Green-Refactor cycle

## Development Approach

This feature will be developed using Test-Driven Development (TDD) methodology:

1. **First Phase**: Write failing tests for data model changes (PromptMeta.Parent, Index.ImportedEntries)
2. **Second Phase**: Implement data model changes to make tests pass
3. **Third Phase**: Write failing tests for share command functionality
4. **Fourth Phase**: Implement share command to make tests pass
5. **Fifth Phase**: Write failing tests for import command functionality
6. **Sixth Phase**: Implement import command to make tests pass
7. **Final Phase**: Integration testing and refactoring

---

需求文档已更新，加入了 TDD 开发方法的详细要求。如果满意的话，我们可以继续进行设计阶段。