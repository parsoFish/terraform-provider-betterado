---
title: Environment-level secret variables produce a perpetual diff
description: flattenVariables preserves secret values from d.GetOk("variable") — the DEFINITION-level state path — for every call site, including the per-environment call in flattenEnvironments. Env-scoped secrets therefore never recover their state value (ADO masks them to null) and re-plan shows a perpetual diff. Found standing up the 2026-06-12 live demo.
category: defect
created_at: 2026-06-12T15:00:00Z
updated_at: 2026-06-12T15:00:00Z
---

# Environment-level secret variables — perpetual diff

## Observation

While standing up the 2026-06-12 demo (`demo-standing-2026-06-12/`), an
environment-scoped `variable { is_secret = true }` produced a non-empty plan on
every re-plan. Definition-level secrets behave correctly.

## Root cause

`resource_release_definition.go`:

- `flattenVariables(def.Variables, d)` (line ~1749, definition level) and
  `envMap["variable"] = flattenVariables(env.Variables, d)` (line ~1895, per
  environment) share one helper.
- The helper's secret-preservation step reads `d.GetOk("variable")` — the
  top-level definition variable set — regardless of call site. ADO returns
  `null` for secret values, so env-scoped secrets find no state value to
  preserve and flatten to empty → diff against config forever.

## Fix direction

Thread the state path (or the prior env's variable set) into the helper: the
environment call site must preserve from
`environment.<i>.variable`, not `variable`. Add a live acceptance test with an
env-scoped secret + `ExpectNonEmptyPlan: false`.

## Sources

- `demo-standing-2026-06-12/DEMO.md` (demo descope note)
- `azuredevops/internal/service/release/resource_release_definition.go` lines ~1778–1800
