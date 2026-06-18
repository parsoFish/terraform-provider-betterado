package permissions

// ReleaseDefinitionPermissions — security namespace token format (confirmed via live ADO probe)
//
// Namespace: ReleaseManagement2
// Namespace ID: c788c23e-1b46-4162-8f5e-d7585343b5de
//
// Spike methodology (2026-06-06):
//   - Queried _apis/accesscontrollists/c788c23e-1b46-4162-8f5e-d7585343b5de on org
//     davidgparsonson; project-level tokens are plain project UUIDs.
//   - Created a release definition (ID=1) in project 21cff396-a36f-4d05-bccf-91e3a2a8b4bb.
//   - Verified via POST to _apis/accesscontrolentries/{namespaceId} with
//     token="21cff396-a36f-4d05-bccf-91e3a2a8b4bb/1" — the API accepted it and
//     returned a valid ACE (HTTP 200). Subsequent GET confirmed the token exists in the ACL.
//   - Conclusion: definition-level token = "{projectId}/{releaseDefinitionId}"
//     (identical structure to the Build namespace, no "ReleaseManagement2/Project/" prefix).
//
// Token format:
//
//	Project-level:    {projectId}
//	Definition-level: {projectId}/{releaseDefinitionId}
//
// Examples:
//
//	"21cff396-a36f-4d05-bccf-91e3a2a8b4bb"    → project-level
//	"21cff396-a36f-4d05-bccf-91e3a2a8b4bb/7"  → definition ID 7 in that project
//
// Note: The "ReleaseManagement2/Project/{projectId}/{definitionId}" format suggested
// in the WI spec was NOT observed in the live API; the actual format is simpler.

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	securityhelper "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/permissions/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// releaseDefinitionTokenFormat is the confirmed token format for ReleaseManagement2.
// Use fmt.Sprintf(releaseDefinitionTokenFormat, projectID, releaseDefinitionID) to produce
// a definition-scoped token, or just projectID alone for a project-scoped token.
const releaseDefinitionTokenFormat = "%s/%d"

// ResourceReleaseDefinitionPermissions schema and implementation for release definition permission resource
func ResourceReleaseDefinitionPermissions() *schema.Resource {
	return &schema.Resource{
		Create: resourceReleaseDefinitionPermissionsCreateOrUpdate,
		Read:   resourceReleaseDefinitionPermissionsRead,
		Update: resourceReleaseDefinitionPermissionsCreateOrUpdate,
		Delete: resourceReleaseDefinitionPermissionsDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: securityhelper.CreatePermissionResourceSchema(map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Required:     true,
				ForceNew:     true,
			},
			"release_definition_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
		}),
	}
}

func resourceReleaseDefinitionPermissionsCreateOrUpdate(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)

	sn, err := securityhelper.NewSecurityNamespace(d, clients, securityhelper.SecurityNamespaceIDValues.ReleaseManagement2, createReleaseDefinitionToken)
	if err != nil {
		return err
	}

	if err := securityhelper.SetPrincipalPermissions(d, sn, nil, false); err != nil {
		return err
	}

	return resourceReleaseDefinitionPermissionsRead(d, m)
}

func resourceReleaseDefinitionPermissionsRead(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)

	sn, err := securityhelper.NewSecurityNamespace(d, clients, securityhelper.SecurityNamespaceIDValues.ReleaseManagement2, createReleaseDefinitionToken)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	principalPermissions, err := securityhelper.GetPrincipalPermissions(d, sn)
	if err != nil {
		return err
	}
	if principalPermissions == nil {
		d.SetId("")
		log.Printf("[INFO] Permissions for ACL token %q not found. Removing from state", sn.GetToken())
		return nil
	}

	d.Set("permissions", principalPermissions.Permissions)
	return nil
}

func resourceReleaseDefinitionPermissionsDelete(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)

	sn, err := securityhelper.NewSecurityNamespace(d, clients, securityhelper.SecurityNamespaceIDValues.ReleaseManagement2, createReleaseDefinitionToken)
	if err != nil {
		return err
	}

	if err := securityhelper.SetPrincipalPermissions(d, sn, &securityhelper.PermissionTypeValues.NotSet, true); err != nil {
		return err
	}
	return nil
}

// createReleaseDefinitionToken creates the ACL token for a release definition permission
// in the ReleaseManagement2 security namespace.
//
// Token format (confirmed via live probe): "{projectId}/{releaseDefinitionId}"
// Examples:
//
//	projectID="abc123", definitionID=7  →  "abc123/7"
//	projectID="abc123", no definitionID →  "abc123"
//
// NOTE: We use GetOkExists (deprecated) rather than GetOk for release_definition_id
// because GetOk returns false for any zero value (including definitionID=0), which
// would incorrectly produce a project-scoped token for a valid definition with ID=0.
// GetOkExists checks only Exists && !Computed, which correctly distinguishes between
// "field not set" and "field set to 0". See SA1019 suppression below.
func createReleaseDefinitionToken(d *schema.ResourceData, _ *client.AggregatedClient) (string, error) {
	projectID, ok := d.GetOk("project_id")
	if !ok {
		return "", fmt.Errorf("failed to get 'project_id' from schema")
	}

	//nolint:staticcheck // SA1019: GetOkExists is deprecated but required to distinguish unset from zero (definitionID=0 must produce "projectId/0", not "projectId")
	releaseDefinitionID, ok := d.GetOkExists("release_definition_id")
	if !ok {
		// project-level token (no definition scoping)
		return projectID.(string), nil
	}

	return fmt.Sprintf(releaseDefinitionTokenFormat, projectID.(string), releaseDefinitionID.(int)), nil
}
