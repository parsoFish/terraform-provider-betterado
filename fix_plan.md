# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN no project-scoped token test exists WHEN TestReleaseDefinitionPermissions_ProjectScopedToken runs THEN the function createReleaseDefinitionToken returns only the projectID (no slash suffix) when release_definition_id is not set in schema data
- [x] AC2: GIVEN no edge-case tests exist for the token function WHEN TestReleaseDefinitionPermissions_TokenEdgeCases runs THEN a definition-scoped token with definitionID=0 still formats as projectId/0 (not project-only path)
