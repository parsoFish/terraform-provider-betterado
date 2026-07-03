package workitemtracking

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*IterationDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*IterationDataSource)(nil)
)

// IterationDataSource is the terraform-plugin-framework data source for betterado_iteration.
type IterationDataSource struct {
	client *client.AggregatedClient
}

// NewIterationDataSource returns a new framework data source for betterado_iteration.
func NewIterationDataSource() datasource.DataSource {
	return &IterationDataSource{}
}

// iterationDataModel is the tfsdk model for the betterado_iteration data source.
type iterationDataModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Path          types.String `tfsdk:"path"`
	FetchChildren types.Bool   `tfsdk:"fetch_children"`
	Name          types.String `tfsdk:"name"`
	HasChildren   types.Bool   `tfsdk:"has_children"`
	Children      types.List   `tfsdk:"children"`
}

func (d *IterationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iteration"
}

func (d *IterationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an iteration (classification node) from an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier of the classification node.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the iteration belongs to.",
				Validators: []validator.String{
					isUUIDValidator{},
				},
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The path of the iteration node.",
				Validators: []validator.String{
					notWhitespaceValidator{},
				},
			},
			"fetch_children": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to fetch child nodes. Defaults to true.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the iteration node.",
			},
			"has_children": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the iteration node has children.",
			},
			"children": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The child classification nodes.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The identifier of the child classification node.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the child classification node.",
						},
						"project_id": schema.StringAttribute{
							Computed:    true,
							Description: "The project ID of the child classification node.",
						},
						"path": schema.StringAttribute{
							Computed:    true,
							Description: "The path of the child classification node.",
						},
						"has_children": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the child classification node has children.",
						},
					},
				},
			},
		},
	}
}

func (d *IterationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IterationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_iteration data source Read: provider client not configured")
		return
	}

	var model iterationDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default fetch_children to true when not explicitly set.
	fetchChildren := true
	if !model.FetchChildren.IsNull() && !model.FetchChildren.IsUnknown() {
		fetchChildren = model.FetchChildren.ValueBool()
	}

	projectID := model.ProjectID.ValueString()
	depth := 0
	if fetchChildren {
		depth = 1
	}

	structureType := workitemtracking.TreeStructureGroupValues.Iterations
	params := workitemtracking.GetClassificationNodeArgs{
		Project:        &projectID,
		StructureGroup: &structureType,
		Depth:          converter.Int(depth),
	}

	if !model.Path.IsNull() && !model.Path.IsUnknown() && model.Path.ValueString() != "" {
		trimmed := strings.TrimSpace(model.Path.ValueString())
		params.Path = &trimmed
	}

	node, err := d.client.WorkItemTrackingClient.GetClassificationNode(d.client.Ctx, params)
	if err != nil {
		js, parseErr := json.Marshal(params)
		if parseErr != nil {
			resp.Diagnostics.AddError("Read error", fmt.Sprintf("getting iteration ClassificationNode failed. Error: %+v", err))
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("getting iteration ClassificationNode failed. %s. Error: %+v", js, err))
		return
	}

	if node == nil || node.Identifier == nil {
		resp.Diagnostics.AddError("Read error", "iteration classification node returned nil or has no identifier")
		return
	}

	model.ID = types.StringValue(node.Identifier.String())
	if node.Name != nil {
		model.Name = types.StringValue(*node.Name)
	}
	if node.Path != nil {
		model.Path = types.StringValue(classificationConvertNodePath(node.Path))
	}
	if node.HasChildren != nil {
		model.HasChildren = types.BoolValue(converter.ToBool(node.HasChildren, false))
	} else {
		model.HasChildren = types.BoolValue(false)
	}
	model.FetchChildren = types.BoolValue(fetchChildren)

	children, errs := classificationFlattenChildrenFramework(projectID, node.Children)
	for _, e := range errs {
		resp.Diagnostics.AddError("Read error", e.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}
	model.Children = children

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
