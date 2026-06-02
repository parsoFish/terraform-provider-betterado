package acceptancetests

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// TestAccReleaseDefinition_basic tests creating and importing a minimal release definition.
func TestAccReleaseDefinition_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionBasic(name),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttr(tfNode, "path", `\`),
					resource.TestCheckResourceAttr(tfNode, "environment.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.name", "Production"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.name", "Agent job"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccReleaseDefinition_withDeploymentInput tests that deployment_input (queue_id) is correctly
// sent to and read back from the API.
func TestAccReleaseDefinition_withDeploymentInput(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithDeploymentInput(name),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.#", "1"),
					resource.TestCheckResourceAttrSet(tfNode, "environment.0.deploy_phase.0.deployment_input.0.queue_id"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.condition", "succeeded()"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.timeout_in_minutes", "0"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_withApprovalOptions tests approval_options on pre/post approvals.
func TestAccReleaseDefinition_withApprovalOptions(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithApprovalOptions(name),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.release_creator_can_be_approver", "false"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.timeout_in_minutes", "1440"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.execution_order", "beforeGates"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_withEnvironmentOptions tests environment_options and execution_policy blocks.
func TestAccReleaseDefinition_withEnvironmentOptions(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithEnvironmentOptions(name),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_options.0.publish_deployment_status", "true"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_options.0.email_notification_type", "OnlyOnFailure"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.execution_policy.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.execution_policy.0.concurrency_count", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.execution_policy.0.queue_depth_count", "0"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_update tests that a release definition can be updated (triggers revision
// increment and verifies no revision conflicts).
func TestAccReleaseDefinition_update(t *testing.T) {
	name := testutils.GenerateResourceName()
	updatedName := name + "-updated"
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionBasic(name),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
				),
			},
			{
				Config: hclReleaseDefinitionBasic(updatedName),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(updatedName),
					resource.TestCheckResourceAttr(tfNode, "name", updatedName),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_complete tests all major features together.
func TestAccReleaseDefinition_complete(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionComplete(name),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttr(tfNode, "environment.#", "2"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.name", "Staging"),
					resource.TestCheckResourceAttr(tfNode, "environment.1.name", "Production"),
					resource.TestCheckResourceAttr(tfNode, "variable.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.execution_policy.#", "1"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// --- Check functions ---

func checkReleaseDefinitionExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		def, err := getReleaseDefinitionFromState(s)
		if err != nil {
			return err
		}
		if def.Name == nil || *def.Name != expectedName {
			return fmt.Errorf("expected release definition name %q but got %v", expectedName, def.Name)
		}
		return nil
	}
}

func checkReleaseDefinitionDestroyed(s *terraform.State) error {
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "betterado_release_definition" {
			continue
		}
		if _, err := getReleaseDefinitionFromResource(resource); err == nil {
			return fmt.Errorf("unexpectedly found release definition that should be deleted")
		}
	}
	return nil
}

func getReleaseDefinitionFromState(s *terraform.State) (*release.ReleaseDefinition, error) {
	for _, res := range s.RootModule().Resources {
		if res.Type == "betterado_release_definition" {
			return getReleaseDefinitionFromResource(res)
		}
	}
	return nil, fmt.Errorf("no betterado_release_definition found in state")
}

func getReleaseDefinitionFromResource(res *terraform.ResourceState) (*release.ReleaseDefinition, error) {
	defID, err := strconv.Atoi(res.Primary.ID)
	if err != nil {
		return nil, err
	}
	projectID := res.Primary.Attributes["project_id"]
	clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
	return clients.ReleaseClient.GetReleaseDefinition(clients.Ctx, release.GetReleaseDefinitionArgs{
		Project:      &projectID,
		DefinitionId: &defID,
	})
}

// --- HCL templates ---

func hclReleaseDefinitionProjectBase(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}
`, name)
}

// hclReleaseDefinitionBasic creates a minimal release definition with one environment.
// Uses the default agent queue (queue_id omitted from deployment_input since it is optional).
func hclReleaseDefinitionBasic(name string) string {
	base := hclReleaseDefinitionProjectBase(name)
	return fmt.Sprintf(`
%s

resource "betterado_release_definition" "test" {
  name       = "%[2]s"
  project_id = betterado_project.test.id

  environment {
    name = "Production"
    rank = 1

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }
  }
}
`, base, name)
}

// hclReleaseDefinitionWithDeploymentInput uses data source to get a queue ID for deployment_input.
// Falls back to the "Azure Pipelines" hosted pool (queue 4 is the default in most orgs).
func hclReleaseDefinitionWithDeploymentInput(name string) string {
	base := hclReleaseDefinitionProjectBase(name)
	return fmt.Sprintf(`
%s

data "betterado_agent_queue" "test" {
  name       = "Azure Pipelines"
  project_id = betterado_project.test.id
}

resource "betterado_release_definition" "test" {
  name       = "%[2]s"
  project_id = betterado_project.test.id

  environment {
    name = "Production"
    rank = 1

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"

      deployment_input {
        queue_id                      = data.betterado_agent_queue.test.id
        timeout_in_minutes            = 0
        job_cancel_timeout_in_minutes = 1
        condition                     = "succeeded()"
      }
    }
  }
}
`, base, name)
}

// hclReleaseDefinitionWithApprovalOptions creates a release definition with approval_options.
func hclReleaseDefinitionWithApprovalOptions(name string) string {
	base := hclReleaseDefinitionProjectBase(name)
	return fmt.Sprintf(`
%s

resource "betterado_release_definition" "test" {
  name       = "%[2]s"
  project_id = betterado_project.test.id

  environment {
    name = "Production"
    rank = 1

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }

      approval_options {
        release_creator_can_be_approver                                 = false
        enforce_identity_revalidation                                   = false
        timeout_in_minutes                                              = 1440
        execution_order                                                 = "beforeGates"
        auto_triggered_and_previous_environment_approved_can_be_skipped = false
      }
    }

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }
  }
}
`, base, name)
}

// hclReleaseDefinitionWithEnvironmentOptions creates a release definition with environment_options
// and execution_policy blocks.
func hclReleaseDefinitionWithEnvironmentOptions(name string) string {
	base := hclReleaseDefinitionProjectBase(name)
	return fmt.Sprintf(`
%s

resource "betterado_release_definition" "test" {
  name       = "%[2]s"
  project_id = betterado_project.test.id

  environment {
    name = "Production"
    rank = 1

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }

    environment_options {
      email_notification_type   = "OnlyOnFailure"
      publish_deployment_status = true
      badge_enabled             = false
      auto_link_work_items      = false
    }

    execution_policy {
      concurrency_count = 1
      queue_depth_count = 0
    }
  }
}
`, base, name)
}

// hclReleaseDefinitionComplete creates a release definition with most features configured.
func hclReleaseDefinitionComplete(name string) string {
	base := hclReleaseDefinitionProjectBase(name)
	return fmt.Sprintf(`
%s

resource "betterado_release_definition" "test" {
  name                = "%[2]s"
  project_id          = betterado_project.test.id
  description         = "Acceptance test complete release definition"
  release_name_format = "Release-$(rev:r)"

  variable {
    name  = "MY_VAR"
    value = "hello"
  }

  environment {
    name = "Staging"
    rank = 1

    variable {
      name  = "ENV_VAR"
      value = "staging"
    }

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"

      deployment_input {
        queue_id                      = 4
        timeout_in_minutes            = 30
        job_cancel_timeout_in_minutes = 1
        condition                     = "succeeded()"
        skip_artifacts_download       = false
        enable_access_token           = false
      }
    }

    environment_options {
      email_notification_type   = "OnlyOnFailure"
      publish_deployment_status = true
    }

    execution_policy {
      concurrency_count = 1
      queue_depth_count = 0
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }
  }

  environment {
    name = "Production"
    rank = 2

    condition {
      name           = "Staging"
      condition_type = "environmentState"
      value          = "4"
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }

      approval_options {
        release_creator_can_be_approver = false
        timeout_in_minutes              = 0
        execution_order                 = "beforeGates"
      }
    }

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }
  }
}
`, base, name)
}
