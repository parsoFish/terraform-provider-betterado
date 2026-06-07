---
title: Live-acc gate failed on missing creds, yet the cycle merged unverified-live
description: In the release-folder-data-source cycle WI-2's live acceptance gate failed all 5 iterations in PreCheck (missing AZDO_PERSONAL_ACCESS_TOKEN), WI-2 ended status:failed, but the unifier still opened PR #14 on the offline unit gate + TF_ACC-stripped CI, and it merged. The headline AC ("resolves against live ADO") shipped unverified.
category: antipattern
project: terraform-provider-betterado
created_at: 2026-06-07T13:45:00Z
updated_at: 2026-06-07T13:45:00Z
related_themes:
  - 2026-06-06-acceptance-test-partial-delivery-gate-passes
  - 2026-06-07-data-source-parity-pattern-confirmed
  - 2026-05-23-dogfood-cycle-false-pass-gate
---

# Live-acc gate failed, cycle merged unverified-live

## What happened

`INIT-2026-06-07-release-folder-data-source` shipped `data.betterado_release_folder`
(PR #14 MERGED). Every *static* AC clause was met and the unit tests pass. But the
clause that betterado actually exists to guarantee — **behaviour proven against live
ADO** — was never executed:

- WI-2's per-WI gate was `go test -run TestAccDataReleaseFolder` (a live acc test).
- It failed all 5 iterations in **PreCheck** with `` `AZDO_PERSONAL_ACCESS_TOKEN` must
  be set `` — the test `t.Fatal`-ed before any ADO round-trip.
- WI-2 ended `status: failed` (iteration budget exhausted on an unfixable env gap).
- The cycle still opened PR #14 (offline unit gate green + the CI gate runs with
  `TF_ACC` stripped, so the acc test is build-tag-gated out) and the operator merged it.

## Why it's dangerous

For betterado the whole contract (C7) is "behaviour is only provable against live ADO."
A green offline gate + green GitHub CI is *by design* blind to live behaviour. So a
data source that compiles, has passing unit mocks, and is correctly registered can ship
with its live read path entirely unexercised. Here the code is plausibly correct (it's
a faithful mirror of the proven `data_release_definition` pattern), but "plausibly
correct" is exactly what the live acc gate exists to replace with "proven."

## Two root causes (both now addressed)

1. **Creds gap (FIXED).** `.forge/project.json` `acceptance_gate.requires_env` listed
   only `TF_ACC`. `TF_ACC` *was* set, so the dev-loop's live-env guard stayed silent and
   let the test run, where it `t.Fatal`-ed on the missing PAT. Fix: `requires_env` now
   lists all three vars the PreCheck (`commons.go:56-57`) demands — `TF_ACC`,
   `AZDO_ORG_SERVICE_URL`, `AZDO_PERSONAL_ACCESS_TOKEN` — so the guard errors fast
   ("live env absent") instead of burning 5 iterations on an unfixable failure.
2. **Status-blind merge boundary (OPEN, forge-side).** The unifier's `canOpenPr` checks
   only that each WI's `creates[]` files exist + the final CI gate passes; it does NOT
   require that a live-acc WI actually PASSED its gate. So a `status: failed` live-acc WI
   shipped. Tracked in forge `docs/known-gaps.md` (2026-06-07, item 1).

## How to apply

- When running a betterado cycle that has a live-acc WI, run it from a shell with
  `TF_ACC=1` **and** `AZDO_ORG_SERVICE_URL` + `AZDO_PERSONAL_ACCESS_TOKEN` exported
  (the `secrets.env` pattern), or the live AC cannot be closed.
- Treat "PR merged + CI green" as **necessary but not sufficient** for a betterado
  resource/data-source. The live acc test passing is the real acceptance signal.
- To retroactively close this data source's AC: run
  `TF_ACC=1 … go test -tags all -run TestAccDataReleaseFolder ./azuredevops/internal/acceptancetests/`
  and confirm PASS (create folder → read via data source → assert description → idempotent re-plan → clean destroy).
