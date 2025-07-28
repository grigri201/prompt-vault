# Enhanced Prompt Sharing - Implementation Tasks

## Overview
This document contains the implementation tasks for the enhanced-prompt-sharing feature. Each task follows TDD methodology and builds incrementally on previous tasks.

## Tasks

### 1. Extract and Test Search Logic
- [x] **1.1 Create search package and implement search logic**
  - Create `internal/search/search.go` with `Searcher` struct
  - Extract `matchesKeyword` function from `internal/cli/get.go`
  - Extract search logic into `SearchEntries` method
  - References: Requirement 4.2, 6.2 (keyword matching across fields)

- [x] **1.2 Write comprehensive tests for search functionality**
  - Create `internal/search/search_test.go`
  - Test keyword matching in name, author, category, tags, description
  - Test empty keyword returns all entries
  - Test case-insensitive search
  - References: Requirement 9.4 (search test coverage)

- [x] **1.3 Update get command to use new search package**
  - Import and use `search.Searcher` in `internal/cli/get.go`
  - Remove old `matchesKeyword` function
  - Ensure all existing tests pass
  - References: Requirement 6.2 (consistent search behavior)

### 2. Enhance List Command with Gist URL Display
- [x] **2.1 Write failing tests for gist URL display in list command**
  - Add test cases in `internal/cli/list_test.go` for URL display
  - Test URL in table format
  - Test empty gist URL handling
  - Test pagination with URLs
  - References: Requirement 1.1, 9.1 (list command test scenarios)

- [x] **2.2 Implement gist URL column in list command**
  - Modify table header in `internal/cli/list.go` to include "Gist URL"
  - Add gist URL to table row output
  - Handle empty URLs gracefully
  - References: Requirement 1.1, 1.4 (display gist URL in table)

- [x] **2.3 Implement URL truncation for terminal width**
  - Add function to detect terminal width
  - Implement URL truncation with ellipsis
  - Test with various terminal widths
  - References: Requirement 1.2 (URL truncation)

### 3. Enhance Get Command with Gist URL Display
- [x] **3.1 Write failing tests for gist URL display in get command**
  - Add test cases in `internal/cli/get_test.go`
  - Test URL display in search results
  - Test URL display after selection
  - Test URL in clipboard success message
  - References: Requirement 2.1, 2.4, 9.2 (get command test scenarios)

- [x] **3.2 Add gist URL to prompt details display**
  - Modify prompt details output in `internal/cli/get.go`
  - Display URL after each prompt in search results
  - Show URL before variable filling form
  - References: Requirement 2.1, 2.2, 2.3 (display gist URL in get flow)

- [x] **3.3 Include gist URL in success message**
  - Modify clipboard success message to include gist URL
  - Format message appropriately
  - References: Requirement 2.4 (URL in success message)

### 4. Implement Share Command Selection Interface
- [x] **4.1 Write tests for gist ID detection function**
  - Create test for `isGistID` function
  - Test valid 32-character hex strings
  - Test invalid formats (wrong length, non-hex characters)
  - References: Requirement 4.6, 5.2 (gist ID detection)

- [x] **4.2 Implement gist ID detection utility**
  - Add `isGistID` function in `internal/cli/share.go`
  - Validate 32-character hexadecimal format
  - References: Requirement 4.6, 5.2 (distinguish gist ID from keyword)

- [x] **4.3 Write tests for share command without arguments**
  - Add test cases for empty argument scenario
  - Test prompt list display
  - Test selection and cancellation
  - Mock selector UI interactions
  - References: Requirement 3.1, 3.5, 9.3 (no-argument test scenarios)

- [ ] **4.4 Implement share command without arguments**
  - Modify `share.go` to accept 0 or 1 arguments
  - Load and display all prompts when no arguments
  - Integrate selector UI for prompt selection
  - Handle empty prompt list case
  - References: Requirement 3.1, 3.2, 3.3, 3.4 (share without arguments)

### 5. Implement Share Command with Keyword Search
- [ ] **5.1 Write tests for share command with keyword search**
  - Test keyword search with multiple matches
  - Test single match confirmation
  - Test no matches error
  - Test gist ID vs keyword priority
  - References: Requirement 4.1, 4.4, 4.5, 9.4 (keyword search tests)

- [ ] **5.2 Implement keyword search in share command**
  - Use `search.Searcher` for keyword matching
  - Display selection UI for multiple matches
  - Ask for confirmation on single match
  - Show appropriate error for no matches
  - References: Requirement 4.1, 4.2, 4.3, 4.4, 4.5 (keyword search)

- [ ] **5.3 Integrate share workflow after selection**
  - Connect selected prompt to existing share logic
  - Maintain existing error handling
  - Ensure backward compatibility
  - References: Requirement 3.3, 4.3, 5.1 (share workflow integration)

### 6. Integration Testing and Final Validation
- [ ] **6.1 Write integration tests for enhanced list command**
  - Test complete list command flow with URLs
  - Verify output formatting
  - Test with mock terminal widths
  - References: Requirement 9.1 (list integration tests)

- [ ] **6.2 Write integration tests for enhanced get command**
  - Test complete get command flow with URLs
  - Verify URL display at each stage
  - Test user interactions
  - References: Requirement 9.2 (get integration tests)

- [ ] **6.3 Write integration tests for enhanced share command**
  - Test all three modes: no args, keyword, gist ID
  - Test complete user interaction flows
  - Verify backward compatibility
  - References: Requirement 9.3, 9.4 (share integration tests)

- [ ] **6.4 Run all tests and ensure 80%+ coverage**
  - Execute full test suite
  - Check code coverage metrics
  - Add any missing test cases
  - References: Requirement 8.2, 8.7 (test coverage and passing)

Do the tasks look good?