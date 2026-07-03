# Fix Plan

> Checklist for UWI-6. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the real capture at .forge/live-evidence/dashboard-acceptance-resource.json WHEN the demo artifact is regenerated and committed THEN the committed forge/history .../demo/demo.json carries BOTH the dashboard and extension liveEvidence checkpoints with their real URLs (no null, no unrelated vsrm citation)
  - forge/history/INIT-2026-07-01-migrate-framework-dashboard-extension/demo/demo.json has:
    - `dashboard-acceptance-resource` checkpoint with url=dev.azure.com/.../dashboards/26cd93be-3f74-4846-a25d-13e21ba67443 (capturedAt 2026-07-03T07:51:19Z)
    - `acceptance-resource` checkpoint with url=extmgmt.dev.azure.com/.../ms-securitydevops/microsoft-security-devops-azdevops (capturedAt 2026-07-02T09:52:39Z)
  - Already committed in prior iterations

- [x] AC2: GIVEN the PR conflicts with current main on provider registration fan-in files WHEN the branch is merged/rebased onto origin/main and pushed THEN gh pr view 45 shows MERGEABLE and gh pr checks 45 shows the four workflows green
  - Merged origin/main into branch, resolved conflicts in:
    - framework_provider.go (imports: added build+memberentitlementmanagement alongside dashboard+extension)
    - provider.go (removed dashboard+extension from SDKv2 AND kept build/entitlement/pipeline_auth removed too)
    - CHANGELOG.md (kept Unreleased dashboard+extension, added main's 1.3.0 block)
    - PROVIDER_VERSION.txt (bumped to 1.3.1)
    - .forge/project.json (took main's version)
  - Fixed gofmt alignment issue in provider.go
  - Pushed; gh pr view 45 shows MERGEABLE, mergeStateStatus CLEAN
  - All 4 checks green: depscheck, go-lint, terrafmt, test
