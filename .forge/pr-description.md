## Why

The `betterado_release_definition` resource exposed pipeline stages under the misleading name `environment`, suggesting infrastructure environments rather than pipeline stages. This was a naming debt from the initial implementation that made the resource harder to understand and inconsistent with ADO terminology. Additionally, the block syntax (`environment { ... }` repeated) prevented users from using HCL `for`/`concat` expressions to build stage lists programmatically — a common need for teams managing many environments.

This initiative makes `betterado_release_definition` read naturally and unlocks dynamic HCL composition for release pipelines.

## What

**Breaking changes** (no back-compat alias):

- Renamed the Terraform schema attribute `environment` → `stages` across the provider source, unit tests, acceptance tests, examples, and docs.
- Applied `ConfigMode: schema.SchemaConfigModeAttr` to `stages`, `deploy_phase` (inside `stages`), and `retention_policy` (inside `stages`), converting them from block syntax to assignable list-of-object attributes.
- Renamed Go expand/flatten helpers: `expandEnvironments` → `expandStages`, `flattenEnvironments` → `flattenStages`.
- Updated `examples/resources/betterado_release_definition/resource.tf` and `docs/resources/release_definition.md` to the new `stages = [ { ... } ]` array syntax.
- Added `TestReleaseDefinition_StagesSchemaConfigMode` unit test (WI-2 fail-first gate).
- Added `TestAccReleaseDefinition_stagesArraySyntax` acceptance test with idempotency step (`ExpectNonEmptyPlan: false`).

**Files changed** (anchored to `git diff --name-only main...HEAD`):

- `azuredevops/internal/service/release/resource_release_definition.go`
- `azuredevops/internal/service/release/resource_release_definition_test.go`
- `azuredevops/internal/acceptancetests/resource_release_definition_test.go`
- `docs/resources/release_definition.md`
- `examples/resources/betterado_release_definition/resource.tf`

## How

**WI-1 — Rename** (`behavior_preserving`): Changed the schema map key `"environment"` → `"stages"` in `ResourceReleaseDefinition().Schema`; updated all `d.Get/d.Set/d.GetOk` call sites; renamed expand/flatten helpers; updated all unit test fixtures and `TestCheckResourceAttr` state paths to `stages`/`stages.N.*`.

**WI-2 — ConfigMode**: Added `TestReleaseDefinition_StagesSchemaConfigMode` as a fail-first gate (test was absent on baseline → gate treated as non-passing → dispatched). Applied `ConfigMode: schema.SchemaConfigModeAttr` to `stages`, `deploy_phase`, and `retention_policy`; honoured ConfigModeAttr constraints by ensuring `Required` fields that the API also computes are `Optional+Computed`.

**WI-3 — Acceptance tests**: Authored `TestAccReleaseDefinition_stagesArraySyntax` using `stages = [...]` array HCL with a two-step `apply`→`idempotency-check` structure and `CheckDestroy`. Converted all existing HCL helper functions and `TestCheckResourceAttr` paths to `stages`/`stages.N.*`. WI-3 is `failed` — live TF_ACC run against ADO was not completed in this cycle (no credentials); the test compiles clean and is ready to run.

**WI-4 — Examples + docs** (`behavior_preserving`): Replaced `environment { ... }` blocks with `stages = [ { ... } ]` in `resource.tf`; updated the Example Usage code block and Argument Reference in `release_definition.md`; confirmed no `environment` block remains in either file.

**Quality gate** (all iterations): `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` — 52 release tests + 30 taskagent tests pass, exit 0.
