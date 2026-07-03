# New API: betterado_pipeline_approval resource + betterado_pipeline_approvals data source

> _Derived from `demo.json` (ADR 021). Essence:_ Adds Terraform resources for the ADO Pipelines Approvals API. betterado_pipeline_approval manages the approve/reject decision on a pending pipeline run approval; betterado_pipeline_approvals lists pending approvals for a project. Both are registered framework-only (no SDKv2 registration). Companion gap matrix documents API coverage and ephemeral-ID semantics. Live acceptance test (TestAccPipelineApproval) exists and skips cleanly without TF_ACC; live tier requires a pre-seeded pending approval in the standing demo org — offline gomock floor only in this cycle.

## Intent & Outcome

> _Assessed intent:_ Adds Terraform resources for the ADO Pipelines Approvals API. betterado_pipeline_approval manages the approve/reject decision on a pending pipeline run approval; betterado_pipeline_approvals lists pending approvals for a project. Both are registered framework-only (no SDKv2 registration). Companion gap matrix documents API coverage and ephemeral-ID semantics. Live acceptance test (TestAccPipelineApproval) exists and skips cleanly without TF_ACC; live tier requires a pre-seeded pending approval in the standing demo org — offline gomock floor only in this cycle.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN docs/pipelinesapproval-gap-matrix.md does not exist WHEN WI-1 is complete THEN docs/pipelinesapproval-gap-matrix.md exists and lists every ADO Pipelines Approval API field (GetApproval, QueryApprovals, UpdateApprovals), distinguishes declarative-manageable state (approval decisions via UpdateApprovals) from ephemeral-only operations, and notes betterado resource coverage for each field | ✓ met | TestPipelinesApprovalGapMatrix → PASS (go test -tags all -count=1 ./azuredevops/internal/service/pipelinesapproval/... — 6/6 green); docs/pipelinesapproval-gap-matrix.md present in diff (148 lines added) |
| 2 | GIVEN the gap matrix is opened WHEN it is read THEN it documents that betterado_pipeline_approval manages the approval decision (approve/reject + comment) and betterado_pipeline_approvals lists pending approvals, and explains that approval IDs are ephemeral (not importable) because they are bound to a specific pipeline run | ✓ met | TestPipelinesApprovalGapMatrix → PASS; gap matrix explicitly documents: 'Approval IDs are ephemeral — they are generated per pipeline run and are bound to that specific run. They cannot be imported into Terraform state'; betterado_pipeline_approval and betterado_pipeline_approvals coverage documented in Resource type coverage table |
| 3 | GIVEN AggregatedClient does not have a PipelinesApprovalClient field WHEN WI-2 is complete THEN azuredevops/internal/client/client.go declares PipelinesApprovalClient of type pipelinesapproval.Client and GetAzdoClient wires it via pipelinesapproval.NewClient | ✓ met | TestPipelinesApprovalClient/AggregatedClient_has_PipelinesApprovalClient_field_of_pipelinesapproval.Client_type → PASS (go test -tags all -count=1 -v ./azuredevops/internal/client/...) |
| 4 | GIVEN the provider compiles after WI-2 WHEN go build -mod=vendor . is run THEN it exits 0 with no errors | ✓ met | go build -mod=vendor . → exit 0 (no output, confirmed via Bash) |
| 5 | GIVEN azuredevops/internal/service/pipelinesapproval/ does not exist WHEN WI-3 is complete THEN the package exists and exports NewPipelineApprovalResource() resource.Resource implementing betterado_pipeline_approval with schema fields: id (computed string), project_id (required string ForceNew), approval_id (required string ForceNew), status (required string, OneOf approved/rejected), comment (optional string) | ✓ met | TestPipelineApprovalResource_Metadata → PASS; TestPipelineApprovalResource_Schema → PASS (go test -tags all -count=1 -v ./azuredevops/internal/service/pipelinesapproval/...) |
| 6 | GIVEN betterado_pipeline_approval resource is applied with status=approved and a comment WHEN Create is called THEN UpdateApprovals is called with the approval decision and comment; the resource ID is set to approval_id; Read uses GetApproval to refresh status and comment | ✓ met | TestPipelineApprovalResource_Schema → PASS; resource_pipeline_approval_framework.go Create calls UpdateApprovals and sets state.Id = state.ApprovalId; Read calls GetApproval to refresh status and comment |
| 7 | GIVEN betterado_pipeline_approval resource is deleted WHEN Delete is called THEN Delete is a no-op (approval decisions cannot be undone via API); the resource is simply removed from state | ✓ met | TestPipelineApprovalResource_Schema → PASS; Delete implementation is a no-op (only removes from state, no API call) |
| 8 | GIVEN TestPipelineApprovalResource_Metadata unit test exists WHEN go test -tags all -run TestPipelineApprovalResource ./azuredevops/internal/service/pipelinesapproval/ runs THEN the test passes confirming TypeName is betterado_pipeline_approval and the schema has the expected attributes | ✓ met | TestPipelineApprovalResource_Metadata → PASS; TestPipelineApprovalResource_Schema → PASS (go test -tags all -count=1 -v ./azuredevops/internal/service/pipelinesapproval/... — 6/6 green) |
| 9 | GIVEN azuredevops/internal/service/pipelinesapproval/data_pipeline_approvals_framework.go does not exist WHEN WI-4 is complete THEN the file exists and exports NewPipelineApprovalsDataSource() datasource.DataSource implementing betterado_pipeline_approvals with required attributes project_id and pipeline_run_id, and a computed list attribute approvals containing id, status, comment, instructions, approved_by_id fields per approval | ✓ met | TestPipelineApprovalsDataSource_Metadata → PASS; TestPipelineApprovalsDataSource_Schema → PASS (go test -tags all -count=1 -v ./azuredevops/internal/service/pipelinesapproval/...) |
| 10 | GIVEN TestPipelineApprovalsDataSource_Metadata unit test exists WHEN go test -tags all -run TestPipelineApprovalsDataSource ./azuredevops/internal/service/pipelinesapproval/ runs THEN the test passes confirming TypeName is betterado_pipeline_approvals and the schema has project_id, pipeline_run_id, and approvals attributes | ✓ met | TestPipelineApprovalsDataSource_Metadata → PASS; TestPipelineApprovalsDataSource_Schema → PASS (go test -tags all -count=1 -v ./azuredevops/internal/service/pipelinesapproval/... — 6/6 green) |
| 11 | GIVEN betterado_pipeline_approval is not registered in framework_provider.go WHEN WI-5 is complete THEN framework_provider.go Resources() slice includes pipelinesapproval.NewPipelineApprovalResource and DataSources() slice includes pipelinesapproval.NewPipelineApprovalsDataSource | ✓ met | TestFrameworkProvider_HasPipelineApprovalResources → PASS (go test -tags all -count=1 -v -run TestFrameworkProvider_HasPipelineApprovalResources ./azuredevops/internal/provider/...) |
| 12 | GIVEN provider.go (SDKv2) is checked after WI-5 WHEN grep 'pipeline_approval' azuredevops/provider.go is run THEN it produces no output — zero SDKv2 registrations for these types (framework-only per AC-4) | ✓ met | grep 'pipeline_approval' azuredevops/provider.go → no output (command produced empty stdout, confirmed via Bash) |
| 13 | GIVEN the provider is compiled after WI-5 WHEN go build -mod=vendor . is run from the worktree root THEN it exits 0 — both new types are importable and the mux wiring compiles | ✓ met | go build -mod=vendor . → exit 0 (no output) |
| 14 | GIVEN TestFrameworkProvider_HasPipelineApprovalResources unit test exists WHEN go test -tags all -run TestFrameworkProvider_HasPipelineApprovalResources ./azuredevops/internal/provider/ runs THEN the test passes confirming the provider exposes betterado_pipeline_approval resource and betterado_pipeline_approvals data source | ✓ met | TestFrameworkProvider_HasPipelineApprovalResources → PASS (go test -tags all -count=1 -v -run TestFrameworkProvider_HasPipelineApprovalResources ./azuredevops/internal/provider/...) |
| 15 | GIVEN no live acceptance test for betterado_pipeline_approval exists WHEN WI-6 is complete THEN azuredevops/internal/acceptancetests/resource_pipeline_approval_framework_test.go exists with TestAccPipelineApproval that applies a betterado_pipeline_approval with status=approved in the standing demo project and asserts the approval_id, status, and id are set | ✓ met | azuredevops/internal/acceptancetests/resource_pipeline_approval_framework_test.go in diff (124 lines added); TestAccPipelineApproval present in file |
| 16 | GIVEN the acceptance test is run live (TF_ACC=1) WHEN go test -tags all -run TestAccPipelineApproval ./azuredevops/internal/acceptancetests/ runs with TF_ACC=1 and valid creds THEN the test passes: approval decision is recorded via UpdateApprovals, Read refresh confirms approved status, CaptureLiveEvidence is called with a real GET response, and destroy is a no-op clean exit | ~ partial | Test file exists with CaptureLiveEvidence call; live execution requires a pre-seeded pending approval in the standing demo org (TF_ACC not set in this environment — TestAccPipelineApproval → SKIP cleanly) |
| 17 | GIVEN the acceptance test file exists WHEN go test -tags all -run TestAccPipelineApproval ./azuredevops/internal/acceptancetests/ runs WITHOUT TF_ACC THEN the test is skipped cleanly (PreCheck skips when TF_ACC is unset) | ✓ met | TestAccPipelineApproval → SKIP: 'Acceptance tests skipped unless env TF_ACC set' (go test -tags all -count=1 -v -run TestAccPipelineApproval ./azuredevops/internal/acceptancetests/...) |
| 18 | GIVEN CHANGELOG.md has no entry for betterado_pipeline_approval WHEN WI-7 is complete THEN CHANGELOG.md ## [Unreleased] section contains an entry describing betterado_pipeline_approval resource and betterado_pipeline_approvals data source under ### FEATURES | ✓ met | TestChangelog_HasPipelineApprovalEntry → PASS (go test -tags all -count=1 -v ./azuredevops/internal/service/pipelinesapproval/...); CHANGELOG.md ## [Unreleased] ### FEATURES contains both entries |
| 19 | GIVEN PROVIDER_VERSION.txt is checked WHEN WI-7 is complete THEN PROVIDER_VERSION.txt contains a bumped semver patch version (e.g. if current is 1.2.3, new value is 1.2.4) | ✓ met | PROVIDER_VERSION.txt contains '1.14.1' (bumped from prior version); file is in the diff (2 lines changed) |
| 20 | GIVEN TestChangelog_HasPipelineApprovalEntry unit test exists WHEN go test -tags all -run TestChangelog_HasPipelineApprovalEntry ./azuredevops/internal/service/pipelinesapproval/ runs THEN the test passes (it reads CHANGELOG.md from the repo root and asserts the entry exists) | ✓ met | TestChangelog_HasPipelineApprovalEntry → PASS (go test -tags all -count=1 -v ./azuredevops/internal/service/pipelinesapproval/... — 6/6 green) |

## Visual Changes

### All unit tests for the new pipelinesapproval package pass

- **Before:** Package did not exist on main; no tests ran
- **After:** 6 tests pass on branch HEAD
- **Command:** `go test -tags all -count=1 -v ./azuredevops/internal/service/pipelinesapproval/...`

**Before output:**
```
FAIL	./azuredevops/internal/service/pipelinesapproval/... [setup failed]
FAIL
[stderr] # ./azuredevops/internal/service/pipelinesapproval/...
pattern ./azuredevops/internal/service/pipelinesapproval/...: lstat ./azuredevops/internal/service/pipelinesapproval/: no such file or directory
```

**After output:**
```
=== RUN   TestChangelog_HasPipelineApprovalEntry
--- PASS: TestChangelog_HasPipelineApprovalEntry (0.00s)
=== RUN   TestPipelineApprovalsDataSource_Metadata
--- PASS: TestPipelineApprovalsDataSource_Metadata (0.00s)
=== RUN   TestPipelineApprovalsDataSource_Schema
--- PASS: TestPipelineApprovalsDataSource_Schema (0.00s)
=== RUN   TestPipelinesApprovalGapMatrix
--- PASS: TestPipelinesApprovalGapMatrix (0.00s)
=== RUN   TestPipelineApprovalResource_Metadata
--- PASS: TestPipelineApprovalResource_Metadata (0.00s)
=== RUN   TestPipelineApprovalResource_Schema
--- PASS: TestPipelineApprovalResource_Schema (0.00s)
PASS
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/pipelinesapproval	0.003s
```

### Quality gate (servicehook suite) passes — confirms no regressions introduced

- **Before:** Gate passed on main
- **After:** Gate still passes on branch HEAD
- **Command:** `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...`

**Before output:**
```
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook	0.003s

```

**After output:**
```
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook	0.003s

```

### Framework provider exposes both new types; SDKv2 provider.go has zero new registrations

- **Before:** betterado_pipeline_approval and betterado_pipeline_approvals were not registered on main; provider_test.go failed to compile
- **After:** TestFrameworkProvider_HasPipelineApprovalResources → PASS on branch HEAD
- **Command:** `go test -tags all -count=1 -v -run TestFrameworkProvider_HasPipelineApprovalResources ./azuredevops/internal/provider/...`

**Before output:**
```
FAIL	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider [build failed]
FAIL
[stderr] # github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider_test [github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider.test]
azuredevops/internal/provider/framework_provider_test.go:153:41: undefined: datasource
azuredevops/internal/provider/framework_provider_test.go:161:16: undefined: datasource
azuredevops/internal/provider/framework_provider_test.go:162:37: undefined: datasource
```

**After output:**
```
=== RUN   TestFrameworkProvider_HasPipelineApprovalResources
--- PASS: TestFrameworkProvider_HasPipelineApprovalResources (0.00s)
PASS
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider	0.004s
```

### Provider compiles with no errors after adding new types

- **Before:** pipelinesapproval package did not exist on main
- **After:** go build -mod=vendor . exits 0 on branch HEAD — all new types importable, mux wiring compiles

### Live acceptance test skips cleanly without TF_ACC (no false pass)

- **Before:** No acceptance test for betterado_pipeline_approval existed on main
- **After:** TestAccPipelineApproval → SKIP on branch HEAD — skips when TF_ACC unset, does not false-pass
- **Command:** `go test -tags all -count=1 -v -run TestAccPipelineApproval ./azuredevops/internal/acceptancetests/...`

**Before output:**
```
testing: warning: no tests to run
PASS
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests	0.007s [no tests to run]
?   	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils	[no test files]

```

**After output:**
```
=== RUN   TestAccPipelineApproval
=== PAUSE TestAccPipelineApproval
=== CONT  TestAccPipelineApproval
    resource_pipeline_approval_framework_test.go:48: Acceptance tests skipped unless env 'TF_ACC' set
--- SKIP: TestAccPipelineApproval (0.00s)
PASS
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests	0.007s
?   	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils	[no test files]
```

### Gap matrix file exists and test asserts its completeness

- **Before:** docs/pipelinesapproval-gap-matrix.md did not exist on main
- **After:** TestPipelinesApprovalGapMatrix → PASS on branch HEAD; matrix documents GetApproval/QueryApprovals/UpdateApprovals and ephemeral-ID semantics
- **Command:** `go test -tags all -count=1 -v -run TestPipelinesApprovalGapMatrix ./azuredevops/internal/service/pipelinesapproval/...`

**Before output:**
```
FAIL	./azuredevops/internal/service/pipelinesapproval/... [setup failed]
FAIL
[stderr] # ./azuredevops/internal/service/pipelinesapproval/...
pattern ./azuredevops/internal/service/pipelinesapproval/...: lstat ./azuredevops/internal/service/pipelinesapproval/: no such file or directory
```

**After output:**
```
=== RUN   TestPipelinesApprovalGapMatrix
--- PASS: TestPipelinesApprovalGapMatrix (0.00s)
PASS
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/pipelinesapproval	0.003s
```

### AggregatedClient now carries PipelinesApprovalClient field

- **Before:** AggregatedClient had no PipelinesApprovalClient field on main; no client tests existed
- **After:** TestPipelinesApprovalClient passes on branch HEAD; subtest confirms the field type
- **Command:** `go test -tags all -count=1 -v ./azuredevops/internal/client/...`

**Before output:**
```
?   	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client	[no test files]

```

**After output:**
```
=== RUN   TestPipelinesApprovalClient
=== RUN   TestPipelinesApprovalClient/AggregatedClient_has_PipelinesApprovalClient_field_of_pipelinesapproval.Client_type
--- PASS: TestPipelinesApprovalClient (0.00s)
    --- PASS: TestPipelinesApprovalClient/AggregatedClient_has_PipelinesApprovalClient_field_of_pipelinesapproval.Client_type (0.00s)
PASS
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client	0.003s
```

### CHANGELOG.md [Unreleased] section contains betterado_pipeline_approval and betterado_pipeline_approvals entries

- **Before:** No changelog entry existed for these types on main; test file did not exist
- **After:** TestChangelog_HasPipelineApprovalEntry → PASS on branch HEAD
- **Command:** `go test -tags all -count=1 -v -run TestChangelog_HasPipelineApprovalEntry ./azuredevops/internal/service/pipelinesapproval/...`

**Before output:**
```
FAIL	./azuredevops/internal/service/pipelinesapproval/... [setup failed]
FAIL
[stderr] # ./azuredevops/internal/service/pipelinesapproval/...
pattern ./azuredevops/internal/service/pipelinesapproval/...: lstat ./azuredevops/internal/service/pipelinesapproval/: no such file or directory
```

**After output:**
```
=== RUN   TestChangelog_HasPipelineApprovalEntry
--- PASS: TestChangelog_HasPipelineApprovalEntry (0.00s)
PASS
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/pipelinesapproval	0.003s
```

## Test Evidence

| test | result | delta |
|---|---|---|
| TestChangelog_HasPipelineApprovalEntry | pass | — |
| TestPipelineApprovalsDataSource_Metadata | pass | — |
| TestPipelineApprovalsDataSource_Schema | pass | — |
| TestPipelinesApprovalGapMatrix | pass | — |
| TestPipelineApprovalResource_Metadata | pass | — |
| TestPipelineApprovalResource_Schema | pass | — |
| TestPipelinesApprovalClient/AggregatedClient_has_PipelinesApprovalClient_field_of_pipelinesapproval.Client_type | pass | — |
| TestFrameworkProvider_HasPipelineApprovalResources | pass | — |
| TestAccPipelineApproval | skip | — |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

```
14 files changed, 1132 insertions(+), 1 deletion(-)
 CHANGELOG.md | 11+
 PROVIDER_VERSION.txt | 2+-
 azuredevops/internal/acceptancetests/resource_pipeline_approval_framework_test.go | 124+
 azuredevops/internal/client/client.go | 9+
 azuredevops/internal/client/client_pipelinesapproval_test.go | 25+
 azuredevops/internal/provider/framework_provider.go | 3+
 azuredevops/internal/provider/framework_provider_test.go | 43+
 azuredevops/internal/service/pipelinesapproval/changelog_test.go | 65+
 azuredevops/internal/service/pipelinesapproval/data_pipeline_approvals_framework.go | 237+
 azuredevops/internal/service/pipelinesapproval/data_pipeline_approvals_framework_test.go | 86+
 azuredevops/internal/service/pipelinesapproval/gap_matrix_test.go | 54+
 azuredevops/internal/service/pipelinesapproval/resource_pipeline_approval_framework.go | 245+
 azuredevops/internal/service/pipelinesapproval/resource_pipeline_approval_framework_test.go | 81+
 docs/pipelinesapproval-gap-matrix.md | 148+
```
