---
title: Release definition schema gap matrix — 93 mapped, 1 partial, 39 missing across 12 ADO SDK types
description: The INIT-2026-06-08 audit found 93 mapped / 1 partial / 39 missing fields in betterado_release_definition vs ADO REST 7.2; all 39 missing top-level fields are read-only/computed/deprecated; the actionable gaps are in nested blocks (approvals, gates, environment options, triggers, task parameters).
category: reference
project: terraform-provider-betterado
created_at: 2026-06-08T12:00:00Z
updated_at: 2026-06-08T12:00:00Z
related_themes:
  - 2026-05-31-release-definition-unit-test-substrate
  - 2026-06-05-ado-silent-field-discard-idempotency
---

# Release definition gap matrix — key findings

## Summary counts (from docs/release-definition-gap-matrix.md)

| Layer | Mapped | Partial | Missing |
|---|---|---|---|
| ReleaseDefinition (top) | 9 | 1 (`tags`) | 12 (all read-only or deprecated) |
| ReleaseDefinitionEnvironment | ~25 | several | several (nested blocks) |
| **Total across 12 types** | **93** | **1** | **39** |

Of the 39 missing: the majority are read-only/computed metadata (`_links`, `createdBy`, `modifiedBy`, `url`, `source`, etc.). The `tags` partial is intentional (ADO silently discards tags on write — Computed field to prevent perpetual diff).

## Actionable gaps (implementation targets)

Prioritised in `docs/release-definition-roadmap.md`:

**P1 (required-for-create / high-demand):**
- `WI-A`: `retention_policy` acceptance tests (stale since ADO 7.1 required it — 2026-05-31-forge-onboarding-findings)
- `WI-B`: `post_deploy_approval` acceptance tests (same stale issue)
- `WI-C`: `workflowTask.timeoutInMinutes` + `retryCount` fields missing from task schema

**P2 (config-surface parity):**
- `WI-D`: `overrideInputs` on WorkflowTask (env-level variable overrides per task)
- `WI-E`: `environmentTriggers` (trigger on environment completion)
- `WI-F`: artifact trigger `tagFilter` + `createReleaseOnBuildTagging` fields

**P3 (rarely-used / complex):**
- `WI-I` through `WI-K`: containerImageTrigger, full data source parity, revision/history data sources

## Data source gaps

- `GetDefinitionRevision` (revision history): **Recommend** — not yet surfaced
- `GetReleaseDefinitionHistory` (audit trail): **Recommend** — not yet surfaced
- Runtime methods (GetRelease, GetApprovals, etc.): **Out-of-scope**

## Test coverage gaps

13 fields with no acceptance coverage. Known stale tests: `TestAccReleaseDefinition_retentionPolicy` (missing `retention_policy` block per ADO REST 7.1), `TestAccReleaseDefinition_approvals` (missing `pre_deploy_approval`).

## Sources

- `projects/terraform-provider-betterado/docs/release-definition-gap-matrix.md`
- `projects/terraform-provider-betterado/docs/release-definition-roadmap.md`
- `_logs/2026-06-08T11-00-43_INIT-2026-06-08-release-definition-schema-audit/events.jsonl` (WI-1 iteration event `EV_mq53yrl6_3lqwfuwa`)
- `brain/cycles/_raw/2026-06-08T11-00-43_INIT-2026-06-08-release-definition-schema-audit.md`
