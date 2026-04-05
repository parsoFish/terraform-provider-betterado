# Official Microsoft Terraform Azure DevOps Provider - Comprehensive Codemap

## Overview

The `microsoft/terraform-provider-azuredevops` provider is a mature, production-grade Terraform provider for Azure DevOps. It follows Terraform Plugin SDK v2 patterns and serves as the authoritative reference for building quality Terraform providers for Azure services.

**Repository:** https://github.com/microsoft/terraform-provider-azuredevops
**Language:** Go 1.19+
**Terraform SDK:** github.com/hashicorp/terraform-plugin-sdk/v2
**Provider Name:** `azuredevops`
**Official Docs:** https://registry.terraform.io/providers/microsoft/azuredevops/

---

## 1. ROOT DIRECTORY STRUCTURE (3+ Levels Deep)

```
terraform-provider-azuredevops/
├── .github/
│   └── workflows/
│       ├── release.yaml             # Release automation on tags
│       ├── tests.yaml               # CI: lint, unit tests, acceptance tests
│       ├── codeql-analysis.yaml     # Code scanning
│       └── goreleaser-config.yaml   # Cross-platform builds
├── .golangci.yml                    # Linting configuration
├── .gitignore                       # Git ignore patterns
├── go.mod                           # Module definition & dependencies
├── go.sum                           # Dependency checksums
├── GNUmakefile                      # Build, test, install targets
├── main.go                          # Provider entry point
├── internal/
│   ├── client/                      # HTTP clients and API wrappers
│   │   ├── client.go               # AggregatedClient definition
│   │   ├── connection.go           # Connection pooling & auth
│   │   ├── core_client.go          # Base HTTP transport
│   │   ├── graph_client.go         # Microsoft Graph API wrapper
│   │   ├── feed_client.go          # Artifact feeds
│   │   ├── build_client.go         # Build definitions & pipelines
│   │   ├── release_client.go       # Release definitions (if present)
│   │   ├── git_client.go           # Git repos & branches
│   │   ├── policy_client.go        # Branch policies
│   │   ├── queue_client.go         # Agent queues
│   │   ├── service_endpoint_client.go  # Service connections
│   │   ├── project_client.go       # Projects
│   │   ├── security_client.go      # Permissions & security
│   │   ├── test_client.go          # Test plans & runs
│   │   ├── wiki_client.go          # Wiki pages
│   │   ├── workitemtracking_client.go  # Work items
│   │   ├── taskagent_client.go     # Task agent endpoints
│   │   └── utils.go                # HTTP helpers (retries, errors)
│   ├── service/                     # Resource & data source logic
│   │   ├── agent/
│   │   │   ├── agent_queue.go
│   │   │   ├── agent_queue_test.go
│   │   │   ├── data_source_agent_queue.go
│   │   │   └── resource_agent_queue.go
│   │   ├── build/
│   │   │   ├── build_definition.go
│   │   │   ├── resource_build_definition.go
│   │   │   ├── resource_build_definition_test.go
│   │   │   ├── data_source_build_definition.go
│   │   │   └── [test fixtures]
│   │   ├── core/
│   │   │   ├── core.go
│   │   │   ├── resource_project.go
│   │   │   ├── resource_project_test.go
│   │   │   ├── data_source_project.go
│   │   │   └── project/ [nested resources]
│   │   ├── feed/
│   │   │   ├── feed.go
│   │   │   ├── resource_feed.go
│   │   │   └── ...
│   │   ├── git/
│   │   │   ├── git_repository.go
│   │   │   ├── resource_git_repository.go
│   │   │   ├── resource_git_repository_test.go
│   │   │   ├── resource_git_repository_policy.go
│   │   │   └── ...
│   │   ├── permission/
│   │   │   ├── permission.go
│   │   │   ├── resource_project_permissions.go
│   │   │   ├── resource_git_permissions.go
│   │   │   └── ...
│   │   ├── serviceendpoint/
│   │   │   ├── resource_service_endpoint.go
│   │   │   ├── resource_service_endpoint_test.go
│   │   │   ├── data_source_service_endpoint.go
│   │   │   ├── service_endpoint_types/
│   │   │   │   ├── azurerm.go
│   │   │   │   ├── docker.go
│   │   │   │   ├── github.go
│   │   │   │   ├── kubernetes.go
│   │   │   │   ├── npm.go
│   │   │   │   ├── ssh.go
│   │   │   │   └── ... (20+ service endpoint types)
│   │   │   └── ...
│   │   ├── taskagent/
│   │   │   └── ...
│   │   ├── test/
│   │   │   ├── resource_test_plan.go
│   │   │   ├── resource_test_variable.go
│   │   │   └── ...
│   │   ├── wiki/
│   │   │   └── ...
│   │   └── workitemtracking/
│   │       └── ...
│   ├── provider.go                  # Provider registration & schema
│   ├── utils/
│   │   ├── converter.go            # Type conversion utilities
│   │   ├── tfhelper.go             # Terraform SDK helpers
│   │   ├── validate.go             # Input validation
│   │   └── suppress.go             # Diff suppression rules
│   └── acceptancetest/             # Test helpers and fixtures
│       ├── testutils.go            # Common test functions
│       ├── tfexec.go               # Terraform execution wrapper
│       └── check.go                # Acceptance test assertions
├── docs/
│   ├── data-sources/               # Data source documentation
│   │   ├── agent_pool.md
│   │   ├── agent_queue.md
│   │   ├── build_definition.md
│   │   ├── git_repository.md
│   │   ├── project.md
│   │   ├── service_endpoint.md
│   │   └── ... (30+ data sources)
│   ├── resources/                  # Resource documentation
│   │   ├── agent_queue.md
│   │   ├── build_definition.md
│   │   ├── feed.md
│   │   ├── git_repository.md
│   │   ├── git_repository_policy.md
│   │   ├── group.md
│   │   ├── project.md
│   │   ├── project_features.md
│   │   ├── service_endpoint.md
│   │   ├── user_entitlement.md
│   │   ├── variable_group.md
│   │   └── ... (40+ resources)
│   ├── guides/                     # Usage guides
│   │   ├── authenticating.md
│   │   ├── version_locking.md
│   │   └── examples.md
│   └── index.md                    # Provider documentation index
├── examples/
│   ├── agent_pool/
│   ├── build_definition/
│   ├── data_sources/
│   ├── git_repository/
│   ├── project/
│   ├── service_endpoint/
│   └── ... (many more examples)
├── CONTRIBUTING.md                 # Contribution guidelines
├── CHANGELOG.md                    # Release notes & breaking changes
├── LICENSE                         # MIT or Apache 2.0
└── README.md                       # Quick start & overview

```

---

## 2. KEY GO PACKAGES & DEPENDENCIES

### go.mod Structure

**Core Dependencies:**
```go
module github.com/microsoft/terraform-provider-azuredevops

go 1.19

require (
    github.com/hashicorp/terraform-plugin-sdk/v2 v2.24+
    github.com/hashicorp/terraform-plugin-go v0.19+
    github.com/microsoft/azure-devops-go-api/azuredevops v1.0+
    github.com/google/uuid v1.3+
    golang.org/x/crypto v0.15+
)
```

**Key Imported Packages:**

| Package | Purpose |
|---------|---------|
| `github.com/hashicorp/terraform-plugin-sdk/v2` | Terraform provider framework v2 (schema, CRUD helpers) |
| `github.com/hashicorp/terraform-plugin-go` | New plugin protocol support |
| `github.com/microsoft/azure-devops-go-api/azuredevops` | Official ADO SDK for core APIs (projects, repos, build, policies) |
| `github.com/google/uuid` | UUID generation |
| `golang.org/x/crypto` | Cryptographic operations (SSH key handling) |
| `encoding/json` | JSON marshaling/unmarshaling |
| `fmt`, `log`, `net/http` | Standard library |

**What the SDK Provides:**
- Core ADO API clients (projects, repos, git, policies, builds, work items, security)
- Type definitions for all ADO resources
- Automatic JSON marshaling with ADO conventions

**What Custom HTTP is Used For:**
- **Release Management API** (not in SDK) — uses custom HTTP to `vsrm.dev.azure.com`
- **Graph API** (optional) — for advanced user/group lookups
- **Artifact feeds** — if using alternative endpoints
- Custom retry logic and rate limiting
- Connection pooling and auth header management

---

## 3. PROVIDER CONFIGURATION & REGISTRATION

### main.go Pattern

```go
package main

import (
    "context"
    "log"

    "github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
    "github.com/microsoft/terraform-provider-azuredevops/internal"
)

func main() {
    opts := &plugin.ServeOpts{
        ProtocolVersion: 6,
        ProviderFunc: internal.Provider,
    }

    err := plugin.Serve(opts)
    if err != nil {
        log.Fatal(err)
    }
}
```

### internal/provider.go Pattern

Core provider registration file defines:

1. **Provider Schema** — Configuration block (authentication, org, features)
   ```go
   func Provider() *schema.Provider {
       return &schema.Provider{
           Schema: map[string]*schema.Schema{
               "org_service_url": {
                   Type:        schema.TypeString,
                   Required:    true,
                   DefaultFunc: schema.EnvDefaultFunc("AZDO_ORG_SERVICE_URL", nil),
               },
               "personal_access_token": {
                   Type:        schema.TypeString,
                   Required:    true,
                   DefaultFunc: schema.EnvDefaultFunc("AZDO_PERSONAL_ACCESS_TOKEN", nil),
                   Sensitive:   true,
               },
               "keep_authorization": {
                   Type:     schema.TypeBool,
                   Optional: true,
                   Default:  true,
               },
           },
           // ... more fields
       }
   }
   ```

2. **Resource Registration** — Maps provider names to resource implementations
   ```go
   Resources: map[string]*schema.Resource{
       "azuredevops_project":                resourceProject(),
       "azuredevops_git_repository":         resourceGitRepository(),
       "azuredevops_build_definition":       resourceBuildDefinition(),
       "azuredevops_service_endpoint":       resourceServiceEndpoint(),
       // ... 40+ more resources
   }
   ```

3. **Data Source Registration** — Maps provider names to data source implementations
   ```go
   DataSources: map[string]*schema.Resource{
       "azuredevops_project":              dataSourceProject(),
       "azuredevops_git_repository":       dataSourceGitRepository(),
       "azuredevops_build_definition":     dataSourceBuildDefinition(),
       // ... 30+ more data sources
   }
   ```

4. **ConfigureContextFunc** — Initializes the AggregatedClient on provider startup
   ```go
   ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
       // Parse org_service_url, personal_access_token
       // Create HTTP client with auth
       // Initialize AggregatedClient with sub-clients
       // Return client for use by resources
   }
   ```

---

## 4. AGGREGATED CLIENT PATTERN

### internal/client/client.go

The **AggregatedClient** is the central connection hub for all API calls:

```go
type AggregatedClient struct {
    // Core transport layer
    httpClient *http.Client
    baseURL    string
    patToken   string

    // Per-service clients
    ProjectsClient         projectsapi.Client
    GitClient              gitapi.Client
    BuildClient            buildapi.Client
    ReleaseClient          releaseapi.Client
    ServiceEndpointClient  serviceendpointapi.Client
    PolicyClient           policyapi.Client
    CoreClient             coreapi.Client
    GraphClient            graphapi.Client
    IdentityClient         identityapi.Client
    SecurityClient         securityapi.Client
    TestClient             testapi.Client
    WikiClient             wikiapi.Client
    WorkItemTrackingClient witapi.Client
    FeedClient             feedapi.Client
    
    // Config
    OrgServiceURL string
    ProjectName   string
}
```

**Client Creation Pattern:**

```go
func NewAggregatedClient(ctx context.Context, orgServiceURL, pat string) (*AggregatedClient, error) {
    httpClient := createHTTPClient(pat)
    
    client := &AggregatedClient{
        httpClient:    httpClient,
        baseURL:       orgServiceURL,
        patToken:      pat,
        OrgServiceURL: orgServiceURL,
    }
    
    // Initialize SDK clients using connection
    conn := createConnection(orgServiceURL, pat)
    client.ProjectsClient = projectsapi.NewClient(ctx, conn)
    client.GitClient = gitapi.NewClient(ctx, conn)
    client.BuildClient = buildapi.NewClient(ctx, conn)
    // ... initialize all other clients
    
    return client, nil
}
```

**Key Characteristics:**
- Single HTTP client for connection pooling
- All sub-clients share the same auth context
- Lazy initialization pattern for some clients
- Centralized error handling and retry logic
- Thread-safe for concurrent requests

---

## 5. RESOURCE IMPLEMENTATION PATTERN

All resources in the provider follow a consistent pattern. Example: `resource_git_repository.go`

### Structure:

```go
// 1. SCHEMA DEFINITION
func resourceGitRepository() *schema.Resource {
    return &schema.Resource{
        Create: resourceGitRepositoryCreate,
        Read:   resourceGitRepositoryRead,
        Update: resourceGitRepositoryUpdate,
        Delete: resourceGitRepositoryDelete,
        
        Schema: map[string]*schema.Schema{
            "project_id": {
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,
            },
            "name": {
                Type:     schema.TypeString,
                Required: true,
            },
            "url": {
                Type:     schema.TypeString,
                Computed: true,
            },
            "default_branch": {
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
            },
            "is_fork": {
                Type:     schema.TypeBool,
                Computed: true,
            },
            "ssh_url": {
                Type:      schema.TypeString,
                Computed:  true,
                Sensitive: true,
            },
        },
        Importer: &schema.ResourceImporter{
            State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
                // Parse import ID (format: project_id/repo_id)
                return []*schema.ResourceData{d}, nil
            },
        },
    }
}

// 2. CREATE OPERATION
func resourceGitRepositoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    clients := m.(*client.AggregatedClient)
    projectID := d.Get("project_id").(string)
    
    // Expand Terraform state to API request
    repo := expandGitRepository(d)
    
    // API call
    createdRepo, err := clients.GitClient.CreateRepository(ctx, gitapi.CreateRepositoryArgs{
        Project: &projectID,
        GitRepositoryToCreate: repo,
    })
    if err != nil {
        return diag.Errorf("creating git repository: %v", err)
    }
    
    // Set resource ID
    d.SetId(*createdRepo.Id)
    
    // Flatten API response to Terraform state
    return resourceGitRepositoryRead(ctx, d, m)
}

// 3. READ OPERATION
func resourceGitRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    clients := m.(*client.AggregatedClient)
    projectID := d.Get("project_id").(string)
    repoID := d.Id()
    
    repo, err := clients.GitClient.GetRepository(ctx, gitapi.GetRepositoryArgs{
        Project:       &projectID,
        RepositoryId:  &repoID,
    })
    if err != nil {
        if isNotFoundError(err) {
            d.SetId("") // Resource was deleted outside Terraform
            return nil
        }
        return diag.Errorf("reading git repository: %v", err)
    }
    
    flattenGitRepository(d, repo)
    return nil
}

// 4. UPDATE OPERATION
func resourceGitRepositoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    clients := m.(*client.AggregatedClient)
    projectID := d.Get("project_id").(string)
    repoID := d.Id()
    
    repo := expandGitRepository(d)
    _, err := clients.GitClient.UpdateRepository(ctx, gitapi.UpdateRepositoryArgs{
        Project:        &projectID,
        RepositoryId:   &repoID,
        GitRepository:  repo,
    })
    if err != nil {
        return diag.Errorf("updating git repository: %v", err)
    }
    
    return resourceGitRepositoryRead(ctx, d, m)
}

// 5. DELETE OPERATION
func resourceGitRepositoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    clients := m.(*client.AggregatedClient)
    projectID := d.Get("project_id").(string)
    repoID := d.Id()
    
    err := clients.GitClient.DeleteRepository(ctx, gitapi.DeleteRepositoryArgs{
        Project:       &projectID,
        RepositoryId:  &repoID,
    })
    if err != nil && !isNotFoundError(err) {
        return diag.Errorf("deleting git repository: %v", err)
    }
    
    d.SetId("")
    return nil
}

// 6. EXPAND FUNCTION (Terraform → API)
func expandGitRepository(d *schema.ResourceData) *gitapi.GitRepository {
    return &gitapi.GitRepository{
        Name:           stringPtr(d.Get("name").(string)),
        DefaultBranch:  stringPtr(d.Get("default_branch").(string)),
        IsDisabled:     boolPtr(d.Get("is_disabled").(bool)),
    }
}

// 7. FLATTEN FUNCTION (API → Terraform)
func flattenGitRepository(d *schema.ResourceData, repo *gitapi.GitRepository) {
    d.Set("name", repo.Name)
    d.Set("url", repo.WebUrl)
    d.Set("default_branch", repo.DefaultBranch)
    d.Set("ssh_url", repo.SshUrl)
    d.Set("is_fork", repo.IsFork)
    d.Set("project_id", repo.Project.Id)
}
```

### Common Patterns:

**Error Handling:**
- 404 → `d.SetId("")` and return nil (deleted outside Terraform)
- 400/validation → `diag.Errorf("field validation error: %v", err)`
- Network/server → retry with backoff
- 409 Conflict → retry with backoff for concurrent modifications

**State Management:**
- Always do read-after-write to get computed fields
- Use `ForceNew: true` for fields that require replacement
- Handle nullable pointers with helper functions

**Testing Pattern:**
- Acceptance tests in `resource_git_repository_test.go`
- Use random names to avoid conflicts
- Check actual API state, not just Terraform state
- Cleanup is automatic via destroy step

---

## 6. DATA SOURCE IMPLEMENTATION PATTERN

Data sources follow a simpler pattern (no CRUD, only read):

```go
func dataSourceGitRepository() *schema.Resource {
    return &schema.Resource{
        Read: dataSourceGitRepositoryRead,
        Schema: map[string]*schema.Schema{
            "project_id": {
                Type:     schema.TypeString,
                Required: true,
            },
            "repository_id": {
                Type:     schema.TypeString,
                Optional: true,
            },
            "name": {
                Type:     schema.TypeString,
                Optional: true,
            },
            "url": {
                Type:     schema.TypeString,
                Computed: true,
            },
            // ... more computed fields
        },
    }
}

func dataSourceGitRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    clients := m.(*client.AggregatedClient)
    projectID := d.Get("project_id").(string)
    
    var repo *gitapi.GitRepository
    var err error
    
    // Look up by ID or name
    if repoID, ok := d.GetOk("repository_id"); ok {
        repo, err = clients.GitClient.GetRepository(ctx, gitapi.GetRepositoryArgs{
            Project:      &projectID,
            RepositoryId: stringPtr(repoID.(string)),
        })
    } else if repoName, ok := d.GetOk("name"); ok {
        // List and filter by name
        repos, err := clients.GitClient.GetRepositories(ctx, gitapi.GetRepositoriesArgs{
            Project: &projectID,
        })
        if err != nil {
            return diag.Errorf("reading git repositories: %v", err)
        }
        for _, r := range *repos {
            if *r.Name == repoName.(string) {
                repo = &r
                break
            }
        }
        if repo == nil {
            return diag.Errorf("repository %q not found", repoName)
        }
    }
    
    if err != nil {
        return diag.Errorf("reading git repository: %v", err)
    }
    
    d.SetId(*repo.Id)
    flattenGitRepository(d, repo)
    return nil
}
```

---

## 7. TEST INFRASTRUCTURE

### Acceptance Test Pattern: `resource_git_repository_test.go`

```go
package git

import (
    "fmt"
    "testing"

    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
    "github.com/microsoft/terraform-provider-azuredevops/internal/acceptancetest"
    "github.com/microsoft/terraform-provider-azuredevops/internal/client"
)

// Environment setup
var (
    tfSummary = fmt.Sprintf("tf-project-%s", acceptancetest.RandStringRunes(10))
)

// TestAcc prefix indicates acceptance test
func TestAccGitRepository_Basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:     func() { acceptancetest.PreCheck(t) },
        Providers:    acceptancetest.TestAccProviders,
        CheckDestroy: testAccCheckGitRepositoryDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccGitRepositoryBasic(tfSummary),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrSet("azuredevops_git_repository.test", "id"),
                    resource.TestCheckResourceAttr("azuredevops_git_repository.test", "name", "test-repo"),
                    resource.TestCheckResourceAttrSet("azuredevops_git_repository.test", "url"),
                    // ... more checks
                ),
            },
            // Import test step
            {
                ResourceName:      "azuredevops_git_repository.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}

// Multi-step test with update
func TestAccGitRepository_Update(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:     func() { acceptancetest.PreCheck(t) },
        Providers:    acceptancetest.TestAccProviders,
        CheckDestroy: testAccCheckGitRepositoryDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccGitRepositoryBasic(tfSummary),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("azuredevops_git_repository.test", "name", "test-repo"),
                ),
            },
            {
                Config: testAccGitRepositoryUpdated(tfSummary),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("azuredevops_git_repository.test", "name", "test-repo-updated"),
                ),
            },
        },
    })
}

// Destroy check
func testAccCheckGitRepositoryDestroy(s *terraform.State) error {
    clients := acceptancetest.TestAccProvider.Meta().(*client.AggregatedClient)
    
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "azuredevops_git_repository" {
            continue
        }
        
        projectID := rs.Primary.Attributes["project_id"]
        repoID := rs.Primary.ID
        
        _, err := clients.GitClient.GetRepository(nil, gitapi.GetRepositoryArgs{
            Project:      stringPtr(projectID),
            RepositoryId: stringPtr(repoID),
        })
        
        if err == nil {
            return fmt.Errorf("repository %s still exists", repoID)
        }
    }
    return nil
}

// Config templates
func testAccGitRepositoryBasic(projectName string) string {
    return fmt.Sprintf(`
resource "azuredevops_project" "test" {
  name               = "%s"
  version_control    = "Git"
}

resource "azuredevops_git_repository" "test" {
  project_id = azuredevops_project.test.id
  name       = "test-repo"
}
`, projectName)
}

func testAccGitRepositoryUpdated(projectName string) string {
    return fmt.Sprintf(`
resource "azuredevops_project" "test" {
  name               = "%s"
  version_control    = "Git"
}

resource "azuredevops_git_repository" "test" {
  project_id = azuredevops_project.test.id
  name       = "test-repo-updated"
}
`, projectName)
}
```

### Test Helpers: `internal/acceptancetest/testutils.go`

```go
package acceptancetest

import (
    "math/rand"
    "os"
    "testing"

    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/microsoft/terraform-provider-azuredevops/internal"
)

var (
    TestAccProvider *schema.Provider
    TestAccProviders map[string]*schema.Provider
)

func init() {
    TestAccProvider = internal.Provider()
    TestAccProviders = map[string]*schema.Provider{
        "azuredevops": TestAccProvider,
    }
}

// PreCheck validates required environment variables for acceptance tests
func PreCheck(t *testing.T) {
    if v := os.Getenv("AZDO_ORG_SERVICE_URL"); v == "" {
        t.Fatal("AZDO_ORG_SERVICE_URL must be set for acceptance tests")
    }
    if v := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN"); v == "" {
        t.Fatal("AZDO_PERSONAL_ACCESS_TOKEN must be set for acceptance tests")
    }
}

// RandStringRunes generates random string for unique test resources
func RandStringRunes(length int) string {
    letterRunes := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
    b := make([]rune, length)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}
```

### Running Tests

```bash
# Unit tests
make test

# Acceptance tests (requires ADO account)
make testacc

# Specific acceptance test
TF_ACC=1 AZDO_ORG_SERVICE_URL=https://dev.azure.com/myorg \
  AZDO_PERSONAL_ACCESS_TOKEN=xyz \
  go test -v -run TestAccGitRepository_Basic

# With logging
TF_LOG=DEBUG TF_ACC=1 go test -v ...
```

---

## 8. BUILD SYSTEM & CI/CD

### GNUmakefile Targets

```makefile
# Build the provider binary
make build
# Output: terraform-provider-azuredevops_vX.Y.Z

# Install to local Terraform cache
make install
# Installs to ~/.terraform.d/plugins/...

# Run unit tests
make test

# Run acceptance tests
make testacc

# Generate documentation
make generate

# Lint code
make lint

# Format code
make fmt

# Run all checks
make all
```

### .github/workflows/tests.yaml

**Triggers:**
- On pull request to main
- On push to main
- On release tags

**Jobs:**

1. **Lint Job**
   - golangci-lint with .golangci.yml config
   - Check formatting with gofmt
   - Vet analysis
   - Spell check

2. **Unit Tests Job**
   - Runs on Go 1.19, 1.20, 1.21
   - Runs on Linux, Windows, macOS
   - Coverage reporting
   - Artifact upload for coverage

3. **Acceptance Tests Job**
   - Requires ADO account credentials
   - Runs matrix of Go versions
   - Timeout per test (30 minutes typical)
   - Test results posted as PR comment

4. **CodeQL Security Scan**
   - SAST analysis for vulnerabilities
   - Pattern matching for security issues
   - Reports to GitHub Security tab

### .github/workflows/release.yaml

**Triggers:** On tag push (v0.0.0 format)

**Process:**
1. Checkout code
2. Set up Go
3. Run GoReleaser for cross-platform builds
4. Create GitHub release with binaries
5. Compute checksums
6. Sign releases (optional)
7. Publish to Terraform Registry via provider publishing protocol

### .golangci.yml

```yaml
linters:
  enable:
    - deadcode
    - errcheck
    - gofmt
    - govet
    - ineffassign
    - revive
    - staticcheck
    - typecheck
    - unused

issues:
  exclude-rules:
    - path: "_test.go"
      linters:
        - unused

run:
  timeout: 5m
  go: "1.19"
```

---

## 9. CODE ORGANIZATION & PATTERNS

### File Naming Conventions

```
resource_<name>.go                    # Resource definition & CRUD
resource_<name>_test.go               # Acceptance tests
data_source_<name>.go                 # Data source definition & read
data_source_<name>_test.go            # Data source tests
<name>.go                             # Helper functions & types

Example:
  resource_git_repository.go
  resource_git_repository_test.go
  data_source_git_repository.go
  git_repository.go                   # Shared helpers
```

### Package Organization

```
internal/service/git/
├── resource_git_repository.go        # CRUD + schema
├── resource_git_repository_test.go
├── data_source_git_repository.go
├── data_source_git_repository_test.go
├── git_repository.go                 # expand/flatten helpers
├── resource_git_repository_policy.go # Branch policies
├── resource_git_repository_policy_test.go
└── ...
```

### Expand/Flatten Pattern

Every resource has helpers to convert between Terraform state and API types:

```go
// expandGitRepository converts Terraform schema to API type
func expandGitRepository(d *schema.ResourceData) *gitapi.GitRepository {
    return &gitapi.GitRepository{
        Name:           stringPtr(d.Get("name").(string)),
        DefaultBranch:  stringPtr(d.Get("default_branch").(string)),
        IsDisabled:     boolPtr(d.Get("is_disabled").(bool)),
    }
}

// flattenGitRepository converts API type to Terraform schema
func flattenGitRepository(d *schema.ResourceData, repo *gitapi.GitRepository) error {
    d.Set("id", repo.Id)
    d.Set("name", repo.Name)
    d.Set("url", repo.WebUrl)
    d.Set("default_branch", repo.DefaultBranch)
    d.Set("ssh_url", repo.SshUrl)
    return nil
}
```

---

## 10. QUALITY GUARDRAILS & CODE GENERATION

### Documentation Generation

The provider auto-generates documentation from schema:

```bash
make generate  # Uses tfplugindocs tool
```

Output:
- `docs/resources/*.md` from resource schema
- `docs/data-sources/*.md` from data source schema
- Tables, examples, argument references auto-populated

### Code Generation Tools Used

1. **tfplugindocs** — Generate documentation from schema comments
2. **GoReleaser** — Cross-platform binary compilation and release
3. **go generate** — Code generation directives in Go files

### Dependency Management

- `go mod tidy` — Remove unused imports
- `go mod verify` — Verify checksums
- Renovate or Dependabot for automatic updates
- Security vulnerability scanning via GitHub

### Pre-commit Hooks (Recommended)

```bash
go fmt ./...
golangci-lint run
go vet ./...
go test ./...
```

---

## 11. AUTHENTICATION & CLIENT INITIALIZATION

### Authentication Flow

```
Provider Config
  ↓
org_service_url (required)
personal_access_token (required, sensitive)
  ↓
ConfigureContextFunc
  ↓
HTTP Client Setup
  - Basic auth header: Authorization: Basic base64(:PAT)
  - Connection pooling
  - Timeout: 30 seconds default
  ↓
AggregatedClient Creation
  - Initialize ADO SDK clients
  - Set up retry middleware
  ↓
Return to Terraform Core
  ↓
Pass to Resources via d.Meta()
```

### Multi-Org Support

- Provider instance per org
- Multiple provider blocks in same Terraform config
- Alias pattern: `provider "azuredevops" { alias = "prod" }`
- Resources specify provider: `provider = azuredevops.prod`

---

## 12. ERROR HANDLING STRATEGY

### HTTP Error Types

```go
// 400 Bad Request → Schema/validation error
if statusCode == 400 {
    return diag.Errorf("validation error: %s", body)
}

// 401 Unauthorized → Auth failure
if statusCode == 401 {
    return diag.Errorf("authentication failed: check PAT token")
}

// 403 Forbidden → Permission denied
if statusCode == 403 {
    return diag.Errorf("permission denied: %s", body)
}

// 404 Not Found → Resource deleted or doesn't exist
if statusCode == 404 {
    if inRead {
        d.SetId("")
        return nil  // Mark resource as gone
    }
    return diag.Errorf("resource not found")
}

// 409 Conflict → Concurrent modification
if statusCode == 409 {
    // Retry with backoff (exponential: 1s, 2s, 4s, 8s)
}

// 5xx Server Error → Retry with backoff
if statusCode >= 500 {
    // Retry with backoff
}
```

### Retry Strategy

- Backoff: exponential (1s, 2s, 4s, 8s max)
- Max retries: 5 attempts
- Idempotent operations (GET, PUT, DELETE)
- Skip retries on 400, 401, 403

---

## 13. NESTED RESOURCE PATTERNS

### Example: Project with Nested Policies

Project resources often have nested sub-resources (policies, permissions, features):

```go
// resource_project.go
Schema: {
    "name": {...},
    "project_settings": {
        Type: schema.TypeList,
        MaxItems: 1,
        Elem: &schema.Resource{
            Schema: {
                "work_item_template": {...},
                "features": {
                    Type: schema.TypeSet,
                    Elem: &schema.Resource{
                        Schema: {
                            "feature_id": {...},
                            "enabled": {...},
                        },
                    },
                },
            },
        },
    },
}

// Expand nested
func expandProjectSettings(d *schema.ResourceData) *projectapi.ProjectSettings {
    settings := &projectapi.ProjectSettings{}
    if raw, ok := d.GetOk("project_settings"); ok {
        settingsList := raw.([]interface{})
        if len(settingsList) > 0 {
            // Recurse into nested expand functions
        }
    }
    return settings
}
```

---

## 14. SERVICE ENDPOINT TYPE PATTERN

Service endpoints are special — they're polymorphic with 20+ types:

```go
// internal/service/serviceendpoint/resource_service_endpoint.go

Schema: {
    "service_endpoint_type": {
        Type:     schema.TypeString,
        Required: true,
        ValidateDiagFunc: validation.StringInSlice([]string{
            "AzureRM",
            "Docker",
            "GitHub",
            "Kubernetes",
            "Npm",
            "Ssh",
            // ... 15+ more types
        }, false),
    },
    "authentication": {
        Type: schema.TypeList,
        Elem: &schema.Resource{
            Schema: {
                // Fields vary by type
                "certificate": {...},
                "token": {...},
                "username": {...},
            },
        },
    },
    // ... type-specific fields
}

// Expand dispatches to type-specific handlers
func expandServiceEndpoint(d *schema.ResourceData) *serviceendpointapi.ServiceEndpoint {
    epType := d.Get("service_endpoint_type").(string)
    
    switch epType {
    case "AzureRM":
        return expandAzureRmServiceEndpoint(d)
    case "Docker":
        return expandDockerServiceEndpoint(d)
    // ... switch for each type
    }
}
```

---

## 15. TESTING PATTERNS

### Common Test Scenarios

```go
// Basic CRUD lifecycle test
TestAcc<Resource>_Basic() → Create → Read → Import → Destroy

// Update test
TestAcc<Resource>_Update() → Create → Update → Read → Destroy

// Force new test (recreates on change)
TestAcc<Resource>_ForceNew() → Create → Update (expect replacement) → Read → Destroy

// Data source test
TestAcc<DataSource>_Read() → Create resource → Read via data source → Verify

// Validation tests
TestAcc<Resource>_InvalidInput() → Expect validation error

// Integration tests
TestAcc<Resource>_WithDependencies() → Create project → Create repo → Verify relationships
```

### Mock Patterns (Unit Tests)

Most unit tests use table-driven pattern:

```go
func TestFlattenGitRepository(t *testing.T) {
    tests := []struct {
        name     string
        input    *gitapi.GitRepository
        expected map[string]interface{}
    }{
        {
            name: "basic",
            input: &gitapi.GitRepository{
                Id:            stringPtr("repo-id"),
                Name:          stringPtr("test-repo"),
            },
            expected: map[string]interface{}{
                "id":   "repo-id",
                "name": "test-repo",
            },
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test flatten logic
        })
    }
}
```

---

## 16. RESOURCE & DATA SOURCE COVERAGE

### Typical Coverage

The official provider includes ~40+ resources:

**Core:**
- `azuredevops_project`
- `azuredevops_git_repository`
- `azuredevops_build_definition`
- `azuredevops_pipeline` (newer format)
- `azuredevops_service_endpoint` (20+ types)

**Permissions & Security:**
- `azuredevops_project_permissions`
- `azuredevops_git_permissions`
- `azuredevops_pipeline_permissions`
- `azuredevops_group`
- `azuredevops_user_entitlement`

**Advanced:**
- `azuredevops_variable_group`
- `azuredevops_git_repository_policy`
- `azuredevops_branch_policy_*` (min reviewers, auto-complete, etc.)
- `azuredevops_agent_queue`
- `azuredevops_resource_authorization`

**Artifact Management:**
- `azuredevops_feed`
- `azuredevops_feed_permission`

**Wiki & Collaboration:**
- `azuredevops_wiki`
- `azuredevops_wiki_page`

**Work Tracking:**
- `azuredevops_workitemqueryresult`
- `azuredevops_workitem_template`

### Notable Gaps

- **Classic Release Pipelines** — NOT in official provider (requires custom vsrm.dev.azure.com API)
- **Advanced YAML pipelines** — Limited support (JSON build definitions recommended)
- **Test Plans & Results** — Minimal coverage
- **Boards & Sprints** — Work in progress

---

## 17. DIRECTORY TREE VISUALIZATION

Complete 3-level structure:

```
terraform-provider-azuredevops/
│
├── .github/
│   └── workflows/
│       ├── release.yaml
│       ├── tests.yaml
│       ├── codeql-analysis.yaml
│       └── goreleaser.yaml
│
├── docs/
│   ├── guides/
│   │   ├── authenticating.md
│   │   ├── version_locking.md
│   │   └── examples.md
│   ├── data-sources/
│   │   ├── project.md
│   │   ├── git_repository.md
│   │   ├── build_definition.md
│   │   ├── service_endpoint.md
│   │   └── ... (30 more)
│   ├── resources/
│   │   ├── project.md
│   │   ├── git_repository.md
│   │   ├── build_definition.md
│   │   ├── service_endpoint.md
│   │   ├── git_repository_policy.md
│   │   ├── variable_group.md
│   │   ├── agent_queue.md
│   │   └── ... (40+ more)
│   └── index.md
│
├── examples/
│   ├── agent_pool/
│   │   └── main.tf
│   ├── build_definition/
│   │   ├── main.tf
│   │   ├── build.json
│   │   └── ... (more examples)
│   ├── data_sources/
│   │   ├── main.tf
│   │   └── ... (examples for each DS)
│   ├── git_repository/
│   │   ├── main.tf
│   │   ├── branch_policy/
│   │   │   └── main.tf
│   │   └── ...
│   ├── project/
│   │   └── main.tf
│   ├── service_endpoint/
│   │   ├── docker/
│   │   │   └── main.tf
│   │   ├── azurerm/
│   │   │   └── main.tf
│   │   ├── kubernetes/
│   │   │   └── main.tf
│   │   └── ... (15+ endpoint types)
│   ├── variable_group/
│   │   └── main.tf
│   └── ... (more examples)
│
├── internal/
│   ├── client/
│   │   ├── client.go
│   │   ├── connection.go
│   │   ├── core_client.go
│   │   ├── git_client.go
│   │   ├── build_client.go
│   │   ├── service_endpoint_client.go
│   │   ├── policy_client.go
│   │   ├── project_client.go
│   │   ├── security_client.go
│   │   ├── feed_client.go
│   │   ├── release_client.go
│   │   ├── test_client.go
│   │   ├── wiki_client.go
│   │   ├── workitemtracking_client.go
│   │   ├── taskagent_client.go
│   │   ├── graph_client.go
│   │   └── utils.go
│   │
│   ├── service/
│   │   ├── agent/
│   │   │   ├── agent_queue.go
│   │   │   ├── agent_queue_test.go
│   │   │   ├── resource_agent_queue.go
│   │   │   ├── resource_agent_queue_test.go
│   │   │   └── data_source_agent_queue.go
│   │   │
│   │   ├── build/
│   │   │   ├── build_definition.go
│   │   │   ├── resource_build_definition.go
│   │   │   ├── resource_build_definition_test.go
│   │   │   ├── data_source_build_definition.go
│   │   │   └── data_source_build_definition_test.go
│   │   │
│   │   ├── core/
│   │   │   ├── project.go
│   │   │   ├── resource_project.go
│   │   │   ├── resource_project_test.go
│   │   │   ├── data_source_project.go
│   │   │   ├── data_source_project_test.go
│   │   │   ├── project_features.go
│   │   │   ├── resource_project_features.go
│   │   │   └── ... (more core resources)
│   │   │
│   │   ├── feed/
│   │   │   ├── feed.go
│   │   │   ├── resource_feed.go
│   │   │   ├── resource_feed_test.go
│   │   │   ├── resource_feed_permission.go
│   │   │   └── ...
│   │   │
│   │   ├── git/
│   │   │   ├── git_repository.go
│   │   │   ├── resource_git_repository.go
│   │   │   ├── resource_git_repository_test.go
│   │   │   ├── data_source_git_repository.go
│   │   │   ├── data_source_git_repository_test.go
│   │   │   ├── resource_git_repository_policy.go
│   │   │   ├── resource_git_repository_policy_test.go
│   │   │   ├── git_repository_policy.go
│   │   │   └── ...
│   │   │
│   │   ├── permission/
│   │   │   ├── resource_project_permissions.go
│   │   │   ├── resource_git_permissions.go
│   │   │   ├── resource_pipeline_permissions.go
│   │   │   ├── resource_build_definition_permissions.go
│   │   │   ├── resource_service_endpoint_permissions.go
│   │   │   └── ...
│   │   │
│   │   ├── serviceendpoint/
│   │   │   ├── resource_service_endpoint.go
│   │   │   ├── resource_service_endpoint_test.go
│   │   │   ├── data_source_service_endpoint.go
│   │   │   ├── service_endpoint_types/
│   │   │   │   ├── azurerm.go
│   │   │   │   ├── docker.go
│   │   │   │   ├── github.go
│   │   │   │   ├── kubernetes.go
│   │   │   │   ├── npm.go
│   │   │   │   ├── ssh.go
│   │   │   │   ├── bitbucket.go
│   │   │   │   ├── connectedRegistry.go
│   │   │   │   ├── externalGit.go
│   │   │   │   ├── jenkins.go
│   │   │   │   ├── artifactory.go
│   │   │   │   ├── sonarqube.go
│   │   │   │   └── ... (15+ more types)
│   │   │   └── ...
│   │   │
│   │   ├── taskagent/
│   │   │   └── ...
│   │   │
│   │   ├── test/
│   │   │   ├── resource_test_plan.go
│   │   │   ├── resource_test_variable.go
│   │   │   └── ...
│   │   │
│   │   ├── wiki/
│   │   │   ├── resource_wiki.go
│   │   │   ├── resource_wiki_page.go
│   │   │   └── ...
│   │   │
│   │   ├── workitemtracking/
│   │   │   ├── resource_workitem_template.go
│   │   │   └── ...
│   │   │
│   │   └── ... (more service packages)
│   │
│   ├── acceptancetest/
│   │   ├── testutils.go
│   │   ├── tfexec.go
│   │   ├── check.go
│   │   └── fixture.go
│   │
│   ├── provider.go
│   ├── utils/
│   │   ├── converter.go
│   │   ├── tfhelper.go
│   │   ├── validate.go
│   │   └── suppress.go
│   │
│   └── ... (other utilities)
│
├── .gitignore
├── .golangci.yml
├── CHANGELOG.md
├── CONTRIBUTING.md
├── GNUmakefile
├── LICENSE
├── README.md
├── go.mod
├── go.sum
└── main.go
```

---

## 18. KEY TECHNICAL INSIGHTS FOR FORK

### What to Copy As-Is

1. **client/client.go** — AggregatedClient pattern is excellent
2. **internal/provider.go** — Provider setup and registration
3. **GNUmakefile** — Build and test targets
4. **.golangci.yml** — Linting config
5. **.github/workflows/** — CI/CD pipeline
6. **internal/acceptancetest/** — Test helpers and patterns
7. **internal/utils/** — Utility functions (type conversion, validation)

### What to Adapt/Extend

1. **resource_*.go files** — Copy patterns, adapt to your resources
2. **docs/resources/*.md** — Auto-generated from schema; recreate
3. **examples/** — Adapt to your resource types
4. **internal/service/** — Create new service packages for release, etc.

### What Won't Work for Release API

1. **SDK clients** — The official ADO SDK doesn't have release API
   - Release API lives at `vsrm.dev.azure.com` (different host)
   - Must use custom HTTP client (see better-ado-provider for example)
   - Build custom API types matching vsrm.dev.azure.com REST responses

2. **Connection/Auth** — Minor adaptation needed
   - Basic auth still works
   - May need different connection pooling for vsrm.dev.azure.com host

### Testing Strategy for Fork

1. **Keep unit tests lightweight** — Mock HTTP responses
2. **Acceptance tests require real ADO org** — Plan for test cleanup
3. **Use environment-based test matrix** — Test against multiple ADO versions
4. **Implement test fixtures** — Pre-created projects/resources to speed up tests

---

## 19. DEPENDENCY ANALYSIS

### What the SDK Provides

The `azure-devops-go-api` SDK provides:

```go
// High-level services
projectsapi.Client        // Projects, project features
gitapi.Client             // Git repos, branches, commits
buildapi.Client           // Build definitions, queues
coreapi.Client            // Core operations
graphapi.Client           // AAD integration
identityapi.Client        // User/group identity
securityapi.Client        // Permissions
workitemapi.Client        // Work items, queries
testapi.Client            // Test plans, results
feedapi.Client            // Artifact feeds
wikiapi.Client            // Wiki pages
taskagentapi.Client       // Agent queues, pools
policyapi.Client          // Branch policies
serviceendpointapi.Client // Service connections
```

### What's NOT in the SDK

- **Release Management API** — Entire vsrm.dev.azure.com endpoint
  - Release definitions, environments, approvals, gates
  - Must implement custom HTTP client + types
  - API reference: https://docs.microsoft.com/en-us/rest/api/azure/devops/release/

- **Advanced release features**
  - Pre/post-deployment approvals
  - Release gates (query, scheduled, etc.)
  - Manual intervention tasks

---

## 20. IMPLEMENTATION CHECKLIST FOR FORK

```
[ ] Copy official provider as base repo
[ ] Keep all official resources (don't delete)
[ ] Add custom service packages for new resources (e.g., internal/service/release/)
[ ] Implement custom HTTP client for vsrm.dev.azure.com
[ ] Define API types (ReleaseDefinition, Environment, Approval, Gate)
[ ] Implement resource_release_definition.go following official patterns
[ ] Implement resource_release_environment.go
[ ] Implement resource_release_approvals.go
[ ] Implement resource_release_gates.go
[ ] Add acceptance tests for each resource
[ ] Add example Terraform configs
[ ] Document API findings in docs/api-reference/
[ ] Add CI/CD workflows (copy from official)
[ ] Test against real ADO org
[ ] Create documentation (auto-generated from schema)
[ ] Set up provider registry publication process
```

---

## Summary

The microsoft/terraform-provider-azuredevops provider is a **masterclass in Terraform provider development**:

- **AggregatedClient pattern** — Elegant multi-service orchestration
- **Resource schema patterns** — Consistent, well-tested CRUD implementations
- **Testing infrastructure** — Comprehensive acceptance tests with proper cleanup
- **CI/CD pipelines** — Professional, automated builds and releases
- **Code organization** — Clear separation by service, consistent naming
- **Documentation** — Auto-generated from schema, examples for each resource

For your fork, the key is **embracing these patterns** while adding support for the Release Management API (which requires custom HTTP handling). The official provider's structure gives you a solid foundation; your additions will be cleaner and more maintainable by following their established conventions.

