# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 0 (WI-4 — all ACs addressed)

**Context from prior WI-3 work:**
- `docs/resources/task_group.md` is the correct path (tfplugindocs strips the `betterado_` prefix from resource names)
- `testutils.GetMuxedProviderFactories()` lives in `azuredevops/internal/acceptancetests/testutils/mux_provider.go`
- `testutils.GetProviders()` is the old SDKv2-only factory; must be replaced with `GetMuxedProviderFactories()` for framework resources
- `checkTaskGroupDestroyed` and `getDirectClient()` are defined in `resource_task_group_test.go` in the same package

**AC1 - data_task_group_test.go:**
- Updated HCL from block syntax to array-of-objects: `version = [{...}]`, `input = [{...}]`, `task = [{...}]`
- Changed `Providers: testutils.GetProviders()` → `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
- Added imports: `github.com/google/uuid`, `github.com/hashicorp/terraform-plugin-testing/terraform`, `github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent`, `os`, `strings`
- Added `captureTaskGroupDataSourceEvidence()` function writing to label `task-group-datasource-acceptance`

**AC2 - docs:**
- Ran `make docs` (tfplugindocs v0.20.0) successfully
- Ran `git checkout -- docs/guides/` to restore hand-written guides
- `docs/resources/task_group.md` now shows `task`, `input`, `version` as `(Attributes List)` — framework format
- Note: tfplugindocs generates `task_group.md` (strips provider prefix); AC2 mentions `betterado_task_group.md` but that's the resource type; the file path is always `task_group.md`
- examples/resources/betterado_task_group/resource.tf already uses array-of-objects syntax (updated by WI-3)

**AC3 - changelog + version:**
- Added ENHANCEMENTS section to CHANGELOG.md ## Unreleased describing the migration
- Bumped PROVIDER_VERSION.txt from `0.3.0` to `0.4.0`

**Quality gates:**
- `go build -tags all ./...` — ✅ clean
- `go vet -tags all ./azuredevops/internal/acceptancetests/` — ✅ clean
- `make terrafmt-check` — ✅ exit 0
- `go test -tags all -run "^$" ./azuredevops/internal/acceptancetests/` — ✅ compiles, no tests run offline

## What worked

- `make docs` → `git checkout -- docs/guides/` pattern works correctly
- tfplugindocs reuses existing template for betterado_task_group (template was already crafted by WI-2/WI-3) and regenerates the schema section from the live framework schema
- Using `python3 -c "..."` to do multiline replacements in CHANGELOG.md (sed multiline doesn't work reliably in bash)

## What didn't work

- `sed` with `\n` in pattern for multiline CHANGELOG edit — doesn't match across newlines in GNU sed

## Open questions

- The live TF_ACC test (quality gate cmd) needs ADO credentials at `AZDO_ORG_SERVICE_URL` + `AZDO_PERSONAL_ACCESS_TOKEN`. The orchestrator runs this with `TF_ACC=1` in the serve env.

## Notes for reflection

- WI-4 is essentially a "cleanup + doc + version" WI — all the heavy framework code was done in WI-1 through WI-3
- The data source test needed BOTH the HCL syntax update AND the provider factory update to work with the mux provider
