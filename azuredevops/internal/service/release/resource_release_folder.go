//go:build (all || resource_release_folder) && !exclude_resource_release_folder

package release

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ResourceReleaseFolder returns the schema.Resource for a betterado_release_folder.
func ResourceReleaseFolder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceReleaseFolderCreate,
		ReadContext:   resourceReleaseFolderRead,
		UpdateContext: resourceReleaseFolderUpdate,
		DeleteContext: resourceReleaseFolderDelete,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Description:  "The ID of the project where the release folder resides.",
			},
			"path": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Description: "The backslash-separated path of the folder (e.g. `\\Production\\Web`). " +
					"The root folder `\\` exists by default and cannot be created or deleted.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A description for the release folder.",
			},
		},
	}
}

// expandReleaseFolder converts Terraform state into a release.Folder SDK struct.
func expandReleaseFolder(d *schema.ResourceData) *releaseapi.Folder {
	folder := &releaseapi.Folder{
		Path:        converter.String(d.Get("path").(string)),
		Description: converter.String(d.Get("description").(string)),
	}
	return folder
}

// flattenReleaseFolder writes values from the API response back into Terraform state.
func flattenReleaseFolder(d *schema.ResourceData, folder *releaseapi.Folder) {
	if folder.Path != nil {
		d.Set("path", *folder.Path) //nolint:errcheck
	}
	if folder.Description != nil {
		d.Set("description", *folder.Description) //nolint:errcheck
	}
}

// CRUD stubs — full implementation is provided by FEAT-2 (WI-2).

func resourceReleaseFolderCreate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceReleaseFolderRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceReleaseFolderUpdate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceReleaseFolderDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
