---
title: report.md diff section can show inverted delivery on resume cycles
description: When a cycle is a resume of prior failed work, the report.md unified diff may capture a stale git state and show the new files as deleted rather than added — contradicting the authoritative dev-loop.delivered event. Always cross-check dev-loop.delivered against the report diff before concluding what landed.
category: antipattern
project: terraform-provider-betterado
created_at: 2026-06-06T09:41:00Z
updated_at: 2026-06-06T09:41:00Z
related_themes:
  - 2026-06-06-resume-already-complete-near-zero-cost
---

# report.md diff section can show inverted delivery on resume cycles

## Antipattern

In `INIT-2026-06-06-shared-acceptance-fixture`, the `report.md` unified diff section showed
`shared_fixtures.go` as a **deleted file** (−484 lines) and `resource_release_definition_test.go`
reverting to the old inline-project HCL. This directly contradicts:

- `dev-loop.delivered` event: `files_changed=6, insertions=1141, deletions=1` (EV_mq25q8m9_10kf7p0v)
- `DEMO.md` artifacts: 6 files, +1117 insertions / −1 deletion

## Why it happens

The `report.md` is generated from a `git diff` snapshot of the worktree at the time the report
renderer runs. On a resume cycle:

1. The prior failed cycle committed the adds (e.g. `shared_fixtures.go` added).
2. The resume cycle started and the unifier made corrective commits (updating `demo.json`, `AGENT.md`).
3. The report renderer captured a diff **between two points in the cycle's commit history** rather than
   between `main` and the branch HEAD — or captured the diff before a final unifier push.

The result is a diff that shows the unifier's corrective commit as a **net deletion** of the prior work.

## How to avoid being misled

**The `dev-loop.delivered` event is authoritative.** It is emitted after Ralph's final quality-gate
confirmation against `git diff main...HEAD`, not from a renderer snapshot. If `dev-loop.delivered`
shows `insertions > 0` and `files_changed > 0`, the work landed — regardless of what `report.md`'s
diff section shows.

**Cross-check pattern for resume cycles:**
1. Check `dev-loop.delivered` metadata: `insertions` / `deletions` / `files_changed`.
2. Check `DEMO.md` / `demo.json` diffStat (unifier confirms these against the live branch).
3. If the `report.md` diff contradicts both, the report is stale — treat it as a reporting artefact.

This is a specific instance of the general principle: per-WI status metadata can be stale on a resume
(known antipattern); the diff-based delivery record is always fresher.

## Sources

- `_logs/2026-06-06T09-32-34_INIT-2026-06-06-shared-acceptance-fixture/events.jsonl` (event `dev-loop.delivered` EV_mq25q8m9_10kf7p0v vs report.md diff showing −1141 lines)
- `_logs/2026-06-06T09-32-34_INIT-2026-06-06-shared-acceptance-fixture/report.md` (diff section, lines 136-145)
- `_logs/2026-06-06T09-32-34_INIT-2026-06-06-shared-acceptance-fixture/artifacts/DEMO.md` (correct diff footer)
- `brain/cycles/_raw/2026-06-06T09-32-34_INIT-2026-06-06-shared-acceptance-fixture.md`
