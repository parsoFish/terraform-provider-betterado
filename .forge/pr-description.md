## Why

The `betterado_release_definition` resource used `environment { }` HCL block syntax for pipeline stages — a double misnomer: the keyword `environment` clashes semantically with Terraform's infrastructure-environment concept, and the block syntax prevents HCL `for`/`concat` expressions from constructing stage lists dynamically. Teams authoring multi-stage pipelines had to write N repetitive blocks with no way to DRY them; the provider's API surface was confusing and at odds with how the rest of the ecosystem names pipeline stages.

This initiative fixes both problems simultaneously: it renames the key to `stages` (what it actually models) and applies `ConfigMode: schema.SchemaConfigModeAttr` to `stages`, `deploy_phase`, `artifact`, `variable`, and `retention_policy` so they can be written as assignable list/object attributes — enabling the full Terraform expression surface. It is a breaking schema change with no back-compat alias, following the project's roadmap for a clean break.

## What

Six files changed (1748 insertions / 1177 deletions):

- **`azuredevops/internal/service/release/resource_release_definition.go`** — Schema key renamed `"environment"` → `"stages"`; helper functions renamed `expandEnvironments` → `expandStages`, `flattenEnvironments` → `flattenStages`; `ConfigMode: schema.SchemaConfigModeAttr` added to `stages`, `deploy_phase` (within a stage), `artifact` (top-level), `variable` (top and stage-level), `retention_policy` (MaxItems:1 within a stage).

- **`azuredevops/internal/service/release/resource_release_definition_test.go`** — All `schema.TestResourceDataRaw` maps updated from `"environment"` key to `"stages"`; new test `TestReleaseDefinition_StagesConfigModeAttr_Schema` added — it asserts `ConfigMode == schema.SchemaConfigModeAttr` for `stages`, `deploy_phase`, and `artifact` at schema-inspection time (fails on pre-WI-2 schema, passes after).

- **`azuredevops/internal/acceptancetests/resource_release_definition_test.go`** — All HCL fixture functions (≈30 tests) converted from `environment { }` block syntax to `stages = [{ }]` array syntax; `resource.TestCheckResourceAttr` paths updated from `environment.0.*` to `stages.0.*`; new acceptance test `TestAccReleaseDefinition_stagesArraySyntax` added — two-stage config, non-default `retention_policy`, non-default `deploy_phase`, `ExpectNonEmptyPlan: false`, clean destroy.

- **`azuredevops/internal/acceptancetests/shared_fixtures.go`** — Minor shared fixture updates to support the new acceptance test helper.

- **`examples/resources/betterado_release_definition/resource.tf`** — Converted from `environment { }` / `artifact { }` block syntax to `stages = [{ }]` / `artifact = [{ }]` array syntax; terrafmt-clean.

- **`docs/resources/release_definition.md`** — All HCL code blocks and attribute reference table updated from `environment` to `stages`; array syntax used throughout.

## How

The work was decomposed into four work items executed in dependency order:

1. **WI-1 (rename, behaviour-preserving)** — Changed the schema map key and all `d.Get`/`d.Set`/`d.GetOk` call sites in the Go implementation; renamed expand/flatten helpers. Gate: full `TestReleaseDefinition_*` unit suite (42 tests).

2. **WI-2 (ConfigModeAttr)** — Added `ConfigMode: schema.SchemaConfigModeAttr` to the five schema entries. Wrote `TestReleaseDefinition_StagesConfigModeAttr_Schema` as a discriminating unit test (fails on clean tree, passes after). Gate: `TestReleaseDefinition_StagesConfigModeAttr_Schema` → PASS.

3. **WI-3 (acceptance tests)** — Converted all HCL fixtures to array syntax and added `TestAccReleaseDefinition_stagesArraySyntax`. Gate: live ADO `TF_ACC` run (requires credentials in CI).

4. **WI-4 (examples + docs, behaviour-preserving)** — Updated HCL examples and markdown docs; ran `make terrafmt` then `make terrafmt-check`. Gate: `TestReleaseDefinition_*` unit suite + doc-audit tests.

**Quality gate** (verbatim, as run by forge):
```
go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...
```
Result: `ok github.com/.../service/release 0.022s` · `ok github.com/.../service/taskagent 0.007s` — all 44 tests green.

**Breaking change notice:** Any existing Terraform configuration using `environment { }` blocks with `betterado_release_definition` must be migrated to `stages = [{ }]` syntax. State migration is handled by the normal `terraform state mv` / plan-and-apply cycle; there is no automatic alias.
