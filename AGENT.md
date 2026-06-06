# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete)

- Read WI-2.md, AGENT.md, fix_plan.md, and prior commit log.
- WI-1 committed the mini-client at `azuredevops/utils/sdk/environmenttemplates/` (NOT `azuredevops/internal/utils/sdk/…` — note the path differs from what the spike test import path references; the correct path is `azuredevops/utils/sdk/environmenttemplates/`).
- Created `azuredevops/internal/service/release/resource_release_definition_environment_template.go` with:
  - `ResourceReleaseDefinitionEnvironmentTemplate()` — no `UpdateContext` (immutable).
  - Schema: `project_id` (Required, ForceNew), `name` (Required, ForceNew), `description` (Optional, ForceNew), `category` (Optional, ForceNew), `environment` (Optional, ForceNew), `icon_task_id` (Optional, ForceNew), `can_delete` (Computed), `icon_uri` (Computed).
  - `expandEnvironmentTemplate(d *schema.ResourceData)` → `*ReleaseDefinitionEnvironmentTemplate`.
  - `flattenEnvironmentTemplate(template, d)` → populates state.
  - CRUD handlers using `clients.EnvironmentTemplatesClient`.
- Created `azuredevops/internal/service/release/resource_release_definition_environment_template_test.go` with:
  - `TestReleaseDefinitionEnvironmentTemplate_Expand` — PASSES.
  - `TestReleaseDefinitionEnvironmentTemplate_Flatten` — PASSES.
  - Uses `schema.TestResourceDataRaw` for offline testing (no mocks needed).
- Updated `azuredevops/internal/client/client.go`:
  - Added import for `environmenttemplates`.
  - Added `EnvironmentTemplatesClient environmenttemplates.Client` field to `AggregatedClient`.
  - Initialised `environmentTemplatesClient := environmenttemplates.NewClient(ctx, connection)` in `GetAzdoClient`.
  - Added to `aggregatedClient` struct literal.
- Updated `azuredevops/provider.go`: added `"betterado_release_definition_environment_template": release.ResourceReleaseDefinitionEnvironmentTemplate()`.
- Created `docs/resources/release_definition_environment_template.md` and `examples/resources/betterado_release_definition_environment_template/resource.tf`.
- `go build ./...` — clean.
- Quality gate `go test -mod=vendor -tags all -count=1 -run TestReleaseDefinitionEnvironmentTemplate_ ./azuredevops/internal/service/release/` — **PASS** (both _Expand and _Flatten tests pass).
- Committed as `feat: implement betterado_release_definition_environment_template resource (WI-2)`.
- All 4 ACs are complete.

## What worked

- `schema.TestResourceDataRaw` for offline expand/flatten unit tests — no mock needed, clean pattern.
- `environmenttemplates.ReleaseDefinitionEnvironmentTemplate` is a type alias for `releaseapi.ReleaseDefinitionEnvironmentTemplate` (defined in `models.go`) — fields: `Name`, `Description`, `Category`, `CanDelete`, `IconUri`, `IconTaskId`, `Id`, `Environment`, `IsDeleted`.
- The mini-client lives at `azuredevops/utils/sdk/environmenttemplates/` (NOT `internal/utils/sdk/`).

## What didn't work

_(none — first iteration completed cleanly)_

## Open questions

_(none)_

## Notes for reflection

- The `ReleaseDefinitionEnvironmentTemplate` struct has no dedicated Update ADO API; immutability is the correct model (all fields `ForceNew`).
- `environment` field is stored as a raw string (JSON-encoded `ReleaseDefinitionEnvironment`) — a more ergonomic approach might use a nested TypeList block, but that is scope for a future WI.
