# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_pipeline_approval is not registered in framework_provider.go WHEN WI-5 is complete THEN framework_provider.go Resources() slice includes pipelinesapproval.NewPipelineApprovalResource and DataSources() slice includes pipelinesapproval.NewPipelineApprovalsDataSource
- [x] AC2: GIVEN provider.go (SDKv2) is checked after WI-5 WHEN grep 'pipeline_approval' azuredevops/provider.go is run THEN it produces no output — zero SDKv2 registrations for these types (framework-only per AC-4)
- [x] AC3: GIVEN the provider is compiled after WI-5 WHEN go build -mod=vendor . is run from the worktree root THEN it exits 0 — both new types are importable and the mux wiring compiles
- [x] AC4: GIVEN TestFrameworkProvider_HasPipelineApprovalResources unit test exists WHEN go test -tags all -run TestFrameworkProvider_HasPipelineApprovalResources ./azuredevops/internal/provider/ runs THEN the test passes confirming the provider exposes betterado_pipeline_approval resource and betterado_pipeline_approvals data source
