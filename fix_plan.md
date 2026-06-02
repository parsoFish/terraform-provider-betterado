# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the codebase after all formatting and lint fixes are applied WHEN ./scripts/gofmtcheck.sh is run THEN the script exits 0 with no diff output — no Go files need gofmt reformatting
- [x] AC2: GIVEN the codebase after all formatting and lint fixes are applied WHEN make terrafmt-check is run (which invokes ./scripts/terrafmt.sh) THEN the script exits 0 with no HCL formatting errors in any _test.go terraform blocks
- [x] AC3: GIVEN the codebase after all formatting and lint fixes are applied WHEN golangci-lint run -v ./azuredevops/... is run against the configured .golangci.yml linters THEN the linter exits 0 with no errors reported
- [x] AC4: GIVEN the codebase after all formatting and lint fixes are applied WHEN go build -v ./... and then make test are run THEN both commands exit 0 — the build compiles cleanly and all unit tests pass
