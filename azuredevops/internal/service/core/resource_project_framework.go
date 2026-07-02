package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/operations"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ── Plan-modifier helpers ─────────────────────────────────────────────────────

// projectUseStateForUnknown copies prior state when plan is unknown.
type projectUseStateForUnknown struct{}

func projectStateForUnknown() planmodifier.String { return projectUseStateForUnknown{} }
func (m projectUseStateForUnknown) Description(_ context.Context) string {
	return "use prior state value when plan is unknown"
}

func (m projectUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "use prior state value when plan is unknown"
}

func (m projectUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}

// projectRequiresReplace triggers replacement on change.
type projectRequiresReplace struct{}

func projectForceReplace() planmodifier.String { return projectRequiresReplace{} }
func (m projectRequiresReplace) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m projectRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m projectRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── Resource struct ───────────────────────────────────────────────────────────

// Compile-time interface checks.
var (
	_ resource.Resource                = (*ProjectResource)(nil)
	_ resource.ResourceWithConfigure   = (*ProjectResource)(nil)
	_ resource.ResourceWithImportState = (*ProjectResource)(nil)
)

// ProjectResource is the terraform-plugin-framework implementation of betterado_project.
type ProjectResource struct {
	client *client.AggregatedClient
}

// NewProjectResource returns a new resource.Resource for betterado_project.
func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

// projectModel is the Terraform state model for betterado_project.
type projectModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Visibility        types.String `tfsdk:"visibility"`
	VersionControl    types.String `tfsdk:"version_control"`
	WorkItemTemplate  types.String `tfsdk:"work_item_template"`
	ProcessTemplateID types.String `tfsdk:"process_template_id"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ProjectResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_project"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					projectStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					projectStateForUnknown(),
				},
			},
			"visibility": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					projectStateForUnknown(),
				},
			},
			"version_control": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					projectForceReplace(),
					projectStateForUnknown(),
				},
			},
			"work_item_template": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					projectForceReplace(),
					projectStateForUnknown(),
				},
			},
			"process_template_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					projectStateForUnknown(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model projectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := model.Name.ValueString()
	description := model.Description.ValueString()
	visibility := core.ProjectVisibility(model.Visibility.ValueString())
	if string(visibility) == "" {
		visibility = core.ProjectVisibilityValues.Private
	}
	versionControl := model.VersionControl.ValueString()
	if versionControl == "" {
		versionControl = string(core.SourceControlTypesValues.Git)
	}
	workItemTemplate := model.WorkItemTemplate.ValueString()
	if workItemTemplate == "" {
		workItemTemplate = "Agile"
	}

	processTemplateID, err := lookupProcessTemplateID(r.client, workItemTemplate)
	if err != nil {
		resp.Diagnostics.AddError("lookup process template", err.Error())
		return
	}

	capabilities := map[string]map[string]string{
		"versioncontrol": {
			"sourceControlType": versionControl,
		},
		"processTemplate": {
			"templateTypeId": processTemplateID,
		},
	}

	project := &core.TeamProject{
		Name:         &name,
		Description:  &description,
		Visibility:   &visibility,
		Capabilities: &capabilities,
	}

	operationRef, err := r.client.CoreClient.QueueCreateProject(r.client.Ctx, core.QueueCreateProjectArgs{
		ProjectToCreate: project,
	})
	if err != nil {
		resp.Diagnostics.AddError("creating project", err.Error())
		return
	}

	diags := r.waitForOperation(operationRef)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := getProject(r.client, "", name, defaultTimeout)
	if err != nil {
		resp.Diagnostics.AddError("reading project after create", err.Error())
		return
	}

	r.flattenIntoModel(ctx, created, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model projectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := model.ID.ValueString()

	project, err := r.client.CoreClient.GetProject(r.client.Ctx, core.GetProjectArgs{
		ProjectId:           &identifier,
		IncludeCapabilities: converter.Bool(true),
		IncludeHistory:      converter.Bool(false),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"reading project",
			fmt.Sprintf("looking up project ID %s: %v", identifier, err),
		)
		return
	}

	// Normalize ID to UUID (handles import by name)
	model.ID = types.StringValue(project.Id.String())

	r.flattenIntoModel(ctx, project, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan projectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("parsing project ID", err.Error())
		return
	}

	update := &core.TeamProject{Id: &projectID}
	needsUpdate := false

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		update.Name = &name
		needsUpdate = true
	}
	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		update.Description = &desc
		needsUpdate = true
	}
	if !plan.Visibility.Equal(state.Visibility) {
		vis := core.ProjectVisibility(plan.Visibility.ValueString())
		update.Visibility = &vis
		needsUpdate = true
	}

	if needsUpdate {
		if err := updateProject(r.client, update, defaultTimeout); err != nil {
			resp.Diagnostics.AddError("updating project", err.Error())
			return
		}
	}

	// Re-read
	idStr := state.ID.ValueString()
	project, err := r.client.CoreClient.GetProject(r.client.Ctx, core.GetProjectArgs{
		ProjectId:           &idStr,
		IncludeCapabilities: converter.Bool(true),
		IncludeHistory:      converter.Bool(false),
	})
	if err != nil {
		resp.Diagnostics.AddError("reading project after update", err.Error())
		return
	}

	plan.ID = state.ID
	r.flattenIntoModel(ctx, project, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model projectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := model.ID.ValueString()
	if err := deleteProject(r.client, id, defaultTimeout); err != nil {
		resp.Diagnostics.AddError("deleting project", err.Error())
		return
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

// ImportState supports import by UUID or by project name.
func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	identifier := req.ID

	project, err := r.client.CoreClient.GetProject(r.client.Ctx, core.GetProjectArgs{
		ProjectId:           &identifier,
		IncludeCapabilities: converter.Bool(true),
		IncludeHistory:      converter.Bool(false),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError("project not found", fmt.Sprintf("No project with ID or name %q", identifier))
			return
		}
		resp.Diagnostics.AddError("importing project", err.Error())
		return
	}

	model := projectModel{}
	model.ID = types.StringValue(project.Id.String())

	r.flattenIntoModel(ctx, project, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// defaultTimeout is used for synchronous project create/delete waits.
var defaultTimeout = 10 * projectBusyTimeoutDuration

// flattenIntoModel populates a projectModel from an ADO TeamProject.
func (r *ProjectResource) flattenIntoModel(_ context.Context, project *core.TeamProject, model *projectModel, diags *diag.Diagnostics) {
	if project.Name != nil {
		model.Name = types.StringValue(*project.Name)
	}
	if project.Description != nil {
		model.Description = types.StringValue(*project.Description)
	}
	if project.Visibility != nil {
		model.Visibility = types.StringValue(strings.ToLower(string(*project.Visibility)))
	}

	if project.Capabilities != nil {
		caps := *project.Capabilities
		if vc, ok := caps["versioncontrol"]; ok {
			model.VersionControl = types.StringValue(vc["sourceControlType"])
		}
		if pt, ok := caps["processTemplate"]; ok {
			ptID := pt["templateTypeId"]
			model.ProcessTemplateID = types.StringValue(ptID)
			// resolve process template name
			if ptID != "" {
				name, err := lookupProcessTemplateName(r.client, ptID)
				if err != nil {
					diags.AddWarning("resolving work_item_template", err.Error())
				} else {
					model.WorkItemTemplate = types.StringValue(name)
				}
			}
		}
	}
}

// waitForOperation blocks until an ADO long-running operation finishes.
func (r *ProjectResource) waitForOperation(operationRef *operations.OperationReference) diag.Diagnostics {
	var diags diag.Diagnostics
	for {
		op, err := r.client.OperationsClient.GetOperation(r.client.Ctx, operations.GetOperationArgs{
			OperationId: operationRef.Id,
			PluginId:    operationRef.PluginId,
		})
		if err != nil {
			diags.AddError("polling operation", err.Error())
			return diags
		}
		switch *op.Status {
		case operations.OperationStatusValues.Succeeded:
			return diags
		case operations.OperationStatusValues.Failed, operations.OperationStatusValues.Cancelled:
			diags.AddError("operation failed", fmt.Sprintf("Operation %s: %s", *op.Status, safeStr(op.DetailedMessage)))
			return diags
		}
	}
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
