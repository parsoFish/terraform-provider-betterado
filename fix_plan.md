# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a betterado_release_definition with an ACR artifact source and a container_image_trigger block referencing that alias WHEN TestAccReleaseDefinition_withContainerImageTrigger runs with TF_ACC=1 against real ADO THEN terraform apply succeeds, the provider read-back shows triggers.0.container_image_trigger.0.alias matching the configured alias, and an idempotency re-plan produces no diff (ExpectNonEmptyPlan: false)
- [x] AC2: GIVEN the acceptance test live read-back has completed (before destroy) WHEN testutils.CaptureLiveEvidence is called with label acceptance-resource and the vsrm-host GET URL THEN .forge/live-evidence/acceptance-resource.json is written containing the real ADO REST GET response body for the release definition

## Notes

AC1 + AC2 fully verified live (commit 71de844e, iteration 1):
- TestAccReleaseDefinition_withContainerImageTrigger PASSES with TF_ACC=1
- apply + triggers check + idempotency re-plan (ExpectNonEmptyPlan: false) all green
- .forge/live-evidence/acceptance-resource.json written (captureReleaseEvidence → testutils.CaptureLiveEvidence)
- Quality gate: `go test -tags all -run TestAccReleaseDefinition_withContainerImageTrigger ./azuredevops/internal/acceptancetests/` → PASS

## Iteration 1 fixes required (not in iteration 0 commit 8405e536):
- stages HCL needed all required null fields for SchemaConfigModeAttr structural type (id, owner, variable, etc.)
- is_primary must be true (ADO auto-promotes first artifact; false caused perpetual diff)
- DockerHub definition_reference fields: {connection, definition, namespaces} (defaultTag and registrytype are invalid ADO fields)
- artifact type ValidateFunc expanded to include DockerHub and AzureContainerRepository
