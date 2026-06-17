## Why

The `betterado_release_definition` resource had 8 documented writable fields from the ADO REST schema that the provider could not set — they appeared in `docs/release-definition-gap-matrix.md` as `missing`. The most impactful gap was `containerImageTrigger`: users who want to automatically trigger a release when a Docker Hub or ACR image tag is pushed had no provider-native way to configure this and had to fall back to manual ADO portal configuration or out-of-band scripting. The gap matrix also incorrectly showed `environmentTriggers`, `workflowTask.timeoutInMinutes`, `workflowTask.retryCountOnTaskFailure`, `deploymentInput.overrideInputs`, artifact `tag_filter.tags`, and `createReleaseOnBuildTagging` as missing, even though those fields had already landed in prior initiatives — the doc was stale and misleading to contributors.

## What

- **`container_image_trigger` schema block** added to the `triggers` stanza of `betterado_release_definition`. Configures `alias` (required string, must match a container artifact alias) and `tag_filter[]` (optional list with `pattern` string per entry).
- **`expandTriggers()`** extended: emits `triggerType=containerImageTrigger` entries with `alias` and `tagFilters[].pattern` for each configured block.
- **`flattenTriggers()`** extended: routes `case "containerImageTrigger":` back into Terraform state as `container_image_trigger[].alias` and `container_image_trigger[].tag_filter[].pattern`.
- **Unit test** `TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten` in `resource_release_definition_test.go` — proves expand and flatten round-trip for the new block.
- **Live acceptance test** `TestAccReleaseDefinition_withContainerImageTrigger` in `azuredevops/internal/acceptancetests/resource_release_definition_test.go` — runs `terraform apply` with a DockerHub artifact + `container_image_trigger`, asserts `triggers.0.container_image_trigger.0.alias`, verifies idempotency (`ExpectNonEmptyPlan: false`), and captures live ADO REST evidence before `destroy`.
- **Gap matrix refresh** (`docs/release-definition-gap-matrix.md`): all 8 previously-missing writable fields flipped to `mapped`; section summaries updated; new §1.12 EnvironmentTrigger and §1.13 ContainerImageTrigger sections added; Overall Summary table recalculated.
- **Example HCL** (`examples/resources/betterado_release_definition/resource.tf`): commented `container_image_trigger` block added to the triggers stanza.

Files changed (exactly as in the branch diff):
- `azuredevops/internal/service/release/resource_release_definition.go` — schema + expand + flatten
- `azuredevops/internal/service/release/resource_release_definition_test.go` — new unit test
- `azuredevops/internal/acceptancetests/resource_release_definition_test.go` — new acceptance test + live evidence capture
- `docs/release-definition-gap-matrix.md` — gap matrix to parity
- `examples/resources/betterado_release_definition/resource.tf` — example update

## How

**WI-1 (schema + offline unit test):** Added `container_image_trigger` TypeList to the existing `triggers` Elem.Schema alongside `cd_artifact_trigger` and `schedule_trigger`. `expandTriggers()` iterates the list and appends `map[string]interface{}{"triggerType": "containerImageTrigger", "alias": ..., "tagFilters": [...]}` entries. `flattenTriggers()` adds `case "containerImageTrigger":` that builds the inverse state map. `TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten` constructs an in-memory `schema.TestResourceDataRaw` with one trigger, calls expand, asserts the raw ADO shape, marshals/unmarshals JSON (to simulate the API round-trip), calls flatten, and asserts the recovered state. Gate: `go test -tags all -count=1 ./azuredevops/internal/service/release/...` — 53 tests pass, exit 0.

**WI-2 (live acceptance test):** `TestAccReleaseDefinition_withContainerImageTrigger` uses a DockerHub artifact (`type = "DockerHub"`, `definition = "library/nginx"`) because ADO accepts `containerImageTrigger` with DockerHub sources without requiring a live ACR service endpoint. The acceptance step asserts `triggers.0.container_image_trigger.0.alias = "_myContainer"` via `TestCheckResourceAttr`. A second step with `PlanOnly: true, ExpectNonEmptyPlan: false` confirms idempotency. A `captureAndWriteLiveEvidence` helper constructs the vsrm-host GET URL from org + project + definition ID and writes `.forge/live-evidence/acceptance-resource.json`. Test ran live (TF_ACC=1, org `davidgparsonson`) — apply green, idempotency confirmed, destroy clean.

**WI-3 (docs + examples):** `docs/release-definition-gap-matrix.md` updated by changing 7 stale `missing` rows to `mapped` and adding the new `container_image_trigger` row; section and overall summaries recalculated. New §1.12 and §1.13 sections document the block fields. `examples/resources/betterado_release_definition/resource.tf` gains a commented `container_image_trigger` block. Quality gate confirms no Go regressions from doc-only changes.
