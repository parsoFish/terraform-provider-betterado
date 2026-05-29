<!-- verdict: approve -->

# Architect plan — 2026-05-23T11-19-10

- Project: `terraform-provider-betterado`
- Repo: `/home/parso/forge/projects/terraform-provider-betterado`
- Initiative type: `implementation`

> **Operator review.** Read each section. Leave inline notes by adding an HTML comment of the form `(left-angle bang dash dash review:` your text `dash dash right-angle)` on its own line beside any item. Set the verdict at the top of this file by replacing the placeholder with `approve`, `revise`, or `reject`. Then run `forge architect commit 2026-05-23T11-19-10` (or pass `--via-pr` for PR-comment review).

## Operator brief + interview

Queue INIT-01 for terraform-provider-betterado — the release_definition test substrate plus the pre/post deployment-gates schema extensions. The fork-side resource_release_definition.go exists and is registered as azuredevops_release_definition but has zero unit tests. Build the gomock substrate per the council's 5-test pattern, then layer both gate blocks with expand/flatten + tests, then ship docs + a runnable example. Quality gate is the standard go-test on release/... plus go build -mod=vendor ./.... Defer the other 18 betterado initiatives until this substrate is green.

### Interview

| # | Question | Operator answer |
|---|---|---|
| 1 | What's the scope edge of INIT-01? Substrate only, or substrate + deployment gates in one initiative? | Substrate + both gates — 3-4 features: gomock substrate w/ 5 unit tests + pre_deployment_gates schema + post_deployment_gates schema + docs/example. Matches the 2026-05-18 architect plan. The gates are the load-bearing reason the substrate exists. |
| 2 | What demo.shape should INIT-01 declare for the unifier/reviewer? | harness — Demo = `go test ./azuredevops/internal/service/release/... -v` output: before (0 tests) vs after (5+ tests passing). Matches betterado profile (no live ADO creds; unit tests are the honest gate). |
| 3 | Where do docs + example land? | Inside INIT-01 as final FEAT — docs/resources/release_definition.md + examples/release_definition/ ship as FEAT-4. Atomic with the substrate they document; bench cost is trivial. |

## Brain context

- `brain/projects/terraform-provider-betterado/profile.md` — Project taste signals + hard constraints (single-branch model, no acceptance tests without live ADO creds, additive-only).
- `brain/projects/terraform-provider-betterado/themes/council-constraints.md` — Binding council constraints shared across all betterado initiatives — quality gate, 5-test pattern, docs shape, fixtures rule, PM scope-guard.
- `brain/projects/terraform-provider-betterado/themes/release-substrate-context.md` — Gap analysis: release_definition exists+registered+zero-tests; substrate-first ordering unblocks 17 dependent initiatives.
- `brain/projects/terraform-provider-betterado/themes/2026-05-18-stack-and-test-layout.md` — Test-layout + go-test substrate patterns specific to the betterado fork.
- `brain/projects/terraform-provider-betterado/themes/2026-05-18-go-test-harness-demos.md` — Demo shape conventions for go-test substrate initiatives.

## Council transcript

Total cost: `$0.0000`

### Flags (auto-applied)

- `same-file-coupling-feat-2-3` — FEAT-2 (pre_deployment_gates) and FEAT-3 (post_deployment_gates) both extend resource_release_definition.go and its test file. Risk of hidden coupling if PM emits parallel WIs.. _Applied:_ Resolved at architect layer: FEAT-3 declares depends_on: [FEAT-2] so PM's topological order serialises them. FEAT-4 (docs) declares depends_on: [FEAT-2, FEAT-3] so docs ship last.

### CEO critic

Cost: `$0.0000`

- _no mechanical flags_

- _no taste escalations_

### Eng critic

Cost: `$0.0000`

**Flags (auto-resolved):**

- `same-file-coupling-feat-2-3` — FEAT-2 and FEAT-3 both extend resource_release_definition.go. PM must not emit parallel WIs.. _Applied:_ depends_on chain on the feature shape enforces serialisation; PM's detectHiddenCoupling will also flag.

- _no taste escalations_

### Design critic

Cost: `$0.0000`

- _no mechanical flags_

- _no taste escalations_

### DX critic

Cost: `$0.0000`

- _no mechanical flags_

- _no taste escalations_

## Proposed initiatives

| ID | Title | Features | Iteration budget | Depends on |
|---|---|---|---|---|
| `INIT-2026-05-23-release-def-substrate-gates` | release_definition — test substrate + deployment gates | 4 | 15 | — |

### INIT-2026-05-23-release-def-substrate-gates — drawer

```markdown
---
initiative_id: INIT-2026-05-23-release-def-substrate-gates
project: terraform-provider-betterado
project_repo_path: /home/parso/forge/projects/terraform-provider-betterado
created_at: '2026-05-23T11:19:10.000Z'
iteration_budget: 15
cost_budget_usd: 1.5
phase: pending
origin: architect
type: implementation
features:
  - feature_id: FEAT-1
    title: gomock test substrate for resource_release_definition
    depends_on: []
    quality_gate_cmd:
      - go
      - test
      - ./azuredevops/internal/service/release/...
      - -run
      - TestReleaseDefinition
    non_goals:
      - adding pre/post deployment_gates schema
      - docs and examples
    hard_constraints:
      - no edits to azdosdkmocks/ (read-only)
      - inline fixtures <20 lines; testdata/*.json otherwise
  - feature_id: FEAT-2
    title: pre_deployment_gates schema + expand/flatten + tests
    depends_on:
      - FEAT-1
    quality_gate_cmd:
      - go
      - test
      - ./azuredevops/internal/service/release/...
    non_goals:
      - post_deployment_gates
      - docs and examples
    hard_constraints:
      - additive — absent config reproduces prior upstream behaviour
  - feature_id: FEAT-3
    title: post_deployment_gates schema + expand/flatten + tests
    depends_on:
      - FEAT-2
    quality_gate_cmd:
      - go
      - test
      - ./azuredevops/internal/service/release/...
    non_goals:
      - docs and examples
    hard_constraints:
      - additive — must not mutate FEAT-2 surface
  - feature_id: FEAT-4
    title: docs + example for release_definition (incl. both gate blocks)
    depends_on:
      - FEAT-2
      - FEAT-3
    quality_gate_cmd:
      - go
      - build
      - -mod=vendor
      - ./...
    non_goals:
      - Go code changes
      - website/ edits
    hard_constraints:
      - docs/resources/release_definition.md + examples/release_definition/ only
demo_hook:
  shape: harness
  cmd:
    - go
    - test
    - ./azuredevops/internal/service/release/...
    - -v
  before_after: true
---
# release_definition — test substrate + deployment gates

## Context

The fork's `azuredevops/internal/service/release/resource_release_definition.go`
exists and is registered as `azuredevops_release_definition` (inherited
from upstream) but ships with **zero unit tests** in the betterado tree.
INIT-01 closes that gap and extends the schema with the two deployment-gate
blocks the upstream provider doesn't expose (`pre_deployment_gates` /
`post_deployment_gates`).

This is the **substrate** initiative for the release-pipelines surface —
INIT-02 through INIT-04 all gate on a green test substrate here.

## Council constraints (binding) — see brain

Per [council-constraints](brain/projects/terraform-provider-betterado/themes/council-constraints.md):
**5-test pattern** (expand/flatten roundtrip + create-error + read-404 +
update-args + delete-error), **quality gate** is
`go test ./azuredevops/internal/service/release/...` + `go build -mod=vendor ./...`,
**docs** template under `docs/resources/` + runnable `examples/<name>/`
(never `website/`), **fixtures** inline <20 lines else `testdata/*.json`,
**additive-and-atomic** (absent config = prior behaviour; quality-gate
failure marks initiative BLOCKED), **PM scope-guard** keeps work-items
within `azuredevops/internal/service/release/` (no scans of `vendor/`
or `website/`).

## Gap analysis — see brain

Per [release-substrate-context](brain/projects/terraform-provider-betterado/themes/release-substrate-context.md):
release_definition is one of three substrate initiatives (01 here, 03
task_group, 04 test plan core) — closing the substrate unblocks the 17
dependent initiatives. `resource_release_definition.go` exists +
registered + zero tests; this initiative adds substrate + gates.

## Features

### FEAT-1 — gomock test substrate

Add `azuredevops/internal/service/release/resource_release_definition_test.go`
with the council's five-test pattern. Uses gomock against
`azdosdkmocks/release_sdk_mock.go` (read-only). Pattern mirrors upstream
`resource_environment_test.go`.

**Acceptance criteria:**

1. **Given** a valid release-definition config, **when** expand-then-flatten
   the SDK type, **then** the round-trip preserves all schema fields.
2. **Given** the Azure DevOps client returning an error on Create,
   **when** the resource Create() runs, **then** Terraform surfaces the
   error and state is empty.
3. **Given** the client returning a 404 on Read, **when** Read() runs,
   **then** state is cleared (no panic, no error).
4. **Given** a valid Update, **when** Update() runs, **then** the client's
   Update SDK method was called with the expected args.
5. **Given** the client returning an error on Delete, **when** Delete()
   runs, **then** Terraform surfaces the error and state is preserved.

### FEAT-2 — pre_deployment_gates schema

Extend `resource_release_definition.go` with a `pre_deployment_gates`
schema block (nested set), implement expand/flatten roundtrip, and extend
the FEAT-1 substrate with two new gate-specific test cases.

**Acceptance criteria:**

1. **Given** a config carrying `pre_deployment_gates`, **when** flatten()
   reads the SDK type, **then** the schema's roundtrip is byte-identical.
2. **Given** a config WITHOUT `pre_deployment_gates`, **when** Read()
   runs, **then** the absent block reproduces upstream behaviour exactly
   (additive contract).

### FEAT-3 — post_deployment_gates schema

Mirror FEAT-2 for the post-deployment side. Strictly additive on top of
FEAT-2 (no mutation of FEAT-2 surface). Same expand/flatten + 2 test
cases as FEAT-2.

**Acceptance criteria:**

1. **Given** a config carrying `post_deployment_gates`, **when**
   flatten() reads the SDK type, **then** the schema's roundtrip is
   byte-identical.
2. **Given** a config carrying BOTH `pre_deployment_gates` and
   `post_deployment_gates`, **when** the resource is created, **then**
   the SDK call's args carry both gate blocks in the order Terraform
   surfaces them.

### FEAT-4 — docs + example

Write `docs/resources/release_definition.md` per the council-constraints
docs template (description + basic + complex example + argument &
attribute reference + import). Write `examples/release_definition/main.tf`
+ `examples/release_definition/README.md` showing both gate blocks in a
runnable example.

Edit `docs/resources/` + `examples/` ONLY. Do NOT touch `website/`.

**Acceptance criteria:**

1. **Given** the new docs file, **when** rendered by Terraform's docs
   tooling (or grepped for required sections), **then** all four
   council-constraint sections are present (description, examples,
   arguments/attributes, import).
2. **Given** the new example dir, **when** `terraform init` is run in
   it (no apply), **then** the config validates against the betterado
   provider build.

## Aggregate cost — informational

Iteration budget: 15. No explicit $-ceiling (per C19). Demo shape:
harness (go-test output before vs after). The demo is the verification.
```

## Aggregate footprint (informational)

_This block surfaces the **informational** footprint of the proposed initiatives — how many cycles + dollars they would consume if every one were queued today. It is informational only; forge does not enforce a budget or block at any number._

- Initiatives proposed: **1**
- Total iteration budget: **15**
- Total estimated cost: **$0.95**

---

_Generated by the architect skill on 2026-05-23T11:20:35.165Z. Edit this file in place; commit with `forge architect commit 2026-05-23T11-19-10`._
