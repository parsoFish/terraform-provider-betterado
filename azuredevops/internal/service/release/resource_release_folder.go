package release

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/tfhelper"
)

// ResourceReleaseFolder schema and implementation for release folder resource
func ResourceReleaseFolder() *schema.Resource {
	return &schema.Resource{
		Create: resourceReleaseFolderCreate,
		Read:   resourceReleaseFolderRead,
		Update: resourceReleaseFolderUpdate,
		Delete: resourceReleaseFolderDelete,
		Importer: tfhelper.ImportProjectQualifiedResource(),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceReleaseFolderCreate(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)

	projectID := d.Get("project_id").(string)
	path := d.Get("path").(string)

	createdFolder, err := clients.ReleaseClient.CreateFolder(clients.Ctx, releaseapi.CreateFolderArgs{
		Folder: &releaseapi.Folder{
			Path: converter.String(path),
		},
		Project: converter.String(projectID),
	})
	if err != nil {
		return fmt.Errorf("failed creating release folder: %w", err)
	}

	if createdFolder == nil || createdFolder.Path == nil {
		return fmt.Errorf("release folder create returned nil or empty path")
	}

	d.SetId(fmt.Sprintf("%s/%s", projectID, *createdFolder.Path))
	return resourceReleaseFolderRead(d, m)
}

func resourceReleaseFolderRead(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)

	projectID := d.Get("project_id").(string)
	path := d.Get("path").(string)

	folders, err := clients.ReleaseClient.GetFolders(clients.Ctx, releaseapi.GetFoldersArgs{
		Project: converter.String(projectID),
		Path:    converter.String(path),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			d.SetId("")
			log.Printf("[TRACE] plugin.terraform-provider-betterado: Release Folder [%s] not found. Removing from state.", path)
			return nil
		}
		return err
	}

	if folders == nil || len(*folders) == 0 {
		d.SetId("")
		log.Printf("[TRACE] plugin.terraform-provider-betterado: Release Folder [%s] not found. Removing from state.", path)
		return nil
	}

	folder := (*folders)[0]

	d.Set("project_id", projectID)

	if folder.Path != nil {
		d.Set("path", folder.Path)
		// name is the last segment of path
		segments := strings.Split(*folder.Path, "\\")
		if len(segments) > 0 {
			d.Set("name", segments[len(segments)-1])
		}
	}

	return nil
}

func resourceReleaseFolderUpdate(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)

	oldPathRaw, newPathRaw := d.GetChange("path")
	oldPath := oldPathRaw.(string)
	newPath := newPathRaw.(string)
	projectID := d.Get("project_id").(string)

	_, err := clients.ReleaseClient.UpdateFolder(clients.Ctx, releaseapi.UpdateFolderArgs{
		Folder: &releaseapi.Folder{
			Path: converter.String(newPath),
		},
		Project: converter.String(projectID),
		Path:    converter.String(oldPath),
	})
	if err != nil {
		return fmt.Errorf("failed updating release folder (project: %s, path: %s): %w", projectID, oldPath, err)
	}

	return resourceReleaseFolderRead(d, m)
}

func resourceReleaseFolderDelete(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)

	err := clients.ReleaseClient.DeleteFolder(clients.Ctx, releaseapi.DeleteFolderArgs{
		Project: converter.String(d.Get("project_id").(string)),
		Path:    converter.String(d.Get("path").(string)),
	})
	if err != nil {
		return fmt.Errorf("failed deleting release folder: %w", err)
	}

	return nil
}
