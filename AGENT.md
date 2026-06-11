# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 0 (final)

- Ran `TestAccDataReleaseDefinitionRevision_Basic` and `TestAccDataReleaseDefinitionHistory_Basic` against live ADO (https://dev.azure.com/davidgparsonson).
- **Both tests PASSED** in ~35 seconds.
- Key: must use `set -a; source secrets.env; set +a` to export env vars before running `go test` — using `source secrets.env &&` in the same shell does NOT export them to the child process properly.
- Offline unit tests also pass (`./azuredevops/internal/service/release/...` + `./azuredevops/internal/service/taskagent/...`).
- All prior iterations (WI-1) had already committed the implementation: resource_release_definition.go sets `d.Set("revision", ...)` at line 1495, hclReleaseDefinitionBasic helper includes all required ADO fields (both pre- and post-deploy approvals + retention_policy), data sources read correctly.
- No code changes were needed this iteration — the implementation from prior work was complete and correct.

## What worked

- `set -a; source secrets.env; set +a` pattern to export vars from secrets.env into the test environment.
- Both data sources (`betterado_release_definition_revision` + `betterado_release_definition_history`) work correctly against live ADO.
- `hclReleaseDefinitionBasic` helper satisfies ADO REST 7.2 requirements (VS402982/VS402877) with both pre- and post-deploy approvals and retention_policy.
- `revision` attribute is correctly exposed in Terraform state via `flattenReleaseDefinition` at line 1495 of resource_release_definition.go.

## What didn't work

- Running `source secrets.env && go test ...` in a single shell command — the env vars are not exported to the child `go test` process (they're only in the current shell).

## Open questions

_(none — both ACs verified)_

## Notes for reflection

- Both live acceptance tests pass against real ADO. WI-2 is complete.
- The `set -a; source secrets.env; set +a` pattern is the correct way to export secrets.env vars when running go tests via bash.
