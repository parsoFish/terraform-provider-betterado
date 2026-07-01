package release

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance.
var (
	_ resource.Resource                = &releaseFolderFrameworkResource{}
	_ resource.ResourceWithImportState = &releaseFolderFrameworkResource{}
)

// releaseFolderFrameworkResource is the terraform-plugin-framework implementation
// of betterado_release_folder.
type releaseFolderFrameworkResource struct {
	client *client.AggregatedClient
}

// NewReleaseFolderResource returns a new framework resource for betterado_release_folder.
func NewReleaseFolderResource() resource.Resource {
	return &releaseFolderFrameworkResource{}
}

// ── Model ────────────────────────────────────────────────────────────────────

type releaseFolderModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Path        types.String `tfsdk:"path"`
	Description types.String `tfsdk:"description"`
}

// ── Metadata / Schema ────────────────────────────────────────────────────────

func (r *releaseFolderFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_release_folder"
}

func (r *releaseFolderFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a release folder within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The path of the release folder (used as the resource ID).",
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the release folder belongs to.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "The path of the release folder.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The description of the release folder.",
				Default:     staticString(""),
			},
		},
	}
}

// ── Provider data injection ───────────────────────────────────────────────────

func (r *releaseFolderFrameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── Create ───────────────────────────────────────────────────────────────────

func (r *releaseFolderFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_folder Create: provider client not configured")
		return
	}

	var model releaseFolderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	folder := &releaseapi.Folder{
		Path:        converter.String(model.Path.ValueString()),
		Description: converter.String(model.Description.ValueString()),
	}

	created, err := r.client.ReleaseClient.CreateFolder(r.client.Ctx, releaseapi.CreateFolderArgs{
		Folder:  folder,
		Project: &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating release folder: %s", err))
		return
	}

	model.ID = types.StringValue(*created.Path)
	model.Path = types.StringValue(*created.Path)
	if created.Description != nil {
		model.Description = types.StringValue(*created.Description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ─────────────────────────────────────────────────────────────────────

func (r *releaseFolderFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_folder Read: provider client not configured")
		return
	}

	var model releaseFolderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := model.ID.ValueString()
	projectID := model.ProjectID.ValueString()

	folders, err := r.client.ReleaseClient.GetFolders(r.client.Ctx, releaseapi.GetFoldersArgs{
		Project: &projectID,
		Path:    &path,
	})
	if err != nil || folders == nil || len(*folders) == 0 {
		// External delete — remove from state.
		resp.State.RemoveResource(ctx)
		return
	}

	folder := (*folders)[0]
	model.Path = types.StringValue(*folder.Path)
	model.ID = types.StringValue(*folder.Path)
	if folder.Description != nil {
		model.Description = types.StringValue(*folder.Description)
	} else {
		model.Description = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ───────────────────────────────────────────────────────────────────

func (r *releaseFolderFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_folder Update: provider client not configured")
		return
	}

	var model releaseFolderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// path is ForceNew — the only updatable field is description.
	var stateModel releaseFolderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	path := stateModel.Path.ValueString()

	folder := &releaseapi.Folder{
		Path:        converter.String(path),
		Description: converter.String(model.Description.ValueString()),
	}

	updated, err := r.client.ReleaseClient.UpdateFolder(r.client.Ctx, releaseapi.UpdateFolderArgs{
		Folder:  folder,
		Project: &projectID,
		Path:    &path,
	})
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating release folder: %s", err))
		return
	}

	model.ID = types.StringValue(*updated.Path)
	model.Path = types.StringValue(*updated.Path)
	if updated.Description != nil {
		model.Description = types.StringValue(*updated.Description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Delete ───────────────────────────────────────────────────────────────────

func (r *releaseFolderFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_folder Delete: provider client not configured")
		return
	}

	var model releaseFolderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := model.ID.ValueString()
	projectID := model.ProjectID.ValueString()

	if err := r.client.ReleaseClient.DeleteFolder(r.client.Ctx, releaseapi.DeleteFolderArgs{
		Project: &projectID,
		Path:    &path,
	}); err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting release folder: %s", err))
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *releaseFolderFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "<project_id>/<path>"  e.g. "abc-123/\MyFolder"
	import_ := req.ID
	// Find the first slash that separates project_id from path.
	slashIdx := -1
	for i, ch := range import_ {
		if ch == '/' {
			slashIdx = i
			break
		}
	}
	if slashIdx < 0 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format '<project_id>/<path>', got %q", import_),
		)
		return
	}
	projectID := import_[:slashIdx]
	path := import_[slashIdx+1:]
	// path in ADO uses backslashes; the user may have provided a forward slash
	// as a URL-safe separator — keep as-is (the API accepts both).

	model := releaseFolderModel{
		ID:          types.StringValue(path),
		ProjectID:   types.StringValue(projectID),
		Path:        types.StringValue(path),
		Description: types.StringValue(""),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
