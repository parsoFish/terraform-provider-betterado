package acceptancetests

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

func TestAccGitRepository_withDefaultBranch(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()
	tfRepoNode := "betterado_git_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGitRepoDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryWithDefaultBranch(projectName, gitRepoName, "Clean"),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "is_fork"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "web_url"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "default_branch", "refs/heads/main"),
					captureGitRepositoryEvidence(tfRepoNode),
				),
			},
			// Idempotency check — no perpetual diff
			{
				Config:             hclGitRepositoryWithDefaultBranch(projectName, gitRepoName, "Clean"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccGitRepository_update(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	gitRepoNameFirst := testutils.GenerateResourceName()
	gitRepoNameSecond := testutils.GenerateResourceName()
	tfRepoNode := "betterado_git_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGitRepoDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryBasic(projectName, gitRepoNameFirst, "Uninitialized"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoNameFirst),
					checkGitRepoExists(gitRepoNameFirst),
					resource.TestCheckResourceAttrSet(tfRepoNode, "is_fork"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "remote_url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "size"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "ssh_url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "web_url"),
				),
			},
			{
				Config: hclGitRepositoryBasic(projectName, gitRepoNameSecond, "Uninitialized"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoNameSecond),
					checkGitRepoExists(gitRepoNameSecond),
					resource.TestCheckResourceAttrSet(tfRepoNode, "is_fork"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "remote_url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "size"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "ssh_url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "web_url"),
				),
			},
		},
	})
}

func TestAccGitRepository_disabled(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()
	tfRepoNode := "betterado_git_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGitRepoDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryDisable(projectName, gitRepoName, true),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "disabled", "true"),
				),
			},
			{
				Config: hclGitRepositoryDisable(projectName, gitRepoName, false),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "disabled", "false"),
				),
			},
		},
	})
}

func TestAccGitRepository_disabledCannotUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()
	gitRepoNameUpdate := gitRepoName + "update"
	tfRepoNode := "betterado_git_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGitRepoDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryDisable(projectName, gitRepoName, true),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "disabled", "true"),
				),
			},
			{
				Config: hclGitRepositoryDisable(projectName, gitRepoNameUpdate, true),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoNameUpdate),
					resource.TestCheckResourceAttr(tfRepoNode, "disabled", "true"),
				),
				ExpectError: regexp.MustCompile(`(?i)disabled repository cannot be updated|Cannot update disabled repository`),
			},
			{
				Config: hclGitRepositoryDisable(projectName, gitRepoName, false),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "disabled", "false"),
				),
			},
		},
	})
}

func TestAccGitRepository_incorrectInitialization(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      hclGitRepositoryIncorrectInitialization(projectName, gitRepoName),
				ExpectError: regexp.MustCompile(`Insufficient initialization blocks`),
			},
		},
	})
}

func TestAccGitRepository_importGitRepository(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()

	tfRepoNode := "betterado_git_repository.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryImport(projectName, gitRepoName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfRepoNode, "is_fork"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "remote_url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "size"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "ssh_url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "web_url"),
					resource.TestCheckResourceAttr(tfRepoNode, "initialization.#", "1"),
					checkGitRepoExists(gitRepoName),
				),
			},
		},
	})
}

func TestAccGitRepository_import_by_name(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()

	tfRepoNode := "betterado_git_repository.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryBasic(projectName, gitRepoName, "Clean"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfRepoNode, "is_fork"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "remote_url"),
					resource.TestCheckResourceAttr(tfRepoNode, "initialization.#", "1"),
					checkGitRepoExists(gitRepoName),
				),
			},
			{
				ResourceName:            tfRepoNode,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"initialization"},
				ImportStateId:           fmt.Sprintf("%s/%s", projectName, gitRepoName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					func(state *terraform.State) error {
						id, ok := state.RootModule().Resources["id"]
						if !ok {
							return fmt.Errorf("Resource `ID` not found in state")
						}
						if err := uuid.Validate(id.String()); err != nil {
							return fmt.Errorf("Resource `ID` is not in UUID format. Current value: %s", id.String())
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccGitRepository_initializationClean(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()
	tfRepoNode := "betterado_git_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGitRepoDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryBasic(projectName, gitRepoName, "Clean"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "default_branch", "refs/heads/master"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "disabled"),
				),
			},
		},
	})
}

func TestAccGitRepository_uninitialized(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()
	tfRepoNode := "betterado_git_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGitRepoDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryBasic(projectName, gitRepoName, "Uninitialized"),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "default_branch", ""),
				),
			},
		},
	})
}

func TestAccGitRepository_forkBranchNotEmpty(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()
	gitForkedRepoName := testutils.GenerateResourceName()
	tfRepoNode := "betterado_git_repository.test"
	tfForkedRepoNode := "betterado_git_repository.fork"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGitRepoDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryForkBranchNotEmpty(projectName, gitRepoName, gitForkedRepoName),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "default_branch", "refs/heads/master"),
					resource.TestCheckResourceAttr(tfForkedRepoNode, "name", gitForkedRepoName),
					resource.TestCheckResourceAttr(tfForkedRepoNode, "default_branch", "refs/heads/master"),
				),
			},
		},
	})
}

func TestAccGitRepository_privateImportServiceEndpointBranchNotEmpty(t *testing.T) {
	if os.Getenv("AZDO_GENERIC_GIT_SERVICE_CONNECTION_USERNAME") == "" ||
		os.Getenv("AZDO_GENERIC_GIT_SERVICE_CONNECTION_PASSWORD") == "" {
		t.Skip("Skipping as AZDO_GENERIC_GIT_SERVICE_CONNECTION_USERNAME or AZDO_GENERIC_GIT_SERVICE_CONNECTION_PASSWORD is not specified")
	}
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()
	gitImportRepoName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	tfRepoNode := "betterado_git_repository.test"
	tfImportRepoNode := "betterado_git_repository.import"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, &[]string{
				"AZDO_GENERIC_GIT_SERVICE_CONNECTION_USERNAME",
				"AZDO_GENERIC_GIT_SERVICE_CONNECTION_PASSWORD",
			})
		},
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGitRepoDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryImportPrivateServiceEndpoint(projectName, gitRepoName, gitImportRepoName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "default_branch", "refs/heads/master"),
					resource.TestCheckResourceAttrSet(tfImportRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfImportRepoNode, "name", gitImportRepoName),
					resource.TestCheckResourceAttr(tfImportRepoNode, "default_branch", "refs/heads/master"),
				),
			},
		},
	})
}

func TestAccGitRepository_privateUserNamePasswordImportBranchNotEmpty(t *testing.T) {
	if os.Getenv("AZDO_GENERIC_GIT_SERVICE_CONNECTION_USERNAME") == "" ||
		os.Getenv("AZDO_GENERIC_GIT_SERVICE_CONNECTION_PASSWORD") == "" {
		t.Skip("Skipping as AZDO_GENERIC_GIT_SERVICE_CONNECTION_USERNAME or AZDO_GENERIC_GIT_SERVICE_CONNECTION_PASSWORD is not specified")
	}
	projectName := testutils.GenerateResourceName()
	gitRepoName := testutils.GenerateResourceName()
	gitImportRepoName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	tfRepoNode := "betterado_git_repository.test"
	tfImportRepoNode := "betterado_git_repository.import"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, &[]string{
				"AZDO_GENERIC_GIT_SERVICE_CONNECTION_USERNAME",
				"AZDO_GENERIC_GIT_SERVICE_CONNECTION_PASSWORD",
			})
		},
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGitRepoDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryImportPrivateUserNamePassword(projectName, gitRepoName, gitImportRepoName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "default_branch", "refs/heads/master"),
					resource.TestCheckResourceAttrSet(tfImportRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfImportRepoNode, "name", gitImportRepoName),
					resource.TestCheckResourceAttr(tfImportRepoNode, "default_branch", "refs/heads/master"),
				),
			},
		},
	})
}

// captureGitRepositoryEvidence captures a live ADO API GET for forge demo evidence.
// It is called from a test Check step while the resource still exists (before destroy).
func captureGitRepositoryEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// best-effort: evidence capture must never fail a test assertion
		captureGitEvidence(s, tfNode)
		return nil
	}
}

// captureGitEvidence performs a live ADO API GET and writes forge demo evidence.
// All errors are silently swallowed — evidence capture is advisory, not a test gate.
func captureGitEvidence(s *terraform.State, tfNode string) {
	rs, ok := s.RootModule().Resources[tfNode]
	if !ok {
		return
	}
	repoID := rs.Primary.ID
	projectID := rs.Primary.Attributes["project_id"]

	clients, err := getDirectClient()
	if err != nil {
		return
	}

	repo, err := clients.GitReposClient.GetRepository(clients.Ctx, git.GetRepositoryArgs{
		RepositoryId: converter.String(repoID),
		Project:      converter.String(projectID),
	})
	if err != nil {
		return
	}

	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	liveURL := fmt.Sprintf("%s/_apis/git/repositories/%s?api-version=7.0", orgURL, repoID)
	_ = testutils.CaptureLiveEvidence("acceptance-resource", liveURL, repo)
}

// checkGitRepoExists verifies (1) the resource is in state and (2) exists in AzDO
// with the correct name. Uses direct client — safe with muxed provider factories.
func checkGitRepoExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("building direct client: %+v", err)
		}

		gitRepo, ok := s.RootModule().Resources["betterado_git_repository.test"]
		if !ok {
			return fmt.Errorf("Did not find a repo definition in the TF state")
		}

		repoID := gitRepo.Primary.ID
		projectID := gitRepo.Primary.Attributes["project_id"]

		repo, err := readGitRepo(clients, repoID, projectID)
		if err != nil {
			return err
		}

		if *repo.Name != expectedName {
			return fmt.Errorf("AzDO Git Repository has Name=%s, but expected Name=%s", *repo.Name, expectedName)
		}

		return nil
	}
}

// checkGitRepoDestroyed verifies every betterado_git_repository resource in the
// state no longer exists in AzDO. Uses direct client — safe with muxed provider factories.
func checkGitRepoDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("building direct client: %+v", err)
	}

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "betterado_git_repository" {
			continue
		}

		repoID := resource.Primary.ID
		projectID := resource.Primary.Attributes["project_id"]

		// indicates the git repository still exists - this should fail the test
		if _, err := readGitRepo(clients, repoID, projectID); err == nil {
			return fmt.Errorf("repository with ID %s should not exist", repoID)
		}
	}

	return nil
}

// readGitRepo looks up an Azure Git Repository by ID, falling back to hidden-repo
// scan when the repository is disabled.
func readGitRepo(clients *client.AggregatedClient, repoID string, projectID string) (*git.GitRepository, error) {
	repo, err := clients.GitReposClient.GetRepository(clients.Ctx, git.GetRepositoryArgs{
		RepositoryId: converter.String(repoID),
		Project:      converter.String(projectID),
	})

	// If the repository is disabled, the repository cannot be obtained through the GET API
	if utils.ResponseWasNotFound(err) {
		var allRepo *[]git.GitRepository
		allRepo, err = clients.GitReposClient.GetRepositories(clients.Ctx, git.GetRepositoriesArgs{
			Project: converter.String(projectID),
			// This flag is used to include disabled repos
			IncludeHidden: converter.Bool(true),
		})
		if err != nil {
			return nil, err
		}
		for _, gitRepo := range *allRepo {
			if strings.EqualFold(gitRepo.Id.String(), repoID) ||
				strings.EqualFold(*gitRepo.Name, repoID) {
				repo = &gitRepo
				break
			}
		}
	}
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func hclGitRepositoryBasic(projectName, repoName, initType string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%s"
  initialization {
    init_type = "%s"
  }
}
`, projectName, repoName, initType)
}

func hclGitRepositoryWithDefaultBranch(projectName, repoName, initType string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_git_repository" "test" {
  project_id     = betterado_project.test.id
  name           = "%s"
  default_branch = "refs/heads/main"
  initialization {
    init_type = "%s"
  }
}
`, projectName, repoName, initType)
}

func hclGitRepositoryDisable(projectName, repoName string, disabled bool) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%s"
  disabled   = %t
  initialization {
    init_type = "Clean"
  }
}
`, projectName, repoName, disabled)
}

func hclGitRepositoryIncorrectInitialization(projectName, repoName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%s"
}
`, projectName, repoName)
}

func hclGitRepositoryImport(projectName, repoName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%s"
  initialization {
    init_type   = "Import"
    source_type = "Git"
    source_url  = "https://github.com/parsoFish/terraform-provider-betterado.git"
  }
}
`, projectName, repoName)
}

func hclGitRepositoryForkBranchNotEmpty(projectName, repoName, forkRepoName string) string {
	repoInit := hclGitRepositoryBasic(projectName, repoName, "Clean")
	return fmt.Sprintf(`
%s

resource "betterado_git_repository" "fork" {
  project_id           = betterado_project.test.id
  parent_repository_id = betterado_git_repository.test.id
  name                 = "%s"
  initialization {
    init_type = "Fork"
  }
}`, repoInit, forkRepoName)
}

func hclGitRepositoryImportPrivateServiceEndpoint(projectName, repoName, importRepoName, serviceEndpointName string) string {
	repoInit := hclGitRepositoryBasic(projectName, repoName, "Clean")
	return fmt.Sprintf(`
%s

resource "betterado_serviceendpoint_generic_git" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%s"
  repository_url        = betterado_git_repository.test.remote_url
}

resource "betterado_git_repository" "import" {
  project_id = betterado_project.test.id
  name       = "%s"
  initialization {
    init_type             = "Import"
    source_type           = "Git"
    source_url            = betterado_git_repository.test.remote_url
    service_connection_id = betterado_serviceendpoint_generic_git.test.id
  }
}
`, repoInit, serviceEndpointName, importRepoName)
}

func hclGitRepositoryImportPrivateUserNamePassword(projectName, repoName, importRepoName, serviceEndpointName string) string {
	repoInit := hclGitRepositoryBasic(projectName, repoName, "Clean")
	return fmt.Sprintf(`
%s

resource "betterado_git_repository" "import" {
  project_id = betterado_project.test.id
  name       = "%s"
  initialization {
    init_type   = "Import"
    source_type = "Git"
    source_url  = betterado_git_repository.test.remote_url
    username    = "%s"
    password    = "%s"
  }
}
`, repoInit, importRepoName, os.Getenv("AZDO_GENERIC_GIT_SERVICE_CONNECTION_USERNAME"), os.Getenv("AZDO_GENERIC_GIT_SERVICE_CONNECTION_PASSWORD"))
}
