# Tasks: Filter Number Parsing Fix

## Implementation Tasks

### Task 1: Add Helper Functions for Digit Detection

- [x] 1.1 Add `endsWithDigit(s string) bool` function to `interface_config.go`
- [x] 1.2 Add `startsWithDigit(s string) bool` function to `interface_config.go`
- [x] 1.3 Ensure functions handle edge cases (empty strings, whitespace)

### Task 2: Modify preprocessWrappedLines Function

- [x] 2.1 Update the line joining logic in `preprocessWrappedLines`
- [x] 2.2 When current line ends with digit AND next line starts with digit, join without space
- [x] 2.3 Otherwise, maintain existing behavior (join with space)
- [x] 2.4 Ensure `dynamic` keyword case is handled correctly

### Task 3: Add Unit Tests for Mid-Number Wrap Scenarios

- [x] 3.1 Test 6-digit filter number split (200100 → 20010 + 0)
- [x] 3.2 Test filter number split into two valid numbers (200027 → 2000 + 27)
- [x] 3.3 Test with `dynamic` keyword after wrapped number
- [x] 3.4 Test IPv6 dynamic filter number wrap (101085 → 1 + 1085)
- [x] 3.5 Test normal continuation (no mid-number split) still works

### Task 4: Add Round-Trip Tests

- [x] 4.1 Test that generated filter commands parse back to original values
- [x] 4.2 Include 6-digit filter numbers in round-trip tests
- [x] 4.3 Test both IPv4 and IPv6 filter commands

### Task 5: Run Full Test Suite

- [x] 5.1 Run `go test ./...` to verify no regressions
- [x] 5.2 Run `go vet ./...` to check for issues
- [x] 5.3 Run linter if configured

### Task 6: Verify Fix with Terraform Plan

- [x] 6.1 Rebuild provider with `go build -o terraform-provider-rtx`
- [x] 6.2 Install provider locally
- [x] 6.3 Run `terraform plan -parallelism=2` in examples/import directory
- [x] 6.4 Verify filter number differences are resolved

## Dependencies

- Task 2 depends on Task 1
- Task 3 depends on Task 2
- Task 4 depends on Task 2
- Task 5 depends on Tasks 3 and 4
- Task 6 depends on Task 5
