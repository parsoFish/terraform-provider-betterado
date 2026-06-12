# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN no TestAccReleaseDefinition_import test exists WHEN a new TestAccReleaseDefinition_import test is added that creates a betterado_release_definition then imports it by project_id/definition_id THEN the test passes live (TF_ACC=1), the imported state matches the created state (ImportStateVerify: true), and the idempotency step (PlanOnly) produces no diff
- [x] AC2: GIVEN the importer is wired via tfhelper.ImportProjectQualifiedResource WHEN the import step uses ComputeProjectQualifiedResourceImportID(tfNode) THEN all attributes in state match after import with no RequiredDuringImport errors
