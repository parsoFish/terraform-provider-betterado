# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN TestAccReleaseDefinition_complete does not exercise environment_trigger, schedule, properties (PR #18) or cd_artifact_trigger.tag_filter, use_build_definition_branch, create_release_on_build_tagging, source_repo_trigger (PR #19) WHEN a new TestAccReleaseDefinition_completeWithNewFields test is added that includes all of those fields in a single combined exhaustive configuration THEN the test passes live (TF_ACC=1), all new-field assertions succeed, and the idempotency step (PlanOnly: true, ExpectNonEmptyPlan: false) produces no diff
- [x] AC2: GIVEN individual tests TestAccReleaseDefinition_environmentConfig and TestAccReleaseDefinition_triggerEnhancements already cover the new fields in isolation WHEN TestAccReleaseDefinition_completeWithNewFields combines them in one resource block alongside the existing complete-test features (gates, parallel_execution, real queue, agent_specification, etc.) THEN no assertion is duplicated verbatim — the combined test adds value by proving field interactions do not cause drift or unexpected API behaviour
