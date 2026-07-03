package workitemtrackingprocess

// data_work_item_type_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_workitemtype data source.

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

var _ datasource.DataSource = &workItemTypeDataSource{}
var _ datasource.DataSourceWithConfigure = &workItemTypeDataSource{}

type workItemTypeDataSource struct {
	client *client.AggregatedClient
}

// NewWorkItemTypeDataSource returns a new datasource.DataSource.
func NewWorkItemTypeDataSource() datasource.DataSource {
	return &workItemTypeDataSource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type workItemTypeDataSourceModel struct {
	ID                          types.String `tfsdk:"id"`
	ProcessID                   types.String `tfsdk:"process_id"`
	ReferenceName               types.String `tfsdk:"reference_name"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	Color                       types.String `tfsdk:"color"`
	Icon                        types.String `tfsdk:"icon"`
	IsEnabled                   types.Bool   `tfsdk:"is_enabled"`
	ParentWorkItemReferenceName types.String `tfsdk:"parent_work_item_reference_name"`
	Customization               types.String `tfsdk:"customization"`
	URL                         types.String `tfsdk:"url"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *workItemTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_workitemtype"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *workItemTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a work item type from an Azure DevOps process.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The reference name of the work item type (used as ID).",
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
			},
			"reference_name": schema.StringAttribute{
				Required:    true,
				Description: "The reference name of the work item type.",
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
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *workItemTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workItemTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_workitemtype data source Read: provider client not configured")
		return
	}

	var model workItemTypeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := model.ProcessID.ValueString()
	referenceName := model.ReferenceName.ValueString()
	expand := workitemtrackingprocess.GetWorkItemTypeExpandValues.None

	wit, err := d.client.WorkItemTrackingProcessClient.GetProcessWorkItemType(ctx, workitemtrackingprocess.GetProcessWorkItemTypeArgs{
		ProcessId:  converter.UUID(processID),
		WitRefName: &referenceName,
		Expand:     &expand,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError("Not found", fmt.Sprintf("work item type %s in process %s not found", referenceName, processID))
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading work item type %s for process %s: %s", referenceName, processID, err))
		return
	}

	if wit.ReferenceName != nil {
		model.ID = types.StringValue(*wit.ReferenceName)
		model.ReferenceName = types.StringValue(*wit.ReferenceName)
	}
	if wit.Name != nil {
		model.Name = types.StringValue(*wit.Name)
	}
	if wit.Description != nil {
		model.Description = types.StringValue(*wit.Description)
	} else {
		model.Description = types.StringValue("")
	}
	if wit.Color != nil {
		model.Color = types.StringValue(witColorToResource(*wit.Color))
	}
	if wit.Icon != nil {
		model.Icon = types.StringValue(*wit.Icon)
	}
	if wit.IsDisabled != nil {
		model.IsEnabled = types.BoolValue(!*wit.IsDisabled)
	} else {
		model.IsEnabled = types.BoolValue(true)
	}
	if wit.Inherits != nil {
		model.ParentWorkItemReferenceName = types.StringValue(*wit.Inherits)
	} else {
		model.ParentWorkItemReferenceName = types.StringValue("")
	}
	if wit.Customization != nil {
		model.Customization = types.StringValue(string(*wit.Customization))
	} else {
		model.Customization = types.StringValue("")
	}
	if wit.Url != nil {
		model.URL = types.StringValue(*wit.Url)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
