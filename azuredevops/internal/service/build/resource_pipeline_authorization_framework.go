package build

// resource_pipeline_authorization_framework.go — terraform-plugin-framework implementation of
// betterado_pipeline_authorization.
//
// The SDKv2 ResourcePipelineAuthorization() in resource_pipeline_authorization.go remains
// in place for state-file compatibility. The mux serves new apply/plan calls through this
// framework resource once it is registered in framework_provider.go Resources().
// The SDKv2 "betterado_pipeline_authorization" entry in provider.go ResourcesMap is removed
// (commented with a migration note) to avoid "Invalid Provider Server Combination".
//
// No build tag on this production file — only _test.go files carry //go:build.
//
// Sub-packages like int64planmodifier / stringplanmodifier are not vendored, so we
// implement the minimal plan-modifier interfaces inline (same pattern as
// resource_build_folder_framework.go).

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelinepermissions"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &PipelineAuthorizationResource{}
	_ resource.ResourceWithImportState = &PipelineAuthorizationResource{}
	_ resource.ResourceWithConfigure   = &PipelineAuthorizationResource{}
)

// PipelineAuthorizationResource is the framework resource for betterado_pipeline_authorization.
type PipelineAuthorizationResource struct {
	client *client.AggregatedClient
}

// NewPipelineAuthorizationResource returns a new framework resource.Resource.
func NewPipelineAuthorizationResource() resource.Resource {
	return &PipelineAuthorizationResource{}
}

// ── Model ────────────────────────────────────────────────────────────────────

type pipelineAuthorizationFrameworkModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.String `tfsdk:"project_id"`
	PipelineProjectID types.String `tfsdk:"pipeline_project_id"`
	ResourceID        types.String `tfsdk:"resource_id"`
	Type              types.String `tfsdk:"type"`
	PipelineID        types.Int64  `tfsdk:"pipeline_id"`
}

// ── Metadata / Schema ────────────────────────────────────────────────────────

func (r *PipelineAuthorizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_authorization"
}

func (r *PipelineAuthorizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages pipeline authorization for a resource in an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the pipeline authorization resource.",
				PlanModifiers: []planmodifier.String{
					paUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					paRequiresReplaceString(),
				},
			},
			"pipeline_project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the project where the pipeline resides (if different from project_id).",
				PlanModifiers: []planmodifier.String{
					paRequiresReplaceString(),
				},
			},
			"resource_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the resource to authorize.",
				Validators: []validator.String{
					paNotWhitespaceValidator{},
				},
				PlanModifiers: []planmodifier.String{
					paRequiresReplaceString(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the resource. Valid values: endpoint, queue, variablegroup, environment, repository.",
				Validators: []validator.String{
					paTypeValidator{},
				},
				PlanModifiers: []planmodifier.String{
					paRequiresReplaceString(),
				},
			},
			"pipeline_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The ID of the pipeline to authorize. If omitted, all pipelines are authorized.",
				Validators: []validator.Int64{
					paInt64AtLeastValidator{min: 1},
				},
				PlanModifiers: []planmodifier.Int64{
					paRequiresReplaceInt64(),
				},
			},
		},
	}
}

// ── Validators ───────────────────────────────────────────────────────────────

type paNotWhitespaceValidator struct{}

func (v paNotWhitespaceValidator) Description(_ context.Context) string {
	return "value must not be empty or whitespace"
}

func (v paNotWhitespaceValidator) MarkdownDescription(_ context.Context) string {
	return v.Description(context.Background())
}

func (v paNotWhitespaceValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.TrimSpace(req.ConfigValue.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Value must not be whitespace", "The value must contain at least one non-whitespace character.")
	}
}

type paTypeValidator struct{}

func (v paTypeValidator) Description(_ context.Context) string {
	return "type must be one of: endpoint, queue, variablegroup, environment, repository"
}

func (v paTypeValidator) MarkdownDescription(_ context.Context) string {
	return v.Description(context.Background())
}

func (v paTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	valid := []string{"endpoint", "queue", "variablegroup", "environment", "repository"}
	val := req.ConfigValue.ValueString()
	for _, allowed := range valid {
		if val == allowed {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Invalid type",
		fmt.Sprintf("type must be one of %v, got: %q", valid, val))
}

type paInt64AtLeastValidator struct{ min int64 }

func (v paInt64AtLeastValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be at least %d", v.min)
}

func (v paInt64AtLeastValidator) MarkdownDescription(_ context.Context) string {
	return v.Description(context.Background())
}

func (v paInt64AtLeastValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.ValueInt64() < v.min {
		resp.Diagnostics.AddAttributeError(req.Path, "Value too small",
			fmt.Sprintf("value must be at least %d, got: %d", v.min, req.ConfigValue.ValueInt64()))
	}
}

// ── Inline plan-modifiers ─────────────────────────────────────────────────────
// Mirrors the pattern in resource_build_folder_framework.go.
// Sub-packages (stringplanmodifier, int64planmodifier) are not vendored.

type paUseStateForUnknownImpl struct{}

func paUseStateForUnknown() planmodifier.String { return paUseStateForUnknownImpl{} }

func (paUseStateForUnknownImpl) Description(_ context.Context) string {
	return "use prior state value for unknown"
}

func (paUseStateForUnknownImpl) MarkdownDescription(_ context.Context) string {
	return "use prior state value for unknown"
}

func (paUseStateForUnknownImpl) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type paRequiresReplaceStringImpl struct{}

func paRequiresReplaceString() planmodifier.String { return paRequiresReplaceStringImpl{} }

func (paRequiresReplaceStringImpl) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (paRequiresReplaceStringImpl) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (paRequiresReplaceStringImpl) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type paRequiresReplaceInt64Impl struct{}

func paRequiresReplaceInt64() planmodifier.Int64 { return paRequiresReplaceInt64Impl{} }

func (paRequiresReplaceInt64Impl) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (paRequiresReplaceInt64Impl) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (paRequiresReplaceInt64Impl) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── Configure ────────────────────────────────────────────────────────────────

func (r *PipelineAuthorizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData))
		return
	}
	r.client = agg
}

// ── CRUD ─────────────────────────────────────────────────────────────────────

func (r *PipelineAuthorizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pipelineAuthorizationFrameworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectId := plan.ProjectID.ValueString()
	pipelineProjectId := projectId
	if !plan.PipelineProjectID.IsNull() && !plan.PipelineProjectID.IsUnknown() && plan.PipelineProjectID.ValueString() != "" {
		pipelineProjectId = plan.PipelineProjectID.ValueString()
	}

	resType := plan.Type.ValueString()
	resId := plan.ResourceID.ValueString()
	if strings.EqualFold(resType, "repository") {
		resId = projectId + "." + resId
	}

	params := pipelinepermissions.UpdatePipelinePermisionsForResourceArgs{
		Project:      &pipelineProjectId,
		ResourceType: &resType,
		ResourceId:   &resId,
	}

	if !plan.PipelineID.IsNull() && !plan.PipelineID.IsUnknown() {
		pipeId := int(plan.PipelineID.ValueInt64())
		params.ResourceAuthorization = &pipelinepermissions.ResourcePipelinePermissions{
			Pipelines: &[]pipelinepermissions.PipelinePermission{{
				Authorized: converter.ToPtr(true),
				Id:         converter.ToPtr(pipeId),
			}},
		}
	} else {
		params.ResourceAuthorization = &pipelinepermissions.ResourcePipelinePermissions{
			AllPipelines: &pipelinepermissions.Permission{
				Authorized: converter.ToPtr(true),
			},
		}
	}

	_, err := r.client.PipelinePermissionsClient.UpdatePipelinePermisionsForResource(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating pipeline authorization", fmt.Sprintf("could not create pipeline authorization: %+v", err))
		return
	}

	// Poll until the authorization is confirmed (ADO API is eventually-consistent).
	stateConf := &retry.StateChangeConf{
		ContinuousTargetOccurence: 1,
		Delay:                     5 * time.Second,
		MinTimeout:                10 * time.Second,
		Pending:                   []string{"waiting"},
		Target:                    []string{"succeed", "failed"},
		Refresh:                   r.checkPipelineAuthorizationFramework(ctx, plan, params),
		Timeout:                   2 * time.Minute,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		resp.Diagnostics.AddError("Error waiting for pipeline authorization", fmt.Sprintf("waiting for pipeline authorization ready: %v", err))
		return
	}

	// Build ID.
	id := pipelineAuthorizationId{
		projectId:  projectId,
		typ:        resType,
		resourceId: plan.ResourceID.ValueString(),
	}
	if !plan.PipelineProjectID.IsNull() && !plan.PipelineProjectID.IsUnknown() && plan.PipelineProjectID.ValueString() != "" {
		v := plan.PipelineProjectID.ValueString()
		id.pipelineProjectId = &v
	}
	if !plan.PipelineID.IsNull() && !plan.PipelineID.IsUnknown() {
		v := int(plan.PipelineID.ValueInt64())
		id.pipelineId = &v
	}
	plan.ID = types.StringValue(id.id())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PipelineAuthorizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pipelineAuthorizationFrameworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectID.ValueString()
	pipelineProjectId := projectId
	if !state.PipelineProjectID.IsNull() && !state.PipelineProjectID.IsUnknown() && state.PipelineProjectID.ValueString() != "" {
		pipelineProjectId = state.PipelineProjectID.ValueString()
	}

	resType := state.Type.ValueString()
	resId := state.ResourceID.ValueString()
	if strings.EqualFold(resType, "repository") {
		resId = projectId + "." + resId
	}

	result, err := r.client.PipelinePermissionsClient.GetPipelinePermissionsForResource(
		ctx,
		pipelinepermissions.GetPipelinePermissionsForResourceArgs{
			Project:      &pipelineProjectId,
			ResourceType: &resType,
			ResourceId:   &resId,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Error reading pipeline authorization", fmt.Sprintf("%+v", err))
		return
	}

	if result == nil || result.Pipelines == nil || (result.AllPipelines == nil && len(*result.Pipelines) == 0) {
		resp.State.RemoveResource(ctx)
		return
	}

	if result.Resource != nil {
		if result.Resource.Type != nil {
			state.Type = types.StringValue(*result.Resource.Type)
		}
		if result.Resource.Id != nil {
			resIdRead := *result.Resource.Id
			if strings.EqualFold(state.Type.ValueString(), "repository") {
				parts := strings.SplitN(resIdRead, ".", 2)
				if len(parts) == 2 {
					resIdRead = parts[1]
				}
			}
			state.ResourceID = types.StringValue(resIdRead)
		}
	}

	// Check if specific pipeline_id is still authorized.
	if !state.PipelineID.IsNull() && !state.PipelineID.IsUnknown() {
		pipeId := int(state.PipelineID.ValueInt64())
		found := false
		if result.Pipelines != nil {
			for _, pipe := range *result.Pipelines {
				if pipe.Id != nil && *pipe.Id == pipeId {
					found = true
					break
				}
			}
		}
		if !found {
			state.PipelineID = types.Int64Null()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PipelineAuthorizationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes use RequiresReplace; Update is unreachable.
	resp.Diagnostics.AddError("Update not supported",
		"betterado_pipeline_authorization does not support in-place updates; all attributes force replacement.")
}

func (r *PipelineAuthorizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pipelineAuthorizationFrameworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectID.ValueString()
	pipelineProjectId := projectId
	if !state.PipelineProjectID.IsNull() && !state.PipelineProjectID.IsUnknown() && state.PipelineProjectID.ValueString() != "" {
		pipelineProjectId = state.PipelineProjectID.ValueString()
	}

	resType := state.Type.ValueString()
	resId := state.ResourceID.ValueString()
	if strings.EqualFold(resType, "repository") {
		resId = projectId + "." + resId
	}

	params := pipelinepermissions.UpdatePipelinePermisionsForResourceArgs{
		Project:      &pipelineProjectId,
		ResourceType: &resType,
		ResourceId:   &resId,
	}

	if !state.PipelineID.IsNull() && !state.PipelineID.IsUnknown() {
		pipeId := int(state.PipelineID.ValueInt64())
		params.ResourceAuthorization = &pipelinepermissions.ResourcePipelinePermissions{
			Pipelines: &[]pipelinepermissions.PipelinePermission{{
				Authorized: converter.ToPtr(false),
				Id:         converter.ToPtr(pipeId),
			}},
		}
	} else {
		params.ResourceAuthorization = &pipelinepermissions.ResourcePipelinePermissions{
			AllPipelines: &pipelinepermissions.Permission{
				Authorized: converter.ToPtr(false),
			},
		}
	}

	_, err := r.client.PipelinePermissionsClient.UpdatePipelinePermisionsForResource(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting pipeline authorization", fmt.Sprintf("deleting authorized resource: %+v", err))
	}
}

// ── Import ───────────────────────────────────────────────────────────────────

func (r *PipelineAuthorizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := parsePipelineAuthorizationId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("cannot parse pipeline authorization ID %q: %v", req.ID, err))
		return
	}

	model := pipelineAuthorizationFrameworkModel{
		ID:         types.StringValue(req.ID),
		ProjectID:  types.StringValue(id.projectId),
		Type:       types.StringValue(id.typ),
		ResourceID: types.StringValue(id.resourceId),
	}
	if id.pipelineProjectId != nil {
		model.PipelineProjectID = types.StringValue(*id.pipelineProjectId)
	} else {
		model.PipelineProjectID = types.StringNull()
	}
	if id.pipelineId != nil {
		model.PipelineID = types.Int64Value(int64(*id.pipelineId))
	} else {
		model.PipelineID = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (r *PipelineAuthorizationResource) checkPipelineAuthorizationFramework(
	ctx context.Context,
	plan pipelineAuthorizationFrameworkModel,
	params pipelinepermissions.UpdatePipelinePermisionsForResourceArgs,
) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		projectId := plan.ProjectID.ValueString()
		pipelineProjectId := projectId
		if !plan.PipelineProjectID.IsNull() && !plan.PipelineProjectID.IsUnknown() && plan.PipelineProjectID.ValueString() != "" {
			pipelineProjectId = plan.PipelineProjectID.ValueString()
		}
		resType := plan.Type.ValueString()
		resId := plan.ResourceID.ValueString()
		if strings.EqualFold(resType, "repository") {
			resId = projectId + "." + resId
		}

		pollResp, err := r.client.PipelinePermissionsClient.GetPipelinePermissionsForResource(
			ctx,
			pipelinepermissions.GetPipelinePermissionsForResourceArgs{
				Project:      &pipelineProjectId,
				ResourceType: &resType,
				ResourceId:   &resId,
			},
		)
		if err != nil {
			return nil, "failed", err
		}

		if !plan.PipelineID.IsNull() && !plan.PipelineID.IsUnknown() {
			pipeId := int(plan.PipelineID.ValueInt64())
			if pollResp.Pipelines != nil && len(*pollResp.Pipelines) > 0 {
				for _, pipe := range *pollResp.Pipelines {
					if pipe.Id != nil && *pipe.Id == pipeId {
						return pollResp, "succeed", nil
					}
				}
				// Reapply for authorization.
				_, err = r.client.PipelinePermissionsClient.UpdatePipelinePermisionsForResource(ctx, params)
				return nil, "waiting", err
			}
		} else {
			if pollResp.AllPipelines != nil && pollResp.AllPipelines.Authorized != nil && *pollResp.AllPipelines.Authorized {
				return pollResp, "succeed", nil
			}
			// Reapply for authorization.
			_, err = r.client.PipelinePermissionsClient.UpdatePipelinePermisionsForResource(ctx, params)
			return nil, "waiting", err
		}

		return pollResp, "succeed", nil
	}
}
