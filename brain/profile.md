---
project: terraform-provider-betterado
created_at: 2026-05-18T10:43:08Z
updated_at: 2026-05-18T10:43:08Z
status: active
domain: infrastructure tooling (Terraform provider for Azure DevOps)
stack: [Go, Terraform Plugin SDK v2, Azure DevOps REST API 7.1]
taste_decay: 0.05
---

# terraform-provider-betterado

A GitHub fork of `microsoft/terraform-provider-azuredevops` ("better ADO
provider"). It inherits the full official provider (100+ resources) and adds
what Microsoft has not implemented — chiefly **classic release pipelines**
(`vsrm.dev.azure.com` REST API) and **task groups**. Success = the new
`betterado_*` resources manage real Azure DevOps release/task-group state
correctly, cleanly track upstream, and stay mergeable back toward upstream.

## Taste signals

- Stay faithful to upstream's package structure and idioms — this is a
  fork meant to (potentially) contribute back, not a rewrite.
- New surface lives behind the `betterado_` prefix; the `release/` service
  package is the home of the net-new value.
- Prefer additive resources/attributes over invasive changes to inherited
  code (keeps upstream merges tractable).

## Hard constraints

- **Acceptance tests need live Azure DevOps** (`TF_ACC=1` + a PAT). As of
  2026-05-31 the operator supplies creds via a gitignored `secrets.env`
  (`AZDO_ORG_SERVICE_URL` + PAT), so cycles CAN do live verification when
  launched with those in the env — confirmed: the provider creates real ADO
  resources end-to-end ([[2026-05-31-forge-onboarding-findings]]). Offline,
  `azdosdkmocks`+gomock unit tests remain the default gate. Both `task_group`
  (5 tests) and `release_definition` (11 tests) now have canonical gomock unit
  substrates ([[2026-05-31-release-definition-unit-test-substrate]]).
- Go 1.24.1; deps are vendored — do not break offline `go build`/`go test`.
- Single-branch model: `main` is the fork. The repo's own `CLAUDE.md`
  describes a now-superseded two-branch workflow — do not trust it
  ([[2026-05-18-branch-model-consolidated]]).

## Active focus

- Both unit substrates landed. **Next:** live acceptance harness (`scripts/forge-acc-harness.sh`) + fix stale `TestAccReleaseDefinition_basic` HCL (stage `retention_policy` + `pre_deploy_approval` now required by ADO REST 7.1 — see [[2026-05-31-forge-onboarding-findings]]). Then resume the createable-resource program in `roadmap.md`.

## Cycles

- **2026-05-31 — task_group unit-test substrate.** Landed 5 passing gomock unit
  tests for `betterado_task_group` (forge dev-loop, quality-gates-pass). Reached
  review; landed clean on `main` after operator stripped cycle PR pollution
  (binary/graphify/scratch — now gitignored). Live ADO confirmation: provider
  created+confirmed+destroyed a real project. See
  [[2026-05-31-forge-onboarding-findings]].
- **2026-05-31 — release_definition unit-test substrate.** Landed 11 gomock
  characterization tests for `betterado_release_definition` (PR #2, merged
  `9f3ac5d5`). 1 iteration resume, $0.80. One production fix: type-switch in
  `expandWorkflowTask` for `map[string]interface{}` inputs from Terraform SDK.
  See [[2026-05-31-release-definition-unit-test-substrate]].
