---
title: Linear dependency chain amplifies single-WI crash into total cycle failure
description: A fully sequential WI-1 → WI-2 → … chain means one WI-1 crash skips all downstream WIs; prefer shallower graphs for parallelizable content.
category: antipattern
created_at: 2026-06-11
updated_at: 2026-06-11
---

## What happened

INIT-2026-06-08 Run 1: PM emitted 5 WIs in a fully linear chain (WI-1 → WI-2 → WI-3 → WI-4 → WI-5). WI-1 agent crashed twice (`Claude Code process exited with code 1`, `iterations=0`). All 4 downstream WIs received `ralph.skipped` with `reason=prerequisite-failed`. Result: 0/5 WIs delivered — total failure.

The downstream WIs were individually thin slices (WI-3 = process_parameters only, WI-4 = properties only) that had no real dependency on each other beyond sharing the same file. They were forced sequential by the graph.

## Why it's an antipattern

- A single agent crash at WI-1 (any cause: API timeout, SDK crash, resource exhaustion) eliminates all remaining work regardless of whether the remaining WIs are actually blocked.
- The resume run re-decomposed to 2 WIs (WI-1: all schema+tests, WI-2: live acceptance) and delivered everything. The original 5-WI decomposition added no value and maximised blast radius.

## Rule of thumb

If WI-3, WI-4, WI-5 all edit the same two files (`resource_release_definition.go`, `resource_release_definition_test.go`), they should be one WI. Declare `depends_on` edges only for genuine sequencing requirements (e.g., unit tests before live acceptance), not for file-sharing granularity.

## Sources

- `_logs/2026-06-08T12-01-16_INIT-2026-06-08-release-definition-environment-config-surface/events.jsonl` — `ralph.skipped reason=prerequisite-failed` × 4 at 12:09:10
- `/home/parso/forge/brain/cycles/_raw/2026-06-08T12-01-16_INIT-2026-06-08-release-definition-environment-config-surface.md`
