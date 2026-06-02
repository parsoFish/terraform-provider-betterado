# Fix Plan — unifier sub-phase

> Initiative-level acceptance criteria. Tick each as you prove it against branch tip. Iteration 1 is initial prep; iterations 2+ react to either gate failures or send-back feedback.

- [x] AC1 (WI-1): GIVEN the codebase after gofmt fixes are applied WHEN running ./scripts/gofmtcheck.sh THEN the script exits 0 with no diff output
- [x] AC2 (WI-1): GIVEN the codebase after terrafmt fixes are applied WHEN running make terrafmt-check THEN the command exits 0 with no formatting errors
- [x] AC3 (WI-1): GIVEN the codebase after golangci-lint fixes are applied WHEN running golangci-lint run ./... (as invoked by the ci_gate) THEN the linter exits 0 with no errors reported
- [x] AC4 (WI-1): GIVEN the codebase after all formatting and lint fixes WHEN running make test (fmtcheck + go test ./...) THEN both fmtcheck and go test exit 0, all unit tests pass
