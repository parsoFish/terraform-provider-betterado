# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a betterado_release_definition with an ACR artifact source and a container_image_trigger block referencing that alias WHEN TestAccReleaseDefinition_withContainerImageTrigger runs with TF_ACC=1 against real ADO THEN terraform apply succeeds, the provider read-back shows triggers.0.container_image_trigger.0.alias matching the configured alias, and an idempotency re-plan produces no diff (ExpectNonEmptyPlan: false)
- [x] AC2: GIVEN the acceptance test live read-back has completed (before destroy) WHEN testutils.CaptureLiveEvidence is called with label acceptance-resource and the vsrm-host GET URL THEN .forge/live-evidence/acceptance-resource.json is written containing the real ADO REST GET response body for the release definition

## Notes

Both ACs addressed in commit 8405e536:
- TestAccReleaseDefinition_withContainerImageTrigger added to resource_release_definition_test.go
- Uses captureReleaseEvidence(tfNode) which calls testutils.CaptureLiveEvidence("acceptance-resource", vsrmURL, def)
- The quality gate (go test -tags all -run TestAccReleaseDefinition_withContainerImageTrigger) requires TF_ACC=1 to run live
