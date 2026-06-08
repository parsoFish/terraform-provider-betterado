# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a cd_artifact_trigger block with a tag_filter block containing pattern and tags fields WHEN expandTriggers processes the block THEN the emitted triggerConditions entry contains tagFilter.pattern and tags slice
- [x] AC2: GIVEN an ADO API response with triggerConditions entries carrying tagFilter.pattern and tags WHEN flattenTriggers processes the triggers THEN the Terraform state contains a tag_filter block with pattern and tags set correctly
- [x] AC3: GIVEN a cd_artifact_trigger block with use_build_definition_branch = true WHEN expandTriggers processes it THEN the emitted triggerConditions entry carries useBuildDefinitionBranch: true
- [x] AC4: GIVEN an ADO API response with useBuildDefinitionBranch in triggerConditions WHEN flattenTriggers processes it THEN the Terraform state reflects use_build_definition_branch = true
- [x] AC5: GIVEN a cd_artifact_trigger block with create_release_on_build_tagging = true WHEN expandTriggers processes it THEN the emitted triggerConditions entry carries createReleaseOnBuildTagging: true
- [x] AC6: GIVEN an ADO API response with createReleaseOnBuildTagging in triggerConditions WHEN flattenTriggers processes it THEN the Terraform state reflects create_release_on_build_tagging = true
- [x] AC7: Tests written — TestReleaseDefinition_ArtifactTagFilter_RoundTrip and TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip both PASS
