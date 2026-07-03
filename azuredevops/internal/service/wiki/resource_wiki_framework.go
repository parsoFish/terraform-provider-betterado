package wiki

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	azwiki "github.com/microsoft/azure-devops-go-api/azuredevops/v7/wiki"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ── Inline plan modifiers ──────────────────────────────────────────────────────

// wikiRequiresReplaceModifier forces resource replacement when the string
// attribute changes (equivalent to SDKv2 ForceNew).
type wikiRequiresReplaceModifier struct{}

func wikiRequiresReplace() planmodifier.String { return wikiRequiresReplaceModifier{} }

func (m wikiRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m wikiRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m wikiRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// wikiUseStateForUnknownModifier keeps the prior state value when the plan
// value is unknown (prevents perpetual diffs for computed-only attributes).
type wikiUseStateForUnknownModifier struct{}

func wikiUseStateForUnknown() planmodifier.String { return wikiUseStateForUnknownModifier{} }

func (m wikiUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m wikiUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m wikiUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// Compile-time interface checks.
var (
	_ resource.Resource                = (*WikiResource)(nil)
	_ resource.ResourceWithConfigure   = (*WikiResource)(nil)
	_ resource.ResourceWithImportState = (*WikiResource)(nil)
)

// ── Inline OneOf string validator ─────────────────────────────────────────────

// oneOfStringValidator validates that a string attribute is one of the allowed values.
type oneOfStringValidator struct {
	allowed []string
}

func (v oneOfStringValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be one of: %v", v.allowed)
}

func (v oneOfStringValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value must be one of: `%v`", v.allowed)
}

func (v oneOfStringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	for _, a := range v.allowed {
		if val == a {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid value",
		fmt.Sprintf("%q is not one of %v", val, v.allowed),
	)
}

// WikiResource is the terraform-plugin-framework implementation of betterado_wiki.
type WikiResource struct {
	client *client.AggregatedClient
}

// NewWikiResource returns a new resource.Resource for betterado_wiki.
func NewWikiResource() resource.Resource {
	return &WikiResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *WikiResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_wiki"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *WikiResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					wikiUseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					oneOfStringValidator{allowed: []string{
						string(azwiki.WikiTypeValues.ProjectWiki),
						string(azwiki.WikiTypeValues.CodeWiki),
					}},
				},
				PlanModifiers: []planmodifier.String{
					wikiRequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					wikiRequiresReplace(),
					wikiUseStateForUnknown(),
				},
			},
			"mapped_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					wikiUseStateForUnknown(),
				},
			},
			"repository_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					wikiUseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					wikiUseStateForUnknown(),
				},
			},
			"remote_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					wikiUseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					wikiUseStateForUnknown(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *WikiResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// wikiModel is the Terraform state model for betterado_wiki.
type wikiModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	ProjectID    types.String `tfsdk:"project_id"`
	MappedPath   types.String `tfsdk:"mapped_path"`
	RepositoryID types.String `tfsdk:"repository_id"`
	Version      types.String `tfsdk:"version"`
	RemoteURL    types.String `tfsdk:"remote_url"`
	URL          types.String `tfsdk:"url"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *WikiResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model wikiModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wikiType := azwiki.WikiType(model.Type.ValueString())
	wikiArgs := &azwiki.WikiCreateParametersV2{
		Name: converter.String(model.Name.ValueString()),
		Type: &wikiType,
	}

	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() && model.ProjectID.ValueString() != "" {
		projectUUID, err := uuid.Parse(model.ProjectID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid project_id", fmt.Sprintf("project_id is not a valid UUID: %s", err.Error()))
			return
		}
		wikiArgs.ProjectId = &projectUUID
	}

	if !model.MappedPath.IsNull() && !model.MappedPath.IsUnknown() && model.MappedPath.ValueString() != "" {
		wikiArgs.MappedPath = converter.String(model.MappedPath.ValueString())
	}

	if !model.RepositoryID.IsNull() && !model.RepositoryID.IsUnknown() && model.RepositoryID.ValueString() != "" {
		repoUUID, err := uuid.Parse(model.RepositoryID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid repository_id", fmt.Sprintf("repository_id is not a valid UUID: %s", err.Error()))
			return
		}
		wikiArgs.RepositoryId = &repoUUID
	}

	if !model.Version.IsNull() && !model.Version.IsUnknown() && model.Version.ValueString() != "" {
		wikiArgs.Version = &git.GitVersionDescriptor{
			Version: converter.String(model.Version.ValueString()),
		}
	}

	created, err := r.client.WikiClient.CreateWiki(r.client.Ctx, azwiki.CreateWikiArgs{WikiCreateParams: wikiArgs})
	if err != nil {
		// ADO allows only one project wiki per project. If one already exists (e.g. from
		// a previous test run that did not complete cleanup), adopt it instead of failing.
		if wikiType == azwiki.WikiTypeValues.ProjectWiki &&
			strings.Contains(err.Error(), "already exists") {
			existing, adoptErr := r.adoptExistingProjectWiki(model.ProjectID.ValueString(), model.Name.ValueString())
			if adoptErr != nil {
				resp.Diagnostics.AddError("Error adopting existing project wiki", adoptErr.Error())
				return
			}
			r.flattenWiki(existing, &model)
			model.ID = types.StringValue(existing.Id.String())
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
		resp.Diagnostics.AddError("Error creating wiki", err.Error())
		return
	}

	model.ID = types.StringValue(created.Id.String())
	r.flattenWiki(created, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *WikiResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model wikiModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wikiResp, err := r.client.WikiClient.GetWiki(r.client.Ctx, azwiki.GetWikiArgs{WikiIdentifier: converter.String(model.ID.ValueString())})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading wiki", err.Error())
		return
	}

	r.flattenWiki(wikiResp, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *WikiResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan wikiModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state wikiModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only name can be updated (type and project_id are ForceNew via RequiresReplace).
	if !plan.Name.Equal(state.Name) {
		_, err := r.client.WikiClient.UpdateWiki(r.client.Ctx, azwiki.UpdateWikiArgs{
			WikiIdentifier: converter.String(state.ID.ValueString()),
			UpdateParameters: &azwiki.WikiUpdateParameters{
				Name: converter.String(plan.Name.ValueString()),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error updating wiki", err.Error())
			return
		}
	}

	// Read back the updated state.
	wikiResp, err := r.client.WikiClient.GetWiki(r.client.Ctx, azwiki.GetWikiArgs{WikiIdentifier: converter.String(state.ID.ValueString())})
	if err != nil {
		resp.Diagnostics.AddError("Error reading wiki after update", err.Error())
		return
	}

	plan.ID = state.ID
	r.flattenWiki(wikiResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WikiResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model wikiModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wikiType := azwiki.WikiType(model.Type.ValueString())
	if wikiType == azwiki.WikiTypeValues.CodeWiki {
		// Code wikis can be deleted directly via the Wiki API.
		projectID := model.ProjectID.ValueString()
		_, err := r.client.WikiClient.DeleteWiki(r.client.Ctx, azwiki.DeleteWikiArgs{
			WikiIdentifier: converter.String(model.ID.ValueString()),
			Project:        converter.String(projectID),
		})
		if err != nil {
			if !utils.ResponseWasNotFound(err) {
				resp.Diagnostics.AddError("Error deleting wiki", err.Error())
			}
		}
	} else if wikiType == azwiki.WikiTypeValues.ProjectWiki {
		// Project wikis cannot be deleted via the Wiki API; they are backed by a
		// special git repository that must be deleted to remove the wiki.
		wikiResp, err := r.client.WikiClient.GetWiki(r.client.Ctx, azwiki.GetWikiArgs{
			WikiIdentifier: converter.String(model.ID.ValueString()),
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				// Already gone — nothing to do.
				return
			}
			resp.Diagnostics.AddError("Error reading wiki before delete", err.Error())
			return
		}

		if wikiResp.RepositoryId == nil {
			// No backing repository — nothing to delete.
			return
		}

		err = r.client.GitReposClient.DeleteRepository(r.client.Ctx, git.DeleteRepositoryArgs{
			RepositoryId: wikiResp.RepositoryId,
			Project:      converter.String(model.ProjectID.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError("Error deleting project wiki repository", err.Error())
		}
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *WikiResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// flattenWiki maps an API wiki response into the state model.
func (r *WikiResource) flattenWiki(w *azwiki.WikiV2, model *wikiModel) {
	if w.Id != nil {
		model.ID = types.StringValue(w.Id.String())
	}
	if w.Name != nil {
		model.Name = types.StringValue(*w.Name)
	}
	if w.Type != nil {
		model.Type = types.StringValue(string(*w.Type))
	}
	if w.ProjectId != nil {
		model.ProjectID = types.StringValue(w.ProjectId.String())
	}
	if w.MappedPath != nil {
		model.MappedPath = types.StringValue(*w.MappedPath)
	} else {
		model.MappedPath = types.StringValue("")
	}
	if w.RemoteUrl != nil {
		model.RemoteURL = types.StringValue(*w.RemoteUrl)
	} else {
		model.RemoteURL = types.StringValue("")
	}
	if w.RepositoryId != nil {
		model.RepositoryID = types.StringValue(w.RepositoryId.String())
	} else {
		model.RepositoryID = types.StringValue("")
	}
	if w.Url != nil {
		model.URL = types.StringValue(*w.Url)
	} else {
		model.URL = types.StringValue("")
	}
	if w.Versions != nil && len(*w.Versions) > 0 {
		v := (*w.Versions)[0]
		if v.Version != nil {
			model.Version = types.StringValue(*v.Version)
		} else {
			model.Version = types.StringValue("")
		}
	} else {
		model.Version = types.StringValue("")
	}
}

// adoptExistingProjectWiki finds the existing project wiki for the given project and renames
// it to desiredName. ADO allows only one project wiki per project. When CreateWiki returns
// "already exists", we adopt the existing one and rename it so Terraform state matches config.
func (r *WikiResource) adoptExistingProjectWiki(projectID, desiredName string) (*azwiki.WikiV2, error) {
	wikis, err := r.client.WikiClient.GetAllWikis(r.client.Ctx, azwiki.GetAllWikisArgs{
		Project: converter.String(projectID),
	})
	if err != nil {
		return nil, fmt.Errorf("listing wikis for project %s: %w", projectID, err)
	}
	if wikis == nil {
		return nil, fmt.Errorf("no wikis found for project %s", projectID)
	}
	var found *azwiki.WikiV2
	for i := range *wikis {
		w := &(*wikis)[i]
		if w.Type != nil && *w.Type == azwiki.WikiTypeValues.ProjectWiki {
			found = w
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("no project wiki found in project %s", projectID)
	}
	// Rename the wiki to match the desired Terraform name so that state == config (idempotency).
	if found.Name == nil || *found.Name != desiredName {
		updated, updateErr := r.client.WikiClient.UpdateWiki(r.client.Ctx, azwiki.UpdateWikiArgs{
			WikiIdentifier: converter.String(found.Id.String()),
			UpdateParameters: &azwiki.WikiUpdateParameters{
				Name: converter.String(desiredName),
			},
		})
		if updateErr != nil {
			return nil, fmt.Errorf("renaming adopted project wiki: %w", updateErr)
		}
		return updated, nil
	}
	return found, nil
}
