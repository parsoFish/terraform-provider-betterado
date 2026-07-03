# Fix Plan

> Checklist for UWI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the SDKv2 betterado_workitem validator baseline WHEN resource_workitem_framework.go Schema() is finalized THEN validators restore parity: UUID pattern on project_id, non-whitespace on title/type/state/area_path/iteration_path, ConflictsWith (resourcevalidator.Conflicting) between custom_fields and additional_fields_json, size floor on tags, int64validator.AtLeast on parent_id — grep on that FILE shows validator usage
- [x] AC2: GIVEN the unreferenced SDKv2 helpers WHEN the package is cleaned THEN azuredevops/internal/service/workitemtracking/utils/classification.go and classification_test.go are deleted (or re-wired with references) — no orphaned SDKv2 helper remains
- [x] AC3: GIVEN the org project-cap policy and the shared fixture WHEN workitem acceptance tests run THEN all TestAccWorkItem_* functions use SharedFixtureProjectName instead of creating ad-hoc betterado_project resources (7 remaining conversions)
- [x] AC4: GIVEN the workitemquery_folder AC currently disclosed as partial WHEN the live gate re-runs THEN a real capture lands under its per-type label OR the AC stays honestly partial with the folder test wiring CaptureLiveEvidence added for the next run — captureQueryFolderEvidence already wired in TestAccWorkItemQueryFolder_UnderArea
- [x] AC5: GIVEN main at 1.8.0+ WHEN packaging finalizes after the fan-in merge THEN PROVIDER_VERSION.txt bumps above main and CHANGELOG retains sibling entries — bumped 1.2.1 → 1.9.1 (main=1.9.0), CHANGELOG updated

## All ACs complete — committed as 70ff5e55
