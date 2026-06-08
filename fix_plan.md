# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the gap matrix in docs/release-definition-gap-matrix.md identifying all missing/partial fields WHEN fields are prioritised by (a) ADO 7.2 required-for-create, (b) operator config-surface parity, (c) complexity THEN docs/release-definition-roadmap.md exists and contains an ordered list of implementation work items (one per logical gap cluster), each with an estimated iteration budget calibrated against work-item-completion-by-domain data
- [x] AC2: GIVEN the ordered implementation work items WHEN schema additions gate test additions THEN docs/release-definition-roadmap.md contains explicit depends_on ordering between implementation work items where applicable
- [x] AC3: GIVEN the scope clarification from the initiative body (no runtime/imperative operations) WHEN the out-of-scope section is written THEN docs/release-definition-roadmap.md contains a clear out-of-scope section listing read-only/computed values and imperative runtime operations (CreateRelease, UpdateApproval, ManualInterventions, Deployments)
- [x] AC4: GIVEN azuredevops/internal/service/release/doc_audit_test.go already exists (created by WI-1) WHEN TestAuditRoadmapDocExists is appended to that file and go test -tags all -run TestAuditRoadmapDocExists ./azuredevops/internal/service/release/ is executed THEN the test passes (docs/release-definition-roadmap.md exists and has at least 20 lines)
