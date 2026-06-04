package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ResourceReleaseFolder schema and implementation for release folder resource
func ResourceReleaseFolder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceReleaseFolderCreate,
		ReadContext:   resourceReleaseFolderRead,
		UpdateContext: resourceReleaseFolderUpdate,
		DeleteContext: resourceReleaseFolderDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
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

// CRUD Operations

func resourceReleaseFolderCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	path := d.Get("path").(string)
	description := d.Get("description").(string)

	folder, err := clients.ReleaseClient.CreateFolder(clients.Ctx, releaseapi.CreateFolderArgs{
		Folder: &releaseapi.Folder{
			Path:        converter.String(path),
			Description: converter.String(description),
		},
		Project: &projectID,
	})
	if err != nil {
		return diag.Errorf("creating release folder '%s': %+v", path, err)
	}

	d.SetId(releaseFolderID(projectID, converter.ToString(folder.Path, path)))
	return nil
}

func resourceReleaseFolderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID, path, err := parseReleaseFolderID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	folders, err := clients.ReleaseClient.GetFolders(clients.Ctx, releaseapi.GetFoldersArgs{
		Project: &projectID,
		Path:    &path,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("reading release folder '%s': %+v", path, err)
	}

	// GetFolders returns a list; find our exact path
	if folders == nil || len(*folders) == 0 {
		d.SetId("")
		return nil
	}

	var found *releaseapi.Folder
	for i := range *folders {
		f := (*folders)[i]
		if f.Path != nil && *f.Path == path {
			found = &f
			break
		}
	}
	if found == nil {
		d.SetId("")
		return nil
	}

	flattenReleaseFolder(d, found, projectID)
	return nil
}

func resourceReleaseFolderUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	path := d.Get("path").(string)
	description := d.Get("description").(string)

	_, err := clients.ReleaseClient.UpdateFolder(clients.Ctx, releaseapi.UpdateFolderArgs{
		Folder: &releaseapi.Folder{
			Path:        converter.String(path),
			Description: converter.String(description),
		},
		Project: &projectID,
		Path:    converter.String(path),
	})
	if err != nil {
		return diag.Errorf("updating release folder '%s': %+v", path, err)
	}

	return resourceReleaseFolderRead(ctx, d, m)
}

func resourceReleaseFolderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	path := d.Get("path").(string)

	if path == "\\" {
		return diag.Errorf("Cannot delete the root release folder '\\'")
	}

	err := clients.ReleaseClient.DeleteFolder(clients.Ctx, releaseapi.DeleteFolderArgs{
		Project: &projectID,
		Path:    converter.String(path),
	})
	if err != nil {
		if strings.Contains(err.Error(), "contains release definitions") ||
			strings.Contains(err.Error(), "release definitions") {
			return diag.Errorf("Cannot delete folder '%s' because it contains release definitions. Move or delete the definitions first.", path)
		}
		return diag.Errorf("deleting release folder '%s': %+v", path, err)
	}

	return nil
}

// Expand/Flatten helpers

func expandReleaseFolder(d *schema.ResourceData) (*releaseapi.Folder, string) {
	projectID := d.Get("project_id").(string)
	path := d.Get("path").(string)
	description := d.Get("description").(string)

	return &releaseapi.Folder{
		Path:        converter.String(path),
		Description: converter.String(description),
	}, projectID
}

func flattenReleaseFolder(d *schema.ResourceData, folder *releaseapi.Folder, projectID string) {
	d.Set("project_id", projectID)
	if folder.Path != nil {
		d.Set("path", *folder.Path)
	}
	if folder.Description != nil {
		d.Set("description", *folder.Description)
	}
}

// ID helpers

func releaseFolderID(projectID, path string) string {
	return fmt.Sprintf("%s/%s", projectID, path)
}

func parseReleaseFolderID(id string) (string, string, error) {
	// ID format: "<projectID>/<path>" where path begins with backslash
	// e.g. "abc-123/\Production\Web"
	// Split on first "/" only
	idx := strings.Index(id, "/")
	if idx < 0 {
		return "", "", fmt.Errorf("invalid release folder ID (expected '<projectID>/<path>'): %s", id)
	}
	projectID := id[:idx]
	path := id[idx+1:]
	if projectID == "" || path == "" {
		return "", "", fmt.Errorf("invalid release folder ID (empty component): %s", id)
	}
	return projectID, path, nil
}
