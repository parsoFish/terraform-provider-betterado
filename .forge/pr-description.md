## Why

The betterado provider has been incrementally migrating every resource and data source from the SDKv2 shim path to terraform-plugin-framework. This initiative completes the migration roadmap by removing the `tf6muxserver` + `tf5to6server` mux scaffold that was holding the two implementations together at runtime. Keeping the mux in place:

- forces `main.go` to maintain two provider factories (SDKv2 + framework) even when the SDKv2 side is empty,
- prevents upgrading the plugin protocol cleanly (mux adds latency and version-negotiation complexity), and
- blocks the eventual removal of the entire `azuredevops/provider.go` SDKv2 surface.

With this PR every `betterado_*` resource and data source is registered exclusively on the framework provider. Users on Terraform >= 1.x (plugin protocol 6) get a leaner, single-path provider binary.

## What

**WI-1 — Port 4 JFrog service endpoint resources to framework**
- Added `resource_serviceendpoint_jfrog_{artifactory,distribution,platform,xray}_v2_framework.go` — each implements `resource.Resource` + `resource.ResourceWithConfigure` with schema 100% attribute-compatible with the SDKv2 originals.
- Registered all 4 constructors in `azuredevops/internal/provider/framework_provider.go`.

**WI-2 — Port 12 remaining SDKv2 service endpoint resources to framework**
- Added `resource_serviceendpoint_{kubernetes,maven,nexus,nuget,octopusdeploy,openshift,runpipeline,servicefabric,snyk,sonarqube,ssh,visualstudiomarketplace}_framework.go`.
- Registered all 12 constructors in `framework_provider.go`.

**WI-3 — Empty SDKv2 ResourcesMap / DataSourcesMap in provider.go**
- Removed all 16 remaining SDKv2 resource registrations from `azuredevops/provider.go`.
- `ResourcesMap` and `DataSourcesMap` are now empty. `go build -mod=vendor .` still passes.

**WI-4 — Rewrite main.go to pure framework; delete framework.go shim**
- `main.go` now calls `tf6server.Serve` with `providerserver.NewProtocol6WithError(internalprovider.NewFrameworkProvider())` — no mux, no SDKv2 upgrade path.
- Deleted `azuredevops/framework.go` (re-export shim no longer needed).
- `azuredevops/internal/acceptancetests/testutils/mux_provider.go`: renamed `GetMuxedProviderFactories()` → `GetProviderFactories()`; removed mux and SDKv2 upgrade step; updated all callers.
- Added `TestFrameworkProvider_MuxFree` to `framework_provider_test.go` as the per-WI gate.

**WI-5 — Live acceptance test + CHANGELOG + version bump**
- Added `TestAccProviderMuxFree` in `resource_project_test.go`: reads `betterado-standing-demo` project via pure-framework binary; calls `CaptureLiveEvidence` for live REST GET evidence.
- `CHANGELOG.md`: added `### BREAKING CHANGES` and `### INTERNAL` entries under `## [Unreleased]`.
- `PROVIDER_VERSION.txt`: bumped `1.22.0` → `2.0.0` (BREAKING: mux scaffold removed, Terraform >= 1.x required).

## How

Key files changed (from `git diff --name-only main...HEAD`):

| File | Change |
|---|---|
| `main.go` | Rewritten: pure tf6server.Serve, no mux |
| `azuredevops/framework.go` | Deleted (re-export shim) |
| `azuredevops/provider.go` | ResourcesMap + DataSourcesMap emptied |
| `azuredevops/provider_test.go` | Updated to match empty provider |
| `azuredevops/internal/provider/framework_provider.go` | +16 service endpoint constructors |
| `azuredevops/internal/provider/framework_provider_test.go` | +TestFrameworkProvider_MuxFree |
| `azuredevops/internal/service/serviceendpoint/resource_serviceendpoint_jfrog_*_framework.go` (×4) | New framework implementations |
| `azuredevops/internal/service/serviceendpoint/resource_serviceendpoint_*_framework.go` (×12) | New framework implementations |
| `CHANGELOG.md` | BREAKING CHANGES + INTERNAL entries |
| `PROVIDER_VERSION.txt` | 1.22.0 → 2.0.0 |

The approach is pattern-consistent with every prior migration initiative: each `*_framework.go` mirrors the SDKv2 schema attribute-for-attribute (same types, optional/required/computed flags, sensitive flags) so existing state files round-trip without plan diffs.

Quality gate (`go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...`) is GREEN on branch HEAD.
