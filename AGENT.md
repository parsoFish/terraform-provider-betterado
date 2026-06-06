# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 0

- Created `azuredevops/internal/acceptancetests/data_release_definition_test.go` with three tests:
  - `TestAccDataReleaseDefinition_ById` — creates minimal release definition via `hclReleaseDefinitionBasic`, reads it back with `release_definition_id`, checks name/path via `TestCheckResourceAttrPair`, plus idempotency step.
  - `TestAccDataReleaseDefinition_ByName` — same fixture, reads by `name`, checks id/name/path.
  - `TestAccDataReleaseDefinitions_List` — same fixture, uses `betterado_release_definitions`, asserts `release_definitions.#` == 1, plus idempotency step.
- Reused `hclReleaseDefinitionBasic(name)` from `resource_release_definition_test.go` (same package — no re-declaration needed).
- Compiled cleanly with `go build -mod=vendor -tags all ./azuredevops/internal/acceptancetests/`.
- All unit tests pass.

## What worked

- Reusing `hclReleaseDefinitionBasic` from the same package (`acceptancetests`) — no duplication needed; it's already declared in `resource_release_definition_test.go`.
- `tonumber(betterado_release_definition.test.id)` to convert the string ID to int for `release_definition_id`.
- `TestCheckResourceAttrPair` is the idiomatic way to compare data source vs. resource attributes.
- For `release_definitions.#` == "1" in the list test: the project is freshly created per test so exactly one definition exists.

## What didn't work

_(no dead-ends encountered)_

## Open questions

_(none — all ACs addressed)_

## Notes for reflection

- The `creates:` path `azuredevops/internal/acceptancetests/data_release_definition_test.go` is now in the diff, satisfying the gate-tightening check.
- The acceptance gate (`go test -tags all -run TestAccDataReleaseDefinition|TestAccDataReleaseDefinitions ./azuredevops/internal/acceptancetests/`) will need `TF_ACC=1` + live ADO creds to actually run; those are provided by the environment.
