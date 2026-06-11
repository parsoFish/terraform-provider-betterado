package release

import (
	"fmt"
	"io"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// DataReleaseDefinitionRevision returns the schema.Resource for the
// betterado_release_definition_revision data source.
func DataReleaseDefinitionRevision() *schema.Resource {
	return &schema.Resource{
		Read: dataReleaseDefinitionRevisionRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"release_definition_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"revision": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"json_content": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataReleaseDefinitionRevisionRead(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	defID := d.Get("release_definition_id").(int)
	rev := d.Get("revision").(int)

	rc, err := clients.ReleaseClient.GetDefinitionRevision(clients.Ctx, releaseapi.GetDefinitionRevisionArgs{
		Project:      &projectID,
		DefinitionId: &defID,
		Revision:     &rev,
	})
	if err != nil {
		d.SetId("")
		return fmt.Errorf("reading release definition revision (def: %d, rev: %d): %+v", defID, rev, err)
	}
	defer rc.Close()

	raw, err := io.ReadAll(rc)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("draining release definition revision response (def: %d, rev: %d): %+v", defID, rev, err)
	}

	d.SetId(fmt.Sprintf("%s-%d-%d", projectID, defID, rev))
	d.Set("json_content", string(raw))
	return nil
}
