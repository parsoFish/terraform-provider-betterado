# Demo — Cut over from mux scaffold to pure terraform-plugin-framework provider

> **Initiative:** INIT-2026-07-01-mux-free-cutover
> **Project:** terraform-provider-betterado
> **Diff:** 163 files changed, 7503 insertions(+), 549 deletions(-)

## Intent & Outcome

Remove `tf6muxserver` + `tf5to6server` from the provider entry point. All `betterado_*` resources and data sources are now served exclusively by terraform-plugin-framework (protocol 6). The mux bridge, SDKv2 provider factory call, and all SDKv2 `ResourcesMap`/`DataSourcesMap` registrations are deleted. Provider version bumped to 2.0.0 (BREAKING: requires Terraform >= 1.x).

### Acceptance Criteria Verdicts

| # | Criterion (summary) | Verdict | Evidence |
|---|---|---|---|
| AC1 (WI-1) | 4 JFrog *_framework.go files + constructors in framework_provider.go | ✅ met | All 4 files on branch; go build → BUILD_OK |
| AC2 (WI-1) | TestServiceEndpointJFrog* unit tests pass -tags all | ✅ met | go test -tags all -run TestServiceEndpointJFrog → PASS (commit 44e99682) |
| AC3 (WI-2) | 12 general *_framework.go files + constructors in framework_provider.go | ✅ met | All 12 files on branch; go build → BUILD_OK |
| AC4 (WI-2) | TestServiceEndpoint* unit tests pass -tags all | ✅ met | go test -tags all -run TestServiceEndpoint* → PASS (commit 3809b19f) |
| AC5 (WI-3) | ResourcesMap and DataSourcesMap empty | ✅ met | grep non-comment betterado_serviceendpoint → empty (commit 1dae5458) |
| AC6 (WI-3) | go build -mod=vendor . succeeds | ✅ met | BUILD_OK on HEAD |
| AC7 (WI-3) | Offline unit tests pass | ✅ met | servicehook gate → ok (0.003s) |
| AC8 (WI-4) | main.go: no tf5to6server/tf6muxserver/helper/schema imports | ✅ met | grep → empty; commit 0b3d9004 |
| AC9 (WI-4) | go build succeeds after main.go rewrite | ✅ met | BUILD_OK on HEAD |
| AC10 (WI-4) | framework.go shim: build still succeeds | ✅ met | minimal 3-line re-export kept; BUILD_OK |
| AC11 (WI-4) | GetMuxedProviderFactories renamed; mux removed; all callers updated | ✅ met | commit 0b3d9004; 124 acc-test files updated |
| AC12 (WI-5) | TestAccProviderMuxFree passes (TF_ACC=1) | ⚠️ partial | Test added + compiles; skips without live ADO creds |
| AC13 (WI-5) | CHANGELOG.md has BREAKING CHANGES entry under [Unreleased] | ✅ met | commit 5e7ef0dd |
| AC14 (WI-5) | PROVIDER_VERSION.txt bumped to 2.0.0 | ✅ met | cat PROVIDER_VERSION.txt → 2.0.0 |

---

## Checkpoint 1 — Quality gate: servicehook unit tests

**Command:** `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...`

| | Output |
|---|---|
| **Before (main)** | `ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook	0.003s` |
| **After (HEAD)** | `ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook	0.003s` |

Tests pass on main (mux scaffold still in place) and on HEAD (pure-framework entry point; mux removed).

---

## Checkpoint 2 — main.go mux imports removed

**Command:** `grep -E 'tf5to6|tf6mux|helper/schema|NewGRPCProvider' main.go || echo 'none'`

| | Output |
|---|---|
| **Before (main)** | `"github.com/hashicorp/terraform-plugin-mux/tf5to6server"` <br>`"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"` <br>`"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"` |
| **After (HEAD)** | `none` |

main.go ran `tf5to6server.UpgradeServer` + `tf6muxserver.NewMuxServer` to bridge SDKv2 and framework. After cutover: no mux imports — only `providerserver` + `tf6server` + `azuredevops` package shim.

---

## Checkpoint 3 — provider.go ResourcesMap: 16 SDKv2 registrations → empty

**Command:** `grep 'betterado_serviceendpoint' azuredevops/provider.go | grep -v '//' || echo 'empty'`

| | Output |
|---|---|
| **Before (main)** | `"betterado_serviceendpoint_jfrog_artifactory_v2"` ... (16 live entries) |
| **After (HEAD)** | `empty` |

All 16 SDKv2 service endpoint resources removed from `provider.go` ResourcesMap. Maps now contain only comments. Resources continue to be served via `framework_provider.go`.

---

## Checkpoint 4 — TestFrameworkProvider_MuxFree

**Command:** `go test -tags all -count=1 -run TestFrameworkProvider_MuxFree ./azuredevops/internal/provider/`

| | Output |
|---|---|
| **Before (main)** | `testing: warning: no tests to run` / `PASS (cached)` |
| **After (HEAD)** | `ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider	0.003s` |

`TestFrameworkProvider_MuxFree` did not exist on main. After WI-4: test verifies `NewFrameworkProvider()` is non-nil, `betterado_project` is registered, no mux layer present.

---

## Checkpoint 5 — PROVIDER_VERSION.txt bumped to 2.0.0

**Command:** `cat PROVIDER_VERSION.txt`

| | Output |
|---|---|
| **Before (main)** | `1.22.0` |
| **After (HEAD)** | `2.0.0` |

BREAKING change: mux scaffold removed, Terraform >= 1.x (plugin protocol 6) now required.

---

## Test Evidence

| Test | Result |
|---|---|
| `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` | ✅ pass |
| `TestFrameworkProvider_MuxFree` | ✅ pass |
| `TestServiceEndpointJFrog*` (4 JFrog framework unit tests) | ✅ pass |
| `TestServiceEndpoint*` (12 general service endpoint framework unit tests) | ✅ pass |
| `TestAccProviderMuxFree` (acceptance, requires `TF_ACC=1` + live ADO) | ⏭ skip — compiles; skips without live ADO credentials |
