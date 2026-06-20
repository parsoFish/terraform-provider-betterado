# Framework Migration Guard

## What it does

`check.sh` scans every Go file under `azuredevops/internal/service/**` for
resource constructor functions and cross-references each against the two
provider registration sites:

- `azuredevops/provider.go` — `ResourcesMap` (SDK v2 resources)
- `azuredevops/internal/provider/framework_provider.go` — `Resources()` (Plugin Framework resources)

Any constructor found in the service tree but absent from **both** registration
sites is printed as an **ORPHAN** and the script exits non-zero.

Patterns detected:

| Constructor form | Meaning |
|---|---|
| `func ResourceXxx() *schema.Resource` | SDK v2 resource (or data source helper — see note below) |
| `func NewXxxResource() ...` | Plugin Framework resource |

> **Data source note:** `DataXxx()` constructors are intentionally excluded
> from this check — they live in `provider.go`'s `DataSourcesMap`, not
> `ResourcesMap`, and are not part of the migration path.

## When to run

- After migrating a resource from SDK v2 to Plugin Framework.
- Before opening a migration PR to confirm no SDK v2 dead code was left behind.
- As a pre-merge sanity check when touching `provider.go` or `framework_provider.go`.

```bash
# Run from repo root:
bash forge/skills/framework-migration-guard/check.sh
```

Exit 0 = clean. Exit 1 = orphans found (output includes file paths).

## How to read the output

A clean run:
```
framework-migration-guard: OK — no orphaned resource constructors found.
```

A failing run:
```
framework-migration-guard: FAIL — orphaned resource constructors detected:

  ORPHAN: ResourceReleaseDefinition  (azuredevops/internal/service/release/resource_release_definition.go)

These constructors are defined in service/ but registered in NEITHER
  azuredevops/provider.go (ResourcesMap)
  azuredevops/internal/provider/framework_provider.go (Resources())
```

Each ORPHAN line shows the function name and the file where it lives.

## How to safely remove an orphan

1. **Delete the dead resource file** — `resource_<name>.go`. This is the SDK v2
   implementation replaced by `resource_<name>_framework.go`.

2. **Delete the in-package unit test** — `resource_<name>_test.go` (the file
   with build tag `resource_<name>`). Do NOT delete acceptance tests under
   `azuredevops/internal/acceptancetests/` — those test the live framework
   resource and must stay.

3. **Check for shared helpers.** Before deleting, grep the rest of the package
   for any function defined in the dead file that is called from sibling files:

   ```bash
   # Identify functions defined in the dead file
   grep -E "^func " azuredevops/internal/service/<pkg>/resource_<name>.go

   # Check for callers in other package files
   grep -rn "<functionName>" azuredevops/internal/service/<pkg>/ \
     --include="*.go" | grep -v resource_<name>.go
   ```

   If a function is used elsewhere (e.g. `flattenTaskGroup` was called from
   `data_task_group.go`), **move just that function** (and its direct helpers)
   into an appropriately named file in the same package — `shared_helpers.go`
   is the convention used here. Do not restore the whole dead resource file.

4. **Check for shared test fixtures.** If the deleted `_test.go` declared
   package-level `var testXxx` fixtures that other test files in the same
   package use, move them into a `test_fixtures_test.go` file with a matching
   build tag.

5. **Run `gofmt -w`** on any relocated Go file.

6. **Verify:**

   ```bash
   go build -mod=vendor .
   go test -tags all -count=1 ./azuredevops/internal/service/<pkg>/...
   bash forge/skills/framework-migration-guard/check.sh
   ```

   All three must be clean before opening a PR.

## Optional deeper pass

The script only catches constructors that are **defined but not registered**.
It does not catch dead helper functions inside a registered file. For that:

```bash
# Unused exported/unexported identifiers (requires full build — disk-heavy):
golangci-lint run --enable=unused ./azuredevops/internal/service/<pkg>/...

# Whole-program dead code analysis (Go 1.21+):
go install golang.org/x/tools/cmd/deadcode@latest
deadcode -test ./...
```

**Disk discipline:** `go build ./...` or `go vet ./...` on the whole tree
generates a multi-GB build cache and can fill the drive. Always scope build and
vet commands to the package under change: `go build -mod=vendor .` (entry point
only) and `go test ./azuredevops/internal/service/<pkg>/...` (targeted package).
See `CLAUDE.md` — "Disk discipline" note.
