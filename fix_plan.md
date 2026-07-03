# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN azuredevops/internal/service/pipelinesapproval/ does not exist WHEN WI-3 is complete THEN the package exists and exports NewPipelineApprovalResource() resource.Resource implementing betterado_pipeline_approval with schema fields: id (computed string), project_id (required string ForceNew), approval_id (required string ForceNew), status (required string, OneOf approved/rejected), comment (optional string)
- [x] AC2: GIVEN betterado_pipeline_approval resource is applied with status=approved and a comment WHEN Create is called THEN UpdateApprovals is called with the approval decision and comment; the resource ID is set to approval_id; Read uses GetApproval to refresh status and comment
- [x] AC3: GIVEN betterado_pipeline_approval resource is deleted WHEN Delete is called THEN Delete is a no-op (approval decisions cannot be undone via API); the resource is simply removed from state
- [x] AC4: GIVEN TestPipelineApprovalResource_Metadata unit test exists WHEN go test -tags all -run TestPipelineApprovalResource ./azuredevops/internal/service/pipelinesapproval/ runs THEN the test passes confirming TypeName is betterado_pipeline_approval and the schema has the expected attributes
