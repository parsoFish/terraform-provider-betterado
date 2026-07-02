# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete — 2025-07-03)

**Task**: Migrate 7 repository policy SDKv2 resources to terraform-plugin-framework.

**Done:**
1. Created `framework_helpers.go` with shared helpers (ImportState, Bool defaults, scope expand/flatten, plan modifiers)
2. Created 7 `*_framework.go` files — one per resource (author_email_patterns, check_credentials, enforce_consistent_case, file_path_patterns, max_file_size, max_path_length, reserved_names)
3. Registered all 7 in `framework_provider.go` Resources(); removed from `provider.go` SDKv2 ResourcesMap; removed unused import
4. Updated `provider_test.go`: removed 7 `betterado_repository_policy_*` entries — `TestProvider_HasChildResources` PASSES
5. Updated 7 acceptance test files:
   - Removed `betterado_project` resource creation
   - Changed `projectName` → `projectID := SharedFixtureProjectID(t)`
   - Changed `Providers: testutils.GetProviders()` → `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
   - HCL templates now only create a `betterado_git_repository` in the shared project

**Commit**: `feat: migrate 7 repository policy resources to terraform-plugin-framework`

## What worked

- Pattern from branch policy framework migration: same `configure → expand → CRUD → flatten` pattern, adapted for repo policy scope structure
- `SharedFixtureProjectID(t)` + `GetMuxedProviderFactories()` pattern (same as branch policies) to avoid project creation failures
- `json.Marshal(pc.Settings)` + `json.Unmarshal` to a `map[string]interface{}` for safe settings extraction from the API response
- `boolPtr()` helper to create `*bool` values for the policy configuration
- Empty `repository_ids` → project-wide policy via `scope: [{"repositoryId": ""}]`

## What didn't work

- **`Providers: testutils.GetProviders()` with framework resources**: Framework resources only work with `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`. Don't mix.
- **Creating `betterado_project` in tests**: Org is at 1000-project cap — tests fail at create. Always use `SharedFixtureProjectID(t)`.
- **Duplicating resources in both SDKv2 and framework provider**: Causes "Invalid Provider Server Combination" at plan time.

## Key implementation details

### Policy settings scope structure (repo policies)

```go
// project-wide policy (no repo_ids)
"scope": [{"repositoryId": ""}]

// per-repo policy
"scope": [{"repositoryId": "<uuid1>"}, {"repositoryId": "<uuid2>"}]
```

### max_file_size unit conversion
- API field: `maximumGitBlobSizeInBytes`
- TF field: `max_file_size` (in MB)
- Conversion: `1024 * 1024` (constant `maxFileSizeUnitBytes`)
- Valid values in TF: 1, 2, 5, 10, 50, 100, 200

### Import ID format
`<project_id>/<policy_id>` — `importRepoPolicyState()` in framework_helpers.go

### Policy type UUIDs
Defined as `uuid.UUID` variables in `common.go`:
- `AuthorEmailPattern`, `FilePathPattern`, `CaseEnforcement`, `CheckCredentials`, `ReservedNames`, `PathLength`, `FileSize`

## Iteration 1 fix (2026-07-02)

**Gate failure**: `TestAccRepositoryPolicyFileSize/ProjectPolicies/*` and `TestAccRepositoryPolicyPathLength/ProjectPolicies/*` failed with:
```
Received unknown value, however the target type cannot handle unknown values.
Path:  Target Type: []string  Suggested Type: basetypes.ListValue
```

**Root cause**: `repository_ids` is `Optional: true, Computed: true` with **no Default**. When the user omits it from config, terraform-plugin-framework marks the plan value as **unknown** (will-be-computed by provider). Then `expandRepositoryIDs` calls `repoIDs.ElementsAs(ctx, &ids, false)` on the unknown list, which fails.

**Fix**: Added `emptyRepoPolicyListDefault` type (implements `defaults.List`) to `framework_helpers.go`. Added `Default: emptyRepoPolicyList()` to `repository_ids` in all 7 resource schemas. Now when omitted from config, plan value is `[]` (known empty list) instead of unknown.

**KEY PATTERN**: For `Optional+Computed` list/set attributes that default to "empty" when not set: ALWAYS add a `Default:` that returns the empty value. Without it, Terraform marks planned value as unknown, causing `ElementsAs` failures. This applies even when the empty default is semantically equivalent to the computed value.

## Iteration 2 fix (2026-07-02)

**Gate failure**: `TestAccRepositoryPolicyFileSize` and `TestAccRepositoryPolicyPathLength` (ProjectPolicies/basic and ProjectPolicies/update) failed with:
```
Error: Provider produced inconsistent result after apply
.repository_ids: was cty.ListValEmpty(cty.String), but now null.
```

**Root cause**: In `flattenRepositoryIDs`, when the scope from the API is `[{"repositoryId": ""}]` (project-wide policy), the empty repositoryId is filtered out, leaving `ids = nil` (nil slice). Then `types.ListValueFrom(ctx, types.StringType, nil)` returns a **null** list, not an empty list. Terraform sees state had `[]` (empty list), read-back returns `null` — inconsistent result error.

**Fix**: Added `if ids == nil { ids = []string{} }` guard in `flattenRepositoryIDs` before calling `types.ListValueFrom`. This ensures read-back always returns a known empty list `[]` for project-wide policies, consistent with the planned value from `emptyRepoPolicyList()`.

**KEY PATTERN**: When using `types.ListValueFrom` with a potentially nil slice, ALWAYS guard with explicit empty slice initialization. `nil` → `null list`, `[]string{}` → `known empty list`. This is critical for `Optional+Computed` attributes that use `Default: emptyList()` in the schema.

## Iteration 3 fix (2026-07-02)

**Gate failure**: `TestAccRepositoryPolicyCheckCredentials` (all 4 sub-tests) failed with:
```
Error creating repository policy check credentials
Type with id 'e67ae10f-cf9a-40bc-8e66-6b3a8216956e' does not exist.
```

**Root cause**: ADO has **completely removed** the `check_credentials` policy type from the live service. The Create API call returns a 404-class error for this policy type. The resource has a `DeprecationMessage` in the SDKv2 code saying it can't be created, and our suspicion from iteration 2 was confirmed.

**Fix**: Added `t.Skip(...)` at the top of `TestAccRepositoryPolicyCheckCredentials` with a clear message explaining the ADO policy type removal. The resource (`betterado_repository_policy_check_credentials`) is kept in the provider for import/read/destroy of any pre-existing policies, but the acceptance test for create cannot pass in the live ADO environment.

**KEY PATTERN**: When an ADO feature is deprecated and the API type is removed from the live service, acceptance tests MUST be skipped with `t.Skip()` — they will always fail with "Type with id '...' does not exist." The resource itself should be retained (with deprecation warning in schema) so users can still manage any pre-existing instances via import.

## Open questions

- Will the remaining 6 TestAccRepositoryPolicy* tests pass? Gate will confirm.

## Notes for reflection

- The repository policy framework migration followed the same pattern as branch policies (WI-1/WI-2). The key difference: repo policies use a flat `scope` list (repository IDs) rather than `{repository_id, repository_ref, match_type}` scopes.
- The test pattern of using `SharedFixtureProjectID` + creating only a git repo (not a project) is now established for all policy tests.
