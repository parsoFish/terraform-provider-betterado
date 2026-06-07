# Add data.betterado_release_folder — read-only lookup for release folders

> _Derived from `demo.json` (ADR 021). Essence:_ Before this change, Terraform configs that did not own a release folder could not reference one. There was no data source — only the resource. After this change, data.betterado_release_folder resolves a folder by project_id + path via the SDK GetFolders call, surfaces its description (and any other API fields), and is registered in the provider so configs can import folder attributes without managing lifecycle.

## Summary

- Adds data.betterado_release_folder — read-only lookup by project_id + path via SDK GetFolders
- Registered in provider DataSourcesMap; provider count test updated
- 2 new unit tests (Read_Populates, Read_NotFound) + 1 acceptance test (TestAccDataReleaseFolder_Basic)
- Docs page and example HCL included
- Quality gate green: 3 packages, 63 tests, 0 failures

## Test Evidence

### Two new unit tests cover the data source read path: success (GetFolders returns a folder) and not-found (empty slice returns error).

- **Before:** 61 tests in the release+taskagent packages; data_release_folder.go did not exist; no data source was registered.
- **After:** 63 tests pass (2 new: TestDataReleaseFolder_Read_Populates, TestDataReleaseFolder_Read_NotFound). Release package: ok in 0.019s. Taskagent unchanged. Zero failures.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| release package — test count | 35 | 37 | +5.7% | match |
| taskagent package — test count | 26 | 26 | 0.0% | match |
| overall gate test count | 61 | 63 | +3.3% | match |
| gate result | ok (no data source file) | ok 0.019s + ok 0.007s + ok 0.004s | 0.0% | match |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### betterado_release_folder now appears in the provider DataSourcesMap; provider count test passes with the updated list.

- **Before:** Provider had 2 release data sources (betterado_release_definition, betterado_release_definitions). betterado_release_folder was absent.
- **After:** Provider has 3 release data sources. TestProvider_HasChildDataSources passes with betterado_release_folder in the expectedDataSources slice and the count assertion satisfied.

### Live ADO acceptance test creates a release folder via resource, reads it back via data source, asserts description match, verifies idempotency, then destroys.

- **Before:** No acceptance test existed for release folder data source lookup; data source didn't exist.
- **After:** TestAccDataReleaseFolder_Basic: resource creates folder → data source reads by project_id+path → description attribute matches → re-plan produces no diff (ExpectNonEmptyPlan: false) → destroy is clean. File: azuredevops/internal/acceptancetests/data_release_folder_test.go.

## API / Behaviour Diff

### data.betterado_release_folder (added)

**Before:**
```
(did not exist)
```
**After:**
```
Required: project_id (string, UUID), path (string). Computed: description (string). Read via SDK GetFolders(project, path).
```

## Test Evidence

| test | result | delta |
|---|---|---|
| TestDataReleaseFolder_Read_Populates | pass | new |
| TestDataReleaseFolder_Read_NotFound | pass | new |
| TestProvider_HasChildDataSources | pass | updated (count +1) |
| TestAccDataReleaseFolder_Basic | pass | new (TF_ACC, live ADO) |
| release package — all pre-existing tests | pass | unchanged |
| taskagent package — all tests | pass | unchanged |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Acceptance criteria

- AC1 (WI-1): dataReleaseFolderRead sets Id=path and populates description when GetFolders returns a matching folder
- AC2 (WI-1): dataReleaseFolderRead returns an error containing the path when GetFolders returns an empty slice
- AC3 (WI-2): betterado_release_folder appears in the provider DataSourcesMap; TestProvider_HasChildDataSources passes with updated count
- AC4 (WI-2): TestAccDataReleaseFolder_Basic creates a folder via resource, reads it back via data source, asserts description match, idempotency re-plan, clean destroy (TF_ACC gate)
- AC5 (WI-2): examples/data-sources/betterado_release_folder/main.tf shows minimal usage; docs/resources/release_folder.md documents the data source

## Files Changed

- `azuredevops/internal/service/release/data_release_folder.go` — New: DataReleaseFolder() schema.Resource + dataReleaseFolderRead implementation
- `azuredevops/internal/service/release/data_release_folder_test.go` — New: unit tests — Read_Populates and Read_NotFound
- `azuredevops/provider.go` — Added betterado_release_folder to DataSourcesMap
- `azuredevops/provider_test.go` — Added betterado_release_folder to expectedDataSources slice
- `azuredevops/internal/acceptancetests/data_release_folder_test.go` — New: TestAccDataReleaseFolder_Basic acceptance test
- `examples/data-sources/betterado_release_folder/main.tf` — New: minimal usage example
- `docs/resources/release_folder.md` — Added Data Source section documenting data.betterado_release_folder

```
.../acceptancetests/data_release_folder_test.go    | 66 ++++++++++++++++
 .../service/release/data_release_folder.go         | 58 ++++++++++++++
 .../service/release/data_release_folder_test.go    | 88 ++++++++++++++++++++++
 azuredevops/provider.go                            |  1 +
 azuredevops/provider_test.go                       |  1 +
 docs/resources/release_folder.md                   | 77 +++++++++++++++++++
 .../data-sources/betterado_release_folder/main.tf  | 12 +++
 7 files changed, 303 insertions(+)
```

## Usage

```
```hcl
# Look up an existing release folder by path.
# The folder must already exist (managed elsewhere or created via
# betterado_release_folder resource).

data "betterado_project" "example" {
  name = "MyProject"
}

data "betterado_release_folder" "example" {
  project_id = data.betterado_project.example.id
  path       = "\\MyFolder"
}

output "folder_description" {
  value = data.betterado_release_folder.example.description
}
```
```

## Impact

- Configs that do not own a release folder can now reference one by path — closing the resource/data-source parity gap for release folders.
- Enables cross-stack references: a separate Terraform root managing deployments can look up a folder created by an infra root.
- Mirrors the existing betterado_release_definition / betterado_release_definitions data-source pattern, making the provider API consistent.
- Unit tests enforce both the happy path and the not-found path, preventing silent regressions on the SDK GetFolders call.
