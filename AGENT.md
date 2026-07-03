# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

**Iteration 0 (this iteration):**
- Read `azdosdkmocks/test_sdk_mock.go` — enumerated all 43 SDK methods on `MockTestClient`.
- Read `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/test/client.go` — confirmed REST route shapes and HTTP verbs.
- Checked vendor for `testplan` package — **not present**. Only the legacy `test` package is vendored. The modern `_apis/testplan/*` endpoints (Plans, Suites, Configurations, Variables) are not exposed by the vendored SDK.
- Read `docs/release-definition-gap-matrix.md` and `docs/task-group-gap-matrix.md` for format reference.
- Wrote `docs/test-gap-matrix.md` — full 16-resource-type classification table, status summary, declarative-vs-ephemeral rationale, SDK ops reference, and WI roadmap.
- Quality gate `go test -tags all -count=1 -run TestProvider_HasChildResources ./azuredevops/` passes (this WI makes no Go code changes).
- Committed: `3da2ef27` — "docs: add ADO Test REST API v7.1 gap matrix"

## What worked

- WI only requires a markdown document; the quality gate (`TestProvider_HasChildResources`) just checks the provider registry and passes trivially with no code changes.
- Format from `release-definition-gap-matrix.md` (6-column table + status summary + rationale sections) worked well for this style of matrix.

## What didn't work

_(nothing to record — first iteration completed successfully)_

## Open questions

- The `testplan` package is not vendored. WI-2+ will need to vendor it or wire direct REST calls for plan/suite/config/variable CRUD.

## Notes for reflection

- The legacy `test` SDK (resource area `c2aa639c-3ccc-4740-b3b6-ce2a1e1d984e`) only covers `_apis/test/*`. The modern plan/suite management surface (`_apis/testplan/*`) requires a different SDK package that is absent from the vendor directory.
- Both ACs are satisfied. WI is done.
