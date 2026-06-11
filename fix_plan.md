# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN TF_ACC=1, AZDO_ORG_SERVICE_URL, and AZDO_PERSONAL_ACCESS_TOKEN are set WHEN TestAccDataReleaseDefinitionRevision_Basic runs against real ADO THEN the test creates a release definition, reads back revision 1 via data.betterado_release_definition_revision, confirms json_content is non-empty, and an idempotency re-plan produces no diff
- [x] AC2: GIVEN TF_ACC=1, AZDO_ORG_SERVICE_URL, and AZDO_PERSONAL_ACCESS_TOKEN are set WHEN TestAccDataReleaseDefinitionHistory_Basic runs against real ADO THEN the test creates a release definition, reads its full history via data.betterado_release_definition_history, confirms at least one revision entry exists with a non-empty revision number, and an idempotency re-plan produces no diff
