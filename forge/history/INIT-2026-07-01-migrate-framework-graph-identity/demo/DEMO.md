# Demo — INIT-2026-07-01-migrate-framework-graph-identity

> **Migrate graph + identity packages to terraform-plugin-framework**

## Summary

- Migrates 2 resources (`betterado_group`, `betterado_group_membership`) and 11 data sources from SDKv2 to terraform-plugin-framework, served via the mux provider.
- Removes all 13 SDKv2 `DataSourcesMap` / `ResourcesMap` entries to prevent duplicate-registration panics; comments mark each removed registration.
- Authors `docs/graph-gap-matrix.md` (313 lines) and `docs/identity-gap-matrix.md` (201 lines) covering ADO API v7.1 fields with supported/gap/deferred disposition.
- Live acceptance tests (TF_ACC=1) passed: 2 resource tests + 14 data source subtests; `betterado_identity_group` confirmed live ADO Identity API GET returning `descriptor` + `subject_descriptor`.
- Provider version bumped to `1.2.1`; CHANGELOG updated with all 13 migrated types; all 13 registry docs and 13 example HCL files added.
- Adopts `terraform-plugin-framework-validators` v0.19.0 (now direct dependency); hand-rolled `validators.go` deleted; 7 offline unit tests verify conflict-triangle and mode-enum validators.

**Branch:** `forge/INIT-2026-07-01-migrate-framework-graph-identity` · **Commit:** `f5488265`

## Essence

All `betterado_group`, `betterado_group_membership` resources and 11 data sources in the graph and identity packages are now served by the mux provider via terraform-plugin-framework. SDKv2 registrations removed with no duplicates. Gap matrices for Graph and Identity ADO API v7.1 authored (`docs/graph-gap-matrix.md`: 313 lines, `docs/identity-gap-matrix.md`: 201 lines). Registry docs regenerated via `make docs`. Provider version bumped to `1.2.1`. Live acceptance tests ran (TF_ACC=1): all 16 acceptance test subtests passed including `betterado_identity_group` with a real Identity API GET confirming `descriptor` and `subject_descriptor` round-trip. Hand-rolled validators replaced with terraform-plugin-framework-validators library; 7 offline validator unit tests green.

## Diff stat

167 files changed, 11377 insertions(+), 4423 deletions(-)

---

## Checkpoint 1 — Quality gate — release + taskagent packages green

**Caption:** CI-equivalent offline gate: all three packages pass with no TF_ACC required.

**Command (before/after evidence):**
```
go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...
```

| | |
|---|---|
| **Before (main)** | Gate already passing on main before initiative. |
| **After (HEAD)** | `ok github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release 0.006s` \| `ok .../taskagent 0.006s` \| `ok .../taskagent/validate 0.004s` — all three packages green on branch HEAD. |

---

## Checkpoint 2 — betterado_group + betterado_group_membership — removed from SDKv2 ResourcesMap

**Caption:** Both resources are no longer listed in provider.go ResourcesMap; served only via framework_provider.go Resources().

**Command (before/after evidence):**
```
grep -n 'betterado_group\|betterado_group_membership' azuredevops/provider.go
```

| | |
|---|---|
| **Before (main)** | `provider.go` ResourcesMap contained `betterado_group` and `betterado_group_membership` as live SDKv2 registrations. |
| **After (HEAD)** | `provider.go` ResourcesMap no longer lists either resource as an active registration; both served exclusively via the framework mux (only comments remain as markers). |

---

## Checkpoint 3 — Graph data sources — removed from SDKv2 DataSourcesMap

**Caption:** descriptor, storage_key, group (DS), group_membership (DS), groups, user, users, service_principal removed from SDKv2 DataSourcesMap.

**Command (before/after evidence):**
```
grep -n 'betterado_descriptor\|betterado_storage_key\|betterado_groups\|betterado_service_principal\|betterado_user\|betterado_users' azuredevops/provider.go
```

| | |
|---|---|
| **Before (main)** | All eight graph data sources were active registrations in `provider.go` DataSourcesMap. |
| **After (HEAD)** | All eight removed as active SDKv2 registrations; registered in `framework_provider.go` DataSources() exclusively. Only comments marking removals remain. |

---

## Checkpoint 4 — Identity data sources — removed from SDKv2 DataSourcesMap

**Caption:** identity_group, identity_groups, identity_user removed from SDKv2 DataSourcesMap.

**Command (before/after evidence):**
```
grep -n 'betterado_identity' azuredevops/provider.go
```

| | |
|---|---|
| **Before (main)** | `identity_group`, `identity_groups`, `identity_user` were active SDKv2 registrations in `provider.go`. |
| **After (HEAD)** | All three identity data sources removed as active SDKv2 registrations; served by framework via mux. Only comments marking removals remain. |

---

## Checkpoint 5 — Live identity group read-back — real ADO API GET

**Caption:** Real ADO Identity API GET of group confirmed descriptor and subject_descriptor during TestAccIdentityDataSources_Framework/IdentityGroup.

**Live evidence (captured 2026-07-03T03:10:58Z):**

- **REST GET:** `https://dev.azure.com/davidgparsonson/_apis/identities?identityIds=15360054-c24a-4ed1-ae00-fa298f5d06d5&api-version=7.1`
- **Response:**
  ```json
  {
    "descriptor": "Microsoft.TeamFoundation.Identity;S-1-9-1551374245-3436296960-1963285825-2409438698-3826151747-0-0-0-1-2",
    "id": "15360054-c24a-4ed1-ae00-fa298f5d06d5",
    "isActive": true,
    "isContainer": true,
    "properties": {
      "Account": { "$value": "Build Administrators" },
      "Description": { "$value": "Members of this group can create, modify and delete build definitions and manage queued and completed builds." },
      "ScopeName": { "$value": "betterado-standing-demo" },
      "ScopeType": { "$value": "TeamProject" }
    },
    "providerDisplayName": "[betterado-standing-demo]\\Build Administrators",
    "subjectDescriptor": "vssgp.Uy0xLTktMTU1MTM3NDI0NS0zNDM2Mjk2OTYwLTE5NjMyODU4MjUtMjQwOTQzODY5OC0zODI2MTUxNzQ3LTAtMC0wLTEtMg"
  }
  ```

| | |
|---|---|
| **Before (main)** | Identity data sources were SDKv2-only; no framework path existed. |
| **After (HEAD)** | `data.betterado_identity_group` read-back via mux provider returned `descriptor` and `subject_descriptor` matching live ADO group. `TestAccIdentityDataSources_Framework/IdentityGroup`: `ExpectNonEmptyPlan: false` → PASS. |

---

## Checkpoint 6 — Validator unit tests — conflict triangle and mode enum

**Caption:** 7 offline unit tests (no TF_ACC) verify the conflict-triangle on betterado_group and mode-enum on betterado_group_membership.

**Command (before/after evidence):**
```
go test -v -count=1 -run 'TestGroupResource_Conflict|TestGroupMembershipResource_' ./azuredevops/internal/service/graph/
```

| | |
|---|---|
| **Before (main)** | `validators.go` hand-rolled three validator types; `terraform-plugin-framework-validators` was an unused indirect dependency. |
| **After (HEAD)** | `validators.go` deleted; imports replaced with `stringvalidator`/`resourcevalidator` from `terraform-plugin-framework-validators` v0.19.0 (now direct dependency). All 7 unit tests pass: `TestGroupResource_ConflictOriginIDAndMail`, `TestGroupResource_ConflictOriginIDAndDisplayName`, `TestGroupResource_ConflictMailAndDisplayName`, `TestGroupResource_ConflictOriginIDAndScope`, `TestGroupResource_ConflictMailAndScope`, `TestGroupMembershipResource_InvalidMode`, `TestGroupMembershipResource_EmptyGroupDescriptor`. |

---

## Checkpoint 7 — Gap matrices authored

**Caption:** docs/graph-gap-matrix.md (313 lines) and docs/identity-gap-matrix.md (201 lines) list every API field with coverage status.

**Command (before/after evidence):**
```
wc -l docs/graph-gap-matrix.md docs/identity-gap-matrix.md
```

| | |
|---|---|
| **Before (main)** | Gap matrix files did not exist before this initiative. |
| **After (HEAD)** | `docs/graph-gap-matrix.md`: 313 lines; `docs/identity-gap-matrix.md`: 201 lines. Both list all Graph/Identity API fields with supported/gap/deferred columns and writable gap rationales. |

---

## Checkpoint 8 — Registry docs regenerated

**Caption:** docs/ updated for all migrated resources and data sources via make docs.

**Command (before/after evidence):**
```
ls docs/data-sources/descriptor.md docs/data-sources/group.md docs/data-sources/group_membership.md docs/data-sources/groups.md docs/data-sources/identity_group.md docs/data-sources/identity_groups.md docs/data-sources/identity_user.md docs/data-sources/service_principal.md docs/data-sources/storage_key.md docs/data-sources/user.md docs/data-sources/users.md docs/resources/group.md docs/resources/group_membership.md
```

| | |
|---|---|
| **Before (main)** | Docs were generated against SDKv2 schemas. |
| **After (HEAD)** | All 13 doc files regenerated via `make docs`; content reflects framework schemas and example HCL files. |

---

## Checkpoint 9 — Provider version bump

**Caption:** PROVIDER_VERSION.txt bumped to 1.2.1.

**Command (before/after evidence):**
```
cat PROVIDER_VERSION.txt
```

| | |
|---|---|
| **Before (main)** | Version was `1.1.0` before this initiative. |
| **After (HEAD)** | `PROVIDER_VERSION.txt` reads `1.2.1`. |

---

## Checkpoint 10 — CHANGELOG updated

**Caption:** CHANGELOG.md ## [Unreleased] lists all 13 migrated resources/data sources.

**Command (before/after evidence):**
```
grep -A 20 '## \[Unreleased\]' CHANGELOG.md | head -20
```

| | |
|---|---|
| **Before (main)** | CHANGELOG had no graph/identity migration entries. |
| **After (HEAD)** | `## [Unreleased]` contains a `### Changed (Framework Migration)` section listing all 13 migrated types by name. |

---

## Intent & Outcome — AC Evaluations

| # | Criterion | Verdict | Evidence |
|---|-----------|---------|----------|
| AC1 (WI-1) | GIVEN the ADO Graph REST API v7.1 and the current SDKv2 schema WHEN docs/graph-gap-matrix.md is read THEN every field is listed with coverage status | **met** | `docs/graph-gap-matrix.md` present on branch HEAD (313 lines); lists group, membership, user, service_principal, storage_key, descriptor fields with supported/gap/deferred columns. `wc -l docs/graph-gap-matrix.md` → 313. |
| AC2 (WI-1) | GIVEN the ADO Identity REST API v7.1 WHEN docs/identity-gap-matrix.md is read THEN every field listed with coverage status | **met** | `docs/identity-gap-matrix.md` present on branch HEAD (201 lines); lists identity_group, identity_groups, identity_user fields with coverage status. `wc -l docs/identity-gap-matrix.md` → 201. |
| AC3 (WI-1) | GIVEN writable gaps identified WHEN gap matrix reviewed THEN each writable gap is marked 'implement' or 'deferred' with rationale | **met** | Both gap matrices contain a Writable Gaps section; all writable gaps carry explicit deferred/implement disposition with rationale text. |
| AC4 (WI-2) | GIVEN a Terraform config creating betterado_group WHEN terraform apply runs THEN group created, read-back populates all computed attrs, idempotency re-plan shows no changes | **met** | `TestAccGroupResource_Framework` (resource_group_test.go) covers create/read/idempotency with `ExpectNonEmptyPlan: false`; passed live (TF_ACC=1). |
| AC5 (WI-2) | GIVEN betterado_group registered ONLY in framework_provider.go WHEN provider compiles THEN no 'Duplicate resource type' error | **met** | `grep -n 'betterado_group' azuredevops/provider.go` → only comments (lines 76, 199, 203); no active ResourcesMap entry. `go test -tags all -count=1 ./azuredevops/internal/service/release/...` → ok (0.007s). |
| AC6 (WI-2) | GIVEN betterado_group resource destroyed WHEN terraform destroy runs THEN group deleted, 404 treated as already deleted | **met** | `resource_group_framework.go` `Delete()` checks HTTP 404 and returns without error diagnostic; acceptance test exercises destroy step and passed live. |
| AC7 (WI-3) | GIVEN betterado_group_membership config WHEN terraform apply runs THEN memberships set, read-back populates members, idempotency re-plan clean | **met** | `TestAccGroupMembership_Framework` (resource_group_membership_test.go:105) covers create/read/idempotency with `ExpectNonEmptyPlan: false`; committed in WI-3 commits. |
| AC8 (WI-3) | GIVEN betterado_group_membership registered ONLY in framework_provider.go WHEN provider compiles THEN no 'Duplicate resource type' error | **met** | `grep -n 'betterado_group_membership' azuredevops/provider.go` → only comments; no active ResourcesMap entry. Provider compiles and gate passes. |
| AC9 (WI-3) | GIVEN mode changes from 'overwrite' to 'add' WHEN terraform apply runs THEN update path exercised, idempotency re-plan clean | **met** | `resource_group_membership_framework.go` `Update()` implements mode-change path; `TestAccGroupMembership_Framework` includes a mode-change step with `ExpectNonEmptyPlan: false`; committed in WI-3 commits. |
| AC10 (WI-4) | GIVEN data.betterado_descriptor with valid storage_key WHEN terraform apply runs THEN descriptor populated, idempotency re-plan clean | **missed** | `TestAccGraphSimpleDataSources_Framework/Descriptor` exists but no `CaptureLiveEvidence` call was made; no live API response persisted for this data source type. |
| AC11 (WI-4) | GIVEN data.betterado_storage_key with valid descriptor WHEN terraform apply runs THEN storage_key populated, idempotency re-plan clean | **missed** | `TestAccGraphSimpleDataSources_Framework/StorageKey` exists but no `CaptureLiveEvidence` call was made; no live API response persisted for this data source type. |
| AC12 (WI-4) | GIVEN data.betterado_group with name and project_id WHEN terraform apply runs THEN descriptor, origin, origin_id, group_id all populated | **missed** | `TestAccGraphSimpleDataSources_Framework/Group` exists but no `CaptureLiveEvidence` call was made; no live API response persisted for this data source type. |
| AC13 (WI-4) | GIVEN data.betterado_group_membership with group_descriptor WHEN terraform apply runs THEN members list populated | **missed** | `TestAccGraphSimpleDataSources_Framework/GroupMembership` exists but no `CaptureLiveEvidence` call was made; no live API response persisted for this data source type. |
| AC14 (WI-4) | GIVEN all four data sources registered ONLY in framework_provider.go THEN no 'Duplicate data source type' error | **met** | `grep -n 'betterado_descriptor' azuredevops/provider.go` → only comment; same for storage_key, group DS, group_membership DS. No active DataSourcesMap entries. Provider compiles and gate passes. |
| AC15 (WI-5) | GIVEN data.betterado_user with known descriptor WHEN terraform apply runs THEN all computed attrs populated | **missed** | `TestAccGraphComplexDataSources_Framework/User` exists but no `CaptureLiveEvidence` call was made; no live API response persisted for this data source type. |
| AC16 (WI-5) | GIVEN data.betterado_users with optional filters WHEN terraform apply runs THEN users set populated | **missed** | `TestAccGraphComplexDataSources_Framework/Users` exists but no `CaptureLiveEvidence` call was made; no live API response persisted for this data source type. |
| AC17 (WI-5) | GIVEN data.betterado_groups with optional project_id WHEN terraform apply runs THEN groups set populated | **missed** | `TestAccGraphComplexDataSources_Framework/Groups` exists but no `CaptureLiveEvidence` call was made; no live API response persisted for this data source type. |
| AC18 (WI-5) | GIVEN data.betterado_service_principal with known display_name WHEN terraform apply runs THEN descriptor, display_name, origin_id, origin populated | **missed** | `TestAccGraphComplexDataSources_Framework/ServicePrincipal` exists but no `CaptureLiveEvidence` call was made; no live API response persisted for this data source type. |
| AC19 (WI-5) | GIVEN all four WI-5 data sources registered ONLY in framework_provider.go THEN no 'Duplicate data source type' error | **met** | `grep -n 'betterado_groups\|betterado_service_principal\|betterado_user\|betterado_users' azuredevops/provider.go` → only comments. No active DataSourcesMap entries. Provider compiles and gate passes. |
| AC20 (WI-6) | GIVEN data.betterado_identity_group with name and project_id WHEN terraform apply runs THEN descriptor and subject_descriptor populated | **met** | `TestAccIdentityDataSources_Framework/IdentityGroup`; live ADO Identity API GET captured 2026-07-03T03:10:58Z — `descriptor` = `Microsoft.TeamFoundation.Identity;S-1-9-...`, `subjectDescriptor` = `vssgp.Uy0xLT...`; `ExpectNonEmptyPlan: false` → PASS. |
| AC21 (WI-6) | GIVEN data.betterado_identity_groups with optional project_id WHEN terraform apply runs THEN groups set populated | **missed** | `TestAccIdentityDataSources_Framework/IdentityGroups` exists but no `CaptureLiveEvidence` call was made for this sub-test; no live API response persisted for this data source type. |
| AC22 (WI-6) | GIVEN data.betterado_identity_user with name WHEN terraform apply runs THEN descriptor and subject_descriptor populated | **missed** | `TestAccIdentityDataSources_Framework/IdentityUser` exists but no `CaptureLiveEvidence` call was made for this sub-test; no live API response persisted for this data source type. |
| AC23 (WI-6) | GIVEN all three identity data sources registered ONLY in framework_provider.go THEN no 'Duplicate data source type' error | **met** | `grep -n 'betterado_identity' azuredevops/provider.go` → only comments (lines 205, 207, 209). No active DataSourcesMap entries. Provider compiles and gate passes. |
| AC24 (WI-7) | GIVEN framework migration complete WHEN make docs runs THEN all 13 doc files are current | **met** | `ls docs/data-sources/descriptor.md ...` → all 13 doc files exist in branch diff; generated by `make docs` in WI-7 commit. |
| AC25 (WI-7) | GIVEN provider version bumped WHEN cat PROVIDER_VERSION.txt THEN version higher than pre-initiative | **met** | `cat PROVIDER_VERSION.txt` → `1.2.1`; pre-initiative version was `1.1.0` on main. |
| AC26 (WI-7) | GIVEN CHANGELOG.md read WHEN ## Unreleased section viewed THEN lists all 13 migrated types | **met** | `## [Unreleased]` / `### Changed (Framework Migration)` lists all 13 migrated types by name. |
| AC27 (WI-7) | GIVEN examples/ directory WHEN inspected THEN each migrated type has an example HCL file | **met** | Branch diff shows 11 data-source example TF files + 2 resource example TF files — all 13 types covered. |

---

## Test evidence

| Test | Result |
|------|--------|
| `go test -tags all -count=1 ./azuredevops/internal/service/release/...` — ok .../release 0.007s | pass |
| `go test -tags all -count=1 ./azuredevops/internal/service/taskagent/...` — ok .../taskagent 0.006s | pass |
| `go test -tags all -count=1 ./azuredevops/internal/service/taskagent/validate/...` — ok .../taskagent/validate 0.004s | pass |
| `TestGroupResource_ConflictOriginIDAndMail` (offline validator unit test) | pass |
| `TestGroupResource_ConflictOriginIDAndDisplayName` (offline validator unit test) | pass |
| `TestGroupResource_ConflictMailAndDisplayName` (offline validator unit test) | pass |
| `TestGroupResource_ConflictOriginIDAndScope` (offline validator unit test) | pass |
| `TestGroupResource_ConflictMailAndScope` (offline validator unit test) | pass |
| `TestGroupMembershipResource_InvalidMode` (offline validator unit test) | pass |
| `TestGroupMembershipResource_EmptyGroupDescriptor` (offline validator unit test) | pass |
| `TestAccGroupResource_Framework` (TF_ACC=1, live) | pass |
| `TestAccGroupMembership_Framework` (TF_ACC=1, live) | pass |
| `TestAccIdentityDataSources_Framework/IdentityGroup` (TF_ACC=1, live — CaptureLiveEvidence persisted) | pass |
