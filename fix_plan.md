# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [ ] AC1: GIVEN a Terraform config creating a betterado_group with display_name, description, scope, and members WHEN terraform apply runs via the muxed provider THEN the group is created in ADO, the provider read-back populates descriptor, origin, principal_name, domain, subject_kind, url, group_id, and the idempotency re-plan shows no changes (ExpectNonEmptyPlan: false)
  - [x] Test infrastructure: hclGroupFramework now uses persistent project data source (no project create → no 1000-cap failure)
  - [ ] Live gate: TestAccGroupResource_Framework must pass end-to-end (create → read-back → idempotency re-plan → destroy)
- [ ] AC2: GIVEN betterado_group is registered ONLY in framework_provider.go Resources() WHEN the provider compiles and terraform apply runs THEN no 'Duplicate resource type betterado_group' error occurs and the SDKv2 provider.go ResourcesMap no longer contains 'betterado_group'
  - [x] provider.go: betterado_group removed from ResourcesMap (done in prior iteration)
  - [x] framework_provider.go: NewGroupResource registered (done in prior iteration)
- [ ] AC3: GIVEN a betterado_group resource is destroyed WHEN terraform destroy runs THEN the group is deleted from ADO and the provider returns clean (no 404 error — treat 404 as already deleted)
  - [x] resource_group_framework.go Delete(): 404 treated as already deleted (done in prior iteration)
  - [x] checkGroupDestroyedFramework: gracefully handles 404 (done in prior iteration)
