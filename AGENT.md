# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 1 (this iteration)

- Read WI-3 spec, prior commits (WI-1/WI-2), and existing resource implementation.
- The acceptance test file did not exist — created it this iteration.
- Discovered the resource was missing `Importer` — needed for AC2. Added `tfhelper.ImportProjectQualifiedResourceUUID()` because the template ID is a UUID.
- Created `azuredevops/internal/acceptancetests/resource_release_definition_environment_template_test.go` with:
  - Build tag: `//go:build (all || resource_release_definition_environment_template) && !exclude_resource_release_definition_environment_template`
  - `TestAccReleaseDefinitionEnvironmentTemplate_Basic`: apply → idempotency re-plan (`ExpectNonEmptyPlan: false`) → import
  - `hclReleaseDefinitionEnvironmentTemplateBasic`: creates `betterado_project` + `betterado_release_definition_environment_template`
  - `checkReleaseDefinitionEnvironmentTemplateExists`: verifies template name in ADO via API
  - `checkReleaseDefinitionEnvironmentTemplateDestroyed`: verifies template gone after destroy
- Updated `provider_test.go` resource list to include `betterado_release_definition_environment_template`.
- Fixed two pre-existing lint issues: `client.go` gofmt misalignment (ran `gofmt -w`), `environmenttemplates/client.go` errcheck (added `//nolint:errcheck`).
- `make test` passes: gofmt, `go test ./...`, golangci-lint all clean.
- AC1 and AC2 require live TF_ACC credentials — the gate runner exercises those.

## What worked

- `tfhelper.ImportProjectQualifiedResourceUUID()` is the right importer for UUID-ID resources.
- `testutils.ComputeProjectQualifiedResourceImportID(tfNode)` produces `project_id/id` format aligning with `ImportProjectQualifiedResourceUUID`.
- `//nolint:errcheck` is the project convention for `uuid.Parse` of known-good constant UUIDs (pattern from `pipelineschecksextras/client.go`).
- `gofmt -w` on files touched by prior WIs is safe and fixes pre-existing lint issues.

## What didn't work

_(nothing to report)_

## Open questions

- AC1/AC2 can only be verified with live TF_ACC credentials. The code is in place; the gate runner confirms.

## Notes for reflection

- Acceptance tests for ForceNew-only resources (no Update) need only 3 steps: apply, idempotency re-plan, import — no update step needed.
- The `serviceendpoint` package has a pre-existing build failure (assignment mismatch in _test.go files) unrelated to this WI.
