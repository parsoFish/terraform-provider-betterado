# Better ADO Provider

## Project Overview

This is a **GitHub fork** of `microsoft/terraform-provider-azuredevops`. It inherits the full official provider (100+ resources) while adding resources Microsoft has not implemented, particularly **classic release pipelines**. The fork is designed to pull upstream changes via `git merge upstream/main` and potentially contribute back.

The official provider covers build definitions, repos, service endpoints, policies, and permissions, but has **zero** support for classic release pipelines despite a fully-featured REST API at `vsrm.dev.azure.com`. This fork adds that support.

## Architecture

- **Language:** Go (Terraform Plugin SDK v2)
- **Go module:** `github.com/parsoFish/terraform-provider-betterado`
- **Provider name:** `betterado` (resource prefix: `betterado_*`)
- **API base (core):** `https://dev.azure.com/{org}/{project}/_apis/` (inherited)
- **API base (release):** `https://vsrm.dev.azure.com/{org}/{project}/_apis/release/` (new)
- **API version:** 7.1
- **Auth:** Personal Access Token (PAT) via `AZDO_PERSONAL_ACCESS_TOKEN` env var

## Fork Workflow

- **`main` branch** — tracks upstream `microsoft/terraform-provider-azuredevops`
- **`betterado` branch** — all fork-specific work (release resources, renames, etc.)
- **`upstream` remote** — `https://github.com/microsoft/terraform-provider-azuredevops.git`
- **`origin` remote** — will be set to `github.com/parsoFish/terraform-provider-betterado` once created

### Pulling upstream changes

```bash
git checkout main
git fetch upstream
git merge upstream/main
git checkout betterado
git merge main
# Resolve conflicts: module path renames + resource prefix changes
```

> **Note:** ~450 files have import path renames (`microsoft/` → `parsoFish/`) and all resources use `betterado_` prefix. Upstream merges will produce conflicts in these files. A merge helper script is a future improvement.

## Reference Material

- **Official provider codemap:** `docs/official-provider-codemap.md`
- **API validation findings:** `docs/api-reference/api-validation-findings.md`
- **Captured API response:** `docs/api-reference/captured-definition-response.json`
- **Release API reference:** `docs/api-reference/release-definitions.md`

## Directory Structure

```
azuredevops/                   # Provider package (upstream structure)
├── provider.go                # Provider registration (all resources)
├── provider_test.go           # Provider tests
├── internal/
│   ├── client/                # ADO API client wrappers (SDK-based)
│   ├── service/               # Resource implementations
│   │   ├── build/             # Build definitions
│   │   ├── core/              # Projects, teams
│   │   ├── git/               # Repos, branches
│   │   ├── release/           # NEW: Classic release definitions
│   │   ├── serviceendpoint/   # Service endpoints
│   │   └── ...                # 20+ other service packages
│   ├── acceptancetests/       # Acceptance tests (TF_ACC=1)
│   └── utils/                 # Shared helpers
├── utils/                     # Additional utilities
docs/
├── api-reference/             # ADO REST API mapping docs
├── resources/                 # Terraform resource documentation
├── contributing.md            # Upstream contribution guide
└── official-provider-codemap.md  # Upstream structural analysis
examples/                      # Example Terraform configurations
scripts/                       # Build, test, and API helper scripts
.claude/skills/                # Agent skills for development workflow
vendor/                        # Vendored Go dependencies
```

## Conventions

### Resource Implementation Pattern

Every resource follows this pattern (matching the official provider):

1. **Schema definition** — `resource_<name>.go` with `*schema.Resource` return
2. **CRUD functions** — `Create`, `Read`, `Update`, `Delete` with `Context` suffix
3. **Expand functions** — Convert Terraform state → API request objects
4. **Flatten functions** — Convert API response → Terraform state
5. **Tests** — Acceptance tests in `azuredevops/internal/acceptancetests/`

### Naming

- Resources: `betterado_release_definition`, `betterado_task_group`
- Files: `resource_release_definition.go`, `resource_task_group.go`
- Go functions: `ResourceReleaseDefinition()`, `expandReleaseDefinition()`, `flattenReleaseDefinition()`
- Test functions: `TestAccReleaseDefinition_Basic`, `TestAccReleaseDefinition_Complete`

### API Client Pattern

- Release resources use the official `release.Client` from `microsoft/azure-devops-go-api/azuredevops/v7/release`
- Client is initialized in `azuredevops/internal/client/client.go` alongside all other SDK clients
- The release API runs on `vsrm.dev.azure.com` (different host) — the SDK handles this
- Handle 404 gracefully in Read (clear resource ID, don't error)

### Error Handling

- Wrap API errors with context: `fmt.Errorf("creating release definition: %w", err)`
- On 404 in Read: `d.SetId("")` and return nil (resource was deleted outside TF)
- On revision conflict in Update: API returns **HTTP 400** (not 409) with `typeKey: InvalidRequestException` and message containing "old copy of the release pipeline". The Update function detects this, re-reads the definition to get the current revision, and retries once with the fresh revision.

## Development Workflow

### Disk Space Warning

> **IMPORTANT for AI agents:** `go build ./...` compiles all packages including every
> test file and generates 2–4 GB of build cache in `%LOCALAPPDATA%\go-build`.
> Running it multiple times per session can fill a drive.
>
> **Always use targeted builds:**
> ```powershell
> # Verify compilation — builds entry point only (fast, ~50 MB cache delta)
> & "C:\Program Files\Go\bin\go.exe" build -mod=vendor .
>
> # Vet a specific package under development (not ./...)
> & "C:\Program Files\Go\bin\go.exe" vet -mod=vendor ./azuredevops/internal/service/release/...
>
> # Clean build cache when done with a session
> & "C:\Program Files\Go\bin\go.exe" clean -cache -testcache
> # or on Windows: scripts/clean-build-cache.ps1
> ```
>
> **Never run `go build ./...` or `go vet ./...` (full tree) unless explicitly asked.**
> One targeted `go build -mod=vendor .` at the end of a session is sufficient.

### Building

```bash
make build                     # Build the provider
make install                   # Build and install locally
make test                      # Run unit tests
make testacc                   # Run acceptance tests (needs TF_ACC=1)
make clean-cache               # Clear Go build/test cache to reclaim disk space
```

### Testing API calls

Use `scripts/ado-api.sh` to test API endpoints directly:
```bash
./scripts/ado-api.sh GET "release/definitions" --project MyProject
./scripts/ado-api.sh POST "release/definitions" --project MyProject --body @definition.json
```

### Discovering new API surfaces

1. Use the `ado-browser-inspector` skill to capture network traces from ADO UI
2. Use the `ado-api-explorer` skill to systematically test endpoints
3. Document findings in `docs/api-reference/`

### Adding a new resource

1. Use the `resource-scaffolder` skill to generate boilerplate
2. Map the API types in `docs/api-reference/`
3. Implement expand/flatten for the nested structure
4. Register in `azuredevops/provider.go`
5. Write acceptance tests in `azuredevops/internal/acceptancetests/`
6. Add example configurations

## Key Technical Notes

### Release API Host

The release API lives at a **different host** than the core ADO API:
- Core API: `https://dev.azure.com/{org}/`
- Release API: `https://vsrm.dev.azure.com/{org}/`

The official Go SDK handles the host routing via the `release.Client`.

### Nested Complexity

Release definitions are deeply nested:
```
ReleaseDefinition
├── Environments[]
│   ├── DeployPhases[]
│   │   ├── DeploymentInput
│   │   └── WorkflowTasks[]
│   ├── PreDeployApprovals
│   │   └── Approvals[] + ApprovalOptions
│   ├── PostDeployApprovals
│   ├── Conditions[]
│   ├── EnvironmentOptions
│   ├── ExecutionPolicy
│   ├── RetentionPolicy
│   └── Variables{}
├── Artifacts[]
│   └── DefinitionReference{}
├── Variables{}
└── VariableGroups[]
```

Each expand/flatten function handles one layer. The expand functions construct API request objects from Terraform state; flatten functions convert API responses into the `[]map[string]interface{}` structures Terraform expects.

### API-Computed Fields in Artifacts

The artifact `definition_reference` map comes back from the API with extra keys (e.g., `artifactSourceDefinitionUrl`, `defaultVersionSpecific`) that weren't in user config. The `flattenArtifacts` function filters these out, only persisting keys the user actually configured, preventing perpetual diff.

## Implemented Resource Status

### betterado_release_definition (feature-complete for P0/P1)

**File:** `azuredevops/internal/service/release/resource_release_definition.go` (~1,490 lines)

Fully tested with Create → Read → Update → Delete lifecycle. Confirmed idempotent (no drift on re-plan).

**Implemented features:**
- Environments with agent-based deployment phases
- Pre/post deploy approvals with approval_options (6 fields)
- Variables (plain, secret, allow_override) at definition + environment level
- Variable groups at definition + environment level
- Artifacts (Build type with definition_reference map)
- Workflow tasks (task + metaTask/task group references)
- Environment options (notifications, badge, auto-link, PR deploy)
- Execution policy (concurrency, queue depth)
- Retention policy per environment
- Environment conditions (event trigger, environment state dependency)
- Revision-aware update with retry on stale revision (HTTP 400)
- 404 tolerance in Read (graceful external delete handling)
- Import support (projectID/definitionID format)
- Agent specification (free-form string, future-proof)

**Known limitations:**
- Tags: API doesn't persist via definitions endpoint (separate API may be needed)

### betterado_task_group

**File:** `azuredevops/internal/service/taskagent/resource_task_group.go`

Full CRUD lifecycle for reusable task group definitions. Referenced from release definitions as `definition_type = "metaTask"` workflow tasks.

**Features:**
- Name, description, category, author, instance_name_format
- Version block (major ForceNew, minor, patch, is_test)
- Parameterized inputs (name, label, type, default_value, required, help, group_name)
- Task steps (task_id, version, display_name, inputs, condition, enabled, etc.)
- runs_on configuration
- Computed: id (UUID), revision, definition_type

**API host:** `dev.azure.com` (core host, uses TaskAgentClient)

## Roadmap

See `docs/feature-plan.md` for the full implementation plan and `docs/api-reference/` for API mappings. Key future features include agentless jobs, deployment group jobs, gates, and multi-config parallelism — detailed in repo memory at `/memories/repo/pr178-analysis-and-roadmap.md`.
make clean-cache               # Clear Go build/test cache to reclaim disk space
```

### Testing API calls

Use `scripts/ado-api.sh` to test API endpoints directly:
```bash
./scripts/ado-api.sh GET "release/definitions" --project MyProject
./scripts/ado-api.sh POST "release/definitions" --project MyProject --body @definition.json
```

### Discovering new API surfaces

1. Use the `ado-browser-inspector` skill to capture network traces from ADO UI
2. Use the `ado-api-explorer` skill to systematically test endpoints
3. Document findings in `docs/api-reference/`

### Adding a new resource

1. Use the `resource-scaffolder` skill to generate boilerplate
2. Map the API types in `docs/api-reference/`
3. Implement expand/flatten for the nested structure
4. Register in `azuredevops/provider.go`
5. Write acceptance tests in `azuredevops/internal/acceptancetests/`
6. Add example configurations

## Key Technical Notes

### Release API Host

The release API lives at a **different host** than the core ADO API:
- Core API: `https://dev.azure.com/{org}/`
- Release API: `https://vsrm.dev.azure.com/{org}/`

The official Go SDK handles the host routing via the `release.Client`.

### Nested Complexity

Release definitions are deeply nested:
```
ReleaseDefiniti
