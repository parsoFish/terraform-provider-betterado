# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete)

- Read all four feed source files (`resource_feed.go`, `resource_feed_permission.go`, `resource_feed_retention_policy.go`, `data_feed.go`) to inventory the current TF schema.
- Read `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed/models.go` for the complete Go SDK structs: `Feed`, `FeedPermission`, `FeedRetentionPolicy`, and `FeedUpdate`.
- Read `docs/release-definition-gap-matrix.md` as the canonical format reference.
- Wrote `docs/feed-gap-matrix.md` covering all four surfaces with 149 lines, classifying every API field as `implemented` / `writable-gap` / `read-only` / `deferred`.
- Gate test `TestFeed_Create_DoesNotSwallowError` passes.
- Committed as `10c007ee`.

## What worked

- Reading `vendor/.../feed/models.go` gives the authoritative SDK field list — no need for external API docs.
- The `release-definition-gap-matrix.md` format (legend table → per-model section → summary table) transfers directly.
- `FeedRetentionPolicy.AgeLimitInDays` is explicitly deprecated in the SDK comment — classify as `deferred` rather than `writable-gap`.
- `features.permanent_delete` and `features.restore` are provider-internal flags (not ADO API fields); they still need to appear in the matrix as `implemented`.

## What didn't work

_(nothing failed in iteration 0)_

## Open questions

_(none)_

## Notes for reflection

- WI-1 is a docs-only AC (`behavior_preserving: true`). The gate only runs offline unit tests; no TF_ACC.
- Key writable gaps to track for future WIs: `upstream_enabled`, `upstream_sources`, `description`, `hide_deleted_package_versions`, `badges_enabled`, `default_view_id` on `betterado_feed`; the same fields as computed outputs on `data.betterado_feed`.
