# betterado — API-coverage roadmap (north star + first initiative)

> **Status: active — gap registry initialized 2026-08-23.**
> INIT-2026-08-14-betterado-gap-registry completed the 31-area gap matrix and priority
> backlog (docs/gap-registry.md). The Release initiative substrate (FEAT-1/2/4) shipped;
> FEAT-3 (environment template) remains not-started. The forge-hardening refinement
> concerns flagged in 2026-05-31 have been resolved. Next: execute priority backlog
> items from docs/gap-registry.md ## Priority backlog (Tier 1 → Tier 2 → Tier 3).

## North star

`betterado` becomes a **feature-complete representation of the ADO REST API**
([7.2](https://learn.microsoft.com/en-us/rest/api/azure/devops/?view=azure-devops-rest-7.2)):
every *manageable* resource in the API is represented as a `betterado_*` resource
(or data source). Once complete, **keeping in sync with new API releases** becomes
mechanical: diff the published spec against the coverage matrix (below) and file an
initiative per delta.

### Fork reality — don't reimplement upstream
betterado is a fork of `microsoft/terraform-provider-azuredevops`, which already
covers ~132 resources + ~44 data sources. "Feature-complete" therefore =
**inherited upstream coverage (maintain parity)** + **the net-new surface Microsoft
doesn't cover** (the fork's reason to exist: classic release pipelines, task
groups, and the long tail). The build program targets net-new + parity gaps; it
does **not** rewrite inherited resources.

## The mapping model (the user's structure, made concrete)

| ADO API concept | forge unit | what it is |
|---|---|---|
| An API **path / area** (Release, Build, Git, Pipelines, Test Plans, Distributed Task, Service Hooks, Audit, Core, Graph, Policy, Security, Wiki, …) | **Initiative** | one scoped, releasable program run roadmap-scale |
| A **manageable resource** within that area | **Feature** | the full Terraform lifecycle of one resource |
| A **slice** of that resource | **Work item** | (a) schema + expand/flatten; (b) CRUD wiring + the canonical 5 unit tests; (c) live acceptance test; (d) docs + example. Split per-endpoint only if the resource is large. |

### The Terraform-shape rule — not every endpoint is a resource
- **Declarative, lifecycle-managed** (create/read/update/delete + import) → a
  `betterado_*` **resource**.
- **Read-only / list** → a **data source**.
- **Imperative / runtime** (queue a build, create a release *run*, approve/reject,
  manual intervention) → **out of declarative scope.** Documented and deferred to a
  separate design decision (a possible future imperative escape-hatch); NOT a
  resource. This is the line the existing `roadmap.md` flags for vsrm runtime.

### Per-feature definition-of-done (the forge↔project contract, applied)
Every feature must, before it's "done":
- **In-loop gate** — the canonical 5 mock unit tests (`azdosdkmocks`+gomock,
  creds-free), gated `go test -tags all -run <TestPrefix> ./<pkg>/` (exact package
  dir, no `/...`). Mirror `resource_environment_test.go`.
- **Confirmation layer** — a live `TF_ACC` acceptance test that creates the resource
  in an isolated provider-created project, confirms via a live API GET, and
  auto-destroys (randomized `test-acc-*` names).
- **Docs + runnable example**; resource registered in `azuredevops/provider.go`.
- Additive; new fields Optional/Computed; preserve state shape; `go build -mod=vendor .` green.

---

## First-mover initiative — the **Release** API path

`INIT-<date>-release-api-coverage` — bring the Release Management API
(`vsrm.dev.azure.com`, 7.2) to complete **declarative** coverage. Chosen first
because it's the fork's reason to exist and already has partial substrate.

**Current state (as of 2026-05-31):**
- `release_definition`: resource built (~1,490 LOC). 6 acceptance tests exist but
  are **stale** — live `apply` fails on current ADO (`VS402982` stage-level
  `retention_policy` now required; `VS402877` pre/post approvals now required).
  **No unit tests.**
- `release_folder`, `release_definition_environment_template`: **not built.**

**Features (dependency-ordered):**

- **FEAT-1 — `release_definition`: complete the substrate.** `depends_on: []`
  - WI: 5-test unit substrate `resource_release_definition_test.go` (the item
    deferred from the 2026-05-31 onboarding run). Gate:
    `go test -tags all -run ^TestReleaseDefinition ./azuredevops/internal/service/release/`.
  - WI: acceptance refresh — add a stage `retention_policy` + a `pre_deploy_approval`
    with a valid approver so `TestAccReleaseDefinition_*` pass live. Gate:
    `env TF_ACC=1 go test -tags all -run TestAccReleaseDefinition ./azuredevops/internal/acceptancetests/`.
  - WI: schema parity audit vs the 7.2 Release Definitions schema (gates,
    approvalOptions, properties, tags) + docs/example refresh.

- **FEAT-2 — `release_folder`: new resource.** `depends_on: [FEAT-1]`
  - WIs: schema + expand/flatten + provider registration → CRUD + 5 unit tests →
    acceptance test + docs/example. (This was the old roadmap's INIT-02.)

- **FEAT-3 — `release_definition_environment_template`: new resource.** `depends_on: [FEAT-1]`
  - Create/read/delete only (templates are immutable → no Update; major fields
    ForceNew). WIs: schema → CRD + unit tests → acceptance + docs.

- **FEAT-4 — read surface as data sources.** `depends_on: [FEAT-1]`
  - `data.betterado_release_definition` (by id/name) and
    `data.betterado_release_definitions` (list). WIs: schema + read + unit tests →
    docs.

**Out of declarative scope (documented, NOT built this initiative):** Releases
(runtime create), Deployments, Approvals (approve/reject), runtime Gates, Manual
Interventions — imperative; needs a separate design decision before any imperative
escape-hatch.

**Budgets:** roadmap-scale, multi-feature (NOT single-WI). Size per the
work-item-completion-by-domain calibration; the live-acceptance WIs need creds in
the cycle env (`AZDO_*`) + the orphan-sweep discipline.

---

## The coverage matrix (the sync mechanism)

Maintain a tracked table — **7.2 API path → resource → {upstream | fork | gap} →
status** — regenerated from the published 7.2 spec. Diffing it against a new API
release yields the delta set; each delta becomes an initiative or feature. This is
the engine of "feature-complete, then auto-synced." Seed it from the upstream
provider's resource list + this fork's net-new resources; the per-path initiatives
above fill the `gap` rows.

## Sequencing across the whole roadmap (sketch, not built here)
After Release lands as the proven template, the same path→initiative pattern
applies to the other net-new / parity-gap areas (Test Plans, Service Hooks,
Notification Subscriptions, Audit, Pipelines, Dashboards, Secure Files, …) — the
areas already enumerated in `roadmap.md`, now reframed as one initiative per API
path. Inherited-from-upstream areas are parity-maintenance, not net build.
