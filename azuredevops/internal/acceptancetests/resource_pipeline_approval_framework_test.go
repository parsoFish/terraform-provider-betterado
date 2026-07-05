//go:build (all || resource_pipeline_approval) && !exclude_resource_pipeline_approval

// Package acceptancetests contains live acceptance tests for the betterado
// Terraform provider.
//
// # Pre-seeded approval required
//
// This test does NOT create a pipeline run or approval — it resolves an
// existing pending approval.  Before running the live gate you must seed a
// pending approval in the standing demo org and export its UUID:
//
//	export BETTERADO_TEST_APPROVAL_ID=<uuid-of-pending-approval>
//
// The approval is looked up by UUID and resolved to status=approved.
// Because ADO approval decisions are immutable, the destroy phase is a no-op
// (the resource is simply removed from Terraform state).
package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelinesapproval"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccPipelineApproval exercises the betterado_pipeline_approval resource
// via the mux provider path (terraform-plugin-framework):
//
//  1. apply — records an approval decision (status=approved) for a pre-seeded
//     pending approval in the standing demo org via UpdateApprovals.
//  2. read-back — asserts approval_id, status, and id are set.
//  3. evidence — CaptureLiveEvidence performs a real GET and persists the
//     response as forge demo live-evidence.
//  4. destroy — no-op clean exit (ADO approval decisions are immutable).
//
// The test requires BETTERADO_TEST_APPROVAL_ID to be set to the UUID of a
// pre-seeded pending approval in the standing demo org.
func TestAccPipelineApproval(t *testing.T) {
	approvalID := os.Getenv("BETTERADO_TEST_APPROVAL_ID")
	tfNode := "betterado_pipeline_approval.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"BETTERADO_TEST_APPROVAL_ID"}) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkPipelineApprovalDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclPipelineApprovalBasic(approvalID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "approval_id", approvalID),
					resource.TestCheckResourceAttr(tfNode, "status", "approved"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					capturePipelineApprovalEvidence(tfNode),
				),
			},
		},
	})
}

// hclPipelineApprovalBasic returns HCL that looks up the standing demo project
// and records an approval decision for the pre-seeded pending approval.
func hclPipelineApprovalBasic(approvalID string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_pipeline_approval" "test" {
  project_id  = data.betterado_project.test.id
  approval_id = %[1]q
  status      = "approved"
  comment     = "Approved by Terraform acceptance test"
}
`, approvalID, SharedFixtureProjectName)
}

// checkPipelineApprovalDestroyed is the CheckDestroy function for
// betterado_pipeline_approval.  Approval decisions cannot be revoked via the
// ADO API — destroy removes the resource from Terraform state only.
// This check always returns nil (expected no-op behaviour).
func checkPipelineApprovalDestroyed(s *terraform.State) error {
	return nil
}

// capturePipelineApprovalEvidence performs a real live GET of the approval
// and persists it as forge demo live-evidence.
// Best-effort: a capture failure never fails the test.
func capturePipelineApprovalEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		approvalID := res.Primary.Attributes["approval_id"]
		projectID := res.Primary.Attributes["project_id"]

		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort; never fail the test
		}

		approvalUUID, err := uuid.Parse(approvalID)
		if err != nil {
			return nil
		}

		approval, err := clients.PipelinesApprovalClient.GetApproval(clients.Ctx,
			pipelinesapproval.GetApprovalArgs{Project: &projectID, ApprovalId: &approvalUUID})
		if err != nil || approval == nil {
			return nil
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/pipelines/approvals/%s?api-version=7.1", orgURL, projectID, approvalID)
		_ = testutils.CaptureLiveEvidence("pipeline-approval-create", url, approval)
		return nil
	}
}
