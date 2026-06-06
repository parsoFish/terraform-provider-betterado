package release

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ResourceReleaseFolder returns the schema and CRUD implementation for the betterado_release_folder resource.
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
			},
			"path": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func resourceReleaseFolderCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	path := d.Get("path").(string)
	description := d.Get("description").(string)

	_, err := clients.ReleaseClient.CreateFolder(clients.Ctx, releaseapi.CreateFolderArgs{
		Folder: &releaseapi.Folder{
			Path:        converter.String(path),
			Description: converter.String(description),
		},
		Project: &projectID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("creating release folder: %w", err))
	}

	d.SetId(path)
	return resourceReleaseFolderRead(ctx, d, m)
}

func resourceReleaseFolderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	path := d.Id()

	folders, err := clients.ReleaseClient.GetFolders(clients.Ctx, releaseapi.GetFoldersArgs{
		Project: &projectID,
		Path:    &path,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("reading release folder (path: %s): %w", path, err))
	}

	if folders == nil || len(*folders) == 0 {
		d.SetId("")
		return nil
	}

	folder := (*folders)[0]
	flattenReleaseFolder(d, &folder, projectID)
	return nil
}

func resourceReleaseFolderUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	path := d.Id()
	description := d.Get("description").(string)

	_, err := clients.ReleaseClient.UpdateFolder(clients.Ctx, releaseapi.UpdateFolderArgs{
		Folder: &releaseapi.Folder{
			Path:        converter.String(path),
			Description: converter.String(description),
		},
		Project: &projectID,
		Path:    &path,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("updating release folder (path: %s): %w", path, err))
	}

	return resourceReleaseFolderRead(ctx, d, m)
}

func resourceReleaseFolderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	path := d.Id()

	err := clients.ReleaseClient.DeleteFolder(clients.Ctx, releaseapi.DeleteFolderArgs{
		Project: &projectID,
		Path:    &path,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("deleting release folder (path: %s): %w", path, err))
	}

	return nil
}

// expandReleaseFolder builds a Folder struct from Terraform resource data.
func expandReleaseFolder(d *schema.ResourceData) (*releaseapi.Folder, string) {
	projectID := d.Get("project_id").(string)
	path := d.Get("path").(string)
	description := d.Get("description").(string)

	return &releaseapi.Folder{
		Path:        converter.String(path),
		Description: converter.String(description),
	}, projectID
}

// flattenReleaseFolder writes an API Folder back into Terraform resource data.
func flattenReleaseFolder(d *schema.ResourceData, folder *releaseapi.Folder, projectID string) {
	if folder.Path != nil {
		d.SetId(*folder.Path)
		d.Set("path", *folder.Path)
	}
	d.Set("project_id", projectID)
	if folder.Description != nil {
		d.Set("description", *folder.Description)
	} else {
		d.Set("description", "")
	}
}
