## Why

The `betterado_feed`, `betterado_feed_permission`, `betterado_feed_retention_policy`, and `data.betterado_feed` resources were still on the legacy terraform-plugin-sdk/v2 path. Migrating them to terraform-plugin-framework completes the feed package's integration with the mux ProtoV6 provider, ensuring these resources are served consistently alongside other already-migrated resources. Framework resources receive first-class plan modifier support and eliminate the need for `d.Set`/`d.Get` patterns, reducing the risk of silent schema drift. The field-level gap matrix (`docs/feed-gap-matrix.md`) documents every ADO Artifacts Feed API v7.1 field with its implementation status so future contributors know exactly where gaps remain.

## What

- **`docs/feed-gap-matrix.md`** — new field-level gap matrix comparing ADO Artifacts Feed API v7.1 against each resource's schema (fields: implemented / writable-gap / read-only / deferred).
- **`azuredevops/internal/service/feed/resource_feed_framework.go`** — framework implementation of `betterado_feed` (Create/Read/Update/Delete/Import); soft-deleted feeds treated as destroyed.
- **`azuredevops/internal/service/feed/resource_feed_permission_framework.go`** — framework implementation of `betterado_feed_permission`; identity resolution via GraphClient + IdentityClient; role polling loop.
- **`azuredevops/internal/service/feed/resource_feed_retention_policy_framework.go`** — framework implementation of `betterado_feed_retention_policy`; in-place update supported; import splits on `/`.
- **`azuredevops/internal/service/feed/datasource_feed_framework.go`** — framework implementation of `data.betterado_feed`; lookup by name or feed_id.
- **`azuredevops/internal/service/feed/framework_defaults.go`** — shared framework helpers (defaults, plan modifiers).
- **`azuredevops/internal/provider/framework_provider.go`** — registers all four new framework resources/data-sources.
- **`azuredevops/provider.go`** — deregisters all four SDKv2 entries from ResourcesMap/DataSourcesMap.
- **`azuredevops/provider_test.go`** — updated resource count.
- **`azuredevops/internal/acceptancetests/resource_feed_framework_test.go`** — `TestAccFeedFramework_orgScopedBasic`, `TestAccFeedFramework_projectScopedBasic` (apply → read-back → idempotency → destroy; CaptureLiveEvidence called).
- **`azuredevops/internal/acceptancetests/resource_feed_permission_test.go`** — `TestAccFeedPermissionFramework_basic` added (apply → read-back → idempotency → destroy).
- **`azuredevops/internal/acceptancetests/resource_feed_retention_policy_test.go`** — `TestAccFeedRetentionPolicyFramework_projectBasic` and `TestAccFeedRetentionPolicyFramework_update` added.
- **`azuredevops/internal/acceptancetests/data_feed_test.go`** — `TestAccFeedDataSourceFramework_byName` and `TestAccFeedDataSourceFramework_byId` added.
- **`docs/resources/feed.md`**, **`docs/resources/feed_permission.md`**, **`docs/resources/feed_retention_policy.md`**, **`docs/data-sources/feed.md`** — regenerated via `make docs`.
- **`examples/resources/betterado_feed/resource.tf`**, **`examples/resources/betterado_feed_permission/resource.tf`**, **`examples/resources/betterado_feed_retention_policy/resource.tf`**, **`examples/data-sources/betterado_feed/data-source.tf`** — new HCL examples for docs embedding.
- **`CHANGELOG.md`** — draft `## [Unreleased]` entry added.
- **`PROVIDER_VERSION.txt`** — patch version bumped.

## How

1. **Gap matrix first (WI-1):** Read each SDKv2 resource file, cross-referenced against the ADO Artifacts Feed API v7.1 schema, and produced `docs/feed-gap-matrix.md` following the format of the release-definition gap matrix.

2. **Framework resource migration (WI-2–5):** Each resource follows the same checklist:
   - New `*_framework.go` file implementing `resource.Resource` (or `datasource.DataSource`) with full schema parity.
   - Registered in `framework_provider.go` (Resources()/DataSources() slices).
   - Deregistered from `provider.go` ResourcesMap/DataSourcesMap.
   - `Configure()` asserts `req.ProviderData` is `*client.AggregatedClient`.
   - New `TestAcc*Framework*` acceptance tests use `GetMuxProviderFactories()` (ProtoV6 path); existing SDKv2 tests left unchanged.
   - `CaptureLiveEvidence("acceptance-resource", url, response)` called in the check step, writing `.forge/live-evidence/acceptance-resource.json` with a real REST GET URL.

3. **Docs, examples, changelog, version (WI-6):**
   - HCL example files created before `make docs` so tfplugindocs can embed them.
   - `make docs` run; `docs/guides/` restored via `git checkout -- docs/guides/`.
   - CHANGELOG.md updated under `## [Unreleased]`.
   - PROVIDER_VERSION.txt patch-bumped.
   - `make terrafmt` run to format HCL examples.

4. **Quality gate:** `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` — all packages pass (release, taskagent, taskagent/validate).
