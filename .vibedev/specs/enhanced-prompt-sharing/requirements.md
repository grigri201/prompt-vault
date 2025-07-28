# Enhanced Prompt Sharing Requirements

## Introduction

This feature enhances the prompts-vault (pv) command-line tool to improve user experience by adding gist URL visibility in list and get commands, and extending the share command functionality to support interactive prompt selection with and without keyword search.

## Requirements

### 1. Display Gist URL in List Command

**User Story**: As a user, I want to see the gist URL for each prompt when listing prompts, so that I can quickly access or share the gist link without additional commands.

**Acceptance Criteria**:
1. **WHEN** the user executes `pv list`, **THEN** the system SHALL display the gist URL for each prompt entry in the table output.
2. **IF** the gist URL is longer than the terminal width, **THEN** the system SHALL truncate the URL with ellipsis or display it on a separate line.
3. **WHERE** pagination is used, **THEN** the system SHALL maintain consistent formatting across all pages.
4. **WHILE** displaying the gist URL, **THEN** the system SHALL preserve the existing table structure and readability.

### 2. Display Gist URL in Get Command

**User Story**: As a user, I want to see the gist URL when viewing prompt details in the get command, so that I can reference or share the original gist.

**Acceptance Criteria**:
1. **WHEN** the user executes `pv get` and views prompt details, **THEN** the system SHALL display the gist URL along with other prompt information.
2. **AFTER** the user selects a prompt, **THEN** the system SHALL show the gist URL before the variable filling form.
3. **WHERE** multiple prompts are displayed in search results, **THEN** the system SHALL show the gist URL for each prompt in the list.
4. **IF** the prompt is successfully copied to clipboard, **THEN** the system SHALL display the gist URL in the success message.

### 3. Share Command Without Arguments

**User Story**: As a user, I want to execute `pv share` without arguments to see all my prompts and select one to share, so that I don't need to remember gist IDs.

**Acceptance Criteria**:
1. **WHEN** the user executes `pv share` without arguments, **THEN** the system SHALL display a list of all available prompts.
2. **WHERE** the prompt list is displayed, **THEN** the system SHALL show the same information as `pv list` (name, author, category, tags, description).
3. **AFTER** the user selects a prompt, **THEN** the system SHALL proceed with the existing share workflow.
4. **IF** no prompts are available, **THEN** the system SHALL display an appropriate message suggesting to run `pv sync`.
5. **WHEN** the user cancels the selection, **THEN** the system SHALL exit gracefully without sharing.

### 4. Share Command with Keyword Search

**User Story**: As a user, I want to execute `pv share <keyword>` to search and select a prompt to share, so that I can quickly find and share specific prompts.

**Acceptance Criteria**:
1. **WHEN** the user executes `pv share <keyword>`, **THEN** the system SHALL search for prompts matching the keyword.
2. **WHERE** the keyword matches prompt name, author, category, tags, or description, **THEN** the system SHALL include that prompt in the results.
3. **IF** multiple prompts match the keyword, **THEN** the system SHALL display a selection list similar to `pv get`.
4. **IF** only one prompt matches, **THEN** the system SHALL ask for confirmation before sharing.
5. **IF** no prompts match the keyword, **THEN** the system SHALL display a "no prompts found" message.
6. **WHEN** the keyword is a valid gist ID format, **THEN** the system SHALL prioritize exact gist ID match over keyword search.

### 5. Backward Compatibility

**User Story**: As an existing user, I want the current `pv share <gist-id>` functionality to remain unchanged, so that my existing workflows continue to work.

**Acceptance Criteria**:
1. **WHEN** the user executes `pv share <gist-id>` with a valid gist ID, **THEN** the system SHALL directly share the prompt without showing a selection interface.
2. **IF** the provided argument is a valid gist ID, **THEN** the system SHALL bypass keyword search and selection.
3. **WHERE** the gist ID is invalid or not found, **THEN** the system SHALL display the existing error messages.

### 6. User Experience Consistency

**User Story**: As a user, I want consistent interaction patterns across list, get, and share commands, so that the tool is intuitive to use.

**Acceptance Criteria**:
1. **WHERE** prompt selection is required, **THEN** the system SHALL use the same selector UI component as the get command.
2. **WHEN** displaying search results, **THEN** the system SHALL use the same format across get and share commands.
3. **IF** keyboard navigation is used, **THEN** the system SHALL support the same shortcuts (arrow keys, enter, escape) across all commands.
4. **WHERE** error messages are displayed, **THEN** the system SHALL maintain consistent formatting and tone.

### 7. Performance Requirements

**User Story**: As a user, I want the enhanced features to perform efficiently, so that my workflow is not slowed down.

**Acceptance Criteria**:
1. **WHEN** displaying gist URLs in list or get commands, **THEN** the system SHALL NOT introduce noticeable delay (< 100ms additional overhead).
2. **WHERE** search is performed in share command, **THEN** the system SHALL use the same efficient search algorithm as the get command.
3. **IF** the prompt list is large (> 100 entries), **THEN** the system SHALL maintain responsive performance for selection and navigation.

### 8. Test-Driven Development Requirements

**User Story**: As a developer, I want to implement all new features using TDD methodology, so that the code is reliable, maintainable, and well-tested.

**Acceptance Criteria**:
1. **BEFORE** implementing any new functionality, **THEN** the developer SHALL write failing unit tests that define the expected behavior.
2. **WHERE** new functions or methods are created, **THEN** the developer SHALL have corresponding test cases with at least 80% code coverage.
3. **WHEN** modifying existing functions, **THEN** the developer SHALL update or add tests to cover the new behavior.
4. **IF** integration between components is required, **THEN** the developer SHALL write integration tests following the existing test patterns.
5. **WHILE** implementing features, **THEN** the developer SHALL follow the Red-Green-Refactor cycle for each functionality.
6. **WHERE** user interaction is involved (UI components), **THEN** the developer SHALL write tests that simulate user input and verify output.
7. **AFTER** implementation is complete, **THEN** all tests SHALL pass and no existing tests SHALL be broken.

### 9. Testing Strategy

**User Story**: As a quality assurance engineer, I want comprehensive test coverage for all new features, so that we can ensure reliability and prevent regressions.

**Acceptance Criteria**:
1. **FOR** list command enhancements, **THEN** tests SHALL verify gist URL display in various scenarios (empty list, single page, multiple pages).
2. **FOR** get command enhancements, **THEN** tests SHALL verify gist URL display in search results and after selection.
3. **FOR** share command without arguments, **THEN** tests SHALL cover empty prompt list, single prompt, multiple prompts, and user cancellation scenarios.
4. **FOR** share command with keyword, **THEN** tests SHALL cover no matches, single match, multiple matches, and gist ID vs keyword priority.
5. **WHERE** error conditions exist, **THEN** tests SHALL verify appropriate error messages and graceful handling.
6. **WHEN** testing UI components, **THEN** tests SHALL use the existing test helpers and mock interfaces.
7. **IF** external dependencies are involved (gist client), **THEN** tests SHALL use mocks to ensure isolation and reliability.

Do the requirements look good? If so, we can move on to the design.