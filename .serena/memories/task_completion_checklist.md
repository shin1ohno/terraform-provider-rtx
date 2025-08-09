# Task Completion Checklist for Terraform Provider RTX

When completing any development task, follow this checklist to ensure code quality and maintainability:

## 1. Code Implementation
- [ ] Follow TDD approach - write tests first
- [ ] Implement feature/fix according to requirements
- [ ] Ensure all tests pass

## 2. Testing
- [ ] Unit tests written and passing
- [ ] Acceptance tests written (if applicable)
- [ ] Test coverage adequate (check with `go test -cover`)
- [ ] Edge cases covered
- [ ] Error scenarios tested

## 3. Code Quality Checks
```bash
# Format code
make fmt
# or
gofmt -s -w .
terraform fmt -recursive ./examples/

# Run all tests
make test

# Run linter (when golangci-lint is configured)
make lint

# Check for race conditions
go test -race ./...
```

## 4. Documentation
- [ ] Update relevant documentation files
- [ ] Add/update code comments where necessary
- [ ] Update SESSION_PROGRESS.md with completed work
- [ ] Generate docs if provider interface changed:
  ```bash
  make docs
  ```

## 5. Validation Steps
- [ ] Build succeeds: `make build`
- [ ] No compilation warnings
- [ ] No linting errors (when configured)
- [ ] Manual testing with Docker environment if needed

## 6. Git Hygiene
- [ ] Changes are focused and atomic
- [ ] Commit message follows project conventions
- [ ] No sensitive information in commits
- [ ] Code follows existing patterns

## 7. Specific Checks by Task Type

### For New Data Sources/Resources
- [ ] Schema properly defined with descriptions
- [ ] Read/Create/Update/Delete operations implemented (as applicable)
- [ ] Import functionality considered
- [ ] Example in `examples/` directory
- [ ] Acceptance tests cover CRUD lifecycle

### For Client/Parser Changes
- [ ] Interface compatibility maintained
- [ ] Backward compatibility considered
- [ ] Error handling comprehensive
- [ ] Retry logic appropriate
- [ ] Parser tests with real RTX output samples

### For Security-Related Changes
- [ ] No hardcoded credentials
- [ ] Sensitive fields marked appropriately
- [ ] SSH security not compromised
- [ ] Host key verification maintained

## 8. Performance Considerations
- [ ] No resource leaks (connections, goroutines)
- [ ] Proper context handling
- [ ] Efficient algorithms used
- [ ] No unnecessary allocations

## 9. Final Checks
- [ ] All TODOs addressed or documented
- [ ] No commented-out code
- [ ] No debug print statements
- [ ] Dependencies updated if needed: `go mod tidy`

## 10. Communication
- [ ] Update SESSION_PROGRESS.md
- [ ] Note any decisions or trade-offs made
- [ ] Document any known limitations
- [ ] List follow-up tasks if any