# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2025 — first read)

All work was already committed in prior iterations of this initiative branch. WI-2 arrived as status=complete in the WI spec frontmatter but the fix_plan checklist was blank. Upon inspection:

- `resource_group_framework.go` (636 lines): full terraform-plugin-framework Resource implementation of `betterado_group` with all schema fields: scope, origin_id, mail, display_name, description, members (set), url, origin, subject_kind, domain, principal_name, descriptor (ID), group_id.
- `resource_group_test.go` already contains `TestAccGroupResource_Framework` (line 325) with muxed provider factories, idempotency step, evidence capture, and a checkGroupDestroyedFramework helper using a direct ADO client.
- `framework_provider.go`: `graph.NewGroupResource` is in Resources() at line 211.
- `provider.go`: `"betterado_group"` key is ABSENT from ResourcesMap (only a comment remains).
- `CHANGELOG.md`: betterado_group resource migration bullet under ## [Unreleased] FEATURES (line 12).
- `docs/resources/group.md`: exists.

All quality gates passed:
- `make test`: PASS (no FAIL lines)
- `golangci-lint run --new-from-rev=main ./azuredevops/...`: 0 issues
- `make terrafmt-check`: Exit 0

## What worked

- All implementation was pre-existing from earlier work on this branch. No code changes were needed in this iteration.

## What didn't work

_(nothing to record — work was already done)_

## Open questions

_(none)_

## Notes for reflection

- WI-2 was already fully implemented before iteration 0 ran. The WI spec had `status: complete` in frontmatter. Fix plan was just blank. Iteration 0 confirmed completion and updated the plan.
- The `checkGroupDestroyedFramework` helper uses `groupGetDirectClient()` (direct ADO client from env vars) instead of `testutils.GetProvider().Meta()` because ProtoV6ProviderFactories doesn't wire SDKv2 provider singleton Meta — this is the correct pattern for framework acceptance tests.
- Live acceptance test requires `TF_ACC=1` env var; without it the test skips (prints `ok ... 0.00s`), which is NOT a pass — the live gate runs it with TF_ACC set.
