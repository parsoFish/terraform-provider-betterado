---
project: terraform-provider-betterado
updated_at: 2026-05-23T11:19:10Z
---

# terraform-provider-betterado Roadmap

> **North star (2026-05-31):** feature-complete representation of the ADO REST API
> 7.2, structured as **one initiative per API path** (area → feature → work item),
> then auto-synced as the API evolves via a coverage matrix. The model + the
> first-mover **Release** initiative are detailed in
> [`docs/api-coverage-roadmap.md`](./docs/api-coverage-roadmap.md). The phase below
> is the prior 7.1 createable-resource framing, now reframed under that north star
> (the listed initiatives become per-path features). Parked pending forge-hardening
> (forge `docs/known-gaps.md` §2026-05-31). The `task_group` unit substrate landed
> 2026-05-31; `release_definition` still needs unit tests + an acceptance refresh.

## Current phase

**Theme:** Full ADO REST API 7.2 coverage — one initiative per API path (north star above)
**Status:** active
**Rationale:** The fork already inherits ~132 resources + ~44 data sources
from upstream; its reason to exist is the createable surface Microsoft does
not cover. This phase drives toward representing *every* createable Azure
DevOps REST API 7.1 resource as a long program of small, releasable
initiatives the scheduler grinds through unattended. Net-new / absent whole
areas (release/task_group test substrate, Test Management, service-hook &
notification subscriptions, audit streams, Pipelines API) run before
upstream-parity polish. Every resource ships `azdosdkmocks`+gomock unit
tests — the only behavioural verification available without live ADO creds
— so the two substrate initiatives (01, 03) gate the rest to lock the
canonical mock-test pattern before ~60 resources are written against it.

## Initiatives

| ID | Title | Status | Manifest | Depends on |
|----|-------|--------|----------|------------|
| INIT-2026-05-23-release-def-substrate-gates | release_definition — test substrate + deployment gates (cwc-refined dogfood) | pending | [link](./_architect/2026-05-23T11-19-10/manifests/INIT-2026-05-23-release-def-substrate-gates.md) | — |
| INIT-2026-05-18-betterado-01-release-def-test-substrate | Release def test substrate + deployment gates | superseded | [link](../../_queue/pending/INIT-2026-05-18-betterado-01-release-def-test-substrate.md) | — (superseded by INIT-2026-05-23-release-def-substrate-gates; pre-cwc-amendment plan) |
| INIT-2026-05-18-betterado-02-release-folder | betterado_release_folder | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-02-release-folder.md) | 01 |
| INIT-2026-05-18-betterado-03-task-group-test-substrate | task_group test substrate + completeness | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-03-task-group-test-substrate.md) | — |
| INIT-2026-05-18-betterado-04-test-plan-core | Test Management — plans/suites/configurations | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-04-test-plan-core.md) | 01, 03 |
| INIT-2026-05-18-betterado-05-test-case-and-variables | Test Management — cases & variables | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-05-test-case-and-variables.md) | 04 |
| INIT-2026-05-18-betterado-06-servicehook-subscriptions | Service-hook subscriptions (composite) | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-06-servicehook-subscriptions.md) | 01, 03 |
| INIT-2026-05-18-betterado-07-notification-subscriptions | Notification subscriptions | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-07-notification-subscriptions.md) | 01, 03 |
| INIT-2026-05-18-betterado-08a-audit-client | Audit SDK client + MockAuditClient | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-08a-audit-client.md) | 01, 03 |
| INIT-2026-05-18-betterado-08b-audit-streams | betterado_audit_stream | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-08b-audit-streams.md) | 08a |
| INIT-2026-05-18-betterado-09-pipelines-api | betterado_pipeline (Pipelines API) | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-09-pipelines-api.md) | 01, 03 |
| INIT-2026-05-18-betterado-10-git-pull-request | Git pull requests / threads / labels | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-10-git-pull-request.md) | 01, 03 |
| INIT-2026-05-18-betterado-11-git-collaboration-extras | Git annotated tags / import / commit status | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-11-git-collaboration-extras.md) | 10 |
| INIT-2026-05-18-betterado-12-dashboard-widgets | betterado_dashboard_widget | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-12-dashboard-widgets.md) | 01, 03 |
| INIT-2026-05-18-betterado-13-secure-files | betterado_secure_file | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-13-secure-files.md) | 01, 03 |
| INIT-2026-05-18-betterado-14-environment-vm-resources | Environment VM / generic resources | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-14-environment-vm-resources.md) | 01, 03 |
| INIT-2026-05-18-betterado-15-build-retention-and-settings | Build retention leases / settings / tags | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-15-build-retention-and-settings.md) | 01, 03 |
| INIT-2026-05-18-betterado-16-feed-views-upstream | Feed views & upstream sources | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-16-feed-views-upstream.md) | 01, 03 |
| INIT-2026-05-18-betterado-17-workitem-collaboration | Work-item tags / comments / relations | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-17-workitem-collaboration.md) | 01, 03 |
| INIT-2026-05-18-betterado-18-serviceendpoint-longtail | Service-endpoint long tail (first batch) | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-18-serviceendpoint-longtail.md) | 01, 03 |
| INIT-2026-05-18-betterado-19-pat-token-management | betterado_personal_access_token | pending | [link](../../_queue/pending/INIT-2026-05-18-betterado-19-pat-token-management.md) | 01, 03 |

> Status reflects `_queue/` location at last architect pass; the scheduler
> moves files, a later architect/PM pass reconciles links. "Depends on"
> uses the ordinal fragment (e.g. `01`) of the full `INIT-…-betterado-NN-…`
> ID; full IDs + feature-level edges live in each manifest. Aggregate
> worst-case unattended spend if all initiatives run: ≈ $534 (operator
> chose "queue everything now" + comprehensive docs, informed).

## Backlog

- **Parity polish (revisit via a later /forge-architect):** missing data sources where the resource already exists; deeper attribute parity on inherited resources.
- **Service-endpoint long tail (beyond the first batch):** remaining ~dozen newer endpoint types after INIT-18 lands.
- **Release/vsrm runtime:** releases, deployments, deployment approvals — imperative/runtime, not declarative-Terraform-shaped; needs a design decision before it becomes initiatives.
- **Platform extras:** symbol server; search; profile/billing (not Terraform-appropriate); an upstream-sync mechanism (currently out of scope — single-branch fork).
