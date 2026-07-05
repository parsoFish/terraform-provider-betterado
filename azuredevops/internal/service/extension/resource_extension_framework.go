package extension

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/extensionmanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*ExtensionResource)(nil)
	_ resource.ResourceWithConfigure   = (*ExtensionResource)(nil)
	_ resource.ResourceWithImportState = (*ExtensionResource)(nil)
)

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// extensionUseStateForUnknownString copies the prior state value when the plan
// value is unknown (prevents perpetual diffs for computed-only string attributes).
type extensionUseStateForUnknownString struct{}

func extensionUseStateForUnknown() planmodifier.String { return extensionUseStateForUnknownString{} }

func (m extensionUseStateForUnknownString) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m extensionUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m extensionUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// extensionRequiresReplaceString marks the resource for replacement when the
// attribute value changes from what is in state.
type extensionRequiresReplaceString struct{}

func extensionRequiresReplace() planmodifier.String { return extensionRequiresReplaceString{} }
func (m extensionRequiresReplaceString) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m extensionRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m extensionRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// extensionUseStateForUnknownBoolMod copies the prior state value for Bool attributes.
type extensionUseStateForUnknownBoolMod struct{}

func extensionUseStateForUnknownBool() planmodifier.Bool {
	return extensionUseStateForUnknownBoolMod{}
}

func (m extensionUseStateForUnknownBoolMod) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m extensionUseStateForUnknownBoolMod) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m extensionUseStateForUnknownBoolMod) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// extensionUseStateForUnknownList copies the prior state value for List attributes.
type extensionUseStateForUnknownList struct{}

func extensionUseStateForUnknownListMod() planmodifier.List { return extensionUseStateForUnknownList{} }

func (m extensionUseStateForUnknownList) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m extensionUseStateForUnknownList) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m extensionUseStateForUnknownList) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ExtensionResource is the terraform-plugin-framework implementation of betterado_extension.
type ExtensionResource struct {
	client *client.AggregatedClient
}

// NewExtensionResource returns a new resource.Resource for betterado_extension.
func NewExtensionResource() resource.Resource {
	return &ExtensionResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ExtensionResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_extension"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ExtensionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure DevOps Extension installation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource (publisher_id/extension_id).",
				PlanModifiers: []planmodifier.String{
					extensionUseStateForUnknown(),
				},
			},
			"extension_id": schema.StringAttribute{
				Required:    true,
				Description: "The Azure DevOps Marketplace extension identifier, for example `vss-code-search`.",
				PlanModifiers: []planmodifier.String{
					extensionRequiresReplace(),
				},
			},
			"publisher_id": schema.StringAttribute{
				Required:    true,
				Description: "The Azure DevOps Marketplace publisher identifier, for example `ms`.",
				PlanModifiers: []planmodifier.String{
					extensionRequiresReplace(),
				},
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the extension is disabled. Defaults to false when not set.",
				PlanModifiers: []planmodifier.Bool{
					extensionUseStateForUnknownBool(),
				},
			},
			"version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The version of the extension to install. If not specified, the latest version is installed.",
				PlanModifiers: []planmodifier.String{
					extensionUseStateForUnknown(),
				},
			},
			"extension_name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the extension.",
				PlanModifiers: []planmodifier.String{
					extensionUseStateForUnknown(),
				},
			},
			"publisher_name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the publisher.",
				PlanModifiers: []planmodifier.String{
					extensionUseStateForUnknown(),
				},
			},
			"scope": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The list of OAuth scopes required by the extension.",
				PlanModifiers: []planmodifier.List{
					extensionUseStateForUnknownListMod(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ExtensionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// extensionModel is the Terraform state model for betterado_extension.
type extensionModel struct {
	ID            types.String `tfsdk:"id"`
	ExtensionID   types.String `tfsdk:"extension_id"`
	PublisherID   types.String `tfsdk:"publisher_id"`
	Disabled      types.Bool   `tfsdk:"disabled"`
	Version       types.String `tfsdk:"version"`
	ExtensionName types.String `tfsdk:"extension_name"`
	PublisherName types.String `tfsdk:"publisher_name"`
	Scope         types.List   `tfsdk:"scope"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *ExtensionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model extensionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	publisherID := model.PublisherID.ValueString()
	extensionID := model.ExtensionID.ValueString()

	installArgs := extensionmanagement.InstallExtensionByNameArgs{
		PublisherName: &publisherID,
		ExtensionName: &extensionID,
	}
	if !model.Version.IsNull() && !model.Version.IsUnknown() && model.Version.ValueString() != "" {
		installArgs.Version = converter.String(model.Version.ValueString())
	}

	installed, err := r.client.ExtensionManagementClient.InstallExtensionByName(r.client.Ctx, installArgs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Installing Azure DevOps Extension",
			fmt.Sprintf("Installing extension for Publisher: %s, Name: %s. Error: %v", publisherID, extensionID, err),
		)
		return
	}

	model.ID = types.StringValue(fmt.Sprintf("%s/%s", *installed.PublisherId, *installed.ExtensionId))

	// Apply disabled flag if explicitly set (not null).
	if !model.Disabled.IsNull() && !model.Disabled.IsUnknown() {
		if err := r.setDisabledState(publisherID, extensionID, model.Version.ValueString(), model.Disabled.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Setting extension disabled state", err.Error())
			return
		}
	}

	found := r.readIntoModel(&model)
	if !found {
		resp.Diagnostics.AddError("Reading Azure DevOps Extension after create", "Extension was not found immediately after installation")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ExtensionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model extensionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found := r.readIntoModel(&model)
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ExtensionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan extensionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state extensionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID

	// Only update disabled state when explicitly set (not null).
	// This preserves the SDKv2 GetRawConfig().AsValueMap()["disabled"] semantics:
	// an unset (null) disabled attribute does not trigger an update API call.
	if !plan.Disabled.IsNull() && !plan.Disabled.IsUnknown() {
		version := plan.Version.ValueString()
		if err := r.setDisabledState(plan.PublisherID.ValueString(), plan.ExtensionID.ValueString(), version, plan.Disabled.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Updating Azure DevOps Extension disabled state", err.Error())
			return
		}
	}

	found := r.readIntoModel(&plan)
	if !found {
		resp.Diagnostics.AddError("Reading Azure DevOps Extension after update", "Extension was not found after update")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ExtensionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model extensionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	publisherID := model.PublisherID.ValueString()
	extensionID := model.ExtensionID.ValueString()

	err := r.client.ExtensionManagementClient.UninstallExtensionByName(r.client.Ctx, extensionmanagement.UninstallExtensionByNameArgs{
		PublisherName: &publisherID,
		ExtensionName: &extensionID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Uninstalling Azure DevOps Extension",
			fmt.Sprintf("Uninstalling extension for Publisher: %s, Name: %s. Error: %v", publisherID, extensionID, err),
		)
	}
}

// ── Import ────────────────────────────────────────────────────────────────────

func (r *ExtensionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in the format <publisher_id>/<extension_id>, got: %s", req.ID),
		)
		return
	}

	model := extensionModel{
		ID:          types.StringValue(req.ID),
		PublisherID: types.StringValue(parts[0]),
		ExtensionID: types.StringValue(parts[1]),
		Scope:       types.ListValueMust(types.StringType, []attr.Value{}),
	}

	found := r.readIntoModel(&model)
	if !found {
		resp.Diagnostics.AddError("Reading Azure DevOps Extension during import", fmt.Sprintf("Extension %s not found", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// readIntoModel fetches the current state from ADO and fills model.
// Returns false if the resource was not found (caller removes from state).
func (r *ExtensionResource) readIntoModel(model *extensionModel) bool {
	publisherID := model.PublisherID.ValueString()
	extensionID := model.ExtensionID.ValueString()

	extension, err := r.client.ExtensionManagementClient.GetInstalledExtensionByName(r.client.Ctx, extensionmanagement.GetInstalledExtensionByNameArgs{
		PublisherName: &publisherID,
		ExtensionName: &extensionID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return false
		}
		return false
	}
	if extension == nil {
		return false
	}

	if extension.ExtensionId != nil {
		model.ExtensionID = types.StringValue(*extension.ExtensionId)
	}
	if extension.PublisherId != nil {
		model.PublisherID = types.StringValue(*extension.PublisherId)
	}
	if extension.Version != nil {
		model.Version = types.StringValue(*extension.Version)
	}
	if extension.ExtensionName != nil {
		model.ExtensionName = types.StringValue(*extension.ExtensionName)
	}
	if extension.PublisherName != nil {
		model.PublisherName = types.StringValue(*extension.PublisherName)
	}

	// Map scopes.
	scopeVals := []attr.Value{}
	if extension.Scopes != nil {
		for _, s := range *extension.Scopes {
			scopeVals = append(scopeVals, types.StringValue(s))
		}
	}
	model.Scope = types.ListValueMust(types.StringType, scopeVals)

	// Map disabled flag from InstallState.Flags.
	disabled := false
	if extension.InstallState != nil && extension.InstallState.Flags != nil {
		flagsStr := string(*extension.InstallState.Flags)
		for _, flag := range strings.Split(flagsStr, ",") {
			if flag == string(extensionmanagement.ExtensionStateFlagsValues.Disabled) {
				disabled = true
				break
			}
		}
	}
	model.Disabled = types.BoolValue(disabled)

	return true
}

// setDisabledState calls UpdateInstalledExtension to enable or disable the extension.
func (r *ExtensionResource) setDisabledState(publisherID, extensionID, version string, disabled bool) error {
	flags := extensionmanagement.ExtensionStateFlagsValues.None
	if disabled {
		flags = extensionmanagement.ExtensionStateFlagsValues.Disabled
	}

	ext := &extensionmanagement.InstalledExtension{
		PublisherId: &publisherID,
		ExtensionId: &extensionID,
		InstallState: &extensionmanagement.InstalledExtensionState{
			Flags: converter.ToPtr(flags),
		},
	}
	if version != "" {
		ext.Version = converter.String(version)
	}

	_, err := r.client.ExtensionManagementClient.UpdateInstalledExtension(r.client.Ctx, extensionmanagement.UpdateInstalledExtensionArgs{
		Extension: ext,
	})
	if err != nil {
		return fmt.Errorf("Update extension for Publisher: %s, Name: %s. Error: %v", publisherID, extensionID, err)
	}
	return nil
}
