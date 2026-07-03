package extensionmanagement

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/extensionmanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*ExtensionInstallResource)(nil)
	_ resource.ResourceWithConfigure = (*ExtensionInstallResource)(nil)
)

// ExtensionInstallResource is the terraform-plugin-framework implementation of
// betterado_extension_install.
type ExtensionInstallResource struct {
	client *client.AggregatedClient
}

// NewExtensionInstallResource returns a new resource.Resource for betterado_extension_install.
func NewExtensionInstallResource() resource.Resource {
	return &ExtensionInstallResource{}
}

// ── stringNotEmpty validator ──────────────────────────────────────────────────

// stringNotEmptyValidator rejects empty or whitespace-only strings.
type stringNotEmptyValidator struct{}

func (v stringNotEmptyValidator) Description(_ context.Context) string {
	return "string must not be empty or whitespace"
}

func (v stringNotEmptyValidator) MarkdownDescription(_ context.Context) string {
	return "string must not be empty or whitespace"
}

func (v stringNotEmptyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.TrimSpace(req.ConfigValue.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Empty String",
			"Value must not be empty or consist only of whitespace characters.",
		)
	}
}

// ── requiresReplaceString plan modifier ──────────────────────────────────────

type requiresReplaceStringModifier struct{}

func requiresReplaceString() planmodifier.String { return requiresReplaceStringModifier{} }

func (m requiresReplaceStringModifier) Description(_ context.Context) string {
	return "Forces replacement when value changes."
}

func (m requiresReplaceStringModifier) MarkdownDescription(_ context.Context) string {
	return "Forces replacement when value changes."
}

func (m requiresReplaceStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── useStateForUnknownString plan modifier ───────────────────────────────────

type useStateForUnknownStringModifier struct{}

func useStateForUnknownString() planmodifier.String { return useStateForUnknownStringModifier{} }

func (m useStateForUnknownStringModifier) Description(_ context.Context) string {
	return "Uses prior state value when plan value is unknown."
}

func (m useStateForUnknownStringModifier) MarkdownDescription(_ context.Context) string {
	return "Uses prior state value when plan value is unknown."
}

func (m useStateForUnknownStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── State model ───────────────────────────────────────────────────────────────

// extensionInstallModel is the Terraform state model for betterado_extension_install.
type extensionInstallModel struct {
	ID          types.String `tfsdk:"id"`
	PublisherID types.String `tfsdk:"publisher_id"`
	ExtensionID types.String `tfsdk:"extension_id"`
	Version     types.String `tfsdk:"version"`
	Disabled    types.Bool   `tfsdk:"disabled"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ExtensionInstallResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_extension_install"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ExtensionInstallResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the installation of an Azure DevOps Marketplace extension at the organization level.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					useStateForUnknownString(),
				},
			},
			"publisher_id": schema.StringAttribute{
				Required:    true,
				Description: "The Azure DevOps Marketplace publisher identifier, for example `ms`.",
				Validators: []validator.String{
					stringNotEmptyValidator{},
				},
				PlanModifiers: []planmodifier.String{
					requiresReplaceString(),
				},
			},
			"extension_id": schema.StringAttribute{
				Required:    true,
				Description: "The Azure DevOps Marketplace extension identifier, for example `vss-code-search`.",
				Validators: []validator.String{
					stringNotEmptyValidator{},
				},
				PlanModifiers: []planmodifier.String{
					requiresReplaceString(),
				},
			},
			"version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The version of the extension to install. If omitted, the latest version is installed.",
				PlanModifiers: []planmodifier.String{
					useStateForUnknownString(),
				},
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the extension is disabled. Defaults to false.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ExtensionInstallResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ExtensionInstallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model extensionInstallModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	publisherID := model.PublisherID.ValueString()
	extensionID := model.ExtensionID.ValueString()

	var versionPtr *string
	if !model.Version.IsNull() && !model.Version.IsUnknown() && model.Version.ValueString() != "" {
		versionPtr = converter.String(model.Version.ValueString())
	}

	installed, err := r.client.ExtensionManagementClient.InstallExtensionByName(ctx,
		extensionmanagement.InstallExtensionByNameArgs{
			PublisherName: &publisherID,
			ExtensionName: &extensionID,
			Version:       versionPtr,
		})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error installing extension",
			fmt.Sprintf("Installing extension for Publisher: %s, Name: %s. Error: %v", publisherID, extensionID, err),
		)
		return
	}

	model.ID = types.StringValue(fmt.Sprintf("%s/%s", *installed.PublisherId, *installed.ExtensionId))
	r.flattenIntoModel(&model, installed)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ExtensionInstallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model extensionInstallModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	publisherID := model.PublisherID.ValueString()
	extensionID := model.ExtensionID.ValueString()

	installed, err := r.client.ExtensionManagementClient.GetInstalledExtensionByName(ctx,
		extensionmanagement.GetInstalledExtensionByNameArgs{
			PublisherName: &publisherID,
			ExtensionName: &extensionID,
		})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading extension",
			fmt.Sprintf("Get extension for Publisher: %s, Name: %s. Error: %v", publisherID, extensionID, err),
		)
		return
	}

	if installed == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.flattenIntoModel(&model, installed)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ExtensionInstallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan extensionInstallModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state extensionInstallModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	publisherID := plan.PublisherID.ValueString()
	extensionID := plan.ExtensionID.ValueString()

	flags := extensionmanagement.ExtensionStateFlagsValues.None
	if !plan.Disabled.IsNull() && !plan.Disabled.IsUnknown() && plan.Disabled.ValueBool() {
		flags = extensionmanagement.ExtensionStateFlagsValues.Disabled
	}

	updateArgs := extensionmanagement.UpdateInstalledExtensionArgs{
		Extension: &extensionmanagement.InstalledExtension{
			PublisherId: &publisherID,
			ExtensionId: &extensionID,
			InstallState: &extensionmanagement.InstalledExtensionState{
				Flags: converter.ToPtr(flags),
			},
		},
	}

	if !plan.Version.IsNull() && !plan.Version.IsUnknown() && plan.Version.ValueString() != "" {
		updateArgs.Extension.Version = converter.String(plan.Version.ValueString())
	}

	updated, err := r.client.ExtensionManagementClient.UpdateInstalledExtension(ctx, updateArgs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating extension",
			fmt.Sprintf("Update extension for Publisher: %s, Name: %s. Error: %v", publisherID, extensionID, err),
		)
		return
	}

	plan.ID = state.ID
	if updated != nil {
		r.flattenIntoModel(&plan, updated)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ExtensionInstallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model extensionInstallModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	publisherID := model.PublisherID.ValueString()
	extensionID := model.ExtensionID.ValueString()

	err := r.client.ExtensionManagementClient.UninstallExtensionByName(ctx,
		extensionmanagement.UninstallExtensionByNameArgs{
			PublisherName: &publisherID,
			ExtensionName: &extensionID,
		})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error uninstalling extension",
			fmt.Sprintf("Uninstalling extension for Publisher: %s, Name: %s. Error: %v", publisherID, extensionID, err),
		)
	}
}

// ── Expand / Flatten helpers ──────────────────────────────────────────────────

// expandExtensionInstall converts the Terraform model into an InstalledExtension struct
// suitable for API calls (e.g. UpdateInstalledExtension).
func expandExtensionInstall(model *extensionInstallModel) *extensionmanagement.InstalledExtension {
	publisherID := model.PublisherID.ValueString()
	extensionID := model.ExtensionID.ValueString()

	ext := &extensionmanagement.InstalledExtension{
		PublisherId: &publisherID,
		ExtensionId: &extensionID,
	}

	if !model.Version.IsNull() && !model.Version.IsUnknown() {
		ext.Version = converter.String(model.Version.ValueString())
	}

	flags := extensionmanagement.ExtensionStateFlagsValues.None
	if !model.Disabled.IsNull() && !model.Disabled.IsUnknown() && model.Disabled.ValueBool() {
		flags = extensionmanagement.ExtensionStateFlagsValues.Disabled
	}
	ext.InstallState = &extensionmanagement.InstalledExtensionState{
		Flags: converter.ToPtr(flags),
	}

	return ext
}

// flattenExtensionInstall populates the model from an InstalledExtension API response.
func flattenExtensionInstall(model *extensionInstallModel, ext *extensionmanagement.InstalledExtension) {
	if ext.PublisherId != nil {
		model.PublisherID = types.StringValue(*ext.PublisherId)
	}
	if ext.ExtensionId != nil {
		model.ExtensionID = types.StringValue(*ext.ExtensionId)
	}
	if ext.Version != nil {
		model.Version = types.StringValue(*ext.Version)
	} else {
		model.Version = types.StringValue("")
	}

	disabled := false
	if ext.InstallState != nil && ext.InstallState.Flags != nil {
		flagsStr := string(*ext.InstallState.Flags)
		flags := strings.Split(flagsStr, ",")
		for _, flag := range flags {
			if flag == string(extensionmanagement.ExtensionStateFlagsValues.Disabled) {
				disabled = true
				break
			}
		}
	}
	model.Disabled = types.BoolValue(disabled)
}

// flattenIntoModel is a convenience wrapper for CRUD methods.
func (r *ExtensionInstallResource) flattenIntoModel(model *extensionInstallModel, ext *extensionmanagement.InstalledExtension) {
	flattenExtensionInstall(model, ext)
}
