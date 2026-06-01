# Add betterado_release_folder resource for organizing Azure DevOps release definitions

> _Derived from `demo.json` (ADR 021). Essence:_ A new Terraform resource betterado_release_folder is now available, enabling practitioners to create, update, and delete release-definition folders in Azure DevOps. Before this change there was no Terraform-managed way to organise release definitions into folders; after it, the full CRUD lifecycle is covered with import support.

## All 5 canonical unit tests for the release_folder resource pass against the gomock-backed fake Release API client.

- **Before:** No resource_release_folder package existed; no tests existed.
- **After:** 5 tests pass: ExpandFlatten_Roundtrip, Create_DoesNotSwallowError, Read_ClearsIdOn404, Update_CallsSDKWithArgs, Delete_SurfacesAPIError.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestReleaseFolder_ExpandFlatten_Roundtrip | N/A (file did not exist) | PASS | — | new |
| TestReleaseFolder_Create_DoesNotSwallowError | N/A (file did not exist) | PASS | — | new |
| TestReleaseFolder_Read_ClearsIdOn404 | N/A (file did not exist) | PASS | — | new |
| TestReleaseFolder_Update_CallsSDKWithArgs | N/A (file did not exist) | PASS | — | new |
| TestReleaseFolder_Delete_SurfacesAPIError | N/A (file did not exist) | PASS | — | new |

## Acceptance criteria

- AC1: Resource is registered and builds — go build -mod=vendor ./... succeeds
- AC2: Schema matches build_folder pattern — path (Required, ForceNew), project_id (Required, ForceNew), name (Computed)
- AC3: All 5 canonical unit tests pass
- AC4: Documentation exists at docs/resources/release_folder.md

## Changed files

```
 azuredevops/internal/service/release/resource_release_folder.go      | 151 +++++++++++++++
 azuredevops/internal/service/release/resource_release_folder_test.go | 202 +++++++++++++++++++++
 azuredevops/provider.go                                              |   1 +
 docs/resources/release_folder.md                                     |  45 +++++
 examples/release_folder/main.tf                                      |  28 +++
 5 files changed, 427 insertions(+)
```
