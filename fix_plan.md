# Fix Plan

> Checklist for UWI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: Add stringvalidator parity to framework data sources
  - `stringvalidator.RegexMatches(UUID pattern)` on `namespace_id` in data_security_namespace_token_framework.go
  - `stringvalidator.LengthAtLeast(1)` on `namespace_name` in data_security_namespace_token_framework.go
  - `stringvalidator.LengthAtLeast(1)` on `scope` in data_securityrole_definitions_framework.go
  - go build ./... green

- [x] AC2: Delete 19 superseded SDKv2 source files
  - Deleted 14 permissions SDKv2 files (including resource_release_definition_permissions.go)
  - Deleted 4 security SDKv2 files
  - Deleted 2 securityroles SDKv2 files
  - Shared helpers relocated to new files: namespace_token_helpers.go, security_helpers.go, permissions_helpers.go
  - go build ./... stays green

- [x] AC3: Fix stale c0ac3757/2026-07-02T12:44:24Z references in demo artifacts
  - Fixed both stale references in demo.json (lines 66 and 91)
  - Fixed both stale references in DEMO.md (lines 14 and 19)
  - All now reference 6ddb680c / 2026-07-03T02:14:27Z

- [ ] AC4: Live evidence for betterado_security_permissions AND betterado_securityrole_assignment
  - Requires TF_ACC=1 live test run against real Azure DevOps
  - Needs AZDO_ORG_SERVICE_URL + AZDO_PERSONAL_ACCESS_TOKEN env vars
  - BLOCKED: cannot run live tests without credentials

- [x] AC5: Vendor clean, committed, pushed
  - Vendor was already clean (go mod vendor produced no changes)
  - All AC1-AC3 changes committed and pushed
  - local HEAD == origin HEAD ✓
