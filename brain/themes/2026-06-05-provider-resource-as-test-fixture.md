---
title: Use provider resources as test fixtures in acceptance tests, not pre-created ADO objects
description: Acceptance tests that need ADO objects (e.g. a shared work-item query for gate tasks) should create them via provider resources in the test HCL, not rely on pre-existing ADO state — keeps tests self-contained and declarative.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-06T00:00:00Z
updated_at: 2026-06-06T00:00:00Z
related_themes:
  - 2026-05-31-forge-onboarding-findings
---

# Use provider resources as test fixtures in acceptance tests

## Pattern

When an acceptance test needs an ADO object (e.g. a shared work-item query to use as a gate task's `queryId`), create it via a provider resource in the test's HCL config — do NOT:
- Rely on a pre-existing ADO object (breaks if the org is reset or the object is deleted).
- Hard-code a static GUID (brittle, won't work in a fresh org).
- Skip the field (leaves the gate task incomplete, masks the real behaviour).

## Applied in WI-9

`TestAccReleaseDefinition_complete` needed a real `queryId` for the "Query Work Items" gate task. The test project had no saved queries. Solution:

```hcl
resource "betterado_workitemquery" "gate_query" {
  project_id = azuredevops_project.project.id
  name       = "All Work Items - Gate Check"
  wiql       = "SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project"
  path       = "Shared Queries"
}
```

Then reference: `queryId = betterado_workitemquery.gate_query.id`

This is self-contained, declarative, destroyed with the rest of the test state on `terraform destroy`, and works in any fresh org.

## Generalisation

Any acceptance test that requires an ADO dependency (agent queue ID, variable group, environment, query, etc.) should provision it declaratively in the HCL config via the corresponding provider resource. If no provider resource exists yet, that itself signals a missing resource to implement.

## Sources

- `_logs/2026-06-05T15-06-02_INIT-2026-06-05-complete-release-definition/work-items-snapshot/WI-9.md` (AC3 — real gate queryId)
- `_logs/2026-06-05T15-06-02_INIT-2026-06-05-complete-release-definition/pr-description.md` (WI-9 section)
- `brain/cycles/_raw/2026-06-05T15-06-02_INIT-2026-06-05-complete-release-definition.md`
