---
title: Resume with already-complete WI produces near-zero-cost cycle
description: When a prior failed cycle implemented the work but died before review, the resume detects gate-pass at iter-0 and routes directly to the unifier — total cycle cost is ~$0.66 (unifier-only), not a full dev-loop spend.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-06T02:36:29Z
updated_at: 2026-06-06T02:36:29Z
related_themes:
  - 2026-05-31-release-definition-unit-test-substrate
---

# Resume with already-complete WI produces near-zero-cost cycle

## Pattern

When an initiative has prior `requeued-from-failed` history and the work was actually
implemented (the branch has commits), a `resume_from: developer` cycle will:

1. Ralph starts the WI and immediately finds the quality gate passes at iter-0.
2. Ralph exits with `stop_reason: already-complete`, cost $0.00, zero tool use.
3. Unifier runs a single iteration to correct/validate demo artifacts and push.
4. CI gate runs and confirms green.
5. PR opens and merges normally.

Total cost is ~$0.66 (unifier pass only), not $3–8 (typical dev-loop with implementation).

## Observed in this cycle

`INIT-2026-06-05-release-folder` had 4× `requeued-from-failed`. The `betterado_release_folder`
resource was already fully implemented from a prior failed cycle. This cycle:

- Ralph WI-1: iter-0 gate pass → `already-complete` → $0.00
- Unifier: 1 iteration → $0.66 (demo.json correction + commit + push)
- CI gate: green (`ran_fixer: true`)
- PR #9: opened and merged

**Takeaway:** The resume machinery works as intended. Prior failures that implemented
the work are not wasted — the next resume recovers cheaply.

## Implication for failure analysis

If a cycle shows 4× `requeued-from-failed` but the final resume is cheap ($0.66),
the prior failures were likely environmental (CI state, worktree lock, creds) rather
than agent quality issues. Check the failure mode labels before concluding the agent
was struggling.

## Sources

- `_logs/2026-06-06T02-00-02_INIT-2026-06-05-release-folder/events.jsonl` (events:
  `ralph.end` WI-1 `already-complete` EV_mq1ph3bm; `dev-loop.delivered` EV_mq1pkp79;
  `unifier.end` cost_usd=0.6579797 EV_mq1pkp6i)
- `brain/cycles/_raw/2026-06-06T02-00-02_INIT-2026-06-05-release-folder.md`
