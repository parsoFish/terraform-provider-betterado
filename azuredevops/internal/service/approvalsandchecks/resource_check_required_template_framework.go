package approvalsandchecks

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/pipelineschecksextras"
)

var (
	_ resource.Resource                = (*RequiredTemplateResource)(nil)
	_ resource.ResourceWithConfigure   = (*RequiredTemplateResource)(nil)
	_ resource.ResourceWithImportState = (*RequiredTemplateResource)(nil)
)

type RequiredTemplateResource struct {
	client *client.AggregatedClient
}

func NewRequiredTemplateResource() resource.Resource {
	return &RequiredTemplateResource{}
}

func (r *RequiredTemplateResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_check_required_template"
}

func (r *RequiredTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData))
		return
	}
	r.client = agg
}

func requiredTemplateAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"repository_type": types.StringType,
		"repository_name": types.StringType,
		"repository_ref":  types.StringType,
		"template_path":   types.StringType,
	}
}

func (r *RequiredTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{checkUseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{checkRequiresReplace()},
			},
			"target_resource_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{checkRequiresReplace()},
			},
			"target_resource_type": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{checkRequiresReplace()},
			},
			"version": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{checkUseStateForUnknownInt64Val()},
			},
		},
		Blocks: map[string]schema.Block{
			"required_template": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"repository_type": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  staticCheckString("azuregit"),
						},
						"repository_name": schema.StringAttribute{Required: true},
						"repository_ref":  schema.StringAttribute{Required: true},
						"template_path":   schema.StringAttribute{Required: true},
					},
				},
			},
		},
	}
}

type requiredTemplateModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	TargetResourceID   types.String `tfsdk:"target_resource_id"`
	TargetResourceType types.String `tfsdk:"target_resource_type"`
	Version            types.Int64  `tfsdk:"version"`
	RequiredTemplate   types.List   `tfsdk:"required_template"`
}

type requiredTemplateItemModel struct {
	RepositoryType types.String `tfsdk:"repository_type"`
	RepositoryName types.String `tfsdk:"repository_name"`
	RepositoryRef  types.String `tfsdk:"repository_ref"`
	TemplatePath   types.String `tfsdk:"template_path"`
}

func (r *RequiredTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model requiredTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	check, diags := expandRequiredTemplateFW(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.PipelinesChecksClientExtras.AddCheckConfiguration(r.client.Ctx, pipelineschecksextras.AddCheckConfigurationArgs{
		Project: converter.String(model.ProjectID.ValueString()), Configuration: check,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating required template check", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *RequiredTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model requiredTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *RequiredTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan requiredTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state requiredTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	plan.Version = state.Version
	check, diags := expandRequiredTemplateFW(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.PipelinesChecksClientExtras.UpdateCheckConfiguration(r.client.Ctx, pipelineschecksextras.UpdateCheckConfigurationArgs{
		Project: converter.String(plan.ProjectID.ValueString()), Configuration: check, Id: check.Id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating required template check", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RequiredTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model requiredTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid check ID", err.Error())
		return
	}
	projectID := model.ProjectID.ValueString()
	if err := r.client.PipelinesChecksClientExtras.DeleteCheckConfiguration(r.client.Ctx, pipelineschecksextras.DeleteCheckConfigurationArgs{
		Project: &projectID, Id: &id,
	}); err != nil {
		resp.Diagnostics.AddError("Error deleting required template check", err.Error())
	}
}

func (r *RequiredTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importCheckState(ctx, req, resp)
}

func (r *RequiredTemplateResource) readIntoModel(ctx context.Context, model *requiredTemplateModel) diag.Diagnostics {
	var diags diag.Diagnostics
	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid check ID", err.Error())
		return diags
	}
	check, notFound, err := readCheckFromAPI(r.client, model.ProjectID.ValueString(), id)
	if err != nil {
		diags.AddError("Error reading required template check", err.Error())
		return diags
	}
	if notFound {
		model.ID = types.StringNull()
		return diags
	}
	diags.Append(flattenRequiredTemplateFW(ctx, model, check)...)
	return diags
}

func expandRequiredTemplateFW(ctx context.Context, model *requiredTemplateModel) (*pipelineschecksextras.CheckConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	checkID := 0
	if !model.ID.IsNull() && model.ID.ValueString() != "" {
		checkID, _ = strconv.Atoi(model.ID.ValueString())
	}
	version := int(model.Version.ValueInt64())
	var items []requiredTemplateItemModel
	diags.Append(model.RequiredTemplate.ElementsAs(ctx, &items, false)...)
	if diags.HasError() {
		return nil, diags
	}
	extendsChecks := make([]interface{}, 0, len(items))
	for _, item := range items {
		repoType := item.RepositoryType.ValueString()
		if repoType == "azuregit" {
			repoType = "git"
		}
		extendsChecks = append(extendsChecks, map[string]interface{}{
			"repositoryType": repoType,
			"repositoryName": item.RepositoryName.ValueString(),
			"repositoryRef":  item.RepositoryRef.ValueString(),
			"templatePath":   item.TemplatePath.ValueString(),
		})
	}
	settings := map[string]interface{}{"extendsChecks": extendsChecks}
	check := &pipelineschecksextras.CheckConfiguration{
		Type:     approvalAndCheckType.ExtendsCheck,
		Settings: settings,
		Resource: &pipelineschecksextras.Resource{
			Id: converter.String(model.TargetResourceID.ValueString()), Type: converter.String(model.TargetResourceType.ValueString()),
		},
		Version: &version,
	}
	if checkID != 0 {
		check.Id = &checkID
	}
	return check, diags
}

func flattenRequiredTemplateFW(ctx context.Context, model *requiredTemplateModel, check *pipelineschecksextras.CheckConfiguration) diag.Diagnostics {
	var diags diag.Diagnostics
	model.ID = types.StringValue(strconv.Itoa(*check.Id))
	if check.Resource != nil {
		if check.Resource.Id != nil {
			model.TargetResourceID = types.StringValue(*check.Resource.Id)
		}
		if check.Resource.Type != nil {
			model.TargetResourceType = types.StringValue(*check.Resource.Type)
		}
	}
	if check.Version != nil {
		model.Version = types.Int64Value(int64(*check.Version))
	}
	settingsMap, err := settingsAsMap(check.Settings)
	if err != nil {
		diags.AddError("Error parsing required template check settings", err.Error())
		return diags
	}
	items := make([]requiredTemplateItemModel, 0)
	if extendsChecksRaw, ok := settingsMap["extendsChecks"]; ok {
		if extendsChecks, ok := extendsChecksRaw.([]interface{}); ok {
			for _, ecRaw := range extendsChecks {
				ec, ok := ecRaw.(map[string]interface{})
				if !ok {
					continue
				}
				repoType := ""
				if v, ok := ec["repositoryType"].(string); ok {
					repoType = v
					if repoType == "git" {
						repoType = "azuregit"
					}
				}
				items = append(items, requiredTemplateItemModel{
					RepositoryType: types.StringValue(repoType),
					RepositoryName: types.StringValue(strVal(ec["repositoryName"])),
					RepositoryRef:  types.StringValue(strVal(ec["repositoryRef"])),
					TemplatePath:   types.StringValue(strVal(ec["templatePath"])),
				})
			}
		}
	}
	listVal, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: requiredTemplateAttrTypes()}, items)
	diags.Append(d...)
	model.RequiredTemplate = listVal
	return diags
}

func strVal(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
