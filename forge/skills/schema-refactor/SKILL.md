# Schema Refactor Skill

## Purpose

Rename a schema field and/or convert a nested **block** collection into an
assignable **list-of-object attribute**, so the generated Terraform reads
naturally (`stages = [ { … }, { … } ]` instead of repeated `stages { … }`
blocks) and can be built with HCL `for` / `concat` / `merge` expressions. The
canonical use is renaming release_definition `environment` → `stages` and making
stages (and other nested collections) array-assignable.

## When to use

- Renaming a user-facing schema field to a friendlier name (no back-compat alias).
- Converting a `TypeList`/`TypeSet` of `*schema.Resource` from block syntax to
  attribute (array) syntax for readability.

## Background: block vs attribute (SDK v2)

A nested `TypeList`/`TypeSet` whose `Elem` is a `*schema.Resource` is, by default,
configured in HCL as repeated **blocks**:
```hcl
stages { name = "Prod" }
stages { name = "QA" }
```
Add `ConfigMode: schema.SchemaConfigModeAttr` to the schema entry and the same
collection becomes an **attribute** assignable as an array of objects:
```hcl
stages = [
  { name = "Prod" /* … */ },
  { name = "QA"   /* … */ },
]
```
The expand/flatten code is unchanged in shape (it still reads
`d.Get("stages").([]interface{})` of `map[string]interface{}`); only the HCL
surface and a few schema constraints change.

## Workflow

### 1. Rename the field (if renaming)
- Change the schema map key (`"environment"` → `"stages"`) in `resource_<name>.go`.
- Update every `d.Get("environment")` / `d.Set("environment", …)` / `d.GetOk(...)`
  and any nested references to the new key. Grep the whole package for the old key.
- Rename the expand/flatten helpers for clarity (`expandEnvironments` →
  `expandStages`) and their call sites.
- Update examples (`examples/resources/<resource>/*.tf`), docs
  (`docs/resources/*` / templates), and acceptance-test HCL fixtures.
- This is a **breaking** change — no alias, no deprecation shim.

### 2. Convert blocks → attribute array
For each nested collection to convert, add to its schema entry:
```go
"stages": {
    Type:       schema.TypeList,
    ConfigMode: schema.SchemaConfigModeAttr,   // <- assignable as an array
    Required:   true,                          // or Optional
    Elem:       &schema.Resource{ Schema: map[string]*schema.Schema{ /* … */ } },
},
```
Apply the same to the inner collections you want array-style (e.g.
`deploy_phase`, `artifact`, `variable`) where it improves readability.

### 3. Honour the ConfigModeAttr constraints (the gotchas)
- A `SchemaConfigModeAttr` collection generally must be `Optional` or `Required`,
  **not** purely `Computed`. A nested element field that is `Optional + Computed`
  can warn/err under attr mode — make server-assigned fields (`id`, `rank` if
  server-owned) `Computed` only, and keep user-set fields `Optional`/`Required`.
- `MaxItems: 1` + `ConfigModeAttr` models a single assignable object
  (`x = { … }`); use it for the single-object blocks (retention_policy,
  execution_policy, approvals) if you want them attribute-style too.
- Don't mix block and attribute syntax for the same field — pick one.
- Re-run `terrafmt` after editing example/doc HCL; attribute arrays must be
  `terrafmt`-clean for the CI gate.

### 4. Prove it (two-gate + demo)
- Update unit tests for the renamed/reshaped schema; non-default values + read-back.
- Update / add a `TF_ACC` acceptance test that applies the array-style HCL against
  real ADO, asserts via provider read, runs an idempotency re-plan
  (`ExpectNonEmptyPlan: false`), and destroys cleanly.
- Capture a live-evidence demo (see the `ado-demo` skill): the array-style `.tf`
  plans + applies, then an API GET shows the created stages.
- Run the CI-equivalent gate (`make test && golangci-lint run ./... &&
  make terrafmt-check`) green.

## Definition of done

The renamed, array-assignable schema applies + round-trips against live ADO, is
idempotent, the CI gate is green, examples + docs use the new array syntax, and
the plan + demo are recorded under `forge/history/<initiative-id>/`.
