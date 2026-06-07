package release

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// DataReleaseFolder returns the schema.Resource for the betterado_release_folder data source.
func DataReleaseFolder() *schema.Resource {
	return &schema.Resource{
		Read: dataReleaseFolderRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"path": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataReleaseFolderRead(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	path := d.Get("path").(string)

	folders, err := clients.ReleaseClient.GetFolders(clients.Ctx, releaseapi.GetFoldersArgs{
		Project: &projectID,
		Path:    &path,
	})
	if err != nil {
		return fmt.Errorf("reading release folder (path: %s): %w", path, err)
	}
	if folders == nil || len(*folders) == 0 {
		return fmt.Errorf("release folder not found: path %q in project %s", path, projectID)
	}
	folder := (*folders)[0]
	// reuse the resource's flatten helper
	flattenReleaseFolder(d, &folder, projectID)
	return nil
}
