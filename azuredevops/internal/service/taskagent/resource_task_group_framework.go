package taskagent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ── Static default helpers ────────────────────────────────────────────────────

// staticStringDefault implements defaults.String for a constant value.
type staticStringDefault struct{ value string }

func defaultString(v string) defaults.String { return staticStringDefault{v} }
func (d staticStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d staticStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%q`", d.value)
}

func (d staticStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// staticBoolDefault implements defaults.Bool for a constant value.
type staticBoolDefault struct{ value bool }

func defaultBool(v bool) defaults.Bool { return staticBoolDefault{v} }
func (d staticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d staticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", d.value)
}

func (d staticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// staticInt64Default implements defaults.Int64 for a constant value.
type staticInt64Default struct{ value int64 }

func defaultInt64(v int64) defaults.Int64 { return staticInt64Default{v} }
func (d staticInt64Default) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %d", d.value)
}

func (d staticInt64Default) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%d`", d.value)
}

func (d staticInt64Default) DefaultInt64(_ context.Context, _ defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(d.value)
}

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// requiresReplaceModifier is a plan modifier that marks the resource for
// replacement when the attribute value changes.
type requiresReplaceModifier struct{}

func requiresReplace() planmodifier.String { return requiresReplaceModifier{} }
func (m requiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m requiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m requiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// useStateForUnknownModifier copies the prior state value when the plan value
// is unknown (prevents perpetual diffs for computed-only attributes).
type useStateForUnknownModifier struct{}

func useStateForUnknown() planmodifier.String { return useStateForUnknownModifier{} }
func (m useStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m useStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m useStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── Resource struct ───────────────────────────────────────────────────────────

// Compile-time interface checks.
var (
	_ resource.Resource                 = (*TaskGroupResource)(nil)
	_ resource.ResourceWithConfigure    = (*TaskGroupResource)(nil)
	_ resource.ResourceWithUpgradeState = (*TaskGroupResource)(nil)
)

// TaskGroupResource is the terraform-plugin-framework implementation of
// betterado_task_group.
type TaskGroupResource struct {
	client *client.AggregatedClient
}

// NewTaskGroupResource returns a new resource.Resource for betterado_task_group.
func NewTaskGroupResource() resource.Resource {
	return &TaskGroupResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *TaskGroupResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_task_group"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *TaskGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			// ── Top-level scalars ──────────────────────────────────────────
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"friendly_name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  defaultString(""),
			},
			"category": schema.StringAttribute{
				Required: true,
			},
			"author": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  defaultString(""),
			},
			"icon_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  defaultString(""),
			},
			"instance_name_format": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  defaultString(""),
			},
			"runs_on": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			// ── Computed ──────────────────────────────────────────────────
			"revision": schema.Int64Attribute{
				Computed: true,
			},
			"definition_type": schema.StringAttribute{
				Computed: true,
			},
			// ── Nested: version ───────────────────────────────────────────
			"version": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"major": schema.Int64Attribute{Required: true},
						"minor": schema.Int64Attribute{Required: true},
						"patch": schema.Int64Attribute{Required: true},
						"is_test": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultBool(false),
						},
					},
				},
			},
			// ── Nested: input ─────────────────────────────────────────────
			"input": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":  schema.StringAttribute{Required: true},
						"label": schema.StringAttribute{Required: true},
						"type": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultString("string"),
						},
						"default_value": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultString(""),
						},
						"required": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultBool(false),
						},
						"help_markdown": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultString(""),
						},
						"group_name": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultString(""),
						},
						"options": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
						},
						"visible_rule": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultString(""),
						},
						"properties": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
						},
						"aliases": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			// ── Nested: task ──────────────────────────────────────────────
			"task": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"display_name": schema.StringAttribute{Required: true},
						"task_id":      schema.StringAttribute{Required: true},
						"task_version": schema.StringAttribute{Required: true},
						"task_definition_type": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultString("task"),
						},
						"enabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultBool(true),
						},
						"always_run": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultBool(false),
						},
						"continue_on_error": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultBool(false),
						},
						"condition": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  defaultString("succeeded()"),
						},
						"timeout_in_minutes": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							Default:  defaultInt64(0),
						},
						"retry_count_on_task_failure": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							Default:  defaultInt64(0),
						},
						"inputs": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
						},
						"environment": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *TaskGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── State model ───────────────────────────────────────────────────────────────

// taskGroupModel is the Terraform state model for betterado_task_group.
type taskGroupModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	Name               types.String `tfsdk:"name"`
	FriendlyName       types.String `tfsdk:"friendly_name"`
	Description        types.String `tfsdk:"description"`
	Category           types.String `tfsdk:"category"`
	Author             types.String `tfsdk:"author"`
	IconURL            types.String `tfsdk:"icon_url"`
	InstanceNameFormat types.String `tfsdk:"instance_name_format"`
	RunsOn             types.List   `tfsdk:"runs_on"`
	Revision           types.Int64  `tfsdk:"revision"`
	DefinitionType     types.String `tfsdk:"definition_type"`
	Version            types.List   `tfsdk:"version"`
	Input              types.List   `tfsdk:"input"`
	Task               types.List   `tfsdk:"task"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *TaskGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model taskGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	createParam, diags := expandTaskGroupFrameworkCreate(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.TaskAgentClient.AddTaskGroup(r.client.Ctx, taskagent.AddTaskGroupArgs{
		TaskGroup: createParam,
		Project:   &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating task group", err.Error())
		return
	}

	tgIDStr := created.Id.String()
	model.ID = types.StringValue(tgIDStr)

	readDiags := r.readIntoModel(ctx, &model, projectID, tgIDStr)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TaskGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model taskGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tgIDStr := model.ID.ValueString()
	projectID := model.ProjectID.ValueString()

	readDiags := r.readIntoModel(ctx, &model, projectID, tgIDStr)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If readIntoModel cleared ID it means resource was removed.
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TaskGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan taskGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateModel taskGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tgIDStr := stateModel.ID.ValueString()
	tgID, err := uuid.Parse(tgIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid task group ID", err.Error())
		return
	}
	projectID := plan.ProjectID.ValueString()

	updateParam, diags := expandTaskGroupFrameworkUpdate(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateParam.Id = &tgID
	if !stateModel.Revision.IsNull() && !stateModel.Revision.IsUnknown() {
		rev := int(stateModel.Revision.ValueInt64())
		updateParam.Revision = &rev
	}

	_, err = r.client.TaskAgentClient.UpdateTaskGroup(r.client.Ctx, taskagent.UpdateTaskGroupArgs{
		TaskGroup:   updateParam,
		Project:     &projectID,
		TaskGroupId: &tgID,
	})
	if err != nil {
		// On stale-revision (HTTP 400 with "revision" in error) try once more.
		if utils.ResponseWasStatusCode(err, 400) && strings.Contains(err.Error(), "revision") {
			freshGroups, rerr := r.client.TaskAgentClient.GetTaskGroups(r.client.Ctx, taskagent.GetTaskGroupsArgs{
				Project:     &projectID,
				TaskGroupId: &tgID,
			})
			if rerr == nil && freshGroups != nil && len(*freshGroups) > 0 {
				updateParam.Revision = (*freshGroups)[0].Revision
				_, err = r.client.TaskAgentClient.UpdateTaskGroup(r.client.Ctx, taskagent.UpdateTaskGroupArgs{
					TaskGroup:   updateParam,
					Project:     &projectID,
					TaskGroupId: &tgID,
				})
			}
		}
		if err != nil {
			resp.Diagnostics.AddError("Error updating task group", err.Error())
			return
		}
	}

	plan.ID = stateModel.ID
	readDiags := r.readIntoModel(ctx, &plan, projectID, tgIDStr)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TaskGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model taskGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tgID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid task group ID", err.Error())
		return
	}
	projectID := model.ProjectID.ValueString()

	err = r.client.TaskAgentClient.DeleteTaskGroup(r.client.Ctx, taskagent.DeleteTaskGroupArgs{
		Project:     &projectID,
		TaskGroupId: &tgID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting task group", err.Error())
	}
}

// ── Internal read helper ──────────────────────────────────────────────────────

// readIntoModel fetches the task group by ID and populates model. If resource
// is gone (404/empty), model.ID is set to null so callers can remove from state.
func (r *TaskGroupResource) readIntoModel(ctx context.Context, model *taskGroupModel, projectID, tgIDStr string) diag.Diagnostics {
	var diags diag.Diagnostics

	tgID, err := uuid.Parse(tgIDStr)
	if err != nil {
		diags.AddError("Invalid task group ID", err.Error())
		return diags
	}

	taskGroups, err := r.client.TaskAgentClient.GetTaskGroups(r.client.Ctx, taskagent.GetTaskGroupsArgs{
		Project:     &projectID,
		TaskGroupId: &tgID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return diags
		}
		diags.AddError("Error reading task group", err.Error())
		return diags
	}
	if taskGroups == nil || len(*taskGroups) == 0 {
		model.ID = types.StringNull()
		return diags
	}

	tg := (*taskGroups)[0]
	diags.Append(flattenTaskGroupFramework(ctx, model, &tg)...)
	return diags
}

// ── Nested state models ───────────────────────────────────────────────────────

// versionModel is the state model for the version block.
type versionModel struct {
	Major  types.Int64 `tfsdk:"major"`
	Minor  types.Int64 `tfsdk:"minor"`
	Patch  types.Int64 `tfsdk:"patch"`
	IsTest types.Bool  `tfsdk:"is_test"`
}

// inputModel is the state model for an input block.
type inputModel struct {
	Name         types.String `tfsdk:"name"`
	Label        types.String `tfsdk:"label"`
	Type         types.String `tfsdk:"type"`
	DefaultValue types.String `tfsdk:"default_value"`
	Required     types.Bool   `tfsdk:"required"`
	HelpMarkdown types.String `tfsdk:"help_markdown"`
	GroupName    types.String `tfsdk:"group_name"`
	Options      types.Map    `tfsdk:"options"`
	VisibleRule  types.String `tfsdk:"visible_rule"`
	Properties   types.Map    `tfsdk:"properties"`
	Aliases      types.List   `tfsdk:"aliases"`
}

// taskStepModel is the state model for a task step block.
type taskStepModel struct {
	DisplayName             types.String `tfsdk:"display_name"`
	TaskID                  types.String `tfsdk:"task_id"`
	TaskVersion             types.String `tfsdk:"task_version"`
	TaskDefinitionType      types.String `tfsdk:"task_definition_type"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	AlwaysRun               types.Bool   `tfsdk:"always_run"`
	ContinueOnError         types.Bool   `tfsdk:"continue_on_error"`
	Condition               types.String `tfsdk:"condition"`
	TimeoutInMinutes        types.Int64  `tfsdk:"timeout_in_minutes"`
	RetryCountOnTaskFailure types.Int64  `tfsdk:"retry_count_on_task_failure"`
	Inputs                  types.Map    `tfsdk:"inputs"`
	Environment             types.Map    `tfsdk:"environment"`
}

// ── AttrType maps ─────────────────────────────────────────────────────────────

func versionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"major":   types.Int64Type,
		"minor":   types.Int64Type,
		"patch":   types.Int64Type,
		"is_test": types.BoolType,
	}
}

func inputAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":          types.StringType,
		"label":         types.StringType,
		"type":          types.StringType,
		"default_value": types.StringType,
		"required":      types.BoolType,
		"help_markdown": types.StringType,
		"group_name":    types.StringType,
		"options":       types.MapType{ElemType: types.StringType},
		"visible_rule":  types.StringType,
		"properties":    types.MapType{ElemType: types.StringType},
		"aliases":       types.ListType{ElemType: types.StringType},
	}
}

func taskAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"display_name":                types.StringType,
		"task_id":                     types.StringType,
		"task_version":                types.StringType,
		"task_definition_type":        types.StringType,
		"enabled":                     types.BoolType,
		"always_run":                  types.BoolType,
		"continue_on_error":           types.BoolType,
		"condition":                   types.StringType,
		"timeout_in_minutes":          types.Int64Type,
		"retry_count_on_task_failure": types.Int64Type,
		"inputs":                      types.MapType{ElemType: types.StringType},
		"environment":                 types.MapType{ElemType: types.StringType},
	}
}

// ── Expand ────────────────────────────────────────────────────────────────────

func expandTaskGroupFrameworkCreate(ctx context.Context, model *taskGroupModel) (*taskagent.TaskGroupCreateParameter, diag.Diagnostics) {
	var diags diag.Diagnostics

	param := &taskagent.TaskGroupCreateParameter{
		Name:               converter.String(model.Name.ValueString()),
		FriendlyName:       converter.String(model.FriendlyName.ValueString()),
		Category:           converter.String(model.Category.ValueString()),
		Description:        converter.String(model.Description.ValueString()),
		Author:             converter.String(model.Author.ValueString()),
		IconUrl:            converter.String(model.IconURL.ValueString()),
		InstanceNameFormat: converter.String(model.InstanceNameFormat.ValueString()),
	}

	if !model.RunsOn.IsNull() && !model.RunsOn.IsUnknown() {
		var runsOnList []types.String
		diags.Append(model.RunsOn.ElementsAs(ctx, &runsOnList, false)...)
		runsOn := make([]string, len(runsOnList))
		for i, s := range runsOnList {
			runsOn[i] = s.ValueString()
		}
		param.RunsOn = &runsOn
	}

	if !model.Version.IsNull() && !model.Version.IsUnknown() {
		versionParam, d := expandVersionFramework(ctx, model.Version)
		diags.Append(d...)
		param.Version = versionParam
	}

	if !model.Input.IsNull() && !model.Input.IsUnknown() {
		inputs, d := expandInputsFramework(ctx, model.Input)
		diags.Append(d...)
		param.Inputs = &inputs
	}

	if !model.Task.IsNull() && !model.Task.IsUnknown() {
		tasks, d := expandTasksFramework(ctx, model.Task)
		diags.Append(d...)
		param.Tasks = &tasks
	}

	return param, diags
}

func expandTaskGroupFrameworkUpdate(ctx context.Context, model *taskGroupModel) (*taskagent.TaskGroupUpdateParameter, diag.Diagnostics) {
	var diags diag.Diagnostics

	param := &taskagent.TaskGroupUpdateParameter{
		Name:               converter.String(model.Name.ValueString()),
		FriendlyName:       converter.String(model.FriendlyName.ValueString()),
		Category:           converter.String(model.Category.ValueString()),
		Description:        converter.String(model.Description.ValueString()),
		Author:             converter.String(model.Author.ValueString()),
		IconUrl:            converter.String(model.IconURL.ValueString()),
		InstanceNameFormat: converter.String(model.InstanceNameFormat.ValueString()),
	}

	if !model.RunsOn.IsNull() && !model.RunsOn.IsUnknown() {
		var runsOnList []types.String
		diags.Append(model.RunsOn.ElementsAs(ctx, &runsOnList, false)...)
		runsOn := make([]string, len(runsOnList))
		for i, s := range runsOnList {
			runsOn[i] = s.ValueString()
		}
		param.RunsOn = &runsOn
	}

	if !model.Version.IsNull() && !model.Version.IsUnknown() {
		versionParam, d := expandVersionFramework(ctx, model.Version)
		diags.Append(d...)
		param.Version = versionParam
	}

	if !model.Input.IsNull() && !model.Input.IsUnknown() {
		inputs, d := expandInputsFramework(ctx, model.Input)
		diags.Append(d...)
		param.Inputs = &inputs
	}

	if !model.Task.IsNull() && !model.Task.IsUnknown() {
		tasks, d := expandTasksFramework(ctx, model.Task)
		diags.Append(d...)
		param.Tasks = &tasks
	}

	return param, diags
}

func expandVersionFramework(ctx context.Context, vList types.List) (*taskagent.TaskVersion, diag.Diagnostics) {
	var diags diag.Diagnostics
	var versions []versionModel
	diags.Append(vList.ElementsAs(ctx, &versions, false)...)
	if diags.HasError() || len(versions) == 0 {
		return nil, diags
	}
	v := versions[0]
	major := int(v.Major.ValueInt64())
	minor := int(v.Minor.ValueInt64())
	patch := int(v.Patch.ValueInt64())
	return &taskagent.TaskVersion{
		Major:  &major,
		Minor:  &minor,
		Patch:  &patch,
		IsTest: converter.Bool(v.IsTest.ValueBool()),
	}, diags
}

func expandInputsFramework(ctx context.Context, iList types.List) ([]taskagent.TaskInputDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics
	var models []inputModel
	diags.Append(iList.ElementsAs(ctx, &models, false)...)

	result := make([]taskagent.TaskInputDefinition, len(models))
	for i, m := range models {
		inp := taskagent.TaskInputDefinition{
			Name:         converter.String(m.Name.ValueString()),
			Label:        converter.String(m.Label.ValueString()),
			Type:         converter.String(m.Type.ValueString()),
			DefaultValue: converter.String(m.DefaultValue.ValueString()),
			Required:     converter.Bool(m.Required.ValueBool()),
			HelpMarkDown: converter.String(m.HelpMarkdown.ValueString()),
			GroupName:    converter.String(m.GroupName.ValueString()),
			VisibleRule:  converter.String(m.VisibleRule.ValueString()),
		}

		if !m.Options.IsNull() && !m.Options.IsUnknown() {
			var opts map[string]string
			diags.Append(m.Options.ElementsAs(ctx, &opts, false)...)
			inp.Options = &opts
		}
		if !m.Properties.IsNull() && !m.Properties.IsUnknown() {
			var props map[string]string
			diags.Append(m.Properties.ElementsAs(ctx, &props, false)...)
			inp.Properties = &props
		}
		if !m.Aliases.IsNull() && !m.Aliases.IsUnknown() {
			var aliases []types.String
			diags.Append(m.Aliases.ElementsAs(ctx, &aliases, false)...)
			al := make([]string, len(aliases))
			for j, a := range aliases {
				al[j] = a.ValueString()
			}
			inp.Aliases = &al
		}

		result[i] = inp
	}
	return result, diags
}

func expandTasksFramework(ctx context.Context, tList types.List) ([]taskagent.TaskGroupStep, diag.Diagnostics) {
	var diags diag.Diagnostics
	var models []taskStepModel
	diags.Append(tList.ElementsAs(ctx, &models, false)...)

	result := make([]taskagent.TaskGroupStep, len(models))
	for i, m := range models {
		taskID, err := uuid.Parse(m.TaskID.ValueString())
		if err != nil {
			diags.AddError("Invalid task_id", fmt.Sprintf("task[%d].task_id is not a valid UUID: %s", i, err.Error()))
			continue
		}
		timeout := int(m.TimeoutInMinutes.ValueInt64())
		retry := int(m.RetryCountOnTaskFailure.ValueInt64())

		step := taskagent.TaskGroupStep{
			DisplayName:     converter.String(m.DisplayName.ValueString()),
			Enabled:         converter.Bool(m.Enabled.ValueBool()),
			AlwaysRun:       converter.Bool(m.AlwaysRun.ValueBool()),
			ContinueOnError: converter.Bool(m.ContinueOnError.ValueBool()),
			Condition:       converter.String(m.Condition.ValueString()),
			Task: &taskagent.TaskDefinitionReference{
				Id:             &taskID,
				VersionSpec:    converter.String(m.TaskVersion.ValueString()),
				DefinitionType: converter.String(m.TaskDefinitionType.ValueString()),
			},
		}
		if timeout > 0 {
			step.TimeoutInMinutes = &timeout
		}
		if retry > 0 {
			step.RetryCountOnTaskFailure = &retry
		}

		if !m.Inputs.IsNull() && !m.Inputs.IsUnknown() {
			var inputs map[string]string
			diags.Append(m.Inputs.ElementsAs(ctx, &inputs, false)...)
			step.Inputs = &inputs
		}
		if !m.Environment.IsNull() && !m.Environment.IsUnknown() {
			var env map[string]string
			diags.Append(m.Environment.ElementsAs(ctx, &env, false)...)
			step.Environment = &env
		}

		result[i] = step
	}
	return result, diags
}

// ── Flatten ───────────────────────────────────────────────────────────────────

func flattenTaskGroupFramework(ctx context.Context, model *taskGroupModel, tg *taskagent.TaskGroup) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(tg.Id.String())
	model.Name = types.StringValue(converter.ToString(tg.Name, ""))
	model.FriendlyName = types.StringValue(converter.ToString(tg.FriendlyName, ""))
	model.Description = types.StringValue(converter.ToString(tg.Description, ""))
	model.Category = types.StringValue(converter.ToString(tg.Category, ""))
	model.Author = types.StringValue(converter.ToString(tg.Author, ""))
	model.IconURL = types.StringValue(converter.ToString(tg.IconUrl, ""))
	model.InstanceNameFormat = types.StringValue(converter.ToString(tg.InstanceNameFormat, ""))
	model.DefinitionType = types.StringValue(converter.ToString(tg.DefinitionType, ""))

	if tg.Revision != nil {
		model.Revision = types.Int64Value(int64(*tg.Revision))
	} else {
		model.Revision = types.Int64Value(0)
	}

	// runs_on
	if tg.RunsOn != nil {
		elems := make([]types.String, len(*tg.RunsOn))
		for i, s := range *tg.RunsOn {
			elems[i] = types.StringValue(s)
		}
		listVal, d := types.ListValueFrom(ctx, types.StringType, elems)
		diags.Append(d...)
		model.RunsOn = listVal
	} else {
		listVal, d := types.ListValueFrom(ctx, types.StringType, []types.String{})
		diags.Append(d...)
		model.RunsOn = listVal
	}

	// version
	if tg.Version != nil {
		major := int64(converter.ToInt(tg.Version.Major, 0))
		minor := int64(converter.ToInt(tg.Version.Minor, 0))
		patch := int64(converter.ToInt(tg.Version.Patch, 0))
		isTest := converter.ToBool(tg.Version.IsTest, false)
		vObj := versionModel{
			Major:  types.Int64Value(major),
			Minor:  types.Int64Value(minor),
			Patch:  types.Int64Value(patch),
			IsTest: types.BoolValue(isTest),
		}
		listVal, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: versionAttrTypes()}, []versionModel{vObj})
		diags.Append(d...)
		model.Version = listVal
	} else {
		listVal, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: versionAttrTypes()}, []versionModel{})
		diags.Append(d...)
		model.Version = listVal
	}

	// input
	if tg.Inputs != nil && len(*tg.Inputs) > 0 {
		inputModels := make([]inputModel, len(*tg.Inputs))
		for i, inp := range *tg.Inputs {
			m := inputModel{
				Name:         types.StringValue(converter.ToString(inp.Name, "")),
				Label:        types.StringValue(converter.ToString(inp.Label, "")),
				Type:         types.StringValue(converter.ToString(inp.Type, "string")),
				DefaultValue: types.StringValue(converter.ToString(inp.DefaultValue, "")),
				Required:     types.BoolValue(converter.ToBool(inp.Required, false)),
				HelpMarkdown: types.StringValue(converter.ToString(inp.HelpMarkDown, "")),
				GroupName:    types.StringValue(converter.ToString(inp.GroupName, "")),
				VisibleRule:  types.StringValue(converter.ToString(inp.VisibleRule, "")),
			}
			if inp.Options != nil {
				optsMap, d := types.MapValueFrom(ctx, types.StringType, *inp.Options)
				diags.Append(d...)
				m.Options = optsMap
			} else {
				optsMap, d := types.MapValueFrom(ctx, types.StringType, map[string]string{})
				diags.Append(d...)
				m.Options = optsMap
			}
			if inp.Properties != nil {
				propsMap, d := types.MapValueFrom(ctx, types.StringType, *inp.Properties)
				diags.Append(d...)
				m.Properties = propsMap
			} else {
				propsMap, d := types.MapValueFrom(ctx, types.StringType, map[string]string{})
				diags.Append(d...)
				m.Properties = propsMap
			}
			if inp.Aliases != nil {
				alElems := make([]types.String, len(*inp.Aliases))
				for j, a := range *inp.Aliases {
					alElems[j] = types.StringValue(a)
				}
				alList, d := types.ListValueFrom(ctx, types.StringType, alElems)
				diags.Append(d...)
				m.Aliases = alList
			} else {
				alList, d := types.ListValueFrom(ctx, types.StringType, []types.String{})
				diags.Append(d...)
				m.Aliases = alList
			}
			inputModels[i] = m
		}
		listVal, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: inputAttrTypes()}, inputModels)
		diags.Append(d...)
		model.Input = listVal
	} else {
		listVal, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: inputAttrTypes()}, []inputModel{})
		diags.Append(d...)
		model.Input = listVal
	}

	// task
	if tg.Tasks != nil && len(*tg.Tasks) > 0 {
		taskModels := make([]taskStepModel, len(*tg.Tasks))
		for i, step := range *tg.Tasks {
			m := taskStepModel{
				DisplayName:             types.StringValue(converter.ToString(step.DisplayName, "")),
				Enabled:                 types.BoolValue(converter.ToBool(step.Enabled, true)),
				AlwaysRun:               types.BoolValue(converter.ToBool(step.AlwaysRun, false)),
				ContinueOnError:         types.BoolValue(converter.ToBool(step.ContinueOnError, false)),
				Condition:               types.StringValue(converter.ToString(step.Condition, "succeeded()")),
				TimeoutInMinutes:        types.Int64Value(int64(converter.ToInt(step.TimeoutInMinutes, 0))),
				RetryCountOnTaskFailure: types.Int64Value(int64(converter.ToInt(step.RetryCountOnTaskFailure, 0))),
				TaskID:                  types.StringValue(""),
				TaskVersion:             types.StringValue(""),
				TaskDefinitionType:      types.StringValue("task"),
			}
			if step.Task != nil {
				if step.Task.Id != nil {
					m.TaskID = types.StringValue(step.Task.Id.String())
				}
				m.TaskVersion = types.StringValue(converter.ToString(step.Task.VersionSpec, ""))
				m.TaskDefinitionType = types.StringValue(converter.ToString(step.Task.DefinitionType, "task"))
			}
			if step.Inputs != nil {
				inputsMap, d := types.MapValueFrom(ctx, types.StringType, *step.Inputs)
				diags.Append(d...)
				m.Inputs = inputsMap
			} else {
				inputsMap, d := types.MapValueFrom(ctx, types.StringType, map[string]string{})
				diags.Append(d...)
				m.Inputs = inputsMap
			}
			if step.Environment != nil {
				envMap, d := types.MapValueFrom(ctx, types.StringType, *step.Environment)
				diags.Append(d...)
				m.Environment = envMap
			} else {
				envMap, d := types.MapValueFrom(ctx, types.StringType, map[string]string{})
				diags.Append(d...)
				m.Environment = envMap
			}
			taskModels[i] = m
		}
		listVal, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: taskAttrTypes()}, taskModels)
		diags.Append(d...)
		model.Task = listVal
	} else {
		listVal, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: taskAttrTypes()}, []taskStepModel{})
		diags.Append(d...)
		model.Task = listVal
	}

	return diags
}

// ── State upgraders ───────────────────────────────────────────────────────────

// UpgradeState implements resource.ResourceWithUpgradeState.
// Version 0 → 1: renames task.environment map to task.env (see state_upgrade_v0.go).
func (r *TaskGroupResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: taskGroupStateUpgraderV0(),
	}
}
