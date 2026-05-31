---
title: release_definition unit-test substrate (11 tests)
slug: 2026-05-31-release-definition-unit-test-substrate
description: Canonical 11-test gomock substrate for betterado_release_definition — 5 baseline CRUD + 6 characterization tests; gate pattern, fixtures, and a production fix discovered during testing.
category: pattern
project: terraform-provider-betterado
created_at: 2026-05-31T11:30:00Z
updated_at: 2026-05-31T11:30:00Z
related_themes:
  - 2026-05-31-forge-onboarding-findings
  - 2026-05-18-stack-and-test-layout
---

# release_definition unit-test substrate (11 tests)

`azuredevops/internal/service/release/resource_release_definition_test.go` (1008 lines) is now the canonical substrate for `betterado_release_definition`, mirroring the pattern from `resource_task_group_test.go`.

## Baseline CRUD tests (WI-1, 5 tests)

1. `TestReleaseDefinition_ExpandFlatten_Roundtrip` — lossless round-trip through flatten→expand.
2. `TestReleaseDefinition_Create_DoesNotSwallowError` — SDK error surfaces as non-nil Diagnostics.
3. `TestReleaseDefinition_Read_ClearsIdOn404` — 404 WrappedError clears resource ID cleanly.
4. `TestReleaseDefinition_Update_CallsSDKWithArgs` — UpdateReleaseDefinition called with correct project + definition args.
5. `TestReleaseDefinition_Delete_SurfacesAPIError` — SDK error on delete surfaces as Diagnostics.

## Characterization tests (WI-2, 6 tests)

6. `TestReleaseDefinition_Update_RevisionRetryOnConflict` — 409 retry path: 1× GetReleaseDefinition + 2× UpdateReleaseDefinition.
7. `TestReleaseDefinition_SecretVariables_PreserveOnFlatten` — null API value for secret → preserves Terraform state value, not empty string.
8. `TestReleaseDefinition_DeepNestedEnvironment_ExpandFlatten` — full env→phase→step→input chain round-trip.
9. `TestReleaseDefinition_Artifacts_DefinitionReferenceFiltering` — computed keys not silently dropped during flatten.
10. `TestReleaseDefinition_ApprovalOptions_RoundTrip` — approvalOptions struct survives expand→flatten.
11. `TestReleaseDefinition_DeployPhases_JSONMarshalUnmarshal` — JSON serialisation of DeployPhase objects round-trips without data loss.

## Quality gate

```bash
go test -mod=vendor -tags all -count=1 -run TestReleaseDefinition ./azuredevops/internal/service/release/
```

Same betterado gate pattern as task_group: `-tags all`, `-run` scoped to new tests, exact package dir (no `/...`).

## Production fix discovered

`expandWorkflowTask` expected `string` for `inputs` field but Terraform SDK passes `map[string]interface{}`. Added type-switch to handle both. Found by `TestReleaseDefinition_DeepNestedEnvironment_ExpandFlatten`.

## Sources

- `_logs/2026-05-31T10-57-52_INIT-2026-05-31-release-definition-unit-tests/events.jsonl` (EV_mpto32zo dev-loop.delivered: files_changed=4, insertions=1253)
- `brain/cycles/_raw/2026-05-31T10-57-52_INIT-2026-05-31-release-definition-unit-tests.md`
