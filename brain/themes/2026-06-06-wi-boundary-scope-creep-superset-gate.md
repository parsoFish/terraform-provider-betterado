---
title: When WI-N's gate is a superset of WI-(N-1)'s scope, WI-N becomes already-complete at iter-0
description: If the dev-loop agent writes more than the current WI's minimum deliverable, and the next WI's quality gate prefix matches the already-written tests, the next WI passes at iter-0 as already-complete — saving cost but blurring WI boundaries.
category: antipattern
project: terraform-provider-betterado
created_at: 2026-06-06T06:04:52Z
updated_at: 2026-06-06T06:04:52Z
related_themes:
  - 2026-06-06-resume-already-complete-near-zero-cost
---

# WI boundary scope creep: superset gate causes already-complete

## Antipattern

In INIT-2026-06-05-release-definition-permissions:

- **WI-1 gate:** `go test -run TestReleaseDefinitionPermissions_TokenFormatSpike ./azuredevops/internal/service/permissions/`
- **WI-2 gate:** `go test -run TestReleaseDefinitionPermissions ./azuredevops/internal/service/permissions/`

WI-2's gate is a **prefix superset** of WI-1's gate. The PM spec said WI-1 would create the impl stub file; WI-2 would create the test file. But the WI-1 agent wrote the full impl including all tests (spike test + token tests + error-path tests) — not just the spike test. When WI-2 started, the gate passed at iter-0 because all `TestReleaseDefinitionPermissions*` tests already existed.

Result: WI-2 ran at $0 with zero tool use (already-complete). No implementation iteration ran.

## Why it's an antipattern

The WI boundary was designed to separate risk: WI-1 = probe + stub (uncertain), WI-2 = full impl (buildable once WI-1 confirmed format). The already-complete outcome collapses this separation and leaves the PM's WI-2 spec as dead code. It also means the "full impl" was done inside a WI spec that said it would produce only a stub — reducing traceability.

## When it's acceptable

It's cost-efficient (WI-2 spends $0) and the deliverable is correct. The risk is that if the WI-1 agent had written WRONG tests (not testing the right cases), WI-2's gate would have passed on wrong coverage without any agent review pass on WI-2 scope.

## PM mitigation for future permissions resources

For the pattern `spike-WI → impl-WI` where the impl-WI gate is a superset:

**Option A:** Use a file-specific gate for WI-2 that fails if the test file doesn't exist. E.g.:
```
go test -mod=vendor -tags all -count=1 -run TestReleaseDefinitionPermissions_CreateToken ./azuredevops/internal/service/permissions/
```
(Use the `_CreateToken` suffix, which WI-1 is not expected to write.)

**Option B:** Accept the scope-creep as efficient and document the WI-1/WI-2 pair as "spike + impl in one" for permissions resources — since the token is always a pure function, the full impl fits naturally in the spike WI.

## Sources

- `_logs/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions/events.jsonl` (event: `ralph.end` WI-2 `already-complete`, iterations=0, cost_usd=0)
- `_logs/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions/work-items-snapshot/WI-2.md`
- `brain/cycles/_raw/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions.md`
