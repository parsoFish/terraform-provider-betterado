package wiki

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	azwiki "github.com/microsoft/azure-devops-go-api/azuredevops/v7/wiki"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

/*
To improve concurrent page api response:
"The wiki page has already been updated by another client, so you cannot update it. Please try again."
Add mutex lock to limit terraform provider concurrent create / update / delete request.
*/
var pageLock = sync.Mutex{}

// Compile-time interface checks.
var (
	_ resource.Resource              = (*WikiPageResource)(nil)
	_ resource.ResourceWithConfigure = (*WikiPageResource)(nil)
)

// WikiPageResource is the terraform-plugin-framework implementation of betterado_wiki_page.
type WikiPageResource struct {
	client *client.AggregatedClient
}

// NewWikiPageResource returns a new resource.Resource for betterado_wiki_page.
func NewWikiPageResource() resource.Resource {
	return &WikiPageResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *WikiPageResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_wiki_page"
}

// ── Schema validators ─────────────────────────────────────────────────────────

// wikiPageUUIDValidator validates that a string is a valid UUID.
type wikiPageUUIDValidator struct{}

func (v wikiPageUUIDValidator) Description(_ context.Context) string {
	return "value must be a valid UUID"
}

func (v wikiPageUUIDValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a valid UUID"
}

func (v wikiPageUUIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	// Simple UUID validation: 8-4-4-4-12 hex groups.
	parts := strings.Split(val, "-")
	if len(parts) != 5 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid UUID", fmt.Sprintf("%q is not a valid UUID", val))
		return
	}
	lens := []int{8, 4, 4, 4, 12}
	for i, p := range parts {
		if len(p) != lens[i] {
			resp.Diagnostics.AddAttributeError(req.Path, "Invalid UUID", fmt.Sprintf("%q is not a valid UUID", val))
			return
		}
		for _, c := range p {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				resp.Diagnostics.AddAttributeError(req.Path, "Invalid UUID", fmt.Sprintf("%q is not a valid UUID", val))
				return
			}
		}
	}
}

// wikiPageNotEmptyValidator validates that a string is not empty.
type wikiPageNotEmptyValidator struct{}

func (v wikiPageNotEmptyValidator) Description(_ context.Context) string {
	return "value must not be empty"
}

func (v wikiPageNotEmptyValidator) MarkdownDescription(_ context.Context) string {
	return "value must not be empty"
}

func (v wikiPageNotEmptyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Empty value", "value must not be empty")
	}
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *WikiPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					wikiUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					wikiPageUUIDValidator{},
				},
				PlanModifiers: []planmodifier.String{
					wikiRequiresReplace(),
				},
			},
			"wiki_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					wikiPageUUIDValidator{},
				},
				PlanModifiers: []planmodifier.String{
					wikiRequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					wikiPageNotEmptyValidator{},
				},
			},
			"content": schema.StringAttribute{
				Required: true,
			},
			"version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					wikiUseStateForUnknown(),
				},
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *WikiPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// wikiPageModel is the Terraform state model for betterado_wiki_page.
type wikiPageModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	WikiID    types.String `tfsdk:"wiki_id"`
	Path      types.String `tfsdk:"path"`
	Content   types.String `tfsdk:"content"`
	Version   types.String `tfsdk:"version"`
	ETag      types.String `tfsdk:"etag"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *WikiPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model wikiPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pageLock.Lock()
	defer pageLock.Unlock()

	createArgs := azwiki.CreateOrUpdatePageArgs{
		Parameters: &azwiki.WikiPageCreateOrUpdateParameters{
			Content: converter.String(model.Content.ValueString()),
		},
		Project:        converter.String(model.ProjectID.ValueString()),
		WikiIdentifier: converter.String(model.WikiID.ValueString()),
		Path:           converter.String(model.Path.ValueString()),
	}
	// For code wikis the ADO API requires a versionDescriptor (branch reference).
	// When the caller supplies a version, include it; otherwise omit and let ADO
	// use the default branch (project wikis do not require a descriptor).
	if !model.Version.IsNull() && !model.Version.IsUnknown() && model.Version.ValueString() != "" {
		vt := git.GitVersionTypeValues.Branch
		createArgs.VersionDescriptor = &git.GitVersionDescriptor{
			VersionType: &vt,
			Version:     converter.String(model.Version.ValueString()),
		}
	}

	result, err := r.client.WikiClient.CreateOrUpdatePage(r.client.Ctx, createArgs)
	if err != nil {
		resp.Diagnostics.AddError("Error creating wiki page", err.Error())
		return
	}

	model.ID = types.StringValue(strconv.Itoa(*result.Page.Id))
	r.flattenPage(result, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *WikiPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model wikiPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.WikiClient.GetPage(r.client.Ctx, azwiki.GetPageArgs{
		Project:        converter.String(model.ProjectID.ValueString()),
		WikiIdentifier: converter.String(model.WikiID.ValueString()),
		Path:           converter.String(model.Path.ValueString()),
		IncludeContent: converter.Bool(true),
	})
	if err != nil {
		// If page not found, remove from state.
		if strings.Contains(err.Error(), "WikiPageNotFoundException") ||
			strings.Contains(err.Error(), "does not exist") ||
			strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading wiki page", err.Error())
		return
	}

	r.flattenPage(result, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *WikiPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan wikiPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state wikiPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse wiki page ID", fmt.Sprintf("Parse wiki page ID: %s. Error: %+v", state.ID.ValueString(), err))
		return
	}

	pageLock.Lock()
	defer pageLock.Unlock()

	_, err = r.client.WikiClient.UpdatePageById(r.client.Ctx, azwiki.UpdatePageByIdArgs{
		Parameters: &azwiki.WikiPageCreateOrUpdateParameters{
			Content: converter.String(plan.Content.ValueString()),
		},
		Id:             &id,
		Project:        converter.String(state.ProjectID.ValueString()),
		WikiIdentifier: converter.String(state.WikiID.ValueString()),
		Version:        converter.String(state.ETag.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating wiki page", err.Error())
		return
	}

	// Read back the updated state.
	result, err := r.client.WikiClient.GetPage(r.client.Ctx, azwiki.GetPageArgs{
		Project:        converter.String(state.ProjectID.ValueString()),
		WikiIdentifier: converter.String(state.WikiID.ValueString()),
		Path:           converter.String(plan.Path.ValueString()),
		IncludeContent: converter.Bool(true),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading wiki page after update", err.Error())
		return
	}

	plan.ID = state.ID
	r.flattenPage(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WikiPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model wikiPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse wiki page ID", fmt.Sprintf("Parse wiki page ID: %s. Error: %+v", model.ID.ValueString(), err))
		return
	}

	pageLock.Lock()
	defer pageLock.Unlock()

	_, err = r.client.WikiClient.DeletePageById(r.client.Ctx, azwiki.DeletePageByIdArgs{
		Project:        converter.String(model.ProjectID.ValueString()),
		WikiIdentifier: converter.String(model.WikiID.ValueString()),
		Id:             &id,
	})
	if err != nil {
		if !strings.Contains(err.Error(), "WikiPageNotFoundException") &&
			!strings.Contains(err.Error(), "does not exist") &&
			!strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError("Error deleting wiki page", err.Error())
		}
	}
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// flattenPage maps an API wiki page response into the state model.
func (r *WikiPageResource) flattenPage(result *azwiki.WikiPageResponse, model *wikiPageModel) {
	if result.ETag != nil && len(*result.ETag) > 0 {
		etagValue := strings.Trim(strings.Join(*result.ETag, " "), `\"`)
		model.ETag = types.StringValue(etagValue)
	} else if model.ETag.IsNull() || model.ETag.IsUnknown() {
		model.ETag = types.StringValue("")
	}

	// Version is not returned by the API; preserve whatever the user supplied
	// (or empty string if not set) so state does not drift.
	if model.Version.IsNull() || model.Version.IsUnknown() {
		model.Version = types.StringValue("")
	}

	if result.Page != nil {
		if result.Page.Content != nil {
			model.Content = types.StringValue(*result.Page.Content)
		}
		if result.Page.Path != nil {
			model.Path = types.StringValue(*result.Page.Path)
		}
		if result.Page.Id != nil {
			model.ID = types.StringValue(strconv.Itoa(*result.Page.Id))
		}
	}
}
