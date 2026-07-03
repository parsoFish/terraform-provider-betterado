package acceptancetests

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

func TestAccBranchPolicyMinReviewers_basic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	name := testutils.GenerateResourceName()
	node := "betterado_branch_policy_min_reviewers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclPolicyMinReviewersBasic(projectID, 1, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "id"),
					resource.TestCheckResourceAttr(node, "blocking", "true"),
					resource.TestCheckResourceAttr(node, "enabled", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.submitter_can_vote", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.allow_completion_with_rejects_or_waits", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.last_pusher_cannot_approve", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.on_last_iteration_require_vote", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.on_push_reset_approved_votes", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.on_each_iteration_require_vote", "false"),
					// Capture a real API GET of the live policy config as forge demo evidence
					// (before destroy) — proves the demo shows the actual resource, not just test names.
					captureMinReviewersPolicyEvidence(node),
				),
			}, {
				ResourceName:      node,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(node),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBranchPolicyMinReviewers_update(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	name := testutils.GenerateResourceName()
	node := "betterado_branch_policy_min_reviewers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclPolicyMinReviewersBasic(projectID, 1, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "id"),
					resource.TestCheckResourceAttr(node, "blocking", "true"),
					resource.TestCheckResourceAttr(node, "enabled", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.submitter_can_vote", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.allow_completion_with_rejects_or_waits", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.last_pusher_cannot_approve", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.on_last_iteration_require_vote", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.on_push_reset_approved_votes", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.on_each_iteration_require_vote", "false"),
				),
			}, {
				Config: hclPolicyMinReviewersUpdate(projectID, 2, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "id"),
					resource.TestCheckResourceAttr(node, "blocking", "false"),
					resource.TestCheckResourceAttr(node, "enabled", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.submitter_can_vote", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.allow_completion_with_rejects_or_waits", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.last_pusher_cannot_approve", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.on_last_iteration_require_vote", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.on_push_reset_all_votes", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.on_each_iteration_require_vote", "true"),
				),
			}, {
				ResourceName:      node,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(node),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBranchPolicyMinReviewers_resetAllVote(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	name := testutils.GenerateResourceName()
	node := "betterado_branch_policy_min_reviewers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclPolicyMinReviewersResetAllVote(projectID, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "id"),
					resource.TestCheckResourceAttr(node, "blocking", "true"),
					resource.TestCheckResourceAttr(node, "enabled", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.submitter_can_vote", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.on_push_reset_all_votes", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.on_push_reset_approved_votes", "true"),
				),
			}, {
				ResourceName:      node,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(node),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBranchPolicyMinReviewers_requiresImportError(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	name := testutils.GenerateResourceName()
	node := "betterado_branch_policy_min_reviewers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclPolicyMinReviewersResetAllVote(projectID, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "id"),
					resource.TestCheckResourceAttr(node, "blocking", "true"),
					resource.TestCheckResourceAttr(node, "enabled", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.submitter_can_vote", "false"),
					resource.TestCheckResourceAttr(node, "settings.0.on_push_reset_all_votes", "true"),
					resource.TestCheckResourceAttr(node, "settings.0.on_push_reset_approved_votes", "true"),
				),
			}, {
				Config:      hclPolicyMinReviewersResetRequireImportError(projectID, name),
				ExpectError: regexp.MustCompile(`The update is rejected by policy`),
			},
		},
	})
}

// captureMinReviewersPolicyEvidence performs a live API GET of the created policy
// configuration and writes it to .forge/live-evidence/acceptance-resource.json.
// Best-effort: failure does not fail the test (always returns nil).
func captureMinReviewersPolicyEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if err := doCapturePolicyEvidence(s, tfNode); err != nil {
			// best-effort: evidence capture must never fail the real assertion
			_ = err
		}
		return nil
	}
}

func doCapturePolicyEvidence(s *terraform.State, tfNode string) error {
	res, ok := s.RootModule().Resources[tfNode]
	if !ok {
		return nil
	}
	policyID, err := strconv.Atoi(res.Primary.ID)
	if err != nil {
		return err
	}
	projectID := res.Primary.Attributes["project_id"]

	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
	if err != nil {
		return err
	}

	cfg, err := clients.PolicyClient.GetPolicyConfiguration(clients.Ctx, policy.GetPolicyConfigurationArgs{
		Project:         &projectID,
		ConfigurationId: &policyID,
	})
	if err != nil {
		return err
	}
	if cfg == nil {
		return nil
	}

	url := fmt.Sprintf("%s/%s/_apis/policy/configurations/%d?api-version=7.1",
		orgURL, projectID, policyID)
	return testutils.CaptureLiveEvidence("acceptance-resource-branch_policy_min_reviewers", url, cfg)
}

func hclPolicyMinReviewersTemplate(projectID, name string) string {
	return fmt.Sprintf(`
resource "betterado_git_repository" "test" {
  project_id = %[1]q
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}
`, projectID, name)
}

func hclPolicyMinReviewersBasic(projectID string, reviewers int, name string) string {
	template := hclPolicyMinReviewersTemplate(projectID, name)
	return fmt.Sprintf(`
%s

resource "betterado_branch_policy_min_reviewers" "test" {
  project_id = %[3]q
  enabled    = true
  blocking   = true
  settings {
    reviewer_count                         = %[2]d
    submitter_can_vote                     = false
    allow_completion_with_rejects_or_waits = false
    on_push_reset_approved_votes           = true
    on_each_iteration_require_vote         = false
    scope {
      repository_id  = betterado_git_repository.test.id
      repository_ref = "refs/heads/release"
      match_type     = "Exact"
    }
  }
}
`, template, reviewers, projectID)
}

func hclPolicyMinReviewersUpdate(projectID string, reviewers int, name string) string {
	template := hclPolicyMinReviewersTemplate(projectID, name)
	return fmt.Sprintf(`
%s

resource "betterado_branch_policy_min_reviewers" "test" {
  project_id = %[3]q
  enabled    = false
  blocking   = false
  settings {
    reviewer_count                         = %[2]d
    submitter_can_vote                     = true
    allow_completion_with_rejects_or_waits = true
    last_pusher_cannot_approve             = true
    on_last_iteration_require_vote         = true
    on_each_iteration_require_vote         = true
    scope {
      repository_id  = betterado_git_repository.test.id
      repository_ref = "refs/heads/release"
      match_type     = "Exact"
    }
  }
}
`, template, reviewers, projectID)
}

func hclPolicyMinReviewersResetAllVote(projectID, name string) string {
	template := hclPolicyMinReviewersTemplate(projectID, name)
	return fmt.Sprintf(`
%s

resource "betterado_branch_policy_min_reviewers" "test" {
  project_id = %[2]q
  enabled    = true
  blocking   = true
  settings {
    reviewer_count               = 2
    submitter_can_vote           = false
    on_push_reset_all_votes      = true
    on_push_reset_approved_votes = true
    scope {
      repository_id  = betterado_git_repository.test.id
      repository_ref = "refs/heads/release"
      match_type     = "Exact"
    }
  }
}
`, template, projectID)
}

func hclPolicyMinReviewersResetRequireImportError(projectID, name string) string {
	template := hclPolicyMinReviewersResetAllVote(projectID, name)
	return fmt.Sprintf(`
%s

resource "betterado_branch_policy_min_reviewers" "import" {
  project_id = betterado_branch_policy_min_reviewers.test.project_id
  enabled    = betterado_branch_policy_min_reviewers.test.enabled
  blocking   = betterado_branch_policy_min_reviewers.test.blocking
  settings {
    reviewer_count               = betterado_branch_policy_min_reviewers.test.settings.0.reviewer_count
    submitter_can_vote           = betterado_branch_policy_min_reviewers.test.settings.0.submitter_can_vote
    on_push_reset_all_votes      = betterado_branch_policy_min_reviewers.test.settings.0.on_push_reset_all_votes
    on_push_reset_approved_votes = betterado_branch_policy_min_reviewers.test.settings.0.on_push_reset_approved_votes
    scope {
      repository_id  = betterado_branch_policy_min_reviewers.test.settings.0.scope.0.repository_id
      repository_ref = betterado_branch_policy_min_reviewers.test.settings.0.scope.0.repository_ref
      match_type     = betterado_branch_policy_min_reviewers.test.settings.0.scope.0.match_type
    }
  }
}
`, template)
}
