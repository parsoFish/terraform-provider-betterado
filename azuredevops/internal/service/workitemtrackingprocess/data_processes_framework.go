package workitemtrackingprocess

// data_processes_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_processes data source.

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

var _ datasource.DataSource = &processesDataSource{}
var _ datasource.DataSourceWithConfigure = &processesDataSource{}

type processesDataSource struct {
	client *client.AggregatedClient
}

func NewProcessesDataSource() datasource.DataSource {
	return &processesDataSource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type processesDataSourceModel struct {
	Expand    types.String `tfsdk:"expand"`
	Processes types.List   `tfsdk:"processes"`
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (d *processesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_processes"
}

func (d *processesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	projectAttrTypes := map[string]schema.Attribute{
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
	}

	resp.Schema = schema.Schema{
		Description: "Reads all work item tracking processes from Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"expand": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Specifies the expand option when getting the processes. Valid values: 'none', 'projects'.",
			},
			"processes": schema.ListNestedAttribute{
				Computed:    true,
				Description: "A list of all processes including system and inherited.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the process.",
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
								Attributes: projectAttrTypes,
							},
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *processesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *processesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_processes data source Read: provider client not configured")
		return
	}

	var model processesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	expandStr := "none"
	if !model.Expand.IsNull() && !model.Expand.IsUnknown() && model.Expand.ValueString() != "" {
		expandStr = model.Expand.ValueString()
	}
	model.Expand = types.StringValue(expandStr)
	expand := processExpandLevelMap[expandStr]

	retrievedProcesses, err := d.client.WorkItemTrackingProcessClient.GetListOfProcesses(ctx, workitemtrackingprocess.GetListOfProcessesArgs{
		Expand: &expand,
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading list of processes: %s", err))
		return
	}

	// Build the nested object attr types.
	projectAttrTypes := map[string]attr.Type{
		"id":          types.StringType,
		"description": types.StringType,
		"name":        types.StringType,
		"url":         types.StringType,
	}
	projectObjectType := types.ObjectType{AttrTypes: projectAttrTypes}

	processAttrTypes := map[string]attr.Type{
		"id":                     types.StringType,
		"name":                   types.StringType,
		"description":            types.StringType,
		"parent_process_type_id": types.StringType,
		"reference_name":         types.StringType,
		"is_default":             types.BoolType,
		"is_enabled":             types.BoolType,
		"customization_type":     types.StringType,
		"projects":               types.ListType{ElemType: projectObjectType},
	}
	processObjectType := types.ObjectType{AttrTypes: processAttrTypes}

	processObjs := make([]attr.Value, 0)
	for _, p := range *retrievedProcesses {
		// Build projects sub-list.
		projectObjs := make([]attr.Value, 0)
		if p.Projects != nil {
			for _, proj := range *p.Projects {
				projAttrs := map[string]attr.Value{
					"id":          types.StringNull(),
					"description": types.StringNull(),
					"name":        types.StringNull(),
					"url":         types.StringNull(),
				}
				if proj.Id != nil {
					projAttrs["id"] = types.StringValue(proj.Id.String())
				}
				if proj.Description != nil {
					projAttrs["description"] = types.StringValue(*proj.Description)
				}
				if proj.Name != nil {
					projAttrs["name"] = types.StringValue(*proj.Name)
				}
				if proj.Url != nil {
					projAttrs["url"] = types.StringValue(*proj.Url)
				}
				projObj, diag := types.ObjectValue(projectAttrTypes, projAttrs)
				resp.Diagnostics.Append(diag...)
				projectObjs = append(projectObjs, projObj)
			}
		}
		projectsList := types.ListValueMust(projectObjectType, projectObjs)

		// Build process attrs.
		procAttrs := map[string]attr.Value{
			"id":                     types.StringNull(),
			"name":                   types.StringNull(),
			"description":            types.StringNull(),
			"parent_process_type_id": types.StringNull(),
			"reference_name":         types.StringNull(),
			"is_default":             types.BoolNull(),
			"is_enabled":             types.BoolNull(),
			"customization_type":     types.StringNull(),
			"projects":               projectsList,
		}
		if p.TypeId != nil {
			procAttrs["id"] = types.StringValue(p.TypeId.String())
		}
		if p.Name != nil {
			procAttrs["name"] = types.StringValue(*p.Name)
		}
		if p.Description != nil {
			procAttrs["description"] = types.StringValue(*p.Description)
		} else {
			procAttrs["description"] = types.StringValue("")
		}
		if p.ParentProcessTypeId != nil {
			procAttrs["parent_process_type_id"] = types.StringValue(p.ParentProcessTypeId.String())
		}
		if p.ReferenceName != nil {
			procAttrs["reference_name"] = types.StringValue(*p.ReferenceName)
		}
		if p.IsDefault != nil {
			procAttrs["is_default"] = types.BoolValue(*p.IsDefault)
		}
		if p.IsEnabled != nil {
			procAttrs["is_enabled"] = types.BoolValue(*p.IsEnabled)
		}
		if p.CustomizationType != nil {
			procAttrs["customization_type"] = types.StringValue(string(*p.CustomizationType))
		}

		procObj, diag := types.ObjectValue(processAttrTypes, procAttrs)
		resp.Diagnostics.Append(diag...)
		processObjs = append(processObjs, procObj)
	}

	processList, diag := types.ListValue(processObjectType, processObjs)
	resp.Diagnostics.Append(diag...)
	model.Processes = processList

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
