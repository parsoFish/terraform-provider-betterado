package release

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// DataReleaseDefinition returns the schema.Resource for the betterado_release_definition data source.
func DataReleaseDefinition() *schema.Resource {
	return &schema.Resource{
		Read: dataReleaseDefinitionRead,
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
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				AtLeastOneOf: []string{"release_definition_id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"release_definition_id", "name"},
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"release_name_format": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataReleaseDefinitionRead(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)

	// Look up by ID if provided; otherwise look up by name.
	if defIDRaw, ok := d.GetOk("release_definition_id"); ok {
		defID := defIDRaw.(int)
		def, err := clients.ReleaseClient.GetReleaseDefinition(clients.Ctx, releaseapi.GetReleaseDefinitionArgs{
			Project:      &projectID,
			DefinitionId: &defID,
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				d.SetId("")
				return fmt.Errorf("release definition with ID %d not found in project %s", defID, projectID)
			}
			return fmt.Errorf("reading release definition (ID: %d): %+v", defID, err)
		}
		return flattenDataReleaseDefinition(d, def, projectID)
	}

	// Look up by name.
	name := d.Get("name").(string)
	resp, err := clients.ReleaseClient.GetReleaseDefinitions(clients.Ctx, releaseapi.GetReleaseDefinitionsArgs{
		Project:    &projectID,
		SearchText: converter.String(name),
	})
	if err != nil {
		return fmt.Errorf("listing release definitions with name %q in project %s: %+v", name, projectID, err)
	}

	var matches []releaseapi.ReleaseDefinition
	if resp != nil {
		for _, rd := range resp.Value {
			rdCopy := rd
			if rd.Name != nil && *rd.Name == name {
				matches = append(matches, rdCopy)
			}
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("no release definition with name %q found in project %s", name, projectID)
	}
	if len(matches) > 1 {
		return fmt.Errorf("multiple release definitions with name %q found in project %s; use release_definition_id to disambiguate", name, projectID)
	}

	return flattenDataReleaseDefinition(d, &matches[0], projectID)
}

func flattenDataReleaseDefinition(d *schema.ResourceData, def *releaseapi.ReleaseDefinition, projectID string) error {
	d.SetId(strconv.Itoa(*def.Id))
	d.Set("project_id", projectID)
	if def.Name != nil {
		d.Set("name", *def.Name)
	}
	if def.Path != nil {
		d.Set("path", *def.Path)
	}
	if def.Description != nil {
		d.Set("description", *def.Description)
	}
	if def.ReleaseNameFormat != nil {
		d.Set("release_name_format", *def.ReleaseNameFormat)
	}
	return nil
}
