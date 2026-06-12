---
title: process_parameters does not reliably round-trip on basic ADO pipeline definitions
description: ADO silently ignores ProcessParameters on non-task-group release definitions; unit tests pass but live acceptance cannot prove idempotency for this field.
category: decision
created_at: 2026-06-11
updated_at: 2026-06-11
---

## What was decided

AC-3 of INIT-2026-06-08 required a live acceptance test creating a release definition with `process_parameters` configured and proving `ExpectNonEmptyPlan: false`. During implementation the agent discovered ADO does not reliably store `ProcessParameters` on basic pipeline definitions — the field is primarily consumed by task-group template inheritance (where the task-group definition carries the parameter schema and the release definition references it).

**Decision**: `process_parameters` is excluded from `TestAccReleaseDefinition_environmentConfig`. Unit test `TestReleaseDefinition_ProcessParameters_RoundTrip` covers the expand/flatten code path. The acceptance test rationale is documented inline in the HCL helper comment.

## Implication for future acceptance tests

When writing acceptance tests that include `process_parameters`:
- Use a release definition that references a task group (via `task_group_id` in the deploy_phase) — only then does ADO store and return the parameter values.
- A standalone pipeline definition will produce drift even if the expand/flatten code is correct.

## Sources

- `_logs/2026-06-08T12-01-16_INIT-2026-06-08-release-definition-environment-config-surface/events.jsonl` — `gate.pass` WI-2 `iteration=0` (test omits process_parameters)
- `/home/parso/forge/brain/cycles/_raw/2026-06-08T12-01-16_INIT-2026-06-08-release-definition-environment-config-surface.md`
- `projects/terraform-provider-betterado/azuredevops/internal/acceptancetests/resource_release_definition_test.go` — `hclReleaseDefinitionEnvironmentConfig` inline comment
