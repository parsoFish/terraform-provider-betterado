package workitemtrackingprocess

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &processResource{}
	_ resource.ResourceWithConfigure   = &processResource{}
	_ resource.ResourceWithImportState = &processResource{}
)

// ── Inline defaults (booldefault / stringdefault sub-packages not vendored) ──

type staticStringDefault struct{ value string }

func staticStringDef(v string) defaults.String { return staticStringDefault{value: v} }

func (s staticStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", s.value)
}

func (s staticStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%q`", s.value)
}

func (s staticStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(s.value)
}

type staticBoolDefault struct{ value bool }

func staticBoolDef(v bool) defaults.Bool { return staticBoolDefault{value: v} }

func (s staticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", s.value)
}

func (s staticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", s.value)
}

func (s staticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(s.value)
}

// processUseStateForUnknown implements UseStateForUnknown for string attributes.
// (stringplanmodifier sub-package is not vendored in this project.)
type processUseStateForUnknown struct{}

func processUseStateForUnknownMod() planmodifier.String { return processUseStateForUnknown{} }

func (processUseStateForUnknown) Description(_ context.Context) string {
	return "use prior state value for unknown"
}

func (processUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "use prior state value for unknown"
}

func (processUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// processRequiresReplace implements RequiresReplace for string attributes.
// (stringplanmodifier sub-package is not vendored in this project.)
type processRequiresReplace struct{}

func processRequiresReplaceMod() planmodifier.String { return processRequiresReplace{} }

func (processRequiresReplace) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (processRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (processRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── Inline validators ─────────────────────────────────────────────────────────

// processNonWhiteSpaceValidator rejects values that are empty or whitespace-only.
// Mirrors the SDKv2 validation.StringIsNotWhiteSpace behaviour.
type processNonWhiteSpaceValidator struct{}

func (v processNonWhiteSpaceValidator) Description(_ context.Context) string {
	return "value must not be empty or whitespace-only"
}

func (v processNonWhiteSpaceValidator) MarkdownDescription(_ context.Context) string {
	return "value must not be empty or whitespace-only"
}

func (v processNonWhiteSpaceValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.TrimSpace(req.ConfigValue.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid value", "Value must not be empty or whitespace-only.")
	}
}

// processUUIDValidator rejects values that are not valid UUIDs.
// Mirrors the SDKv2 validation.IsUUID behaviour.
type processUUIDValidator struct{}

func (v processUUIDValidator) Description(_ context.Context) string {
	return "value must be a valid UUID"
}

func (v processUUIDValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a valid UUID"
}

func (v processUUIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := uuid.Parse(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid UUID", fmt.Sprintf("Value must be a valid UUID: %s", err))
	}
}

// ── Resource struct ───────────────────────────────────────────────────────────

// processResource is the terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_process.
type processResource struct {
	client *client.AggregatedClient
}

// NewProcessResource returns a new resource.Resource.
func NewProcessResource() resource.Resource {
	return &processResource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type processResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	ParentProcessTypeID types.String `tfsdk:"parent_process_type_id"`
	ReferenceName       types.String `tfsdk:"reference_name"`
	IsDefault           types.Bool   `tfsdk:"is_default"`
	IsEnabled           types.Bool   `tfsdk:"is_enabled"`
	CustomizationType   types.String `tfsdk:"customization_type"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *processResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_process"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *processResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a work item tracking process in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					processUseStateForUnknownMod(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the process.",
				Validators:  []validator.String{processNonWhiteSpaceValidator{}},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     staticStringDef(""),
				Description: "Description of the process.",
			},
			"parent_process_type_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the parent process.",
				PlanModifiers: []planmodifier.String{
					processRequiresReplaceMod(),
				},
				Validators: []validator.String{processUUIDValidator{}},
			},
			"reference_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Reference name of process being created. If not specified, server will assign a unique reference name.",
				PlanModifiers: []planmodifier.String{
					processRequiresReplaceMod(),
					processUseStateForUnknownMod(),
				},
				Validators: []validator.String{processNonWhiteSpaceValidator{}},
			},
			"is_default": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     staticBoolDef(false),
				Description: "Is the process default?",
			},
			"is_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     staticBoolDef(true),
				Description: "Is the process enabled?",
			},
			"customization_type": schema.StringAttribute{
				Computed:    true,
				Description: "Indicates the type of customization on this process. System Process is default process. Inherited Process is modified process that was System process before.",
				PlanModifiers: []planmodifier.String{
					processUseStateForUnknownMod(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *processResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = c
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *processResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_process Create: provider client not configured")
		return
	}

	var model processResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createModel := &workitemtrackingprocess.CreateProcessModel{
		Name:                converter.String(model.Name.ValueString()),
		Description:         converter.String(model.Description.ValueString()),
		ParentProcessTypeId: converter.UUID(model.ParentProcessTypeID.ValueString()),
	}
	if !model.ReferenceName.IsNull() && !model.ReferenceName.IsUnknown() && model.ReferenceName.ValueString() != "" {
		createModel.ReferenceName = converter.String(model.ReferenceName.ValueString())
	}

	created, err := r.client.WorkItemTrackingProcessClient.CreateNewProcess(ctx, workitemtrackingprocess.CreateNewProcessArgs{
		CreateRequest: createModel,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating process: %s", err))
		return
	}

	// The ADO API always creates processes with is_enabled=true (and is_default=false).
	// If the plan differs, we must update immediately.
	needsUpdate := false
	if created.IsDefault == nil || *created.IsDefault != model.IsDefault.ValueBool() {
		needsUpdate = true
	}
	if created.IsEnabled == nil || *created.IsEnabled != model.IsEnabled.ValueBool() {
		needsUpdate = true
	}

	if needsUpdate {
		processTypeID := created.TypeId
		_, err2 := r.client.WorkItemTrackingProcessClient.EditProcess(ctx, workitemtrackingprocess.EditProcessArgs{
			ProcessTypeId: processTypeID,
			UpdateRequest: &workitemtrackingprocess.UpdateProcessModel{
				Name:        converter.String(model.Name.ValueString()),
				Description: converter.String(model.Description.ValueString()),
				IsDefault:   converter.Bool(model.IsDefault.ValueBool()),
				IsEnabled:   converter.Bool(model.IsEnabled.ValueBool()),
			},
		})
		if err2 != nil {
			resp.Diagnostics.AddError("Post-create update error", fmt.Sprintf("updating process after create: %s", err2))
			// Save the partial state anyway.
			r.flattenProcess(&model, created)
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
		// Poll until GetProcessByItsId reflects the intended is_enabled value.
		// The ADO API has eventual consistency: the PATCH response and
		// GetProcessByItsId can both return stale is_enabled immediately after
		// an update, causing Terraform's post-apply consistency check to fail.
		polled, pollErr := r.waitForProcessIsEnabled(ctx, processTypeID, model.IsEnabled.ValueBool())
		if pollErr != nil || polled == nil {
			// Timed out or poll failed — fall back to plan values so state is
			// at least internally consistent; the next plan/apply will reconcile.
			r.flattenProcess(&model, created)
			model.IsEnabled = types.BoolValue(model.IsEnabled.ValueBool())
			model.IsDefault = types.BoolValue(model.IsDefault.ValueBool())
		} else {
			r.flattenProcess(&model, polled)
		}
	} else {
		r.flattenProcess(&model, created)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *processResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_process Read: provider client not configured")
		return
	}

	var model processResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := model.ID.ValueString()
	process, err := r.client.WorkItemTrackingProcessClient.GetProcessByItsId(ctx, workitemtrackingprocess.GetProcessByItsIdArgs{
		ProcessTypeId: converter.UUID(processID),
		Expand:        &workitemtrackingprocess.GetProcessExpandLevelValues.None,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading process %s: %s", processID, err))
		return
	}

	r.flattenProcess(&model, process)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *processResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_process Update: provider client not configured")
		return
	}

	var model processResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateModel processResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processTypeID, err := uuid.Parse(stateModel.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("cannot parse process ID %q: %s", stateModel.ID.ValueString(), err))
		return
	}

	_, err2 := r.client.WorkItemTrackingProcessClient.EditProcess(ctx, workitemtrackingprocess.EditProcessArgs{
		ProcessTypeId: &processTypeID,
		UpdateRequest: &workitemtrackingprocess.UpdateProcessModel{
			Name:        converter.String(model.Name.ValueString()),
			Description: converter.String(model.Description.ValueString()),
			IsDefault:   converter.Bool(model.IsDefault.ValueBool()),
			IsEnabled:   converter.Bool(model.IsEnabled.ValueBool()),
		},
	})
	if err2 != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating process: %s", err2))
		return
	}

	// Poll until GetProcessByItsId reflects the intended is_enabled value.
	// The ADO API has eventual consistency: a GET immediately after a PATCH can
	// return stale is_enabled, causing Terraform's post-apply consistency check
	// to fail ("refresh plan was not empty").
	polled, pollErr := r.waitForProcessIsEnabled(ctx, &processTypeID, model.IsEnabled.ValueBool())

	model.ID = stateModel.ID
	if pollErr != nil || polled == nil {
		// Timed out or poll failed — store plan values directly so state is
		// at least consistent with the intended configuration; the next plan
		// will reconcile against the live API.
		model.IsEnabled = types.BoolValue(model.IsEnabled.ValueBool())
		model.IsDefault = types.BoolValue(model.IsDefault.ValueBool())
	} else {
		r.flattenProcess(&model, polled)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *processResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_process Delete: provider client not configured")
		return
	}

	var model processResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.WorkItemTrackingProcessClient.DeleteProcessById(ctx, workitemtrackingprocess.DeleteProcessByIdArgs{
		ProcessTypeId: converter.UUID(model.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting process: %s", err))
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *processResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	processID := req.ID
	model := processResourceModel{
		ID: types.StringValue(processID),
	}
	process, err := r.client.WorkItemTrackingProcessClient.GetProcessByItsId(ctx, workitemtrackingprocess.GetProcessByItsIdArgs{
		ProcessTypeId: converter.UUID(processID),
		Expand:        &workitemtrackingprocess.GetProcessExpandLevelValues.None,
	})
	if err != nil {
		resp.Diagnostics.AddError("Import error", fmt.Sprintf("reading process %s: %s", processID, err))
		return
	}
	r.flattenProcess(&model, process)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// processMinPollInterval is the minimum time to wait between polls when waiting
// for eventual consistency after a process update. Set to 0 in tests.
var processMinPollInterval = 2 * time.Second

// waitForProcessIsEnabled polls GetProcessByItsId until the is_enabled field
// matches wantEnabled, or the timeout elapses. This handles the ADO eventual
// consistency issue where the read API lags behind a PATCH update.
// Returns the final ProcessInfo on success.
func (r *processResource) waitForProcessIsEnabled(ctx context.Context, processTypeID *uuid.UUID, wantEnabled bool) (*workitemtrackingprocess.ProcessInfo, error) {
	var result *workitemtrackingprocess.ProcessInfo

	stateConf := &sdkretry.StateChangeConf{
		Pending:                   []string{"inconsistent"},
		Target:                    []string{"consistent"},
		Timeout:                   2 * time.Minute,
		MinTimeout:                processMinPollInterval,
		ContinuousTargetOccurence: 2,
		Refresh: func() (interface{}, string, error) {
			p, err := r.client.WorkItemTrackingProcessClient.GetProcessByItsId(ctx, workitemtrackingprocess.GetProcessByItsIdArgs{
				ProcessTypeId: processTypeID,
				Expand:        &workitemtrackingprocess.GetProcessExpandLevelValues.None,
			})
			if err != nil {
				return nil, "", err
			}
			// isEnabled=nil from ADO means disabled (omitempty on the API response).
			gotEnabled := p.IsEnabled != nil && *p.IsEnabled
			if gotEnabled == wantEnabled {
				result = p
				return p, "consistent", nil
			}
			return p, "inconsistent", nil
		},
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		// Timeout: return last known state and a soft error so caller can fall
		// back to the plan value rather than failing hard.
		return result, fmt.Errorf("timed out waiting for process is_enabled=%v: %w", wantEnabled, err)
	}
	return result, nil
}

func (r *processResource) flattenProcess(model *processResourceModel, process *workitemtrackingprocess.ProcessInfo) {
	if process.TypeId != nil {
		model.ID = types.StringValue(process.TypeId.String())
	}
	if process.Name != nil {
		model.Name = types.StringValue(*process.Name)
	}
	if process.Description != nil {
		model.Description = types.StringValue(*process.Description)
	} else {
		model.Description = types.StringValue("")
	}
	if process.ParentProcessTypeId != nil {
		model.ParentProcessTypeID = types.StringValue(process.ParentProcessTypeId.String())
	}
	if process.ReferenceName != nil {
		model.ReferenceName = types.StringValue(*process.ReferenceName)
	}
	if process.IsDefault != nil {
		model.IsDefault = types.BoolValue(*process.IsDefault)
	} else {
		model.IsDefault = types.BoolValue(false)
	}
	if process.IsEnabled != nil {
		model.IsEnabled = types.BoolValue(*process.IsEnabled)
	} else {
		// The ADO API omits isEnabled when it is false (omitempty). When the list
		// API returns nil for IsEnabled it means the process is disabled.
		model.IsEnabled = types.BoolValue(false)
	}
	if process.CustomizationType != nil {
		model.CustomizationType = types.StringValue(string(*process.CustomizationType))
	}
}
