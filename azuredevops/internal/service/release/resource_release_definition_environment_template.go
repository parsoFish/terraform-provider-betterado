package release

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/environmenttemplates"
)

// ResourceReleaseDefinitionEnvironmentTemplate returns the schema and CRUD implementation
// for the betterado_release_definition_environment_template resource.
//
// Templates are immutable once created (no UpdateContext).
func ResourceReleaseDefinitionEnvironmentTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceReleaseDefinitionEnvironmentTemplateCreate,
		ReadContext:   resourceReleaseDefinitionEnvironmentTemplateRead,
		DeleteContext: resourceReleaseDefinitionEnvironmentTemplateDelete,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Description:  "The ID of the project.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Description:  "The name of the environment template.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "The description of the environment template.",
			},
			"category": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "The category of the environment template.",
			},
			"environment": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "JSON-encoded ReleaseDefinitionEnvironment used to create this template.",
			},
			"icon_task_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Description:  "The ID of the task used to display the icon for this template.",
			},
			"can_delete": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether the template can be deleted.",
			},
			"icon_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The icon URI of the template.",
			},
		},
	}
}

func resourceReleaseDefinitionEnvironmentTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)

	projectID := d.Get("project_id").(string)
	template := expandEnvironmentTemplate(d)

	result, err := clients.EnvironmentTemplatesClient.SaveEnvironmentTemplate(ctx, &environmenttemplates.SaveEnvironmentTemplateArgs{
		Project:  &projectID,
		Template: template,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("creating release definition environment template: %w", err))
	}

	if result == nil || result.Id == nil {
		return diag.FromErr(fmt.Errorf("creating release definition environment template: API returned nil result"))
	}

	d.SetId(result.Id.String())
	return resourceReleaseDefinitionEnvironmentTemplateRead(ctx, d, m)
}

func resourceReleaseDefinitionEnvironmentTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)

	projectID := d.Get("project_id").(string)
	templateID, err := uuid.Parse(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("parsing template ID %q: %w", d.Id(), err))
	}

	template, err := clients.EnvironmentTemplatesClient.GetEnvironmentTemplate(ctx, &environmenttemplates.GetEnvironmentTemplateArgs{
		Project:    &projectID,
		TemplateId: &templateID,
	})
	if err != nil {
		// On 404 treat as destroyed
		d.SetId("")
		return nil
	}

	if template == nil {
		d.SetId("")
		return nil
	}

	flattenEnvironmentTemplate(template, d)
	return nil
}

func resourceReleaseDefinitionEnvironmentTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)

	projectID := d.Get("project_id").(string)
	templateID, err := uuid.Parse(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("parsing template ID %q: %w", d.Id(), err))
	}

	err = clients.EnvironmentTemplatesClient.DeleteEnvironmentTemplate(ctx, &environmenttemplates.DeleteEnvironmentTemplateArgs{
		Project:    &projectID,
		TemplateId: &templateID,
	})
	if err != nil {
		// 404 means already gone — treat as success
		return nil
	}

	return nil
}

// expandEnvironmentTemplate converts Terraform resource data into a ReleaseDefinitionEnvironmentTemplate.
func expandEnvironmentTemplate(d *schema.ResourceData) *environmenttemplates.ReleaseDefinitionEnvironmentTemplate {
	template := &environmenttemplates.ReleaseDefinitionEnvironmentTemplate{}

	if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		template.Name = &name
	}

	if v, ok := d.GetOk("description"); ok {
		desc := v.(string)
		template.Description = &desc
	}

	if v, ok := d.GetOk("category"); ok {
		cat := v.(string)
		template.Category = &cat
	}

	if v, ok := d.GetOk("icon_task_id"); ok {
		iconTaskIDStr := v.(string)
		iconTaskID, err := uuid.Parse(iconTaskIDStr)
		if err == nil {
			template.IconTaskId = &iconTaskID
		}
	}

	return template
}

// flattenEnvironmentTemplate reads a ReleaseDefinitionEnvironmentTemplate and writes its values into the ResourceData.
func flattenEnvironmentTemplate(template *environmenttemplates.ReleaseDefinitionEnvironmentTemplate, d *schema.ResourceData) {
	if template.Name != nil {
		d.Set("name", *template.Name) //nolint:errcheck
	}
	if template.Description != nil {
		d.Set("description", *template.Description) //nolint:errcheck
	}
	if template.Category != nil {
		d.Set("category", *template.Category) //nolint:errcheck
	}
	if template.CanDelete != nil {
		d.Set("can_delete", *template.CanDelete) //nolint:errcheck
	}
	if template.IconUri != nil {
		d.Set("icon_uri", *template.IconUri) //nolint:errcheck
	}
}
