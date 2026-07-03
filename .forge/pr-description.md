## Why

ADO Feature Management (`_apis/featuremanagement/featureflags`) exposes per-feature enable/disable control at project and host scope for features such as Boards, Repos, Pipelines, Test Plans, and Artifacts. Before this initiative, the only Terraform surface for feature state was `betterado_project_features`, which hard-codes exactly five named features through a composite endpoint and cannot address arbitrary `featureId` values. There was no way to manage a feature flag declaratively via its full contribution ID (e.g. `ms.vss-work.agile`).

This initiative closes that gap by implementing `betterado_feature_flag` as a first-class terraform-plugin-framework resource and data source, registered exclusively on the framework provider in preparation for the forthcoming mux-free cutover.

## What

- **`docs/featuremanagement-gap-matrix.md`** — field-by-field survey of the ADO Feature Management REST API (`ContributedFeature` + `ContributedFeatureState`), scope matrix (host / project / user), state values (`enabled` / `disabled` / `undefined`), and a documented resolution of the overlap with `betterado_project_features`.
- **`azuredevops/internal/service/featuremanagement/resource_feature_flag_framework.go`** — framework `resource.Resource` implementing full CRUD for `betterado_feature_flag` against `GetFeatureStateForScope` / `SetFeatureStateForScope`. Schema: `feature_id` (Required, ForceNew), `scope_name` (Required, ForceNew), `scope_value` (Required, ForceNew), `state` (Required, `OneOf("enabled","disabled")`), `overridden` (Computed), `reason` (Computed/Optional). Delete sets state to `undefined` (restores default).
- **`azuredevops/internal/service/featuremanagement/data_feature_flag_framework.go`** — framework `datasource.DataSource` reading current feature state by key triple.
- **`azuredevops/internal/service/featuremanagement/featuremanagement_sdk_mock.go`** — gomock mock of `featuremanagement.Client` for unit tests.
- **`azuredevops/internal/service/featuremanagement/resource_feature_flag_framework_test.go`** — unit tests: `TestFeatureFlagSchemaHasRequiredFields` + `TestFeatureFlagCRUDCreate/Read/Update/Delete` (all passing).
- **`azuredevops/internal/acceptancetests/resource_feature_flag_test.go`** — live acceptance test `TestAccFeatureFlag_basic`: apply → read-back → idempotency re-plan (`ExpectNonEmptyPlan: false`) → destroy. Calls `testutils.CaptureLiveEvidence("acceptance-resource", ...)` during read-back.
- **`azuredevops/internal/provider/framework_provider.go`** — adds `featuremanagement.NewFeatureFlagResource` to `Resources()` and `featuremanagement.NewFeatureFlagDataSource` to `DataSources()`. No changes to `azuredevops/provider.go` (SDKv2 — AC-4 compliant).
- **`docs/resources/feature_flag.md`** + **`docs/data-sources/feature_flag.md`** — generated via `make docs` (tfplugindocs).
- **`examples/resources/betterado_feature_flag/resource.tf`** + **`examples/data-sources/betterado_feature_flag/data-source.tf`** — embedded by docs.
- **`CHANGELOG.md`** — draft entry under `## [Unreleased]` for `betterado_feature_flag`.
- **`PROVIDER_VERSION.txt`** — patch version bump.

Files changed (15): `.forge/project.json`, `CHANGELOG.md`, `PROVIDER_VERSION.txt`, `azuredevops/internal/acceptancetests/resource_feature_flag_test.go`, `azuredevops/internal/acceptancetests/shared_fixtures.go`, `azuredevops/internal/provider/framework_provider.go`, `azuredevops/internal/service/featuremanagement/data_feature_flag_framework.go`, `azuredevops/internal/service/featuremanagement/featuremanagement_sdk_mock.go`, `azuredevops/internal/service/featuremanagement/resource_feature_flag_framework.go`, `azuredevops/internal/service/featuremanagement/resource_feature_flag_framework_test.go`, `docs/data-sources/feature_flag.md`, `docs/featuremanagement-gap-matrix.md`, `docs/resources/feature_flag.md`, `examples/data-sources/betterado_feature_flag/data-source.tf`, `examples/resources/betterado_feature_flag/resource.tf`.

## How

1. **WI-1** wrote `docs/featuremanagement-gap-matrix.md` by surveying the vendored `featuremanagement` SDK (`ContributedFeature`, `ContributedFeatureState`, scope/state enums) and cross-referencing `resource_project_features.go` to document the overlap and non-goals.
2. **WI-2** scaffolded `azuredevops/internal/service/featuremanagement/` with the framework schema skeleton and a gomock mock of `featuremanagement.Client`; verified by `TestFeatureFlagSchemaHasRequiredFields`.
3. **WI-3** filled the CRUD stubs: `Create`/`Update` call `SetFeatureStateForScope`; `Read` calls `GetFeatureStateForScope` (404 → `RemoveResource`; `undefined` → treat as deleted); `Delete` calls `SetFeatureStateForScope` with `Undefined` state. CRUD verified by four gomock unit tests.
4. **WI-4** added live acceptance test `TestAccFeatureFlag_basic` targeting the `betterado-standing-demo` project, called `testutils.CaptureLiveEvidence` during read-back to write `.forge/live-evidence/acceptance-resource.json`, confirmed `ExpectNonEmptyPlan: false` idempotency.
5. **WI-5** wired the resource and data source into `framework_provider.go` (framework-only, SDKv2 untouched), regenerated docs via `make docs`, restored `docs/guides/` with `git checkout`, updated `provider_test.go` counts, drafted the changelog entry, and bumped `PROVIDER_VERSION.txt`.
