# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete)

Created full framework migration of `betterado_group_membership` in a single iteration:

1. Created `azuredevops/internal/service/graph/resource_group_membership_framework.go` — full `resource.Resource` implementation.
2. Removed `betterado_group_membership` from `azuredevops/provider.go` SDKv2 `ResourcesMap`.
3. Added `graph.NewGroupMembershipResource` to `azuredevops/internal/provider/framework_provider.go` `Resources()`.
4. Added `TestAccGroupMembership_Framework` to `azuredevops/internal/acceptancetests/resource_group_membership_test.go`.
5. Removed `betterado_group_membership` from `azuredevops/provider_test.go` `TestProvider_HasChildResources` expected list.
6. All offline gates pass: `make test`, `golangci-lint run --new-from-rev=main`, `make terrafmt-check`.

## What worked

### Vendor has `defaults.String` interface but NOT `stringdefault` package

The vendor directory has `vendor/github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults/` with `string.go` defining the `String` interface, but there is no pre-built `stringdefault` or `stringplanmodifier` subpackage (unlike the upstream repo's external contrib packages). The solution: implement `staticStringDefault` struct locally (same pattern used throughout this provider's framework resources).

### Reuse plan modifiers from same package

`resource_group_framework.go` defines `groupRequiresReplaceModifier` and `groupUseStateForUnknownModifier` — both exported as functions `groupRequiresReplace()` and `groupUseStateForUnknown()`. Since `resource_group_membership_framework.go` is in the same `graph` package, we can call these directly without re-implementing them.

### Reuse CRUD helpers from resource_group_membership.go

The SDKv2 file's package-level helpers (`applyMembershipUpdate`, `addMembers`, `removeMembers`, `getGroupMemberships`, `expandGroupMembersFromList`, `toStringSet`) are all accessible since they're in the same `graph` package. `expandGroupMembersFromList` is defined in `resource_group_framework.go` and `toStringSet` is also there.

### TestProvider_HasChildResources count check

The `provider_test.go` has a `require.Equal(len(expectedResources), len(resources))` count check. Removing a resource from `provider.go` without removing it from `provider_test.go` fails this test. Pattern: add a comment line like `// betterado_group_membership is now a framework resource` and remove the entry from the string slice.

### ADO project-cap workaround

The ADO org is at the 1000-project cap. Use the shared persistent project `SharedFixtureProjectName` ("betterado-standing-demo") via a data source lookup instead of creating a new project in the test HCL.

### Gate warning about [no tests to run]

The initial gate failure was `[no tests to run]` because `TestAccGroupMembership_Framework` didn't exist yet. After adding the test function, the gate recognises it. The live forge gate runs with `TF_ACC=1` — the offline test prints `ok ... 0.00s [no tests to run]` which is NOT a pass.

### Iteration 2 (complete) — fix `description` null/empty inconsistency

**Gate failure:** Both `betterado_group.test` and `betterado_group.member` failed:
  `.description: was null, but now cty.StringVal("")`

**Root cause:** `description` attribute is `Optional` only (NOT `Computed`) in the group schema.
- When user doesn't set `description`, the plan value is `null`.
- ADO's `CreateGroupVsts` always returns a group with `Description = ""` (empty string pointer).
- `readGroupIntoModel` was writing `types.StringValue("")` into state when `grp.Description != nil`.
- Framework detected plan (`null`) ≠ state (`""`) → "Provider produced inconsistent result after apply".

**Fix (commit 292d42fb):** In `readGroupIntoModel`, treat ADO `""` description as null when the
user hasn't set it. Only write a non-empty ADO description value. If the user explicitly set
`description` (even to `""`), preserve their current model value.

Pattern for `Optional`-only attributes that ADO normalises to `""`:
```go
switch {
case grp.SomeField != nil && *grp.SomeField != "":
    m.SomeField = types.StringValue(*grp.SomeField)
case !current.SomeField.IsNull() && !current.SomeField.IsUnknown():
    m.SomeField = current.SomeField  // preserve user's explicit value
default:
    m.SomeField = types.StringNull()
}
```

## What didn't work

_(none — iteration 2 fix was first attempt)_

## Open questions

_(none)_

## Notes for reflection

- Pattern confirmed: the `defaults` package in the vendor has only the interface definitions; the concrete implementations (e.g. `stringdefault.StaticString`) need to be implemented locally unless the contrib packages are vendored.
- The `provider_test.go` `TestProvider_HasChildResources` count assertion is a strict gate for SDKv2 resource migration — always update it when moving resources to framework.
- **Key pattern for ADO `Optional`-only attributes:** ADO often returns `""` for unset string fields. If an attribute is `Optional` (not `Computed`), the plan is `null` when unset, so the Read must NOT return `""` — it must return `null`. The switch pattern above is the correct idiom.
