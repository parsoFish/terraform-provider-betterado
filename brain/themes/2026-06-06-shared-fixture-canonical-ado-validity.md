---
title: Shared fixture pattern — canonical ADO validity enforced once, reused across tests
description: A single SharedReleaseFixture helper in shared_fixtures.go centralises ADO object provisioning (project, repo, build, variable group, canonical release definition) and enforces all current API validity constraints (VS402877, VS402982, EditReleaseEnvironment) in one place, eliminating per-test duplication that hid bugs.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-06T09:41:00Z
updated_at: 2026-06-06T09:41:00Z
related_themes:
  - 2026-06-06-integration-test-project-idea
  - 2026-06-05-live-acceptance-gate-for-acceptance-wis
  - 2026-06-06-ado-silent-field-discard-idempotency
---

# Shared fixture pattern — canonical ADO validity enforced once

## Pattern

Place all acceptance-test ADO object provisioning in a single `SharedReleaseFixture(t *testing.T)` helper
(`azuredevops/internal/acceptancetests/shared_fixtures.go`) rather than hand-rolling a minimal release
definition inline in each test.

## Why it matters

Per-test HCL fragments that each create their own minimal project/release-definition are fragmentation
vectors: each one can omit a constraint that the ADO REST API now requires, and the omission is only
discovered when that specific test is run live. Three bugs hid this way until the INIT-1..4 close-out:

- **VS402877** — ADO requires BOTH `pre_deploy_approval` AND `post_deploy_approval` per stage.
- **VS402982** — every stage needs a `retention_policy` block.
- **Stale permission key** — `EditReleaseStage` rejected; `EditReleaseEnvironment` required.

A single canonical fixture locked in all three constraints in one place. Updating the fixture
updates every test that consumes it.

## Implementation contract (as of INIT-2026-06-06-shared-acceptance-fixture)

- **File:** `azuredevops/internal/acceptancetests/shared_fixtures.go`
- **Entry point:** `SharedReleaseFixture(t *testing.T) SharedFixtureResult`
- **Returns:** `SharedFixtureResult{ProjectID, RepoID, BuildDefinitionID, VariableGroupID, ReleaseDefinitionID string/int}`
- **Creds guard:** `t.Skip` immediately when `TF_ACC == ""` — offline unit suite stays creds-free.
- **Teardown:** `t.Cleanup` callbacks for all five provisioned objects — no orphaned ADO resources even on failure.
- **Approval pattern:** automated-approver UUID (`00000000-0000-0000-0000-000000000000` + `IsAutomated: true`) satisfies VS402877 without requiring a real identity lookup.
- **Smoke test:** `TestSharedReleaseFixture` in `shared_fixtures_test.go` verifies non-zero IDs + per-stage approval/retention presence via API read-back.

## Future extension

The `2026-06-06-integration-test-project-idea` theme documents the next step: extending the fixture
to include real group/user identities (for permission tests that require a real approver subject) and
a real ADO project that the environment_templates resource can template from. The current automated-
approver pattern is a valid stepping stone.

## Sources

- `_logs/2026-06-06T09-32-34_INIT-2026-06-06-shared-acceptance-fixture/events.jsonl` (events: `dev-loop.delivered` EV_mq25q8m9_10kf7p0v; `unifier.end` EV_mq25q8lj_e7rpjf5b)
- `_logs/2026-06-06T09-32-34_INIT-2026-06-06-shared-acceptance-fixture/artifacts/DEMO.md`
- `brain/cycles/_raw/2026-06-06T09-32-34_INIT-2026-06-06-shared-acceptance-fixture.md`
