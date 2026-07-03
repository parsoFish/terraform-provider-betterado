# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (completed — all ACs done)

**Goal:** Implement `betterado_servicehook_webhook_tfs` as a terraform-plugin-framework resource.

1. Read last-gate-failure.md — gate rejected "no tests to run" because the framework file + test didn't exist yet (this was iteration 0, no prior work).
2. Read the WI spec + tfs_publisher.go to understand all 19 event types.
3. Read the sibling WI-2 resource (`resource_servicehook_storage_queue_pipelines_framework.go`) and its test as a pattern.
4. Wrote `resource_servicehook_webhook_tfs_framework.go` — full framework resource with:
   - `servicehookWebhookTfsResource` struct + `NewServicehookWebhookTfsResource`
   - All 19 event type model structs + `servicehookWebhookTfsModel`
   - Full Schema with ListNestedBlocks for each event type
   - Inline `defaults.String` / `defaults.Bool` implementations (NOT `booldefault`/`stringdefault` — those packages are NOT in the vendor; use `schema/defaults` package directly)
   - Inline `planmodifier.String` implementations for id (use-state-for-unknown) and project_id (requires-replace)
   - `Configure()` with nil-guard + `*client.AggregatedClient` store
   - `Create/Read/Update/Delete` calling `ServiceHooksClient` CRUD
   - `Read` with 404 → `resp.State.RemoveResource(ctx)` + secret preservation
   - `ImportState` passthrough
   - `expandModel` / `flattenSubscription` helpers with http_headers newline-encode/decode
5. Wrote `resource_servicehook_webhook_tfs_framework_test.go` with build tag `//go:build all || resource_servicehook_webhook_tfs`, test func `TestServicehookWebhookTfsFramework_Configure` (3 subtests: non-nil constructor, nil-ProviderData no-op, valid-client stored).
6. Updated `framework_provider.go` Resources() to add `servicehook.NewServicehookWebhookTfsResource`.
7. Updated `provider.go` ResourcesMap — removed `"betterado_servicehook_webhook_tfs"` entry + removed unused `servicehook` import.
8. Updated `provider_test.go` — removed `"betterado_servicehook_webhook_tfs"` from the expected list (the count is recalculated from the slice length automatically).
9. Ran `make fmt` (gofmt) then `make test` — all green.
10. Ran exact gate: `go test -tags all -run TestServicehookWebhookTfsFramework_Configure ./azuredevops/internal/service/servicehook/` — PASS, 3 sub-tests ran, no "no tests to run".
11. Committed as a79292d3.

## What worked

- **Pattern from sibling WI-2**: The storage_queue resource is the exact template to follow. All inline helpers (plan modifiers, defaults) follow the same pattern with different prefixes.
- **Vendor constraint**: The project vendor does NOT have `booldefault` or `stringdefault` packages. Instead use `schema/defaults` package and implement inline `defaults.Bool` / `defaults.String` interfaces exactly like `sqpInt64Default` in the sibling.
- **provider.go import cleanup**: When the last `servicehook.*` call is removed from provider.go, the import must be removed too or build fails.
- **gofmt**: Always run `make fmt` before committing; the tool changes whitespace in struct alignment.

## What didn't work

- Initial attempt used `booldefault.StaticBool` and `stringdefault.StaticString` — those packages are not in vendor (first build error). Fixed by using inline defaults implementing `defaults.Bool` / `defaults.String` from `schema/defaults` package.

## Open questions

_(none)_

## Notes for reflection

- The 19-event-type pattern with ListNestedBlocks works well for the framework migration; the ExactlyOneOf validator from SDKv2 is intentionally omitted (framework's validator system is different; the logic is enforced at expand time by the switch/case falling through to an empty eventType).
