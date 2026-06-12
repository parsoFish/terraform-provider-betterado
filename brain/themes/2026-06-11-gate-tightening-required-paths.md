---
title: Gate-tightening required-paths rejection works correctly for examples/
description: When WI requires `examples/data-sources/<resource>/main.tf`, gate-tightening rejects iteration 0 if the file is absent from git diff; agent creates it in iteration 1. Expected-fail pattern works.
category: pattern
created_at: 2026-06-11
updated_at: 2026-06-11
---

## Observation

WI-1 (INIT-2026-06-08-release-data-sources-completion) required:
- `examples/data-sources/betterado_release_definition_revision/main.tf`
- `examples/data-sources/betterado_release_definition_history/main.tf`

Iteration 0 gate result: `gate_exit_code: -3`, `reject_reason: required-paths-missing`. Gate stderr told the agent exactly what files to create. Agent created both in the same iteration and gate passed iteration 1.

## Pattern

For any new data source in this provider, the WI `creates[]` list MUST include:
- `examples/data-sources/<resource_name>/main.tf`
- `docs/data-sources/<resource_name>.md`

Both are required for `terrafmt-check` and documentation completeness. Gate-tightening on these paths forces the agent to produce them before exiting, preventing silent omission.

## Required WI acceptance criterion template

```
- path: examples/data-sources/betterado_<resource>/main.tf must exist in git diff
- path: docs/data-sources/<resource>.md must exist in git diff
```

## Sources

- `_logs/2026-06-08T11-43-56_INIT-2026-06-08-release-data-sources-completion/events.jsonl` — EV_mq9g5q8w_boeqosl1 (gate.expected-fail, required-paths-missing), EV_mq9g8n3z_srm10q10 (gate.pass iteration 1)
- `/home/parso/forge/brain/cycles/_raw/2026-06-08T11-43-56_INIT-2026-06-08-release-data-sources-completion.md`
