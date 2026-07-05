package workitemtrackingprocess

// data_process_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_process data source.

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

var (
	_ datasource.DataSource              = &processDataSource{}
	_ datasource.DataSourceWithConfigure = &processDataSource{}
)

type processDataSource struct {
	client *client.AggregatedClient
}

func NewProcessDataSource() datasource.DataSource {
	return &processDataSource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type processDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Expand              types.String `tfsdk:"expand"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	ParentProcessTypeID types.String `tfsdk:"parent_process_type_id"`
	ReferenceName       types.String `tfsdk:"reference_name"`
	IsDefault           types.Bool   `tfsdk:"is_default"`
	IsEnabled           types.Bool   `tfsdk:"is_enabled"`
	CustomizationType   types.String `tfsdk:"customization_type"`
	Projects            types.List   `tfsdk:"projects"`
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (d *processDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_process"
}

func (d *processDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a work item tracking process from Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
			},
			"expand": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Specifies the expand option when getting the processes. Valid values: 'none', 'projects'.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the process.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the process.",
			},
			"parent_process_type_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the parent process.",
			},
			"reference_name": schema.StringAttribute{
				Computed:    true,
				Description: "Reference name of the process.",
			},
			"is_default": schema.BoolAttribute{
				Computed:    true,
				Description: "Is the process default?",
			},
			"is_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Is the process enabled?",
			},
			"customization_type": schema.StringAttribute{
				Computed:    true,
				Description: "Indicates the type of customization on this process.",
			},
			"projects": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Returns associated projects when using the 'projects' expand option.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the project.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description of the project.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the project.",
						},
						"url": schema.StringAttribute{
							Computed:    true,
							Description: "Url of the project.",
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *processDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

// ── Read ──────────────────────────────────────────────────────────────────────

var processExpandLevelMap = map[string]workitemtrackingprocess.GetProcessExpandLevel{
	"none":     workitemtrackingprocess.GetProcessExpandLevelValues.None,
	"projects": workitemtrackingprocess.GetProcessExpandLevelValues.Projects,
}

func (d *processDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_process data source Read: provider client not configured")
		return
	}

	var model processDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := model.ID.ValueString()
	expandStr := "none"
	if !model.Expand.IsNull() && !model.Expand.IsUnknown() && model.Expand.ValueString() != "" {
		expandStr = model.Expand.ValueString()
	}
	expand := processExpandLevelMap[expandStr]
	model.Expand = types.StringValue(expandStr)

	process, err := d.client.WorkItemTrackingProcessClient.GetProcessByItsId(ctx, workitemtrackingprocess.GetProcessByItsIdArgs{
		ProcessTypeId: converter.UUID(processID),
		Expand:        &expand,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError("Not found", fmt.Sprintf("process %s not found", processID))
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading process %s: %s", processID, err))
		return
	}

	if process.Name != nil {
		model.Name = types.StringValue(*process.Name)
	}
	if process.Description != nil {
		model.Description = types.StringValue(*process.Description)
	} else {
		model.Description = types.StringValue("")
	}
	if process.ParentProcessTypeId != nil {
		model.ParentProcessTypeID = types.StringValue(process.ParentProcessTypeId.String())
	}
	if process.ReferenceName != nil {
		model.ReferenceName = types.StringValue(*process.ReferenceName)
	}
	if process.IsDefault != nil {
		model.IsDefault = types.BoolValue(*process.IsDefault)
	}
	if process.IsEnabled != nil {
		model.IsEnabled = types.BoolValue(*process.IsEnabled)
	}
	if process.CustomizationType != nil {
		model.CustomizationType = types.StringValue(string(*process.CustomizationType))
	}

	// Build the projects list.
	projectAttrTypes := map[string]attr.Type{
		"id":          types.StringType,
		"description": types.StringType,
		"name":        types.StringType,
		"url":         types.StringType,
	}
	projectObjectType := types.ObjectType{AttrTypes: projectAttrTypes}

	if process.Projects != nil && len(*process.Projects) > 0 {
		projectObjs := make([]attr.Value, 0, len(*process.Projects))
		for _, p := range *process.Projects {
			attrs := map[string]attr.Value{
				"id":          types.StringNull(),
				"description": types.StringNull(),
				"name":        types.StringNull(),
				"url":         types.StringNull(),
			}
			if p.Id != nil {
				attrs["id"] = types.StringValue(p.Id.String())
			}
			if p.Description != nil {
				attrs["description"] = types.StringValue(*p.Description)
			}
			if p.Name != nil {
				attrs["name"] = types.StringValue(*p.Name)
			}
			if p.Url != nil {
				attrs["url"] = types.StringValue(*p.Url)
			}
			obj, diag := types.ObjectValue(projectAttrTypes, attrs)
			resp.Diagnostics.Append(diag...)
			projectObjs = append(projectObjs, obj)
		}
		list, diag := types.ListValue(projectObjectType, projectObjs)
		resp.Diagnostics.Append(diag...)
		model.Projects = list
	} else {
		model.Projects = types.ListValueMust(projectObjectType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
