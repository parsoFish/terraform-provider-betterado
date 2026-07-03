# Agent Memory — WI-6

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (this iteration — fresh start after branch already has all work)

Oriented on prior progress: all implementation was committed in prior Ralph sessions.
- Confirmed no `last-gate-failure.md` exists (fresh gate run coming).
- Verified all files in scope are present and compiling.
- Ran: `go build ./azuredevops/...` → OK; `make test` → 0 failures; `golangci-lint run --new-from-rev=main` → 0 issues; `make terrafmt-check` → OK.
- Test discovery: `go test -tags all -list TestAccEnvironmentResourceKubernetes ./azuredevops/internal/acceptancetests/` → `TestAccEnvironmentResourceKubernetes_createUpdate` found.
- Updated fix_plan.md to reflect all ACs ✅.

## What worked

- Framework resource pattern: all attributes RequiresReplace (no Update method), `requiresReplaceSet()` custom modifier for the `tags` Set attribute.
- `captureEnvironmentKubernetesEvidence()` helper: fetches live API, calls `testutils.CaptureLiveEvidence("acceptance-resource-environment-kubernetes", url, k8sResource)` — matches AC3 label exactly.
- `GetMuxedProviderFactories()` in `ProtoV6ProviderFactories` field — required for framework resources in acceptance tests.
- Using `SharedFixtureProjectName` ("betterado-standing-demo") avoids the 1000-project org limit that caused earlier iterations to fail.
- `strconv.Atoi` / `strconv.Itoa` for resourceID (the API returns int, TF state stores as string).
- Mux acceptance test pattern: `resource.ParallelTest`, `testutils.GetMuxedProviderFactories()`, `ExpectNonEmptyPlan: false`.

## What didn't work

- Prior iterations tried creating a new project per test run — hit the 1000-project ADO org limit. Solution: use `SharedFixtureProjectName`.
- SDKv2 provider's `Meta()` is nil in mux tests — must use `testutils.GetDirectClient()` (direct PAT-based client) in `CheckDestroy` and evidence helpers.

## Open questions

_(none blocking)_

## Notes for reflection

- The `requiresReplaceSet()` custom modifier pattern (needed because `listplanmodifier.RequiresReplace()` doesn't exist for Sets in the SDK) can be reused for other resources with Set attributes that are ForceNew.
- The `captureEnvironmentKubernetesEvidence` / `CaptureLiveEvidence` pattern for best-effort evidence capture (returns nil on error, never fails test) is the project standard for AC3-type requirements.
