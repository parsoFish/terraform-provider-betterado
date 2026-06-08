package release

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// DataReleaseDefinitionHistory returns the schema.Resource for the
// betterado_release_definition_history data source, which lists all revisions
// of a release definition.
func DataReleaseDefinitionHistory() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataReleaseDefinitionHistoryRead,
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
			"revisions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"revision": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"changed_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"changed_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"change_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"comment": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataReleaseDefinitionHistoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	defID := d.Get("release_definition_id").(int)

	revisions, err := clients.ReleaseClient.GetReleaseDefinitionHistory(clients.Ctx, releaseapi.GetReleaseDefinitionHistoryArgs{
		Project:      &projectID,
		DefinitionId: &defID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("reading release definition history (project: %s, def: %d): %+v", projectID, defID, err))
	}

	flattened := flattenReleaseDefinitionHistory(revisions)
	if err := d.Set("revisions", flattened); err != nil {
		return diag.FromErr(fmt.Errorf("setting revisions: %+v", err))
	}

	d.SetId(fmt.Sprintf("release-definition-history-%s-%d", projectID, defID))
	return nil
}

func flattenReleaseDefinitionHistory(revs *[]releaseapi.ReleaseDefinitionRevision) []interface{} {
	if revs == nil {
		return []interface{}{}
	}
	results := make([]interface{}, 0, len(*revs))
	for _, rev := range *revs {
		m := map[string]interface{}{
			"revision":     0,
			"changed_by":   "",
			"changed_date": "",
			"change_type":  "",
			"comment":      "",
		}
		if rev.Revision != nil {
			m["revision"] = *rev.Revision
		}
		if rev.ChangedBy != nil && rev.ChangedBy.DisplayName != nil {
			m["changed_by"] = *rev.ChangedBy.DisplayName
		}
		if rev.ChangedDate != nil {
			m["changed_date"] = rev.ChangedDate.Time.Format(time.RFC3339)
		}
		if rev.ChangeType != nil {
			m["change_type"] = string(*rev.ChangeType)
		}
		if rev.Comment != nil {
			m["comment"] = *rev.Comment
		}
		results = append(results, m)
	}
	return results
}
