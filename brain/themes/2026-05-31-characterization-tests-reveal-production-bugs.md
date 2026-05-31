---
title: Characterization tests reveal production type-handling bugs
slug: 2026-05-31-characterization-tests-reveal-production-bugs
description: WI-2's deep-nested environment test exposed a type-switch gap in expandWorkflowTask — the function expected string for inputs but Terraform SDK passes map[string]interface{}. Fix was a one-liner; signal was unmistakable.
category: pattern
project: terraform-provider-betterado
created_at: 2026-05-31T11:30:00Z
updated_at: 2026-05-31T11:30:00Z
related_themes:
  - 2026-05-31-release-definition-unit-test-substrate
  - 2026-05-18-stack-and-test-layout
---

# Characterization tests reveal production type-handling bugs

`TestReleaseDefinition_DeepNestedEnvironment_ExpandFlatten` (WI-2) is a pure characterization test — its intent is fidelity verification, not feature addition. Running it immediately exposed that `expandWorkflowTask` typed the `inputs` field as `string`, but the Terraform Plugin SDK delivers `map[string]interface{}` at runtime.

## Mechanism

The Terraform SDK passes resource data as `interface{}` through `schema.ResourceData.Get()`. When the resource schema defines `inputs` as `TypeMap`, the SDK gives `map[string]interface{}`, not a pre-cast string. The expand helper assumed the latter; the test fixture used the former (correctly matching real Terraform behaviour).

**Fix:** one-line type-switch in `expandWorkflowTask`:
```go
switch v := raw.(type) {
case string:
    // existing path
case map[string]interface{}:
    // new path: marshal to JSON string or handle directly
}
```

## Lesson

Scaffold the deep-nested expand test early in WI-2 (not as a late addition). It exercises the real Terraform SDK interface, not a simplified mock, and will surface type-handling gaps that only manifest at plan/apply time. The 5-test CRUD substrate alone would NOT have caught this.

## Sources

- `_logs/2026-05-31T10-57-52_INIT-2026-05-31-release-definition-unit-tests/events.jsonl` (unifier iteration 1, `fix(tests): handle map[string]interface{} inputs from Terraform SDK` commit in git log)
- `brain/cycles/_raw/2026-05-31T10-57-52_INIT-2026-05-31-release-definition-unit-tests.md`
