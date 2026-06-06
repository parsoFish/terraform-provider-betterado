# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN TF_ACC=1, AZDO_ORG_SERVICE_URL, and AZDO_PERSONAL_ACCESS_TOKEN are set in the environment WHEN a test calls SharedReleaseFixture(t) from the new shared_fixtures.go helper THEN it provisions a real ADO project, a Git repo, a build definition, a variable group, and a canonical multi-stage release definition — and registers a t.Cleanup that destroys all of them so no orphaned cloud resources remain
- [x] AC2: GIVEN the fixture provisions the canonical release definition WHEN the returned release definition struct is inspected THEN every stage has BOTH a pre_deploy_approval AND a post_deploy_approval (satisfying VS402877), every stage has a retention_policy block (satisfying VS402982), and all approval/permission keys use the current ADO API schema (EditReleaseEnvironment, not the stale EditReleaseStage)
- [x] AC3: GIVEN TF_ACC is NOT set WHEN a test calling SharedReleaseFixture(t) is compiled and executed THEN the test is skipped (t.Skip) before any ADO API call is made, so the offline unit suite remains creds-free
