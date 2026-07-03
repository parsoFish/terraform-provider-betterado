package workitemtrackingprocess

// data_work_item_types_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_workitemtypes data source.

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

var _ datasource.DataSource = &workItemTypesDataSource{}
var _ datasource.DataSourceWithConfigure = &workItemTypesDataSource{}

type workItemTypesDataSource struct {
	client *client.AggregatedClient
}

// NewWorkItemTypesDataSource returns a new datasource.DataSource.
func NewWorkItemTypesDataSource() datasource.DataSource {
	return &workItemTypesDataSource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type workItemTypesDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProcessID     types.String `tfsdk:"process_id"`
	WorkItemTypes types.Set    `tfsdk:"work_item_types"`
}

// ── Attr types ────────────────────────────────────────────────────────────────

var witElemAttrTypes = map[string]attr.Type{
	"reference_name":                  types.StringType,
	"name":                            types.StringType,
	"description":                     types.StringType,
	"color":                           types.StringType,
	"icon":                            types.StringType,
	"is_enabled":                      types.BoolType,
	"parent_work_item_reference_name": types.StringType,
	"customization":                   types.StringType,
	"url":                             types.StringType,
}
var witElemObjectType = types.ObjectType{AttrTypes: witElemAttrTypes}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *workItemTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_workitemtypes"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *workItemTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads all work item types for an Azure DevOps process.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the process (used as data source ID).",
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
			},
			"work_item_types": schema.SetNestedAttribute{
				Computed:    true,
				Description: "A set of work item types for the process.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"reference_name": schema.StringAttribute{
							Computed:    true,
							Description: "Reference name of the work item type.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the work item type.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description of the work item type.",
						},
						"color": schema.StringAttribute{
							Computed:    true,
							Description: "Color hexadecimal code to represent the work item type.",
						},
						"icon": schema.StringAttribute{
							Computed:    true,
							Description: "Icon to represent the work item type.",
						},
						"is_enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Indicates if the work item type is enabled.",
						},
						"parent_work_item_reference_name": schema.StringAttribute{
							Computed:    true,
							Description: "Parent work item type reference name.",
						},
						"customization": schema.StringAttribute{
							Computed:    true,
							Description: "Indicates the type of customization on this work item type.",
						},
						"url": schema.StringAttribute{
							Computed:    true,
							Description: "URL of the work item type.",
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *workItemTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workItemTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_workitemtypes data source Read: provider client not configured")
		return
	}

	var model workItemTypesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := model.ProcessID.ValueString()
	expand := workitemtrackingprocess.GetWorkItemTypeExpandValues.None

	wits, err := d.client.WorkItemTrackingProcessClient.GetProcessWorkItemTypes(ctx, workitemtrackingprocess.GetProcessWorkItemTypesArgs{
		ProcessId: converter.UUID(processID),
		Expand:    &expand,
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading work item types for process %s: %s", processID, err))
		return
	}

	model.ID = types.StringValue(processID)

	elems := make([]attr.Value, 0, len(*wits))
	for _, wit := range *wits {
		attrs := map[string]attr.Value{
			"reference_name":                  types.StringValue(witStrVal(wit.ReferenceName)),
			"name":                            types.StringValue(witStrVal(wit.Name)),
			"description":                     types.StringValue(""),
			"color":                           types.StringValue(""),
			"icon":                            types.StringValue(witStrVal(wit.Icon)),
			"is_enabled":                      types.BoolValue(true),
			"parent_work_item_reference_name": types.StringValue(""),
			"customization":                   types.StringValue(""),
			"url":                             types.StringValue(witStrVal(wit.Url)),
		}
		if wit.Description != nil {
			attrs["description"] = types.StringValue(*wit.Description)
		}
		if wit.Color != nil {
			attrs["color"] = types.StringValue(witColorToResource(*wit.Color))
		}
		if wit.IsDisabled != nil {
			attrs["is_enabled"] = types.BoolValue(!*wit.IsDisabled)
		}
		if wit.Inherits != nil {
			attrs["parent_work_item_reference_name"] = types.StringValue(*wit.Inherits)
		}
		if wit.Customization != nil {
			attrs["customization"] = types.StringValue(string(*wit.Customization))
		}

		obj, d2 := types.ObjectValue(witElemAttrTypes, attrs)
		resp.Diagnostics.Append(d2...)
		elems = append(elems, obj)
	}

	witSet, d2 := types.SetValue(witElemObjectType, elems)
	resp.Diagnostics.Append(d2...)
	model.WorkItemTypes = witSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
