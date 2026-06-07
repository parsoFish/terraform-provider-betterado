---
title: report.md inverted diff on resume — third occurrence, now systemic
description: For the third time, a resume cycle's report.md showed all delivered files as deleted (−N lines) while DEMO.md and dev-loop.delivered confirmed +N insertions. Treat report.md diff as unreliable on any resume; DEMO.md is authoritative.
category: antipattern
project: terraform-provider-betterado
created_at: 2026-06-07T03:25:00Z
updated_at: 2026-06-07T03:25:00Z
related_themes:
  - 2026-06-06-report-diff-stale-on-resume
  - 2026-06-06-resume-already-complete-near-zero-cost
---

# report.md inverted diff on resume — third occurrence, now systemic

## Antipattern

In `INIT-2026-06-07-release-folder-data-source`, `report.md` showed:
```
10 files changed, 878 deletions(-)
```
All 7 substantive files appeared deleted. `DEMO.md` showed:
```
10 files changed, 872 insertions(+)
```
All 7 files added. The DEMO.md / unifier evidence is correct; the report.md diff is stale.

This is the **third occurrence** (see also `INIT-2026-06-06-shared-acceptance-fixture` and at least one prior). The pattern is no longer incidental — it is systemic for resume cycles.

## Why it happens (recap)

`report.md` is generated from a `git diff` snapshot at report-render time. On a resume cycle, the renderer may diff two points within the cycle's own commit history (e.g., branch tip at cycle-start vs. a prior state), showing the new files as net-deleted rather than net-added vs. `main`.

## Authoritative sources (in priority order)

1. `DEMO.md` diff-stat (unifier-confirmed against live branch).
2. `dev-loop.delivered` event metadata (`insertions` / `deletions` / `files_changed`).
3. Direct `git diff main...HEAD` in the worktree.
4. `report.md` unified diff — **unreliable on resume cycles; do not use for delivery conclusions**.

## Implication

Any tooling or human review that reads `report.md` for delivery completeness MUST cross-check against DEMO.md on resume cycles. "Nothing delivered" conclusions from report.md alone are incorrect.

## Sources

- `_logs/2026-06-07T03-20-11_INIT-2026-06-07-release-folder-data-source/report.md`
- `_logs/2026-06-07T03-20-11_INIT-2026-06-07-release-folder-data-source/artifacts/DEMO.md`
- `brain/cycles/_raw/2026-06-07T03-20-11_INIT-2026-06-07-release-folder-data-source.md`
