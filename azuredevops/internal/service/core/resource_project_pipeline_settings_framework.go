package core

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*ProjectPipelineSettingsResource)(nil)
	_ resource.ResourceWithConfigure   = (*ProjectPipelineSettingsResource)(nil)
	_ resource.ResourceWithImportState = (*ProjectPipelineSettingsResource)(nil)
)

// ProjectPipelineSettingsResource is the terraform-plugin-framework implementation
// of betterado_project_pipeline_settings.
type ProjectPipelineSettingsResource struct {
	client *client.AggregatedClient
}

// NewProjectPipelineSettingsResource returns a new resource.Resource.
func NewProjectPipelineSettingsResource() resource.Resource {
	return &ProjectPipelineSettingsResource{}
}

// projectPipelineSettingsModel is the Terraform state model for
// betterado_project_pipeline_settings.
type projectPipelineSettingsModel struct {
	ID                               types.String `tfsdk:"id"`
	ProjectID                        types.String `tfsdk:"project_id"`
	EnforceJobScope                  types.Bool   `tfsdk:"enforce_job_scope"`
	EnforceReferencedRepoScopedToken types.Bool   `tfsdk:"enforce_referenced_repo_scoped_token"`
	EnforceSettableVar               types.Bool   `tfsdk:"enforce_settable_var"`
	PublishPipelineMetadata          types.Bool   `tfsdk:"publish_pipeline_metadata"`
	StatusBadgesArePrivate           types.Bool   `tfsdk:"status_badges_are_private"`
	EnforceJobScopeForRelease        types.Bool   `tfsdk:"enforce_job_scope_for_release"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ProjectPipelineSettingsResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_project_pipeline_settings"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ProjectPipelineSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages pipeline general settings for an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the project.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid UUID",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enforce_job_scope": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Limit job authorization scope to current project for non-release pipelines.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enforce_referenced_repo_scoped_token": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Protect access to repositories in YAML pipelines.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enforce_settable_var": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Limit variables that can be set at queue time.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"publish_pipeline_metadata": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Publish metadata from pipelines.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"status_badges_are_private": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Disable anonymous access to badges.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enforce_job_scope_for_release": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Limit job authorization scope to current project for release pipelines.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ProjectPipelineSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = agg
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *ProjectPipelineSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model projectPipelineSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	if err := r.applySettings(ctx, projectID, &model); err != nil {
		resp.Diagnostics.AddError("creating project pipeline settings", err.Error())
		return
	}

	model.ID = types.StringValue(projectID)
	if err := r.readSettings(ctx, &model); err != nil {
		resp.Diagnostics.AddError("reading project pipeline settings after create", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectPipelineSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model projectPipelineSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readSettings(ctx, &model); err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("reading project pipeline settings", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectPipelineSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan projectPipelineSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	projectID := state.ID.ValueString()

	if err := r.applySettings(ctx, projectID, &plan); err != nil {
		resp.Diagnostics.AddError("updating project pipeline settings", err.Error())
		return
	}

	if err := r.readSettings(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("reading project pipeline settings after update", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op: original settings are unknown so we cannot restore them.
func (r *ProjectPipelineSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *ProjectPipelineSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	model := projectPipelineSettingsModel{
		ID:        types.StringValue(req.ID),
		ProjectID: types.StringValue(req.ID),
	}
	if err := r.readSettings(ctx, &model); err != nil {
		resp.Diagnostics.AddError("importing project pipeline settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (r *ProjectPipelineSettingsResource) applySettings(ctx context.Context, projectID string, model *projectPipelineSettingsModel) error {
	newSettings := &build.PipelineGeneralSettings{}

	if !model.EnforceJobScope.IsNull() && !model.EnforceJobScope.IsUnknown() {
		v := model.EnforceJobScope.ValueBool()
		newSettings.EnforceJobAuthScope = &v
	}
	if !model.EnforceReferencedRepoScopedToken.IsNull() && !model.EnforceReferencedRepoScopedToken.IsUnknown() {
		v := model.EnforceReferencedRepoScopedToken.ValueBool()
		newSettings.EnforceReferencedRepoScopedToken = &v
	}
	if !model.EnforceSettableVar.IsNull() && !model.EnforceSettableVar.IsUnknown() {
		v := model.EnforceSettableVar.ValueBool()
		newSettings.EnforceSettableVar = &v
	}
	if !model.PublishPipelineMetadata.IsNull() && !model.PublishPipelineMetadata.IsUnknown() {
		v := model.PublishPipelineMetadata.ValueBool()
		newSettings.PublishPipelineMetadata = &v
	}
	if !model.StatusBadgesArePrivate.IsNull() && !model.StatusBadgesArePrivate.IsUnknown() {
		v := model.StatusBadgesArePrivate.ValueBool()
		newSettings.StatusBadgesArePrivate = &v
	}
	if !model.EnforceJobScopeForRelease.IsNull() && !model.EnforceJobScopeForRelease.IsUnknown() {
		v := model.EnforceJobScopeForRelease.ValueBool()
		newSettings.EnforceJobAuthScopeForReleases = &v
	}

	_, err := r.client.BuildClient.UpdateBuildGeneralSettings(r.client.Ctx, build.UpdateBuildGeneralSettingsArgs{
		Project:     converter.String(projectID),
		NewSettings: newSettings,
	})
	return err
}

func (r *ProjectPipelineSettingsResource) readSettings(_ context.Context, model *projectPipelineSettingsModel) error {
	projectID := model.ID.ValueString()
	if projectID == "" {
		projectID = model.ProjectID.ValueString()
	}

	s, err := r.client.BuildClient.GetBuildGeneralSettings(r.client.Ctx, build.GetBuildGeneralSettingsArgs{
		Project: converter.String(projectID),
	})
	if err != nil {
		return err
	}

	model.ProjectID = types.StringValue(projectID)

	if s.EnforceJobAuthScope != nil {
		model.EnforceJobScope = types.BoolValue(*s.EnforceJobAuthScope)
	}
	if s.EnforceReferencedRepoScopedToken != nil {
		model.EnforceReferencedRepoScopedToken = types.BoolValue(*s.EnforceReferencedRepoScopedToken)
	}
	if s.EnforceSettableVar != nil {
		model.EnforceSettableVar = types.BoolValue(*s.EnforceSettableVar)
	}
	if s.PublishPipelineMetadata != nil {
		model.PublishPipelineMetadata = types.BoolValue(*s.PublishPipelineMetadata)
	}
	if s.StatusBadgesArePrivate != nil {
		model.StatusBadgesArePrivate = types.BoolValue(*s.StatusBadgesArePrivate)
	}
	if s.EnforceJobAuthScopeForReleases != nil {
		model.EnforceJobScopeForRelease = types.BoolValue(*s.EnforceJobAuthScopeForReleases)
	}
	return nil
}
