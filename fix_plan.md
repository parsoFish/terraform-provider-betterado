# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] Framework provider Configure() — wire AggregatedClient as ProviderData via AZDO env vars
- [x] testutils.GetMuxedProviderFactories() — build proto-v6 mux server for acceptance tests
- [x] hclTaskGroupBasic — update to array HCL syntax (task = [{...}], input = [{...}], version = [{...}])
- [x] hclTaskGroupWithGapFields — update to array HCL syntax
- [x] TestAccTaskGroup_basic — switch to ProtoV6ProviderFactories: GetMuxedProviderFactories()
- [x] TestAccTaskGroup_withGapFields — switch to ProtoV6ProviderFactories: GetMuxedProviderFactories()
- [x] checkTaskGroupDestroyed — switch from GetProvider().Meta() to getDirectClient() (env vars)
- [x] captureTaskGroupEvidence — switch from GetProvider().Meta() to getDirectClient() (env vars)
- [x] examples/resources/betterado_task_group/resource.tf — update to array HCL syntax
- [x] docs/resources/task_group.md — update example + schema to array HCL syntax + framework terminology
- [x] CHANGELOG.md — add DRAFT entry under ## Unreleased
- [x] golangci-lint (gofumpt) — fix trailing newline issue in resource_task_group_framework.go
- [x] terrafmt check — test file Terraform blocks are correctly formatted
- [x] Framework provider Schema() — mirror SDKv2 provider schema in framework provider (mux parity)
- [x] AC1: Live gate — TF_ACC=1 go test -tags all -run TestAccTaskGroup_basic PASSED (23.67s) + evidence captured to .forge/live-evidence/acceptance-resource.json
- [x] AC2: Idempotency — no perpetual diff on second plan — PASSED (Step 2/2 PlanOnly:true ExpectNonEmptyPlan:false)
- [x] AC3: TF_ACC=1 go test -tags all -run TestAccTaskGroup_withGapFields PASSED (23.40s) + idempotency
