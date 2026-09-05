# betterado — API-coverage roadmap (north star + first initiative)

> **Status: active — Release API initiative complete; roadmap-scale execution underway.**
> Written 2026-05-31 from the forge onboarding session. FEAT-1, FEAT-2, and FEAT-4
> of the Release initiative have shipped (see reconciled status below). The gap-registry
> consolidation initiative (INIT-2026-09-05) completed the Coverage index for all 31 ADO
> API areas — see `docs/gap-registry.md` for the full cross-area coverage status and
> Priority backlog. FEAT-3 (`release_definition_environment_template`) remains unbuilt.

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

**Current state (reconciled from CHANGELOG.md and git log, 2026-09-05):**
- `betterado_release_definition`: **shipped 2026-06-14** (v0.1.0). Migrated to
  terraform-plugin-framework (v1.0.0, 2026-06-20). Full schema parity including
  triggers, stages, deploy phases, gates, approval options, tags, and idempotency
  fixes (v1.0.0–v1.0.5). Acceptance tests live-proven. **FEAT-1 complete.**
- `betterado_release_folder`: **shipped in 0.2.0** (pre-1.0.0; framework migration
  date: 2026-07-01, v1.2.0). Resource and data source both live. **FEAT-2 complete.**
- `release_definition_environment_template`: **not built.** FEAT-3 remains open.
- Release data sources: **shipped 2026-07-01** (v1.2.0). Includes
  `data.betterado_release_definition`, `data.betterado_release_definitions`,
  `data.betterado_release_definition_history`, `data.betterado_release_definition_revision`,
  `data.betterado_release_folder`. **FEAT-4 complete** (expanded beyond original scope).

**Features (dependency-ordered):**

- **FEAT-1 — `release_definition`: complete the substrate.** `depends_on: []`
  **✓ SHIPPED** — v0.1.0 (2026-06-14); framework migration v1.0.0 (2026-06-20).
  Full schema parity, live TF_ACC acceptance tests, idempotency proven.

- **FEAT-2 — `release_folder`: new resource.** `depends_on: [FEAT-1]`
  **✓ SHIPPED** — initial resource in 0.2.0; framework migration v1.2.0 (2026-07-01).
  Resource and data source live with acceptance tests.

- **FEAT-3 — `release_definition_environment_template`: new resource.** `depends_on: [FEAT-1]`
  **NOT BUILT** — Create/read/delete only (templates are immutable → no Update; major
  fields ForceNew). WIs: schema → CRD + unit tests → acceptance + docs. Pending.

- **FEAT-4 — read surface as data sources.** `depends_on: [FEAT-1]`
  **✓ SHIPPED** — v1.2.0 (2026-07-01). Delivered: `data.betterado_release_definition`,
  `data.betterado_release_definitions`, `data.betterado_release_definition_history`,
  `data.betterado_release_definition_revision`, `data.betterado_release_folder`.

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
