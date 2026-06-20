# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN hclTaskGroupDataSourceBasic uses the framework array HCL syntax for the betterado_task_group resource block (matching WI-3's hclTaskGroupBasic update) WHEN TF_ACC=1 go test -tags all -run TestAccTaskGroupDataSource_basic ./azuredevops/internal/acceptancetests/ runs against a live ADO org THEN all test steps pass: resource.TestCheckResourceAttrPair assertions hold for name, description, category; idempotency step (PlanOnly:true ExpectNonEmptyPlan:false) shows no diff; evidence captured to .forge/live-evidence/task-group-datasource-acceptance.json
- [x] AC2: GIVEN make docs runs (tfplugindocs) and git checkout -- docs/guides/ restores hand-written guides WHEN docs/resources/betterado_task_group.md is inspected THEN it documents the task, input, and version attributes as list-of-object (framework attribute) format and examples/resources/betterado_task_group/resource.tf uses array-of-objects HCL syntax
- [x] AC3: GIVEN CHANGELOG.md and PROVIDER_VERSION.txt are updated WHEN cat CHANGELOG.md | grep Unreleased and cat PROVIDER_VERSION.txt are inspected THEN CHANGELOG.md has a DRAFT entry under ## Unreleased describing the betterado_task_group migration to terraform-plugin-framework; PROVIDER_VERSION.txt is bumped by a minor version
