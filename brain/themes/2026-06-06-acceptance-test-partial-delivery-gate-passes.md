---
title: Acceptance test gate passes on partial delivery when spec requires multiple test functions
description: When a WI spec requires two test functions (SetPermissions + UpdatePermissions) but the agent only commits one, the quality gate passes if its -run regex matches the single committed function — partial delivery is not caught by the gate.
category: antipattern
project: terraform-provider-betterado
created_at: 2026-06-06T06:04:52Z
updated_at: 2026-06-06T06:04:52Z
related_themes:
  - 2026-06-05-live-acceptance-gate-for-acceptance-wis
  - 2026-06-06-docs-only-wi-gate-mismatch
---

# Acceptance test gate passes on partial delivery

## Antipattern

WI-4 of INIT-2026-06-05-release-definition-permissions required:
- `TestAccReleaseDefinitionPermissions_SetPermissions` — apply + read + idempotency + destroy
- `TestAccReleaseDefinitionPermissions_UpdatePermissions` — update + re-read + idempotency

Gate command: `go test -run TestAccReleaseDefinitionPermissions -timeout 30m ./azuredevops/internal/acceptancetests/`

The agent committed only `SetPermissions`. Gate ran both function names via the prefix regex — but since only `SetPermissions` exists and it passed, the gate exited 0. `UpdatePermissions` was silently absent.

The unifier caught it by reading the committed file and cross-checking against the WI spec. Gate alone did not.

## Why this happens

`go test -run <prefix>` matches whatever functions exist. If `UpdatePermissions` doesn't exist, the run command simply doesn't run it — no error. The gate can't distinguish "ran and passed" from "not present, not run".

## How to catch it

**Option A:** Name tests so that the gate _cannot_ pass unless both exist. Use function-specific `-run` patterns in separate gate invocations:
```bash
go test -run TestAccReleaseDefinitionPermissions_SetPermissions ... && \
go test -run TestAccReleaseDefinitionPermissions_UpdatePermissions ...
```

**Option B:** Use the `creates:` field in the WI spec to list expected function names; add a pre-gate check that greps for them:
```bash
grep -q TestAccReleaseDefinitionPermissions_UpdatePermissions azuredevops/internal/acceptancetests/resource_release_definition_permissions_test.go
```

**Option C:** Accept partial delivery, flag in unifier AGENT.md for operator decision. This is what this cycle did — it's pragmatic but requires the unifier to be read carefully.

## Operator action

A follow-up WI (or an amend to the PR) should add `TestAccReleaseDefinitionPermissions_UpdatePermissions` to complete the spec.

## Sources

- `_logs/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions/events.jsonl` (event: `gate.pass` WI-4, `duration_ms=...`)
- `_logs/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions/work-items-snapshot/WI-4.md`
- `_logs/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions/report.md` (unifier AGENT.md note: "UpdatePermissions was not committed")
- `brain/cycles/_raw/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions.md`
