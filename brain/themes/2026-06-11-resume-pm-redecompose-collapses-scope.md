---
title: Resume PM correctly collapses over-granular WI decomposition when original run failed at iteration 0
description: When the first run produced 0 WIs complete, the resume PM re-reads the initiative spec and emits a leaner, parallel-friendly graph that delivers the same scope in fewer WIs.
category: pattern
created_at: 2026-06-11
updated_at: 2026-06-11
---

## What happened

Run 1 PM (2026-06-08) decomposed INIT-2026-06-08 into 5 WIs (WI-1 through WI-5), all chained. WI-1 crashed → 0/5 delivered.

Run 2 PM (2026-06-11, manifest had `previous_failure_modes: [requeued-from-failed-2026-06-11]`) re-read the initiative spec and emitted only 2 WIs:
- WI-1: all schema + expand/flatten + 4 unit tests (covers original WI-1 through WI-4 content)
- WI-2: live acceptance test (covers original WI-5 content + AC-5)

The 2-WI graph delivered everything the 5-WI graph intended. The PM condensed because the original WI-3, WI-4, WI-5 were thin file-parallel slices of the same two files — not genuinely sequential concerns.

## Why this matters

- The resume PM receives `previous_failure_modes` in the manifest frontmatter; this signal (combined with the initiative spec reread) produces a materially better decomposition on retry.
- Over-granular decomposition on a single-file-pair initiative maximises crash blast radius. The PM self-corrects on resume when there is evidence of prior failure.
- Corollary: if the resume PM *still* produces the same overly-granular graph, that is a PM intelligence gap worth a theme.

## Sources

- `_logs/2026-06-08T12-01-16_INIT-2026-06-08-release-definition-environment-config-surface/events.jsonl` — Run 1: 5 WIs emitted, Run 2: 2 WIs emitted
- `/home/parso/forge/brain/cycles/_raw/2026-06-08T12-01-16_INIT-2026-06-08-release-definition-environment-config-surface.md`
- Manifest: `_queue/done/INIT-2026-06-08-release-definition-environment-config-surface.md` — `previous_failure_modes: [requeued-from-failed-2026-06-11]`
