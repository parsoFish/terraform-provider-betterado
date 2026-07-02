# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 — Framework migration of all 7 branch policy resources (COMPLETE)

**Created:**
- `framework_helpers.go` — shared scope helpers (`scopeModel`, `expandScopesFramework`, `flattenScopesFramework`), coerce helpers (`boolCoerce`, `int64Coerce`, `stringCoerce`), static defaults (`staticPolicyBool`, `staticPolicyInt64`, `staticPolicyString`), plan modifiers (`policyUseStateForUnknown`, `policyRequiresReplace`), import helper (`importPolicyState`)
- All 7 `*_framework.go` files: auto_reviewers, build_validation, comment_resolution, merge_types, min_reviewers, status_check, work_item_linking

**Modified:**
- `provider.go` — removed all 7 from ResourcesMap; removed branch import
- `framework_provider.go` — added branch import + 7 New*Resource constructors to Resources()
- `provider_test.go` — removed 7 branch policy resources from expectedResources list
- All 7 `resource_branchpolicy_*_test.go` — `Providers: testutils.GetProviders()` → `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`

### Iteration 2 — Fix settings/scope schema: ListNestedAttribute → ListNestedBlock (COMPLETE)

**Root cause:** The gate failure was:
```
Error: Missing required argument — The argument "settings" is required
Error: Unsupported block type — Blocks of type "settings" are not expected here
```

**The fix:** The acceptance test HCL uses block syntax `settings { scope { } }` but the framework schemas defined `settings` and `scope` as `schema.ListNestedAttribute` (which requires `= [{ }]` assignment syntax). 

In Terraform Plugin Framework:
- `schema.ListNestedAttribute` → HCL: `attribute = [{ key = val }]` 
- `schema.ListNestedBlock` → HCL: `block { key = val }`

Changed all 7 files:
1. Moved `"settings"` from `schema.Schema.Attributes` to `schema.Schema.Blocks`
2. Changed `schema.ListNestedAttribute{NestedObject: schema.NestedAttributeObject{...}}` → `schema.ListNestedBlock{NestedObject: schema.NestedBlockObject{Attributes: ..., Blocks: ...}}`
3. Moved `"scope"` from `NestedObject.Attributes` to `NestedObject.Blocks` (same treatment) 
4. Regular scalar attributes inside the blocks (reviewer_count, enabled, etc.) remain in `Attributes` — only the block-syntax ones move to `Blocks`

**Verified:**
- `go build -mod=vendor ./...` — PASS
- `go test ./azuredevops/ -run TestProvider_HasChildResources` — PASS

## What worked

- Inline default implementations in framework_helpers.go (same pattern as `release/framework_defaults.go`) — the sub-packages `booldefault`, `int64default`, `stringplanmodifier` are NOT vendored; only use `resource/schema/defaults` and `resource/schema/planmodifier` interfaces
- Re-using the existing SDKv2 policy settings structs (e.g. `autoReviewerPolicySettings`, `buildValidationPolicySettings`, `mergeTypePolicySettings`) from the same package for JSON unmarshal in flatten functions
- `boolPtr` helper function in min_reviewers_framework.go (in the same package, used across all 7 files)
- `importPolicyState` with `<project_id>/<policy_id>` format matching SDKv2's `tfhelper.ImportProjectQualifiedResourceInteger()`

## What didn't work

- `booldefault.StaticBool()`, `int64default.StaticInt64()`, `stringplanmodifier.UseStateForUnknown()` — NOT in vendor; compile fails with `-mod=vendor`
- `scopeSchemaAttributes()` returning `map[string]schema.Attribute` — unused, removed; schema attributes are inlined in each resource
- **`schema.ListNestedAttribute` for `settings`/`scope`** — requires `= [{ }]` HCL syntax; tests use block syntax `{ }`, so must use `schema.ListNestedBlock` instead

## Key Pattern: Block vs Attribute schema

```go
// WRONG — requires HCL: settings = [{ reviewer_count = 1 }]
"settings": schema.ListNestedAttribute{
    Required: true,
    NestedObject: schema.NestedAttributeObject{Attributes: ...},
}

// CORRECT — accepts HCL: settings { reviewer_count = 1 }
// Must go in schema.Schema.Blocks map, not Attributes map
"settings": schema.ListNestedBlock{
    NestedObject: schema.NestedBlockObject{
        Attributes: map[string]schema.Attribute{ /* scalar attrs */ },
        Blocks: map[string]schema.Block{
            "scope": schema.ListNestedBlock{ /* nested block */ },
        },
    },
}
```

## Notes for reflection

- Pattern: branch policy framework resources use `resource/schema/defaults` interfaces directly with inline implementations (not vendored sub-packages) — matches how release/taskagent packages do it
- `NestedBlockObject` has both `Attributes` (for scalar fields) and `Blocks` (for nested block fields)
- `schema.ListNestedBlock` does NOT support `Required`, `Optional`, `Computed` — those only apply to attributes
- The state model struct (`types.List` field + `ElementsAs`) is identical for both Attribute and Block patterns; only the schema declaration differs
