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

- Resources: `betterado_release_definition`, `betterado_release_environment`
- Files: `resource_release_definition.go`
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
- On 409 (conflict): retry with backoff for concurrent modification

## Development Workflow

### Building

```bash
make build                     # Build the provider
make install                   # Build and install locally
make test                      # Run unit tests
make testacc                   # Run acceptance tests (needs TF_ACC=1)
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
├── environments[]
│   ├── preDeployApprovals
│   │   └── approvers[]
│   ├── postDeployApprovals
│   │   └── approvers[]
│   ├── preDeploymentGates
│   ├── postDeploymentGates
│   ├── deployPhases[]
│   │   └── workflowTasks[]
│   ├── variables{}
│   └── environmentTriggers[]
├── artifacts[]
│   └── definitionReference{}
├── triggers[]
├── variables{}
└── retentionPolicy
```

Handle this by building expand/flatten helpers for each nested level, composing them bottom-up.

### State Management

- `revision` field is critical — updates require the current revision number
- Always read-after-write to get computed fields
- Environment IDs are assigned server-side on creation

### go.mod replace directive

The upstream provider uses a `replace` directive for `microsoft/azure-devops-go-api` pointing to `magodo/azure-devops-go-api`. This is an upstream choice — keep it.

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `AZDO_ORG_SERVICE_URL` | ADO org URL (e.g., `https://dev.azure.com/myorg`) | Yes |
| `AZDO_PERSONAL_ACCESS_TOKEN` | PAT with release management scope | Yes |
| `AZDO_TEST_PROJECT` | Project name for acceptance tests | For tests |

## Skills Available

- **ado-api-explorer** — Systematically discover and test ADO API endpoints
- **ado-browser-inspector** — Capture and analyze network traces from ADO UI in Chrome (requires Cowork with Chrome MCP)
- **resource-scaffolder** — Generate Terraform resource boilerplate from API specs
