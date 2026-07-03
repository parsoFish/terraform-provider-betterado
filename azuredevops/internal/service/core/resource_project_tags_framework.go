package core

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*ProjectTagsResource)(nil)
	_ resource.ResourceWithConfigure   = (*ProjectTagsResource)(nil)
	_ resource.ResourceWithImportState = (*ProjectTagsResource)(nil)
)

// ProjectTagsResource is the terraform-plugin-framework implementation of
// betterado_project_tags.
type ProjectTagsResource struct {
	client *client.AggregatedClient
}

// NewProjectTagsResource returns a new resource.Resource.
func NewProjectTagsResource() resource.Resource {
	return &ProjectTagsResource{}
}

// projectTagsModel is the Terraform state model for betterado_project_tags.
type projectTagsModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Tags      types.Set    `tfsdk:"tags"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ProjectTagsResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_project_tags"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ProjectTagsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages tags on an Azure DevOps project.",
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
			"tags": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Set of tag names to assign to the project.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ProjectTagsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectTagsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model projectTagsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, err := uuid.Parse(model.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("parsing project_id", err.Error())
		return
	}

	tags := r.expandTags(ctx, model.Tags)
	if err := r.client.CoreClient.SetProjectProperties(r.client.Ctx, adocore.SetProjectPropertiesArgs{
		PatchDocument: tags,
		ProjectId:     &projectID,
	}); err != nil {
		resp.Diagnostics.AddError("creating project tags", err.Error())
		return
	}

	model.ID = types.StringValue(projectID.String())
	if err := r.readTags(ctx, &model); err != nil {
		resp.Diagnostics.AddError("reading project tags after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectTagsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model projectTagsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readTags(ctx, &model); err != nil {
		resp.Diagnostics.AddError("reading project tags", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectTagsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan projectTagsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	projectID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("parsing project id", err.Error())
		return
	}

	// Compute add/remove sets.
	existingTags := r.flattenTagsFromSet(ctx, state.Tags)
	desiredTags := r.flattenTagsFromSet(ctx, plan.Tags)

	tagsToAdd := difference(desiredTags, existingTags)
	tagsToRemove := difference(existingTags, desiredTags)

	ops := make([]webapi.JsonPatchOperation, 0, len(tagsToAdd)+len(tagsToRemove))
	for _, tag := range tagsToRemove {
		ops = append(ops, webapi.JsonPatchOperation{
			From:  converter.String(""),
			Op:    converter.ToPtr(webapi.OperationValues.Remove),
			Path:  converter.String(fmt.Sprintf("/Microsoft.TeamFoundation.Project.Tag.%s", tag)),
			Value: nil,
		})
	}
	for _, tag := range tagsToAdd {
		ops = append(ops, webapi.JsonPatchOperation{
			From:  converter.String(""),
			Op:    converter.ToPtr(webapi.OperationValues.Add),
			Path:  converter.String(fmt.Sprintf("/Microsoft.TeamFoundation.Project.Tag.%s", tag)),
			Value: converter.String("true"),
		})
	}

	if len(ops) > 0 {
		if err := r.client.CoreClient.SetProjectProperties(r.client.Ctx, adocore.SetProjectPropertiesArgs{
			PatchDocument: &ops,
			ProjectId:     &projectID,
		}); err != nil {
			resp.Diagnostics.AddError("updating project tags", err.Error())
			return
		}
	}

	if err := r.readTags(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("reading project tags after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectTagsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model projectTagsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("parsing project id", err.Error())
		return
	}

	tags := r.expandTagsForRemove(ctx, model.Tags)
	if err := r.client.CoreClient.SetProjectProperties(r.client.Ctx, adocore.SetProjectPropertiesArgs{
		PatchDocument: tags,
		ProjectId:     &projectID,
	}); err != nil {
		resp.Diagnostics.AddError("deleting project tags", err.Error())
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *ProjectTagsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	model := projectTagsModel{
		ID:        types.StringValue(req.ID),
		ProjectID: types.StringValue(req.ID),
	}
	if err := r.readTags(ctx, &model); err != nil {
		resp.Diagnostics.AddError("importing project tags", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (r *ProjectTagsResource) expandTags(ctx context.Context, s types.Set) *[]webapi.JsonPatchOperation {
	var tagNames []string
	s.ElementsAs(ctx, &tagNames, false)
	ops := make([]webapi.JsonPatchOperation, 0, len(tagNames))
	for _, tag := range tagNames {
		ops = append(ops, webapi.JsonPatchOperation{
			From:  converter.String(""),
			Op:    converter.ToPtr(webapi.OperationValues.Add),
			Path:  converter.String(fmt.Sprintf("/Microsoft.TeamFoundation.Project.Tag.%s", tag)),
			Value: converter.String("true"),
		})
	}
	return &ops
}

func (r *ProjectTagsResource) expandTagsForRemove(ctx context.Context, s types.Set) *[]webapi.JsonPatchOperation {
	var tagNames []string
	s.ElementsAs(ctx, &tagNames, false)
	ops := make([]webapi.JsonPatchOperation, 0, len(tagNames))
	for _, tag := range tagNames {
		ops = append(ops, webapi.JsonPatchOperation{
			From:  converter.String(""),
			Op:    converter.ToPtr(webapi.OperationValues.Remove),
			Path:  converter.String(fmt.Sprintf("/Microsoft.TeamFoundation.Project.Tag.%s", tag)),
			Value: nil,
		})
	}
	return &ops
}

func (r *ProjectTagsResource) flattenTagsFromSet(ctx context.Context, s types.Set) []string {
	var tagNames []string
	s.ElementsAs(ctx, &tagNames, false)
	return tagNames
}

func (r *ProjectTagsResource) readTags(_ context.Context, model *projectTagsModel) error {
	projectID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		return err
	}

	props, err := r.client.CoreClient.GetProjectProperties(r.client.Ctx, adocore.GetProjectPropertiesArgs{
		ProjectId: &projectID,
		Keys:      &[]string{"Microsoft.TeamFoundation.Project.Tag.*"},
	})
	if err != nil {
		return err
	}

	tags := make([]string, 0, len(*props))
	if props != nil {
		for _, p := range *props {
			if p.Name != nil && len(*p.Name) > 37 {
				tags = append(tags, (*p.Name)[37:])
			}
		}
	}

	tagsVal, diags := types.SetValueFrom(context.Background(), types.StringType, tags)
	if diags.HasError() {
		return fmt.Errorf("building tags set: %v", diags)
	}
	model.Tags = tagsVal
	model.ProjectID = types.StringValue(projectID.String())
	return nil
}

// difference returns elements in b that are not in a.
func difference(b, a []string) []string {
	aMap := make(map[string]struct{}, len(a))
	for _, v := range a {
		aMap[v] = struct{}{}
	}
	var result []string
	for _, v := range b {
		if _, ok := aMap[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}
