package pipelines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	adoBuild "github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	adoPipelines "github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// ── Compile-time interface checks ─────────────────────────────────────────────

var (
	_ resource.Resource              = (*PipelineResource)(nil)
	_ resource.ResourceWithConfigure = (*PipelineResource)(nil)
)

// ── Static default helpers ────────────────────────────────────────────────────

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

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// requiresReplaceModifier forces resource replacement when value changes.
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

// useStateForUnknownModifier copies prior state when plan value is unknown.
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

// PipelineResource is the terraform-plugin-framework implementation of betterado_pipeline.
type PipelineResource struct {
	client *client.AggregatedClient
}

// NewPipelineResource returns a new resource.Resource for betterado_pipeline.
func NewPipelineResource() resource.Resource {
	return &PipelineResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *PipelineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *PipelineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Pipelines v2 pipeline (`_apis/pipelines`) in Azure DevOps. " +
			"For classic build pipelines managed via the Build API, use `betterado_build_definition`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The pipeline integer ID as a string (set as the resource ID).",
				PlanModifiers:       []planmodifier.String{useStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the project in which the pipeline lives. Changing this forces a new resource.",
				PlanModifiers:       []planmodifier.String{requiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the pipeline.",
			},
			"folder": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The folder path of the pipeline. Defaults to `\\` (root folder).",
				Default:             defaultString(`\`),
			},
			"configuration_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The configuration type. One of: `yaml`, `designerJson`, `justInTime`, `designerHyphenJson`, `unknown`. Defaults to `yaml`.",
				Default:             defaultString("yaml"),
			},
			"repo_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID (GUID) of the Azure Repos Git repository that contains the YAML pipeline file. Required when `configuration_type` is `yaml`.",
				PlanModifiers:       []planmodifier.String{requiresReplace()},
			},
			"yaml_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The path of the YAML pipeline definition file within the repository (e.g. `/azure-pipelines.yml`). Required when `configuration_type` is `yaml`.",
				PlanModifiers:       []planmodifier.String{requiresReplace()},
			},
			"revision": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The pipeline revision number (optimistic concurrency token).",
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL of the pipeline.",
				PlanModifiers:       []planmodifier.String{useStateForUnknown()},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *PipelineResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// pipelineModel is the Terraform state model for betterado_pipeline.
type pipelineModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.String `tfsdk:"project_id"`
	Name              types.String `tfsdk:"name"`
	Folder            types.String `tfsdk:"folder"`
	ConfigurationType types.String `tfsdk:"configuration_type"`
	RepoID            types.String `tfsdk:"repo_id"`
	YAMLPath          types.String `tfsdk:"yaml_path"`
	Revision          types.Int64  `tfsdk:"revision"`
	URL               types.String `tfsdk:"url"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *PipelineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model pipelineModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := model.ProjectID.ValueString()
	pipeline, err := createPipelineRaw(ctx, r.client, project, model)
	if err != nil {
		resp.Diagnostics.AddError("Error creating pipeline", err.Error())
		return
	}

	flattenPipeline(pipeline, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model pipelineModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := model.ProjectID.ValueString()
	pipelineID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid pipeline ID", err.Error())
		return
	}

	pipeline, err := r.client.PipelinesClient.GetPipeline(ctx, adoPipelines.GetPipelineArgs{
		Project:    &project,
		PipelineId: &pipelineID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading pipeline", err.Error())
		return
	}

	flattenPipeline(pipeline, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pipelineModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state pipelineModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := plan.ProjectID.ValueString()
	pipelineID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid pipeline ID", err.Error())
		return
	}

	// Re-read to get current revision for optimistic concurrency.
	current, err := r.client.PipelinesClient.GetPipeline(ctx, adoPipelines.GetPipelineArgs{
		Project:    &project,
		PipelineId: &pipelineID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading pipeline before update", err.Error())
		return
	}

	name := plan.Name.ValueString()
	folder := plan.Folder.ValueString()
	patchBody := map[string]interface{}{
		"name":     name,
		"folder":   folder,
		"revision": current.Revision,
	}

	updated, err := patchPipeline(ctx, r.client, project, pipelineID, patchBody)
	if err != nil {
		// Optimistic concurrency failure — retry once with a fresh read.
		currentRetry, rerr := r.client.PipelinesClient.GetPipeline(ctx, adoPipelines.GetPipelineArgs{
			Project:    &project,
			PipelineId: &pipelineID,
		})
		if rerr != nil {
			resp.Diagnostics.AddError("Error re-reading pipeline for retry", rerr.Error())
			return
		}
		patchBody["revision"] = currentRetry.Revision
		updated, err = patchPipeline(ctx, r.client, project, pipelineID, patchBody)
		if err != nil {
			resp.Diagnostics.AddError("Error updating pipeline", err.Error())
			return
		}
	}

	flattenPipeline(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PipelineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model pipelineModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := model.ProjectID.ValueString()
	pipelineID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid pipeline ID", err.Error())
		return
	}

	err = deletePipeline(ctx, r.client, project, pipelineID)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting pipeline", err.Error())
		return
	}
}

// ── Expand / Flatten helpers ──────────────────────────────────────────────────

// createPipelineRawBody is the extended JSON body for POST _apis/pipelines.
// The SDK's CreatePipelineConfigurationParameters only has Type; for YAML
// pipelines the API also requires Path and Repository.
type createPipelineRawBody struct {
	Name          string                      `json:"name"`
	Folder        string                      `json:"folder"`
	Configuration createPipelineRawConfigBody `json:"configuration"`
}

type createPipelineRawConfigBody struct {
	Type       string                     `json:"type"`
	Path       string                     `json:"path,omitempty"`
	Repository *createPipelineRawRepoBody `json:"repository,omitempty"`
}

type createPipelineRawRepoBody struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// createPipelineRaw POSTs the full pipeline create body via raw HTTP so we can
// include the YAML path and repository fields that the SDK omits.
func createPipelineRaw(ctx context.Context, agg *client.AggregatedClient, project string, model pipelineModel) (*adoPipelines.Pipeline, error) {
	impl, ok := agg.PipelinesClient.(*adoPipelines.ClientImpl)
	if !ok {
		return nil, fmt.Errorf("createPipelineRaw: cannot access underlying azuredevops.Client")
	}

	body := createPipelineRawBody{
		Name:   model.Name.ValueString(),
		Folder: model.Folder.ValueString(),
		Configuration: createPipelineRawConfigBody{
			Type: model.ConfigurationType.ValueString(),
		},
	}

	if !model.RepoID.IsNull() && !model.RepoID.IsUnknown() && model.RepoID.ValueString() != "" {
		body.Configuration.Repository = &createPipelineRawRepoBody{
			ID:   model.RepoID.ValueString(),
			Type: "azureReposGit",
		}
	}
	if !model.YAMLPath.IsNull() && !model.YAMLPath.IsUnknown() && model.YAMLPath.ValueString() != "" {
		body.Configuration.Path = model.YAMLPath.ValueString()
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	routeValues := map[string]string{"project": project}
	resp, err := impl.Client.Send(ctx, http.MethodPost, pipelineLocationID, "7.1-preview.1", routeValues, nil, bytes.NewReader(bodyBytes), "application/json", "application/json", nil)
	if err != nil {
		return nil, err
	}

	var result adoPipelines.Pipeline
	if err := impl.Client.UnmarshalBody(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// flattenPipeline writes the API response into the Terraform state model.
func flattenPipeline(p *adoPipelines.Pipeline, model *pipelineModel) {
	if p.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*p.Id))
	}
	if p.Name != nil {
		model.Name = types.StringValue(*p.Name)
	}
	if p.Folder != nil {
		model.Folder = types.StringValue(*p.Folder)
	}
	if p.Revision != nil {
		model.Revision = types.Int64Value(int64(*p.Revision))
	}
	if p.Url != nil {
		model.URL = types.StringValue(*p.Url)
	}
	if p.Configuration != nil && p.Configuration.Type != nil {
		model.ConfigurationType = types.StringValue(string(*p.Configuration.Type))
	}
}

// ── REST helpers for Update/Delete (not in SDK v7 client interface) ───────────

// pipelineLocationID is the resource location ID for the Pipelines v2 endpoint.
var pipelineLocationID = uuid.MustParse("28e1305e-2afe-47bf-abaf-cbb0e6a91988")

// patchPipeline performs a PATCH on _apis/pipelines/{pipelineId}.
func patchPipeline(ctx context.Context, agg *client.AggregatedClient, project string, pipelineID int, body map[string]interface{}) (*adoPipelines.Pipeline, error) {
	impl, ok := agg.PipelinesClient.(*adoPipelines.ClientImpl)
	if !ok {
		return nil, fmt.Errorf("patchPipeline: cannot access underlying azuredevops.Client")
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	routeValues := map[string]string{
		"project":    project,
		"pipelineId": strconv.Itoa(pipelineID),
	}

	resp, err := impl.Client.Send(ctx, http.MethodPatch, pipelineLocationID, "7.1-preview.1", routeValues, nil, bytes.NewReader(bodyBytes), "application/json", "application/json", nil)
	if err != nil {
		return nil, err
	}

	var result adoPipelines.Pipeline
	if err := impl.Client.UnmarshalBody(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// deletePipeline removes the pipeline by deleting its backing build definition.
//
// The Pipelines v2 REST API (_apis/pipelines) does NOT support HTTP DELETE —
// the endpoint returns "The requested resource does not support http method
// 'DELETE'". Every YAML pipeline created via _apis/pipelines is backed by a
// Build Definition with the same integer ID, so we use the Build Definitions
// API (_apis/build/definitions/{id}) which does support DELETE.
func deletePipeline(ctx context.Context, agg *client.AggregatedClient, project string, pipelineID int) error {
	return agg.BuildClient.DeleteDefinition(ctx, adoBuild.DeleteDefinitionArgs{
		Project:      &project,
		DefinitionId: &pipelineID,
	})
}
