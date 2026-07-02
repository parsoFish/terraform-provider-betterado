package build

// resource_build_folder_framework.go — terraform-plugin-framework implementation of
// betterado_build_folder.
//
// The SDKv2 ResourceBuildFolder() in resource_build_folder.go remains in place for
// state-file compatibility. The mux serves new apply/plan calls through this
// framework resource once it is registered in framework_provider.go Resources().
// The SDKv2 "betterado_build_folder" entry in provider.go ResourcesMap is removed
// (commented with a migration note) to avoid "Invalid Provider Server Combination".
//
// No build tag on this production file — only _test.go files carry //go:build.

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &BuildFolderResource{}
	_ resource.ResourceWithImportState = &BuildFolderResource{}
	_ resource.ResourceWithConfigure   = &BuildFolderResource{}
)

// BuildFolderResource is the framework resource for betterado_build_folder.
type BuildFolderResource struct {
	client *client.AggregatedClient
}

// NewBuildFolderResource returns a new framework resource.Resource for betterado_build_folder.
func NewBuildFolderResource() resource.Resource {
	return &BuildFolderResource{}
}

// ── Model ────────────────────────────────────────────────────────────────────

type buildFolderModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Path        types.String `tfsdk:"path"`
	Description types.String `tfsdk:"description"`
}

// ── Metadata / Schema ────────────────────────────────────────────────────────

func (r *BuildFolderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_folder"
}

func (r *BuildFolderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a build (pipeline) folder within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The project ID of the build folder resource.",
				PlanModifiers: []planmodifier.String{
					bfUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the build folder belongs to.",
				PlanModifiers: []planmodifier.String{
					bfRequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "The folder path within the build definitions tree (e.g. `\\MyFolder`).",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The description of the build folder.",
				Default:     bfStaticString(""),
			},
		},
	}
}

// ── Provider data injection ───────────────────────────────────────────────────

func (r *BuildFolderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("betterado_build_folder: expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

// ── Create ───────────────────────────────────────────────────────────────────

func (r *BuildFolderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_build_folder Create: provider client not configured")
		return
	}

	var model buildFolderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	path := model.Path.ValueString()
	description := model.Description.ValueString()

	created, err := createBuildFolder(r.client, path, projectID, description)
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating build folder: %s", err))
		return
	}

	model.ID = types.StringValue(created.Project.Id.String())
	if created.Path != nil {
		model.Path = types.StringValue(*created.Path)
	}
	if created.Description != nil {
		model.Description = types.StringValue(*created.Description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ─────────────────────────────────────────────────────────────────────

func (r *BuildFolderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_build_folder Read: provider client not configured")
		return
	}

	var model buildFolderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	path := model.Path.ValueString()

	folders, err := r.client.BuildClient.GetFolders(r.client.Ctx, build.GetFoldersArgs{
		Project: &projectID,
		Path:    &path,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading build folder (path: %s): %s", path, err))
		return
	}

	if folders == nil || len(*folders) == 0 {
		// Externally deleted — remove from state.
		resp.State.RemoveResource(ctx)
		return
	}

	folder := (*folders)[0]
	model.ProjectID = types.StringValue(projectID)
	if folder.Path != nil {
		model.Path = types.StringValue(*folder.Path)
	}
	if folder.Description != nil {
		model.Description = types.StringValue(*folder.Description)
	} else {
		model.Description = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ───────────────────────────────────────────────────────────────────

func (r *BuildFolderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_build_folder Update: provider client not configured")
		return
	}

	var plan buildFolderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state buildFolderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	oldPath := state.Path.ValueString()
	newPath := plan.Path.ValueString()
	description := plan.Description.ValueString()

	_, err := r.client.BuildClient.UpdateFolder(r.client.Ctx, build.UpdateFolderArgs{
		Project: &projectID,
		Path:    converter.String(oldPath),
		Folder: &build.Folder{
			Description: converter.String(description),
			Path:        converter.String(newPath),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating build folder (path: %s): %s", oldPath, err))
		return
	}

	plan.Path = types.StringValue(newPath)
	plan.Description = types.StringValue(description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ───────────────────────────────────────────────────────────────────

func (r *BuildFolderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_build_folder Delete: provider client not configured")
		return
	}

	var model buildFolderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	path := model.Path.ValueString()

	if err := r.client.BuildClient.DeleteFolder(r.client.Ctx, build.DeleteFolderArgs{
		Project: converter.ToPtr(projectID),
		Path:    converter.ToPtr(path),
	}); err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting build folder (path: %s): %s", path, err))
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

// ImportState supports `terraform import betterado_build_folder.x "<project_id>/<path>"`.
func (r *BuildFolderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	slashIdx := -1
	for i, ch := range id {
		if ch == '/' {
			slashIdx = i
			break
		}
	}
	if slashIdx < 0 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format '<project_id>/<path>', got %q", id),
		)
		return
	}
	projectID := id[:slashIdx]
	path := id[slashIdx+1:]

	model := buildFolderModel{
		ID:          types.StringValue(projectID),
		ProjectID:   types.StringValue(projectID),
		Path:        types.StringValue(path),
		Description: types.StringValue(""),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Inline plan-modifiers & defaults ─────────────────────────────────────────
// Mirrors the pattern in azuredevops/internal/service/release/framework_defaults.go.
// Sub-packages like stringplanmodifier / stringdefault are not vendored, so we
// implement the minimal interfaces inline here.

// bfStaticStringImpl is a minimal defaults.String that returns a constant.
type bfStaticStringImpl struct{ value string }

func bfStaticString(v string) defaults.String { return bfStaticStringImpl{value: v} }

func (d bfStaticStringImpl) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d bfStaticStringImpl) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%q`", d.value)
}

func (d bfStaticStringImpl) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// bfUseStateForUnknownImpl is equivalent to stringplanmodifier.UseStateForUnknown().
type bfUseStateForUnknownImpl struct{}

func bfUseStateForUnknown() planmodifier.String { return bfUseStateForUnknownImpl{} }

func (bfUseStateForUnknownImpl) Description(_ context.Context) string {
	return "use prior state value for unknown"
}

func (bfUseStateForUnknownImpl) MarkdownDescription(_ context.Context) string {
	return "use prior state value for unknown"
}

func (bfUseStateForUnknownImpl) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// bfRequiresReplaceImpl is equivalent to stringplanmodifier.RequiresReplace().
type bfRequiresReplaceImpl struct{}

func bfRequiresReplace() planmodifier.String { return bfRequiresReplaceImpl{} }

func (bfRequiresReplaceImpl) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (bfRequiresReplaceImpl) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (bfRequiresReplaceImpl) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}
