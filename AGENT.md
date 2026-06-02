# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (this iteration) — 2025-06-01

**State on arrival:** Prior iterations (1–5) had already applied gofmt, terrafmt, golangci-lint auto-fixes and nolint suppressions to the files in scope. Their commits were safety-net autocommits (no agent note was written to AGENT.md).

**Gate status on arrival:**
- AC1 (`./scripts/gofmtcheck.sh`): PASS (exit 0, no diff output)
- AC2 (`make terrafmt-check`): PASS (exit 0)
- AC3 (`golangci-lint run -v ./azuredevops/...`): PASS (exit 0, 0 issues after nolint_filter)
- AC4a (`go build -v ./...`): PASS
- AC4b (`make test`): FAIL — `TestProvider_HasChildResources` expected 131 resources but provider had 132

**Root cause:** `betterado_task_group` was already registered in `provider.go` (on `main`) via `taskagent.ResourceTaskGroup()`, but was never listed in `provider_test.go`'s `expectedResources` slice. The test was broken on `main` as well — this was a pre-existing CI failure.

**Fix applied:** Added `"betterado_task_group"` to `expectedResources` in `azuredevops/provider_test.go` (alphabetically between `betterado_tagging_permissions` and `betterado_team`).

**Gate status after fix:**
- AC1: PASS
- AC2: PASS
- AC3: PASS
- AC4: PASS (all tests pass, `TestProvider_HasChildResources` now shows 132 expected == 132 actual)

**All ACs complete.**

## What worked

- Running the full gate (`./scripts/gofmtcheck.sh && make terrafmt-check && golangci-lint run -v ./azuredevops/... && go build -v ./... && make test`) immediately to identify the exact failure.
- Using `go run /tmp/check_resources.go` to enumerate all registered resources from the live `Provider().ResourcesMap`.
- Comparing registered vs expected lists to isolate `betterado_task_group` as the missing entry.

## What didn't work

_(none this iteration)_

## Open questions

_(none)_

## Notes for reflection

- The `TestProvider_HasChildResources` / `HasChildDataSources` tests are a manual registry — they must be updated whenever a resource is added or removed from `provider.go`. A future improvement would be to generate this list automatically.
- Prior iterations' safety-net autocommits did not write AGENT.md; the loop continued without institutional memory. This is worth noting for future loop hygiene.
