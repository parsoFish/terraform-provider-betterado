# Demo — INIT-2026-07-01-migrate-framework-workitemtrackingprocess

> **Migrate workitemtrackingprocess package (13 resources, 4 data sources) to terraform-plugin-framework**

## Essence

All workitemtrackingprocess resources and data sources (process, workitemtype, state, inherited_state, page, inherited_page, list, field, rule, control, group, inherited_control, system_control, and the four process/workitemtype data sources) are now served by the mux provider via terraform-plugin-framework implementations. The public HCL surface is identical; no behavioural changes. The provider compiles cleanly (go build ./... exit 0; go vet -tags all exit 0). Live TF_ACC runs were completed for process, workitemtype, state, page, list, and rule resources (provenance-clean captures committed). The group, control, inherited_control, system_control, field, inherited_page, and inherited_state acceptance tests were not run live in this rework cycle (no provenance-clean captures available); those test results are marked missed below.

## Diff stat

188 files changed, 13972 insertions(+), 8123 deletions(-)

---

## Checkpoint 1 — Offline quality gate

**Caption:** Offline quality gate: go build ./... and go vet -tags all on workitemtrackingprocess/provider/acceptancetests packages

**Command (before/after evidence):**
```
go build ./... && go vet -tags all ./azuredevops/internal/service/workitemtrackingprocess/... ./azuredevops/internal/provider/... ./azuredevops/internal/acceptancetests/...
```

| | |
|---|---|
| **Before (main)** | Provider compiles and vet passes on main |
| **After (HEAD)** | `go build ./...` exits 0; `go vet -tags all` on workitemtrackingprocess/provider/acceptancetests packages exits 0 after workitemtrackingprocess migration |

---

## Checkpoint 2 — Provider compiles cleanly

**Caption:** Provider binary builds cleanly with zero compilation errors

**Command (before/after evidence):**
```
go build -mod=vendor .
```

| | |
|---|---|
| **Before (main)** | Provider compiles on main with SDKv2 workitemtrackingprocess registrations |
| **After (HEAD)** | Provider compiles on HEAD with all 13 framework resources and 4 data sources registered in `framework_provider.go`; deregistered from `provider.go` |

---

## Checkpoint 3 — Gap matrix produced

**Caption:** docs/workitemtrackingprocess-gap-matrix.md produced, covering all 13 resources and 4 data sources

**Command (before/after evidence):**
```
wc -l docs/workitemtrackingprocess-gap-matrix.md
```

| | |
|---|---|
| **Before (main)** | File absent on main |
| **After (HEAD)** | 413-line gap matrix present; lists every ADO API v7.1 field per resource with `gap_action` (resolved/deferred/n-a-computed) |

---

## Checkpoint 4 — Framework resources registered

**Caption:** All 13 resources and 4 data sources registered in framework_provider.go (not provider.go)

**Command (before/after evidence):**
```
grep -c 'workitemtrackingprocess' azuredevops/internal/provider/framework_provider.go
```

| | |
|---|---|
| **Before (main)** | 0 workitemtrackingprocess entries in framework_provider.go |
| **After (HEAD)** | 17 workitemtrackingprocess entries present (13 resources + 4 data sources) |

---

## Intent & Outcome — AC Evaluations

| # | Criterion | Verdict | Evidence |
|---|-----------|---------|----------|
| AC-1a | GIVEN ADO Work Item Tracking Process REST API v7.1 schema WHEN compared against SDKv2 schemas THEN `docs/workitemtrackingprocess-gap-matrix.md` lists every API field for all 13 resources and 4 data sources | **met** | `docs/workitemtrackingprocess-gap-matrix.md` present (413 lines); covers all 13 resources and 4 data sources with columns: field, api_writable, in_schema, gap_action |
| AC-2a | GIVEN process resource migrated WHEN TF_ACC tests run live THEN TestAccWorkitemtrackingprocessProcess_Basic, CreateDisabled, CreateAndUpdate all pass | **met** | `resource_process_framework.go` committed; per-WI live gate: TestAccWorkitemtrackingprocessProcess_Basic/CreateDisabled/CreateAndUpdate → pass (provenance-clean capture at `.forge/live-evidence/acceptance-resource-workitemtrackingprocess-process.json`) |
| AC-2b | GIVEN process/processes data sources migrated WHEN TF_ACC tests run live THEN TestAccWorkitemtrackingprocessProcess_DataSource_Get and TestAccWorkitemtrackingprocessProcesses_DataSource_AllProcesses pass | **met** | `data_process_framework.go` and `data_processes_framework.go` committed; live gate: both tests → pass |
| AC-2c | GIVEN SDKv2 process files deregistered WHEN provider.go inspected THEN process registered ONLY in framework_provider.go; provider_test.go counts updated | **met** | `framework_provider.go` carries all three constructors; `provider.go` no longer references them; `provider_test.go` updated |
| AC-2d | GIVEN live acceptance test WHEN process read back before destroy THEN CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-process", …) called | **met** | CaptureLiveEvidence call present in `resource_workitemtrackingprocess_process_test.go` Basic step; provenance-clean capture written |
| AC-3a | GIVEN workitemtype resource migrated WHEN TF_ACC tests run live THEN TestAccWorkitemtrackingprocessWorkItemType_Basic and CreateAndUpdate pass | **met** | `resource_work_item_type_framework.go` committed; per-WI live gate: Basic/CreateAndUpdate → pass (provenance-clean capture at `.forge/live-evidence/acceptance-resource-workitemtrackingprocess-workitemtype.json`) |
| AC-3b | GIVEN workitemtype/workitemtypes data sources migrated WHEN TF_ACC tests run live THEN DataSource_Get and DataSource_List pass | **met** | Framework data source files committed; live gate: both tests → pass |
| AC-3c | GIVEN SDKv2 workitemtype files deregistered WHEN provider.go inspected THEN workitemtype registered ONLY in framework_provider.go | **met** | `framework_provider.go` carries all three workitemtype constructors; `provider.go` deregistered |
| AC-3d | GIVEN live acceptance test WHEN workitemtype read back THEN CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-workitemtype", …) called | **met** | CaptureLiveEvidence call present in workitemtype acceptance test Basic step; provenance-clean capture written |
| AC-4a | GIVEN state resource migrated WHEN TF_ACC tests run live THEN TestAccWorkitemtrackingprocessState_Basic and Update pass | **met** | `resource_state_framework.go` committed; per-WI live gate: Basic/Update → pass (provenance-clean capture at `.forge/live-evidence/acceptance-resource-workitemtrackingprocess-state.json`) |
| AC-4b | GIVEN inherited_state resource migrated WHEN TF_ACC tests run live THEN Basic, Update, and RemoveFromState all pass | **missed** | `resource_inherited_state_framework.go` committed (full CRUD + ImportState). Live acceptance gate not run in this rework cycle — no provenance-clean capture file available for TestAccWorkitemtrackingprocessInheritedState_Basic/Update/RemoveFromState. |
| AC-4c | GIVEN SDKv2 state files deregistered WHEN provider.go inspected THEN both registered ONLY in framework_provider.go | **met** | `framework_provider.go` carries both; `provider.go` deregistered |
| AC-4d | GIVEN live acceptance test WHEN state read back THEN CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-state", …) called | **met** | CaptureLiveEvidence call present in state acceptance test Basic step; provenance-clean capture written |
| AC-5a | GIVEN page resource migrated WHEN TF_ACC tests run live THEN TestAccWorkitemtrackingprocessPage_Basic and Update pass | **met** | `resource_page_framework.go` committed; per-WI live gate: Basic/Update → pass (provenance-clean capture at `.forge/live-evidence/acceptance-resource-workitemtrackingprocess-page.json`) |
| AC-5b | GIVEN inherited_page resource migrated WHEN TF_ACC tests run live THEN Basic, Update, and Revert all pass | **missed** | `resource_inherited_page_framework.go` committed (full CRUD + ImportState). Live acceptance gate not run in this rework cycle — no provenance-clean capture file available for TestAccWorkitemtrackingprocessInheritedPage_Basic/Update/Revert. |
| AC-5c | GIVEN SDKv2 page files deregistered WHEN provider.go inspected THEN both registered ONLY in framework_provider.go | **met** | `framework_provider.go` carries both page constructors; `provider.go` deregistered |
| AC-5d | GIVEN live acceptance test WHEN page read back THEN CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-page", …) called | **met** | CaptureLiveEvidence call present in page acceptance test Basic step; provenance-clean capture written |
| AC-6a | GIVEN group resource migrated WHEN TF_ACC tests run live THEN Basic, Update, Move, WithMultipleControlTypes all pass | **missed** | `resource_group_framework.go` committed (full CRUD + ImportState). Live acceptance gate not run in this rework cycle — no provenance-clean capture file available. Resource was delivered in UWI-2 rework, not WI-6 original delivery. |
| AC-6b | GIVEN control resource migrated WHEN TF_ACC tests run live THEN Basic, Update, Move, Contribution all pass | **missed** | `resource_control_framework.go` committed (688 lines, full CRUD + ImportState). Live acceptance gate not run in this rework cycle — no provenance-clean capture file available. Resource was delivered in UWI-2 rework, not WI-6 original delivery. |
| AC-6c | GIVEN inherited_control resource migrated WHEN TF_ACC tests run live THEN Basic, Update, Revert all pass | **missed** | `resource_inherited_control_framework.go` committed (404 lines, full CRUD + ImportState). Live acceptance gate not run in this rework cycle — no provenance-clean capture file available. Resource was delivered in UWI-2 rework, not WI-6 original delivery. |
| AC-6d | GIVEN system_control resource migrated WHEN TF_ACC tests run live THEN Basic, Update, Revert all pass | **missed** | `resource_system_control_framework.go` committed (402 lines, full CRUD + ImportState). Live acceptance gate not run in this rework cycle — no provenance-clean capture file available. Resource was delivered in UWI-2 rework, not WI-6 original delivery. |
| AC-6e | GIVEN all four SDKv2 resource files deregistered WHEN provider.go inspected THEN all four registered ONLY in framework_provider.go | **met** | `framework_provider.go` carries all four; `provider.go` deregistered; `provider_test.go` updated |
| AC-6f | GIVEN live acceptance test WHEN group read back THEN CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-group", …) called | **missed** | CaptureLiveEvidence call is present in `resource_workitemtrackingprocess_group_test.go` Basic test step. However, no provenance-clean capture was produced — the previously committed capture file carried a future-dated capturedAt timestamp (2026-07-05T10:00:00Z, 34+ hours after its mtime) and was therefore fabricated; file has been deleted. |
| AC-7a | GIVEN list resource migrated WHEN TF_ACC tests run live THEN Basic, Update, Integer all pass | **met** | `resource_list_framework.go` committed; per-WI live gate: Basic/Update/Integer → pass (provenance-clean capture at `.forge/live-evidence/acceptance-resource-workitemtrackingprocess-list.json`) |
| AC-7b | GIVEN field resource migrated WHEN TF_ACC tests run live THEN Basic, Identity, Integer, Update all pass | **missed** | `resource_field_framework.go` committed (full CRUD + ImportState). Live acceptance gate not run in this rework cycle — no provenance-clean capture file available for TestAccWorkitemtrackingprocessField_Basic/Identity/Integer/Update. |
| AC-7c | GIVEN SDKv2 list/field files deregistered WHEN provider.go inspected THEN both registered ONLY in framework_provider.go | **met** | `framework_provider.go` carries both constructors; `provider.go` deregistered |
| AC-7d | GIVEN live acceptance test WHEN list read back THEN CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-list", …) called | **met** | CaptureLiveEvidence call present in list acceptance test Basic step; provenance-clean capture written |
| AC-8a | GIVEN rule resource migrated WHEN TF_ACC tests run live THEN Basic, Update, ConditionTypes, ConditionGroupMembership, ActionTypes, HideTargetField, DisallowValue all pass | **met** | `resource_rule_framework.go` committed (with `SetNestedBlock` fix for HCL block syntax); per-WI live gate: all 7 → pass (provenance-clean capture at `.forge/live-evidence/acceptance-resource-workitemtrackingprocess-rule.json`) |
| AC-8b | GIVEN SDKv2 rule file deregistered WHEN provider.go inspected THEN rule registered ONLY in framework_provider.go | **met** | `framework_provider.go` carries the rule constructor; `provider.go` deregistered |
| AC-8c | GIVEN live acceptance test WHEN rule read back THEN CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-rule", …) called | **met** | CaptureLiveEvidence call present in rule acceptance test Basic step; provenance-clean capture written |
| AC-9a | GIVEN all resources migrated WHEN make docs runs THEN all 13 resource docs and 4 data-source docs regenerated; guides restored; examples exist | **met** | Branch diff: all 13 resource docs + 4 data-source docs updated; 13 resource examples + 4 data-source examples added; commit 'docs: regenerate registry docs and add examples for workitemtrackingprocess migration close-out' |
| AC-9b | GIVEN migration complete WHEN CHANGELOG.md and PROVIDER_VERSION.txt inspected THEN CHANGELOG.md has ## Unreleased entry for all 13 resources and 4 data sources; version bumped | **met** | `CHANGELOG.md` `## [Unreleased]` has ENHANCEMENTS entry listing all 13 resources and 4 data sources; `PROVIDER_VERSION.txt` bumped |
| AC-9c | GIVEN provider compiles WHEN go build -mod=vendor . runs THEN exits 0; offline vet gate passes | **met** | `go build -mod=vendor .` → exit 0; `go vet -tags all` on workitemtrackingprocess/provider/acceptancetests packages → exit 0 |

---

## Test evidence

| Test | Result |
|------|--------|
| `go build ./... && go vet -tags all ./azuredevops/internal/service/workitemtrackingprocess/... ./azuredevops/internal/provider/... ./azuredevops/internal/acceptancetests/...` (offline gate) | pass |
| `go build -mod=vendor .` | pass |
| `TestAccWorkitemtrackingprocessProcess_Basic` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessProcess_CreateDisabled` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessProcess_CreateAndUpdate` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessProcess_DataSource_Get` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessProcesses_DataSource_AllProcesses` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessWorkItemType_Basic` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessWorkItemType_CreateAndUpdate` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessWorkItemType_DataSource_Get` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessWorkItemTypes_DataSource_List` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessState_Basic` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessState_Update` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessInheritedState_Basic` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessInheritedState_Update` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessInheritedState_RemoveFromState` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessPage_Basic` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessPage_Update` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessInheritedPage_Basic` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessInheritedPage_Update` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessInheritedPage_Revert` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessGroup_Basic` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessGroup_Update` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessGroup_Move` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessGroup_WithMultipleControlTypes` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessControl_Basic` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessControl_Update` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessControl_Move` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessControl_Contribution` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessInheritedControl_Basic` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessInheritedControl_Update` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessInheritedControl_Revert` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessSystemControl_Basic` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessSystemControl_Update` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessSystemControl_Revert` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessList_Basic` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessList_Update` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessList_Integer` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessField_Basic` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessField_Identity` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessField_Integer` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessField_Update` (not-run — no provenance-clean capture) | missed |
| `TestAccWorkitemtrackingprocessRule_Basic` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessRule_Update` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessRule_ConditionTypes` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessRule_ConditionGroupMembership` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessRule_ActionTypes` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessRule_HideTargetField` (TF_ACC=1, live, per-WI) | pass |
| `TestAccWorkitemtrackingprocessRule_DisallowValue` (TF_ACC=1, live, per-WI) | pass |
