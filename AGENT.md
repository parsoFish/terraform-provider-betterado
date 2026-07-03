# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 2 (current)

**Root cause of gate failure (iter 1):** `hclGroupFramework` created a `resource "betterado_project"` in the HCL config. The ADO org is at the 1000-project cap, so every `terraform apply` failed at Step 1/2 with "organization already has 1000 projects".

**Fix applied:** Changed `hclGroupFramework(projectName, groupName string)` → `hclGroupFramework(groupName string)`. Now uses `data "betterado_project" "shared" { name = "betterado-standing-demo" }` (the persistent shared project = `SharedFixtureProjectName`) as the group scope. No new project is created.

**Result:** Offline build + make test: all pass. golangci-lint --new-from-rev=main: 0 issues. Committed as `fix(test): use persistent shared project in TestAccGroupResource_Framework`.

## What worked

- **Persistent project pattern:** Using `data "betterado_project"` to look up `SharedFixtureProjectName` ("betterado-standing-demo") avoids any project creation. This is the mandatory pattern for all acceptance tests in this org — see `shared_fixtures.go`:`SharedFixtureProjectName` comment.
- The `%[2]q` / `%[1]q` format specifiers with `fmt.Sprintf` correctly quote the Go strings into HCL without needing manual escaping.

## What didn't work

- **Creating new projects in HCL configs:** The org is at 1000-project cap. Resource `"betterado_project"` blocks in acceptance test HCL always fail with "organization already has 1000 projects". NEVER use `resource "betterado_project"` in acceptance test HCL — always use `data "betterado_project"` with `SharedFixtureProjectName`.

## Current state of the implementation (prior iterations)

Based on `git diff --stat main..HEAD`:
- `resource_group_framework.go`: Framework provider resource — full CRUD implementation (630 lines added)
- `framework_provider.go`: `NewGroupResource` registered in `Resources()`
- `provider.go`: `betterado_group` removed from SDKv2 `ResourcesMap` (no duplicate)
- `resource_group_test.go`: `TestAccGroupResource_Framework` + helper functions
- `shared_fixtures.go`: `groupGetDirectClient` helper + related functions

The live acceptance gate (`TestAccGroupResource_Framework`) will exercise all three ACs when it runs. The only blocker was the project creation issue — now fixed.

## Open questions

- Does the `betterado_project` data source return an `id` attribute that maps to the project UUID? (Expected yes — it's a standard SDKv2 resource and the data source schema has standard computed fields.)

## Notes for reflection

- The ADO org's 1000-project cap is a hard constraint for ALL acceptance tests. Any HCL that includes `resource "betterado_project"` must be rewritten to use `data "betterado_project"` with `SharedFixtureProjectName`. This should be captured as a brain theme / lint rule.
