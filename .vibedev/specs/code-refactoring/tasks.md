# Code Refactoring Implementation Tasks

This document outlines the implementation tasks for refactoring the prompts-vault project based on the design document. Each task focuses on writing, modifying, or testing code components.

## Phase 1: Foundation Components

- [x] **1.1 Implement the error handling package**
  - Create `internal/errors/errors.go` with AppError type and error constructors
  - Implement error types: AuthError, NetworkError, FileSystemError, ValidationError, ParsingError
  - Add error wrapping utilities
  - Write comprehensive unit tests for all error types and functions
  - **References**: Requirement 1.1, 1.2, 1.3

- [x] **1.2 Implement the path management module**
  - Create `internal/paths/paths.go` with PathManager struct
  - Implement GetCachePath(), GetConfigPath(), GetIndexPath() methods
  - Implement EnsureDir() and AtomicWrite() utility methods
  - Write unit tests covering all path scenarios and atomic write operations
  - **References**: Requirement 2.1, 2.2, 2.4

- [x] **1.3 Create the base manager interface and implementation**
  - Create `internal/managers/manager.go` with Manager interface
  - Implement BaseManager struct with common functionality
  - Write unit tests for the base manager implementation
  - **References**: Requirement 3.1, 3.2

- [x] **1.4 Implement the dependency injection container**
  - Create `internal/container/container.go` with Container struct
  - Implement NewContainer() and NewTestContainer() constructors
  - Write unit tests for container initialization
  - **References**: Requirement 4.1, 4.2

- [x] **1.5 Create test helper utilities**
  - Create `internal/testhelpers/helpers.go` with SetupTest() function
  - Implement AssertErrorType() and other common test utilities
  - Write tests to validate test helper functionality
  - **References**: Requirement 7.1, 7.3

## Phase 2: Manager Refactoring

- [x] **2.1 Refactor cache manager to use new patterns**
  - Update `internal/cache/cache.go` to embed BaseManager
  - Replace direct path operations with PathManager
  - Update constructor to accept PathManager dependency
  - Implement Manager interface methods (Initialize, Cleanup, IsInitialized)
  - Update all existing cache manager tests to use new patterns
  - **References**: Requirement 3.3, 3.4, 4.2

- [x] **2.2 Refactor config manager to use new patterns**
  - Update `internal/config/config.go` to embed BaseManager
  - Replace direct path operations with PathManager
  - Update constructor to accept PathManager dependency
  - Implement Manager interface methods
  - Update all existing config manager tests
  - **References**: Requirement 3.3, 3.4, 4.2

- [x] **2.3 Create mock implementations for testing**
  - Create `internal/cache/mock.go` with MockManager
  - Create `internal/config/mock.go` with MockManager
  - Implement configurable mock behaviors for all manager methods
  - **References**: Requirement 7.3

## Phase 3: Error Handling Migration

- [ ] **3.1 Replace error handling in CLI commands**
  - Update all files in `internal/cli/` to use the new error package
  - Replace `fmt.Errorf("failed to...")` patterns with typed errors
  - Ensure error messages remain backward compatible
  - Update tests to verify error types
  - **References**: Requirement 1.4, 1.5

- [ ] **3.2 Update error handling in business logic modules**
  - Update error handling in `internal/gist/`, `internal/parser/`, `internal/auth/`
  - Use appropriate error types for each module
  - Maintain existing error messages for compatibility
  - **References**: Requirement 1.2, 1.4

## Phase 4: Dependency Injection Implementation

- [ ] **4.1 Update CLI commands to use dependency container**
  - Modify command constructors to accept Container instead of global functions
  - Update `cmd/pv/main.go` to initialize and pass Container
  - Remove global variable usage (getCachePathFunc)
  - **References**: Requirement 4.1, 4.3, 4.5

- [ ] **4.2 Update all CLI command tests**
  - Replace test setup code with testhelpers.SetupTest()
  - Use mock managers instead of overriding global functions
  - Ensure all tests pass with new dependency injection
  - **References**: Requirement 4.4, 7.1

## Phase 5: UI Component Abstraction

- [ ] **5.1 Create base UI component interface and implementation**
  - Create `internal/ui/component.go` with Component interface
  - Implement BaseComponent with common UI functionality
  - Write unit tests for base component
  - **References**: Requirement 5.1, 5.5

- [ ] **5.2 Refactor form component to use base implementation**
  - Update `internal/ui/form.go` to embed BaseComponent
  - Remove duplicate error handling code
  - Ensure all existing functionality is preserved
  - Update form component tests
  - **References**: Requirement 5.2, 5.4

- [ ] **5.3 Refactor selector component to use base implementation**
  - Update `internal/ui/selector.go` to embed BaseComponent
  - Remove duplicate code
  - Update selector component tests
  - **References**: Requirement 5.2, 5.4

- [ ] **5.4 Refactor paginator component to use base implementation**
  - Update `internal/ui/paginator.go` to embed BaseComponent
  - Remove duplicate code
  - Update paginator component tests
  - **References**: Requirement 5.2, 5.4

## Phase 6: Code Duplication Cleanup

- [ ] **6.1 Extract common validation logic**
  - Create `internal/validation/validation.go` for shared validation functions
  - Move duplicate validation code from CLI commands
  - Write unit tests for validation functions
  - Update CLI commands to use shared validation
  - **References**: Requirement 6.2, 6.3

- [ ] **6.2 Create file operation utilities**
  - Create `internal/fileutil/fileutil.go` for common file operations
  - Extract duplicate file reading/writing patterns
  - Implement using PathManager for atomic operations
  - Write comprehensive tests
  - **References**: Requirement 6.3, 6.5

- [ ] **6.3 Consolidate JSON handling utilities**
  - Create shared JSON marshaling/unmarshaling functions
  - Replace duplicate JSON handling code
  - Add proper error handling with typed errors
  - **References**: Requirement 6.3

## Phase 7: Integration and Verification

- [x] **7.1 Run all existing tests to ensure backward compatibility**
  - Execute full test suite
  - Fix any failing tests
  - Verify no regression in functionality
  - **References**: Requirement 7.2, Technical Constraints

- [ ] **7.2 Add integration tests for refactored components**
  - Write integration tests for Container initialization
  - Test manager interactions through Container
  - Verify error propagation through the system
  - **References**: Requirement 7.4

- [ ] **7.3 Update linting and formatting**
  - Run `go fmt` on all modified files
  - Run `golangci-lint` and fix any issues
  - Ensure code follows Go best practices
  - **References**: Success Criteria

- [ ] **7.4 Verify e2e tests continue to pass**
  - Run all e2e tests in the `e2e/` directory
  - Ensure CLI behavior remains unchanged
  - Fix any issues that arise
  - **References**: Requirement 7.5, Technical Constraints