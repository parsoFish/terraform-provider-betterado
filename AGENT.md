# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2025 — complete)

All four ACs completed in a single iteration.

**Changes made:**

`resource_release_definition.go`:
- Line 115: schema map key `"environment"` → `"stages"`
- Line 1087: `d.GetOk("environment")` → `d.GetOk("stages")`
- Line 1088: `expandEnvironments(...)` call → `expandStages(...)`
- Line 1140: function `expandEnvironments` → `expandStages`
- Line 1822: `d.Set("environment", flattenEnvironments(...))` → `d.Set("stages", flattenStages(...))`
- Line 1899: function `flattenEnvironments` → `flattenStages`
- Line 1966: format string `"environment.%d.variable"` → `"stages.%d.variable"`
- Line 2272: format string `"environment.%d.deploy_phase.%d.deployment_input.#"` → `"stages.%d.deploy_phase.%d.deployment_input.#"`

`resource_release_definition_test.go`:
- All 27 occurrences of `"environment": []interface{}{` → `"stages": []interface{}{`
- All `resourceData.Get("environment")` → `resourceData.Get("stages")`
- Comments updated: `"environment.<i>.variable"` → `"stages.<i>.variable"`, `"environment.0.variable state"` → `"stages.0.variable state"`

**Sub-field names left unchanged** (correct per WI spec):
- `environment_options`, `environment_trigger`, `definition_environment_id`, `environmentState`, `auto_triggered_and_previous_environment_approved_can_be_skipped`
- Internal variables like `envMap`, `envKey`, `envIdx`, `envOpts` (Go identifiers — not schema keys)
- API type names like `ReleaseDefinitionEnvironment`, `EnvironmentOptions`, etc.

**Quality gate result:**
- `go build ./azuredevops/internal/service/release/` — PASS
- `go test -tags all -count=1 -run TestReleaseDefinition_ ./azuredevops/internal/service/release/` — PASS (0.020s)

## What worked

- sed with explicit line numbers for targeted replacements (avoids false-positive matches on subfield names)
- Pattern `"environment": []interface{}` uniquely identifies top-level schema key usage vs subfield names like `"environment_options"`
- Running `go build` first (fast feedback), then `go test` (full gate)

## What didn't work

- Edit tool with tab-indented strings (tab vs space mismatch in old_string detection)
- sed with line-offset guesses that were off by one after earlier edits shifted line numbers

## Open questions

_(things that aren't blocking but would be useful to clarify)_

## Notes for reflection

- The rename was purely mechanical — no logic changes. All 4 ACs satisfied in iteration 0.
- WI-2 (ConfigMode), WI-3 (acceptance tests), WI-4 (examples/docs) are separate and untouched.
