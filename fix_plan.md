# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a live ADO project with at least one existing release definition (created via betterado_release_definition resource) WHEN TestAccDataReleaseDefinition_ById runs with TF_ACC=1 — terraform apply → data source read → assert name and path attributes THEN the data source resolves the known definition's attributes against live ADO; a re-plan produces no diff (ExpectNonEmptyPlan: false)
- [x] AC2: GIVEN a live ADO project with at least one existing release definition WHEN TestAccDataReleaseDefinition_ByName runs with TF_ACC=1 — terraform apply → data source read by name → assert id and attributes THEN the data source resolves correctly by name; re-plan produces no diff
- [x] AC3: GIVEN a live ADO project with at least one existing release definition WHEN TestAccDataReleaseDefinitions_List runs with TF_ACC=1 — terraform apply → data source list → assert release_definitions list non-empty THEN the list data source returns at least the definition created in the test fixture; re-plan produces no diff
