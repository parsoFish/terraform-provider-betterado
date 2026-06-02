# Fix Plan — unifier sub-phase

> Initiative-level acceptance criteria. Tick each as you prove it against branch tip. Iteration 1 is initial prep; iterations 2+ react to either gate failures or send-back feedback.

- [x] AC1 (WI-1): GIVEN the codebase after all formatting and lint fixes are applied WHEN ./scripts/gofmtcheck.sh is run THEN the script exits 0 with no diff output — no Go files need gofmt reformatting
- [x] AC2 (WI-1): GIVEN the codebase after all formatting and lint fixes are applied WHEN make terrafmt-check is run (which invokes ./scripts/terrafmt.sh) THEN the script exits 0 with no HCL formatting errors in any _test.go terraform blocks
- [x] AC3 (WI-1): GIVEN the codebase after all formatting and lint fixes are applied WHEN golangci-lint run -v ./azuredevops/... is run against the configured .golangci.yml linters THEN the linter exits 0 with no errors reported
- [x] AC4 (WI-1): GIVEN the codebase after all formatting and lint fixes are applied WHEN go build -v ./... and then make test are run THEN both commands exit 0 — the build compiles cleanly and all unit tests pass
  - Verified: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → 45 tests, 3 packages, all PASS, exit 0
