package workitemtracking

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ datasource.DataSource              = (*AreaDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*AreaDataSource)(nil)
)

// AreaDataSource is the terraform-plugin-framework data source for betterado_area.
type AreaDataSource struct {
	client *client.AggregatedClient
}

// NewAreaDataSource returns a new framework data source for betterado_area.
func NewAreaDataSource() datasource.DataSource {
	return &AreaDataSource{}
}

// classificationChildAttrTypes is the attribute type map for nested child nodes.
var classificationChildAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"project_id":   types.StringType,
	"path":         types.StringType,
	"has_children": types.BoolType,
}

// areaDataModel is the tfsdk model for the betterado_area data source.
type areaDataModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Path          types.String `tfsdk:"path"`
	FetchChildren types.Bool   `tfsdk:"fetch_children"`
	Name          types.String `tfsdk:"name"`
	HasChildren   types.Bool   `tfsdk:"has_children"`
	Children      types.List   `tfsdk:"children"`
}

func (d *AreaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_area"
}

func (d *AreaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an area (classification node) from an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier of the classification node.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the area belongs to.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegexp, "value must be a valid UUID"),
				},
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The path of the area node.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(nonWhitespaceRegexp, "value must not be empty or whitespace"),
				},
			},
			"fetch_children": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to fetch child nodes. Defaults to true.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the area node.",
			},
			"has_children": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the area node has children.",
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

func (d *AreaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AreaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_area data source Read: provider client not configured")
		return
	}

	var model areaDataModel
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

	structureType := workitemtracking.TreeStructureGroupValues.Areas
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
			resp.Diagnostics.AddError("Read error", fmt.Sprintf("getting area ClassificationNode failed. Error: %+v", err))
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("getting area ClassificationNode failed. %s. Error: %+v", js, err))
		return
	}

	if node == nil || node.Identifier == nil {
		resp.Diagnostics.AddError("Read error", "area classification node returned nil or has no identifier")
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

// classificationConvertNodePath converts an ADO path (backslash-separated with
// leading project/structure segments) to a provider path (forward-slash, trimmed
// to just the area/iteration part). Exported for use by data_iteration_framework.go.
func classificationConvertNodePath(path *string) string {
	itemPath := ""
	if path != nil {
		itemPathList := strings.Split(strings.ReplaceAll(*path, "\\", "/"), "/")
		if len(itemPathList) >= 3 {
			itemPath = strings.Join(itemPathList[3:], "/")
		}
	}
	return "/" + itemPath
}

// classificationFlattenChildrenFramework converts the ADO child node slice into
// a types.List suitable for the framework model.
func classificationFlattenChildrenFramework(projectID string, nodes *[]workitemtracking.WorkItemClassificationNode) (types.List, []error) {
	childObjType := types.ObjectType{AttrTypes: classificationChildAttrTypes}
	emptyList := types.ListValueMust(childObjType, []attr.Value{})

	if nodes == nil || len(*nodes) == 0 {
		return emptyList, nil
	}

	items := make([]attr.Value, 0, len(*nodes))
	for _, n := range *nodes {
		idStr := ""
		if n.Identifier != nil {
			idStr = n.Identifier.String()
		}
		nameStr := ""
		if n.Name != nil {
			nameStr = *n.Name
		}
		pathStr := classificationConvertNodePath(n.Path)
		hasChildren := converter.ToBool(n.HasChildren, false)

		obj, diag := types.ObjectValue(classificationChildAttrTypes, map[string]attr.Value{
			"id":           types.StringValue(idStr),
			"name":         types.StringValue(nameStr),
			"project_id":   types.StringValue(projectID),
			"path":         types.StringValue(pathStr),
			"has_children": types.BoolValue(hasChildren),
		})
		if diag.HasError() {
			var errs []error
			for _, de := range diag {
				errs = append(errs, fmt.Errorf("%s: %s", de.Summary(), de.Detail()))
			}
			return emptyList, errs
		}
		items = append(items, obj)
	}

	list, diag := types.ListValue(childObjType, items)
	if diag.HasError() {
		var errs []error
		for _, de := range diag {
			errs = append(errs, fmt.Errorf("%s: %s", de.Summary(), de.Detail()))
		}
		return emptyList, errs
	}
	return list, nil
}
