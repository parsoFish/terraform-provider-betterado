# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a minimal betterado_release_definition using block syntax (stages { name = 'Prod' … }) WHEN TestAccReleaseDefinition_basic runs live (TF_ACC=1) against real ADO THEN terraform apply succeeds, read-back confirms stages.0.name='Production', idempotency re-plan produces no diff, destroy cleans up
  - Test already existed (converted to block syntax in WI-2/WI-3). captureReleaseEvidence added in iteration 1.
  - Iteration 3 (commit 8c591b62): Added explicit idempotency step (PlanOnly:true, ExpectNonEmptyPlan:false) to _basic.
  - Awaits live gate run by orchestrator.
- [x] AC2: GIVEN a betterado_release_definition with a container_image_trigger block in triggers WHEN TestAccReleaseDefinition_withContainerImageTrigger runs live (TF_ACC=1) THEN apply succeeds, triggers.0.container_image_trigger.0.artifact_alias and .label round-trip cleanly, idempotency re-plan produces no diff, destroy cleans up
  - Test added in iteration 1 (commit 6cedfd8d). Awaits live gate run.
- [x] AC3: GIVEN the complete exhaustive acceptance test TestAccReleaseDefinition_complete WHEN it runs live (TF_ACC=1) against real ADO with block-syntax HCL THEN all assertions pass, idempotency re-plan produces no diff (ExpectNonEmptyPlan: false)
  - Test already existed. Iteration 3 (commit 8c591b62): Made ExpectNonEmptyPlan:false explicit in idempotency step.
  - Awaits live gate run by orchestrator.
- [x] AC4: GIVEN captureReleaseEvidence is called during the live acceptance run WHEN the resource is live (before destroy) THEN .forge/live-evidence/acceptance-resource.json is written with the real vsrm REST GET URL
  - captureReleaseEvidence called in _basic (iteration 1) and _withContainerImageTrigger (iteration 1). Awaits live run.
- [x] AC5: GIVEN docs/release-definition-gap-matrix.md WHEN it is refreshed after this cycle THEN container_image_trigger row is marked 'mapped' and all 8 previously-writable gaps are marked 'mapped'
  - Done in iteration 1 (commit 6cedfd8d). All 8 writable gaps now 'mapped' in matrix.

## Iteration 2 (2026-06-17)

- Fixed terrafmt alignment issue in `hclReleaseDefinitionBlockSyntax` fixture (commit 3dd9975b).
- All offline quality gates now pass: gofmt ✓, go build ✓, go vet ✓, terrafmt-check ✓, acceptance test package compiles ✓.

## Iteration 3 (2026-06-17)

- Added explicit idempotency step to `TestAccReleaseDefinition_basic` (AC1/WI-5): `PlanOnly:true, ExpectNonEmptyPlan:false`.
- Made `ExpectNonEmptyPlan:false` explicit in `TestAccReleaseDefinition_complete` idempotency step (AC3/WI-5).
- All offline gates still pass: gofmt ✓, go build ✓, go vet ✓, terrafmt-check ✓, acceptance package compiles ✓.
- Commit: 8c591b62.

## Remaining

- AC1, AC2, AC3, AC4: live gate run against real ADO (TF_ACC=1). The orchestrator runs these via quality_gate_cmd. No further code changes needed — all offline gates pass.
