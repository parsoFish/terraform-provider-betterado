# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN docs/pipelinesapproval-gap-matrix.md does not exist WHEN WI-1 is complete THEN docs/pipelinesapproval-gap-matrix.md exists and lists every ADO Pipelines Approval API field (GetApproval, QueryApprovals, UpdateApprovals), distinguishes declarative-manageable state (approval decisions via UpdateApprovals) from ephemeral-only operations, and notes betterado resource coverage for each field
- [x] AC2: GIVEN the gap matrix is opened WHEN it is read THEN it documents that betterado_pipeline_approval manages the approval decision (approve/reject + comment) and betterado_pipeline_approvals lists pending approvals, and explains that approval IDs are ephemeral (not importable) because they are bound to a specific pipeline run
