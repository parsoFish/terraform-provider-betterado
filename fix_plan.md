# Fix Plan

> Checklist for WI-7. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN CHANGELOG.md has no entry for betterado_pipeline_approval WHEN WI-7 is complete THEN CHANGELOG.md ## [Unreleased] section contains an entry describing betterado_pipeline_approval resource and betterado_pipeline_approvals data source under ### FEATURES
- [x] AC2: GIVEN PROVIDER_VERSION.txt is checked WHEN WI-7 is complete THEN PROVIDER_VERSION.txt contains a bumped semver patch version (e.g. if current is 1.2.3, new value is 1.2.4)
- [x] AC3: GIVEN TestChangelog_HasPipelineApprovalEntry unit test exists WHEN go test -tags all -run TestChangelog_HasPipelineApprovalEntry ./azuredevops/internal/service/pipelinesapproval/ runs THEN the test passes (it reads CHANGELOG.md from the repo root and asserts the entry exists)
