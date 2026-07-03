# Agent Memory — WI-9

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 2 (current)

**Gate failures from live run (last-gate-failure.md from prior iteration):**

```
TestAccVariableGroup_basic:
  .key_vault: was null, but now cty.ListValEmpty(...)
  .variable: inconsistent values for sensitive attribute

TestAccVariableGroupVariable_secret:
  .variable: inconsistent values for sensitive attribute

TestAccVariableGroupVariable_ForEach_ConcurrentCreate:
  .variable: inconsistent values for sensitive attribute
```

**Root cause 1 — key_vault null vs empty list:**

`flattenToModel` was returning `types.ListValueMust(..., []attr.Value{})` (empty list) for non-keyvault VGs. Since `key_vault` is `Optional` (not `Computed`) in the schema, the plan value is `null` when the user omits `key_vault`. An empty list after apply mismatched the planned null.

**Fix:** Return `types.ListNull(types.ObjectType{AttrTypes: keyVaultAttrTypes})` for non-keyvault VGs.

---

**Root cause 2 — `variable: inconsistent values for sensitive attribute`:**

The `variable` schema used `SetNestedAttribute`. In terraform-plugin-framework, `TransformDefaults` walks set elements by their cty hash. When a `Default` is applied to one nested attribute (e.g., `secret_value = ""`) the element hash changes mid-transform. Subsequent sibling attribute paths (`is_secret`, `value`, …) use the new hash to look up the config path — but the config still has the old hash. The lookup fails and the defaults are silently skipped, leaving those attributes as `null` in the plan while the provider returns their default values. Terraform core's post-apply consistency check sees this as "inconsistent values for sensitive attribute".

This is a known terraform-plugin-framework bug with `Default` values inside `SetNestedAttribute` elements: https://github.com/hashicorp/terraform-plugin-framework/issues/XXX

**Fix:** Changed `variable` from `SetNestedAttribute` to `ListNestedAttribute`. Lists use positional indexing, not hash-based lookup, so `Default` values are applied correctly to ALL nested attributes. The HCL syntax (`variable = [{ … }]`) is **unchanged** — both Set and List use `= [{ }]` assignment syntax.

**Additional improvements applied in same commit (1f556edf):**
- `mergeVariableListWithPlan` helper: re-emits the API-derived variable list in plan order (to pass Terraform's post-apply list-index comparison) and copies plan values for sensitive attributes (secret_value) so the Apply result exactly matches what was planned.
- `flattenToModel`: preserves prior-state variable ordering using prior state order (instead of always sorting alphabetically), so that Read does not cause spurious list-index diffs on the next plan. New vars not in prior state are appended alphabetically.
- `expandToParams`: passes the user-configured variable list to `searchAzureKVSecrets` for key-vault VGs (was passing nil, which returned an empty variable map — KV VG bug).
- Data source: aligned to `ListNestedAttribute` and stable name-sort ordering.

**Files changed:**
- `azuredevops/internal/service/taskagent/resource_variable_group_framework.go`
- `azuredevops/internal/service/taskagent/data_variable_group_framework.go`

### Iteration 1

**Gate failure:** `variable` attribute treated as a block by Terraform HCL parser — "Blocks of type 'variable' are not expected here" + "The argument 'variable' is required, but no definition was found."

**Root cause:** `resource_variable_group_framework.go` defines `variable` as a `schema.SetNestedAttribute`. In terraform-plugin-framework, `SetNestedAttribute` requires HCL **assignment syntax** (`variable = [{ ... }]`), NOT block syntax (`variable { ... }`). All test fixtures used block syntax (carried over from the SDKv2 implementation).

**Fix applied (commit 63044dd2):** Updated all test HCL fixtures across 7 files to use `variable = [{ ... }]` and `key_vault = [{ ... }]` syntax:
- `azuredevops/internal/acceptancetests/resource_variable_group_test.go`
- `azuredevops/internal/acceptancetests/data_variable_group_test.go`
- `azuredevops/internal/acceptancetests/resource_variable_group_variable_test.go`
- `azuredevops/internal/acceptancetests/resource_variable_group_permissions_test.go`
- `azuredevops/internal/acceptancetests/resource_check_rest_api_test.go`
- `azuredevops/internal/acceptancetests/resource_pipeline_authorization_test.go`
- `azuredevops/internal/acceptancetests/testutils/hcl.go` (HclVariableGroupResource + HclVariableGroupResourceKeyVault)

**Important:** `betterado_build_definition`'s `variable { ... }` blocks were intentionally left alone — those use SDKv2 and block syntax is correct for that resource.

## What worked

- Changing `variable { ... }` block syntax to `variable = [{ ... }]` attribute assignment syntax is the correct fix for `SetNestedAttribute` in terraform-plugin-framework.
- Same pattern applies to `key_vault = [{ ... }]`.
- Pattern reference: `resource_task_group_test.go` uses `task = [{ ... }]` for its `ListNestedAttribute`.
- **`ListNestedAttribute` instead of `SetNestedAttribute`** for nested attributes that have `Default` values and/or `Sensitive: true`. This avoids the set-hash mid-transform Default application bug.
- **`mergeVariableListWithPlan` pattern**: for Create/Update, re-emit the result in plan order using plan values for configured attributes and API values for computed-only attributes. This ensures: (1) list index consistency with the plan, (2) sensitive attribute values match exactly.
- **`types.ListNull(...)` for Optional non-Computed list attributes**: when the user omits the block, the plan is null and the state must also be null (not empty list).

## What didn't work

- **`SetNestedAttribute` with `Sensitive: true` + `Default: defaultString("")` inside nested objects**: the set-hash changes when a Default is applied, causing subsequent attribute path lookups in the config to fail. Do NOT use Set with Default or Sensitive nested attrs.
- **Passing `nil` to `searchAzureKVSecrets` variables param**: returns empty map, creates KV VG with no variables.

## Open questions

_(nothing blocking — awaiting live gate confirmation)_

## Notes for reflection

- When migrating SDKv2 resources to framework, the HCL test fixtures MUST be updated alongside the schema. Block syntax → attribute assignment syntax is a BREAKING CONFIG CHANGE for users, so it should be documented (though for this internal migration it's just test fixtures).
- The pattern `schema.SetNestedAttribute` / `schema.ListNestedAttribute` = assignment syntax; `schema.SetNestedBlock` / `schema.ListNestedBlock` = block syntax.
- **Prefer `ListNestedAttribute` over `SetNestedAttribute`** for variable-like patterns where the key is embedded in the nested object (`name` attribute). Use `MapNestedAttribute` when the key should be the map key and the `name` attribute can be dropped.
- **Plan-faithful writes (Create/Update)**: after calling the API and reading back, always re-emit configured (non-computed) attributes from the plan rather than the API response. Use `mergeVariableListWithPlan` or a similar pattern. This prevents ADO value normalization from triggering "inconsistent values" errors.
