# Agent Memory — WI-8

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (this iteration)

**Root cause identified and fixed:**

The gate failure was:
- `Blocks of type "condition" are not expected here` — TF config uses HCL block syntax `condition { ... }` / `action { ... }`
- `The argument "condition" is required, but no definition was found` — TF couldn't find the attribute

**The fix:**
- `schema.SetNestedAttribute` requires **assignment syntax**: `condition = [{ condition_type = "when" }]`
- `schema.SetNestedBlock` accepts **block syntax**: `condition { condition_type = "when" }`

The tests always used HCL block syntax. The initial migration incorrectly used `SetNestedAttribute` (in the `Attributes:` map). Fixed by:
1. Removing `condition` and `action` from `Attributes:` map
2. Adding them to `Blocks:` map (new field) as `schema.SetNestedBlock` with `NestedObject: schema.NestedBlockObject{...}`

The Go model type (`types.Set`) is unchanged — both `SetNestedAttribute` and `SetNestedBlock` use `types.Set` in Go.

**Commit:** `e06aab96` — fix: use SetNestedBlock for condition/action to support HCL block syntax

**AC2 and AC3** were already complete in the prior commit (`2bedf727`) which:
- Removed SDKv2 `resource_rule.go`
- Deregistered from `provider.go`, registered in `framework_provider.go`
- Updated `provider_test.go` resource count
- Added `captureRuleEvidence` / `testutils.CaptureLiveEvidence` in `TestAccWorkitemtrackingprocessRule_Basic`

## What worked

- `schema.SetNestedBlock` in `Blocks:` map = HCL block syntax support (no `=` sign)
- `schema.SetNestedAttribute` in `Attributes:` map = HCL assignment syntax (requires `=`)
- The framework documentation comment on `SetNestedBlock` explicitly says: "Terraform configurations configure this block repeatedly using curly brace syntax without an equals (=) sign"

## What didn't work

- Using `SetNestedAttribute` for schemas where tests use HCL block syntax — generates "Blocks of type X are not expected here" and "Missing required argument" errors

## Open questions

_(none)_

## Notes for reflection

- When migrating SDKv2 resources that use `Schema: map[string]*schema.Schema{"condition": {Type: schema.TypeSet, Elem: &schema.Resource{...}}}`, the SDKv2 TypeSet-of-Resource pattern translates to `schema.SetNestedBlock` in the framework (NOT `SetNestedAttribute`). Nested attributes are a framework-only concept and use assignment syntax.
