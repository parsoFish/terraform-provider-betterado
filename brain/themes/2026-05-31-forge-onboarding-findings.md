---
slug: 2026-05-31-forge-onboarding-findings
project: terraform-provider-betterado
date_added: 2026-05-31T02:30:00.000Z
category: operation
related_themes:
  - 2026-05-23-dogfood-cycle-false-pass-gate.md
  - 2026-05-18-go-test-harness-demos.md
  - 2026-05-18-stack-and-test-layout.md
---

# Forge onboarding run (2026-05-31) — gate pattern, live creds, harness findings

First forge cycle to ship landed code here, and the first run with **live Azure
DevOps creds** available. Durable learnings for future betterado cycles:

## The betterado quality-gate pattern (use this for every test-adding WI)

A test-adding WI's gate MUST be:

```
go test -mod=vendor -tags all -count=1 -run <TestPrefix> ./azuredevops/internal/service/<pkg>/
```

- **`-run <Prefix>`** scoped to the WI's NEW tests — so on a clean tree it prints
  `[no tests to run]` and the gate FAILS pre-work (forge's hollow-iter0 guard
  needs a clean-tree failure; a bare package gate passes because the package has
  sibling tests → the WI dies `gate-too-loose`).
- **`-tags all`** — the gomock unit tests sit behind `//go:build all`; without it
  `go test` runs nothing.
- **Exact package dir, NO `/...`** — `./taskagent/...` also runs the test-less
  `taskagent/validate` subpackage, whose `[no tests to run]` poisons forge's
  no-work-indicator and fails the gate even when the real tests pass.
- `-mod=vendor` keeps the build offline.

This is the corrected form of the [[2026-05-23-dogfood-cycle-false-pass-gate]]
mitigation (the old "never use -run" advice predates forge's `[no tests to run]`
indicator, which now backstops `-run`).

## Canonical unit-test substrate (landed for task_group)

`azuredevops/internal/service/taskagent/resource_task_group_test.go` now holds the
canonical 5-test gomock pattern (expand/flatten roundtrip, create-error,
read-404-clears-id, update-args, delete-error), mirroring `resource_environment_test.go`.
Reuse it as the template for `release_definition` and every future `betterado_*`
resource. The fork's other net-new resource, `release_definition`, still has **no
unit test** (only the 6 live acceptance tests).

## Live testing now possible — and it catches real drift

`secrets.env` (gitignored) holds the PAT; the provider reads `AZDO_ORG_SERVICE_URL`
+ `AZDO_PERSONAL_ACCESS_TOKEN` from env. Confirmed live: `betterado_project`
created a real ADO project in org `davidgparsonson` (~10s), verified via API GET,
destroyed clean. The repeatable apply→confirm→destroy kernel is in
`/tmp/live-confirm/` (dev_overrides → locally-built binary); formalize it into
`scripts/forge-acc-harness.sh`.

**Stale release acceptance HCL (work item):** live `apply` of the basic release
definition failed on current ADO — `VS402982` (stage-level `retention_policy` now
required; pipeline-level deprecated) then `VS402877` (pre/post approvals now
required on the stage). `TestAccReleaseDefinition_basic`'s HCL omits both and would
fail live. The release-acceptance feature must add a stage `retention_policy` block
and a `pre_deploy_approval` with a valid approver.

## Forge-readiness config (committed to main)

`.forge/project.json` + `.forge/quality_gate_cmd` declare the offline unit gate;
`.gitignore` now excludes forge scratch, `secrets.env`, the compiled
`terraform-provider-betterado` binary, and root `graphify-out/` (the binary +
graphify gaps had polluted a cycle PR with a 35 MB blob). `forge preflight` passes.
