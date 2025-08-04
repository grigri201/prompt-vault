# Task Completion Guidelines

## Commands to Run After Task Completion

### Mandatory Steps
1. **Format Code**: `go fmt ./...`
   - Ensures consistent code formatting
   - Should be run before every commit

2. **Generate Dependencies**: `go generate ./internal/di`
   - Required if any dependency injection configuration was modified
   - Updates wire_gen.go with new providers

3. **Lint Code**: `go vet ./...`
   - Checks for potential issues and bugs
   - Must pass before code submission

4. **Run Tests**: `go test ./...`
   - Ensures all existing functionality still works
   - All tests must pass

### Optional but Recommended
5. **Build Application**: `go build -o pv`
   - Verifies that the application compiles correctly
   - Useful for catching compilation errors

6. **Test Manually**: `./pv <command>`
   - Run the specific command you implemented
   - Verify expected behavior

## Quality Checklist
- [ ] Code follows Go naming conventions
- [ ] Interfaces are properly implemented
- [ ] Error handling is consistent with project patterns
- [ ] Tests are written for new functionality
- [ ] Documentation is updated if needed
- [ ] No hardcoded values (use configuration)
- [ ] Dependency injection is properly configured

## Common Issues to Check
- Missing error handling
- Unused imports
- Inconsistent naming
- Missing wire providers for new dependencies
- Incomplete interface implementations