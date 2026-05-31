# ADO Demo Skill — make a betterado change *demonstrable*

## Purpose

This is **betterado's half of the forge demo contract** (the forge half lives in
`forge/skills/demo/SKILL.md`). It codifies how a change to this provider is shown
to the operator at review so the merge decision is made on *evidence*, not on a
table of test names.

The key fact this skill exists to capture: **betterado is NOT a "no-web-UI"
project.** Every resource it manages is visible two independent ways —

1. **The live Azure DevOps REST API** (`scripts/ado-api.sh` GET against
   `dev.azure.com` / `vsrm.dev.azure.com`), and
2. **The Azure DevOps web portal** (`https://dev.azure.com/{org}/...`) — a real
   web UI that can be screenshotted via the `ado-browser-inspector` skill.

Because the provider builds locally and runs against the operator's real ADO org
(creds in the gitignored `secrets.env`, standing sample project), a betterado
demo can show the *actual resource it created in Azure*, not just that some mocks
passed. That live evidence is the demo we want; the offline test harness is the
creds-free floor, not the ceiling.

## When to use

The forge demo phase composes this skill when producing `demo.json` /
`DEMO.html` for any betterado initiative. Pick the mode by the diff:

| Change kind | Demo mode | Evidence |
|---|---|---|
| **New or changed resource behaviour** (new resource, new field, fixed CRUD/expand-flatten path) | **A — Live evidential demo** | `terraform apply` against the sample ADO project → API GET confirming the created object → **portal screenshot** of the resource → `destroy` |
| **Tests-only / internal** (characterization tests for *existing* behaviour, refactor, doc) | **B — Harness + API double-confirm** | Green `go test -v` (what changed) **plus** at least one live API GET (or portal screenshot) that *independently* confirms the real resource exhibits the behaviour the new tests assert |
| **Infra/CI/no observable surface** | `none` | Rationale-only `essence` (rare — most betterado work touches a resource) |

Scale the effort to the change, but **default to showing the live object** for
anything that touches resource behaviour. A reviewer who can see the
release definition sitting in the ADO portal — with the stages, approvals, and
secret variables the change is about — approves in seconds. That is the whole
point of the demo phase.

## Prerequisites (the live layer)

Mode A and the double-confirm in Mode B need a working live path:

- `AZDO_ORG_SERVICE_URL`, `AZDO_PERSONAL_ACCESS_TOKEN`, `AZDO_TEST_PROJECT` —
  exported from the gitignored `secrets.env` (PAT scoped to the sample org).
- A locally-built provider wired via `dev_overrides` (a Terraform CLI config that
  points `betterado` at the freshly-built binary — same setup the onboarding
  live-confirmation used). `make install` + the dev-overrides block.
- The standing **sample ADO project** in the operator's personal org (the demo
  applies into it and destroys clean — never leaves orphans).

If creds are absent (CI, a fork without the PAT), Mode A degrades to Mode B's
harness floor — state that in `essence` rather than fabricating a screenshot.

## Workflow — Mode A (live evidential demo)

1. **Frame the prior state.** Capture the *before*: either the resource doesn't
   exist yet (new resource) or `ado-api.sh GET` the current object (changed
   field). Save the JSON; this is the "before" checkpoint.
2. **Apply the change.** `terraform apply` a minimal HCL exercising the new
   behaviour against `AZDO_TEST_PROJECT`, using the dev-overrides binary.
3. **Confirm via API.** `scripts/ado-api.sh GET "release/definitions/{id}?expand=environments,artifacts"`
   (compose the **`ado-api-explorer`** skill). The response is machine evidence
   that the resource exists with the expected shape. Save it as the "after" JSON.
4. **Screenshot the portal.** Compose the **`ado-browser-inspector`** skill to
   navigate to the resource's ADO page (e.g.
   `https://dev.azure.com/{org}/{project}/_releaseDefinition?definitionId={id}&_a=environments-editor`)
   and capture a screenshot. This is the *visual* before/after the operator sees.
5. **Destroy + verify clean.** `terraform destroy`; `ado-api.sh GET` returns 404.
   Confirm no orphan resources (the onboarding clean-destroy invariant).
6. **Author `demo.json`.** One checkpoint per behavioural delta:
   - `kind: "screenshot"` with `media` paths to the portal captures (before/after),
   - plus a `metrics` row or two from the API GET (e.g. `environments: 0 → 2`,
     `secret_variables_preserved: PASS`), so the harness signal sits alongside the
     visual one.
   Drop the media under the cycle's demo dir (NOT a `.forge/` shadow), per the
   forge demo skill's capture rules.

## Workflow — Mode B (tests-only + double-confirm)

Use when the initiative only *adds tests* to existing, already-shipped behaviour
(e.g. the `release_definition` characterization-tests cycle):

1. **Show what changed — the green run.** `go test -tags all -v -count=1
   ./azuredevops/internal/service/release/...` and capture the real `--- PASS`
   lines (not just a count). This is the harness checkpoint.
2. **Double-confirm against reality.** The mock tests assert "CRUD calls the SDK
   correctly / secrets preserved / retry-on-conflict works." Independently confirm
   at least the headline assertion against *live* ADO: `ado-api.sh` create→read a
   real definition with a secret variable and show the read-back matches what the
   test claims (and, ideally, one portal screenshot of that definition). This
   guards against tests that pass against mocks but drift from real API behaviour.
3. **Author `demo.json`** with the harness checkpoint + a `kind: "screenshot"` or
   API-evidence checkpoint for the double-confirm.

> **`release_definition` live caveat (2026-05-31):** the acceptance HCL for
> `release_definition` is currently stale — current ADO requires stage
> `retention_policy` + approvals that the sample HCL omits (deferred fix, tracked
> in the repo roadmap + forge memory). Until that's fixed, the live double-confirm
> for `release_definition` specifically may need a hand-built minimal definition
> via `ado-api.sh` rather than `terraform apply`. `task_group` and other resources
> with working acc HCL can use the full Mode A flow today.

## Mapping to the forge review UI

`demo.json` checkpoints render on the `/review/{cycleId}` screen
(`DemoComparison`). `kind: "screenshot"` checkpoints show the portal media large
(the review equivalent of the PLAN gate); `kind: "harness"` checkpoints render the
metrics table. A betterado demo should aim for **at least one visual checkpoint**
whenever the change touches a resource — that is the difference between a reviewer
reading a list of test names and a reviewer *seeing the release pipeline the change
built*.

## Tips

- The release API is on `vsrm.dev.azure.com`; core resources on `dev.azure.com`.
  `ado-api.sh` and the portal URLs differ accordingly.
- Secrets are write-only on read-back (`isSecret: true` → null) — the double-confirm
  for secret handling is "the apply succeeded and a re-plan shows no drift," shown
  via the plan output, not an API read of the secret value.
- Keep the sample project tidy: every demo applies AND destroys; a screenshot of a
  half-torn-down state is worse than no screenshot.
- This skill is the project side of the contract; if the forge demo phase ever
  hands you a `demo.shape` that doesn't fit (betterado's evidence is "live external
  system + portal screenshots," which the shape vocab calls `harness` today), prefer
  the richer evidence and note the shape mismatch for the forge-side fix.
