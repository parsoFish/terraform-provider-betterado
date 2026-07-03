# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a Terraform config creating a betterado_group_membership with group, mode=overwrite, and members WHEN terraform apply runs via the muxed provider THEN the memberships are set in ADO, the provider read-back populates members, and the idempotency re-plan shows no changes (ExpectNonEmptyPlan: false)
- [x] AC2: GIVEN betterado_group_membership is registered ONLY in framework_provider.go Resources() WHEN the provider compiles and terraform apply runs THEN no 'Duplicate resource type betterado_group_membership' error occurs and provider.go ResourcesMap no longer contains 'betterado_group_membership'
- [x] AC3: GIVEN mode changes from 'overwrite' to 'add' WHEN terraform apply runs THEN the update path is exercised and the idempotency re-plan is clean

## Implementation summary

### Iteration 1 (commit 03008ebc) — initial framework migration

All 3 ACs addressed:

- `resource_group_membership_framework.go` — new framework `resource.Resource`:
  - Reuses helpers from `resource_group_membership.go` (applyMembershipUpdate, addMembers, removeMembers, getGroupMemberships, expandGroupMembersFromList, toStringSet — all in same `graph` package)
  - Uses plan modifiers from `resource_group_framework.go`: `groupRequiresReplace()` + `groupUseStateForUnknown()` (same package, no extra imports)
  - Implements `staticStringDefault` for mode default "add" (vendor has `defaults.String` interface, not a pre-built `stringdefault` subpackage)
  - ID = group descriptor (stable, deterministic)
  - `waitForMembershipSync()` with configurable `continuousOccurrences`
- `provider.go` — removed `betterado_group_membership` from SDKv2 `ResourcesMap`
- `framework_provider.go` — added `graph.NewGroupMembershipResource` to `Resources()`
- `resource_group_membership_test.go` — `TestAccGroupMembership_Framework` with 4 steps:
  1. Create overwrite + 1 member, assert read-back (AC1 create)
  2. Idempotency re-plan (AC1 idempotency, ExpectNonEmptyPlan: false)
  3. Update mode overwrite→add (AC3 update path)
  4. Idempotency after mode update (AC3 idempotency, ExpectNonEmptyPlan: false)
- `provider_test.go` — removed `betterado_group_membership` from SDKv2 expectedResources count list

### Iteration 2 (commit 292d42fb) — fix description null/empty inconsistency in betterado_group

Gate failure: `betterado_group.test` and `betterado_group.member` failed with:
  `.description: was null, but now cty.StringVal("")`

Root cause: `description` is `Optional` (not Computed). When unset, plan is `null`.
ADO returns `""` after create. `readGroupIntoModel` was setting `types.StringValue("")`
in state → plan vs state mismatch.

Fix in `resource_group_framework.go` `readGroupIntoModel`: treat ADO `""` as null
when the user hasn't set description; only write non-empty ADO value, or preserve
the user's explicit setting (including `""`).

## Remaining

Gate command requires `TF_ACC=1` + live ADO credentials — live forge gate confirms.
