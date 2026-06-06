package release

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// DataReleaseDefinitions returns the schema.Resource for the betterado_release_definitions
// list data source, which reads all release definitions in a project, optionally filtered
// by folder path.
func DataReleaseDefinitions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataReleaseDefinitionsRead,
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
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"release_definitions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataReleaseDefinitionsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	path := d.Get("path").(string)
	name := d.Get("name").(string)

	args := releaseapi.GetReleaseDefinitionsArgs{
		Project: &projectID,
	}
	if path != "" {
		args.Path = &path
	}
	if name != "" {
		args.SearchText = &name
	}

	resp, err := clients.ReleaseClient.GetReleaseDefinitions(clients.Ctx, args)
	if err != nil {
		return diag.FromErr(fmt.Errorf("reading release definitions for project %s: %+v", projectID, err))
	}

	var defs []releaseapi.ReleaseDefinition
	if resp != nil {
		defs = resp.Value
	}

	flattened := flattenReleaseDefinitions(defs)
	if err := d.Set("release_definitions", flattened); err != nil {
		return diag.FromErr(fmt.Errorf("setting release_definitions: %+v", err))
	}

	// Compute a stable ID based on project + path + name so the data source
	// is re-read when filters change.
	h := sha1.New()
	h.Write([]byte(projectID + ":" + path + ":" + name)) //nolint:errcheck
	d.SetId("release-definitions#" + base64.URLEncoding.EncodeToString(h.Sum(nil)))

	return nil
}

func flattenReleaseDefinitions(defs []releaseapi.ReleaseDefinition) []interface{} {
	results := make([]interface{}, 0, len(defs))
	for _, def := range defs {
		m := map[string]interface{}{
			"id":   "",
			"name": "",
			"path": "",
		}
		if def.Id != nil {
			m["id"] = strconv.Itoa(*def.Id)
		}
		if def.Name != nil {
			m["name"] = *def.Name
		}
		if def.Path != nil {
			m["path"] = *def.Path
		}
		results = append(results, m)
	}
	return results
}
