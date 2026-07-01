# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete)

**Goal:** implement framework resource, acceptance test, register in provider.

**Files created/modified:**

| File | Change |
|------|--------|
| `azuredevops/internal/service/release/resource_release_folder_framework.go` | NEW — full framework resource |
| `azuredevops/internal/acceptancetests/resource_release_folder_framework_test.go` | NEW — acceptance test |
| `azuredevops/internal/provider/framework_provider.go` | Added `release.NewReleaseFolderResource` to Resources() slice |
| `CHANGELOG.md` | Draft FEATURES entry under `## [Unreleased]` |

**CI gates:**
- `make test` → PASS (0 failures)
- `golangci-lint run --new-from-rev=main ./azuredevops/...` → 0 issues
- `make terrafmt-check` → PASS
- `go build ./...` → PASS

## What worked

- **`staticString("")`** (package-local in `framework_defaults.go`) — used instead of `stringdefault` which is NOT vendored.
- **`requiresReplace()`** (package-local) — ForceNew plan modifier for `path` attribute.
- **Empty description normalisation** — API returns `nil` description for empty string; Read normalises `nil → ""` to avoid perpetual diff.
- **Resource ID = path** — mirrors SDKv2 (`d.SetId(*folder.Path)`). Idempotency should be clean.
- **Acceptance test build tag** `//go:build (all || resource_release_folder) && !exclude_resource_release_folder` — shares tag with SDKv2 test, which is correct.
- **`ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()`** — mux path, as required.

## What didn't work

_(none yet — iteration 0 succeeded on first pass)_

## Open questions

- Live acceptance (TF_ACC=1) pending orchestrator run. No code changes anticipated unless:
  - API normalises path (e.g. strips trailing backslash) → would need description-normalisation pattern applied to path too.
  - `betterado_release_folder` resource type conflict between SDKv2 and framework (both register the same type name). This WOULD be a problem; but the framework resource replaces the SDKv2 one in the mux — verify mux routing handles this correctly.

## Notes for reflection

_(observations the reflector should capture into the brain; the agent doesn't write them itself, but flags here)_
