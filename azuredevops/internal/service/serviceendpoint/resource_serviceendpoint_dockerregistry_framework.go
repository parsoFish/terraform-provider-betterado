package serviceendpoint

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*ServiceEndpointDockerRegistryResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointDockerRegistryResource)(nil)
)

// ── plan modifier helpers ─────────────────────────────────────────────────────

type seDockerRegistryRequiresReplaceModifier struct{}

func seDockerRegistryRequiresReplace() planmodifier.String {
	return seDockerRegistryRequiresReplaceModifier{}
}

func (m seDockerRegistryRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seDockerRegistryRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seDockerRegistryRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seDockerRegistryUseStateForUnknownModifier struct{}

func seDockerRegistryUseStateForUnknown() planmodifier.String {
	return seDockerRegistryUseStateForUnknownModifier{}
}

func (m seDockerRegistryUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seDockerRegistryUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seDockerRegistryUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── default helpers ───────────────────────────────────────────────────────────

type seDockerRegistryStringDefault struct{ value string }

func seDockerRegistryDefaultString(v string) defaults.String { return seDockerRegistryStringDefault{v} }

func (d seDockerRegistryStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seDockerRegistryStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seDockerRegistryStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointDockerRegistryResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_dockerregistry.
type ServiceEndpointDockerRegistryResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointDockerRegistryResource returns a new resource.Resource.
func NewServiceEndpointDockerRegistryResource() resource.Resource {
	return &ServiceEndpointDockerRegistryResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDockerRegistryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_dockerregistry"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDockerRegistryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Docker Registry Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seDockerRegistryUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seDockerRegistryRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seDockerRegistryDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"docker_registry": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seDockerRegistryDefaultString("https://index.docker.io/v1/"),
				Description: "The DockerRegistry registry which should be used.",
			},
			"docker_username": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seDockerRegistryDefaultString(""),
				Description: "The DockerRegistry username which should be used.",
			},
			"docker_password": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Default:     seDockerRegistryDefaultString(""),
				Description: "The DockerRegistry password which should be used.",
			},
			"docker_email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seDockerRegistryDefaultString(""),
				Description: "The DockerRegistry email address which should be used.",
			},
			"registry_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seDockerRegistryDefaultString("DockerHub"),
				Description: "The registry type. Either 'DockerHub' or 'Others'.",
				PlanModifiers: []planmodifier.String{
					seDockerRegistryRequiresReplace(),
				},
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Specifies the authorization scheme and parameters.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDockerRegistryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointDockerRegistryModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	ServiceEndpointName types.String `tfsdk:"service_endpoint_name"`
	Description         types.String `tfsdk:"description"`
	DockerRegistry      types.String `tfsdk:"docker_registry"`
	DockerUsername      types.String `tfsdk:"docker_username"`
	DockerPassword      types.String `tfsdk:"docker_password"`
	DockerEmail         types.String `tfsdk:"docker_email"`
	RegistryType        types.String `tfsdk:"registry_type"`
	Authorization       types.Map    `tfsdk:"authorization"`
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDockerRegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointDockerRegistryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint := seDockerRegistryBuildEndpoint(&plan, &projectID, name, description)

	created, err := r.client.ServiceEndpointClient.CreateServiceEndpoint(ctx, serviceendpoint.CreateServiceEndpointArgs{
		Endpoint: endpoint,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating docker registry service endpoint", err.Error())
		return
	}

	plan.ID = types.StringValue(created.Id.String())
	plan.Authorization = seDockerRegistryBuildAuthMap(ctx, "UsernamePassword")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDockerRegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointDockerRegistryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := state.ProjectID.ValueString()

	ep, err := r.client.ServiceEndpointClient.GetServiceEndpointDetails(ctx, serviceendpoint.GetServiceEndpointDetailsArgs{
		EndpointId: &endpointID,
		Project:    &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading docker registry service endpoint", err.Error())
		return
	}
	if ep == nil || ep.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if ep.Name != nil {
		state.ServiceEndpointName = types.StringValue(*ep.Name)
	}
	if ep.Description != nil {
		state.Description = types.StringValue(*ep.Description)
	}
	if ep.Authorization != nil && ep.Authorization.Parameters != nil {
		params := *ep.Authorization.Parameters
		if v, ok := params["registry"]; ok {
			state.DockerRegistry = types.StringValue(v)
		}
		if v, ok := params["email"]; ok {
			state.DockerEmail = types.StringValue(v)
		}
		if v, ok := params["username"]; ok {
			state.DockerUsername = types.StringValue(v)
		}
		// NOTE: API never returns password; preserve state value to avoid spurious diffs.
	}
	if ep.Data != nil {
		if v, ok := (*ep.Data)["registrytype"]; ok {
			state.RegistryType = types.StringValue(v)
		}
	}

	scheme := "UsernamePassword"
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seDockerRegistryBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDockerRegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointDockerRegistryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointDockerRegistryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint := seDockerRegistryBuildEndpoint(&plan, &projectID, name, description)
	endpoint.Id = &endpointID

	_, err := r.client.ServiceEndpointClient.UpdateServiceEndpoint(ctx, serviceendpoint.UpdateServiceEndpointArgs{
		Endpoint:   endpoint,
		EndpointId: &endpointID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating docker registry service endpoint", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Authorization = seDockerRegistryBuildAuthMap(ctx, "UsernamePassword")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDockerRegistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointDockerRegistryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := state.ProjectID.ValueString()
	projectIDs := []string{projectID}

	if err := r.client.ServiceEndpointClient.DeleteServiceEndpoint(ctx, serviceendpoint.DeleteServiceEndpointArgs{
		EndpointId: &endpointID,
		ProjectIds: &projectIDs,
	}); err != nil {
		resp.Diagnostics.AddError("Error deleting docker registry service endpoint", err.Error())
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func seDockerRegistryBuildEndpoint(plan *serviceEndpointDockerRegistryModel, projectID *uuid.UUID, name, description string) *serviceendpoint.ServiceEndpoint {
	return &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("dockerregistry"),
		Url:         converter.String("https://hub.docker.com/"),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"registry": plan.DockerRegistry.ValueString(),
				"username": plan.DockerUsername.ValueString(),
				"password": plan.DockerPassword.ValueString(),
				"email":    plan.DockerEmail.ValueString(),
			},
			Scheme: converter.String("UsernamePassword"),
		},
		Data: &map[string]string{
			"registrytype": plan.RegistryType.ValueString(),
		},
		ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
			{
				ProjectReference: &serviceendpoint.ProjectReference{Id: projectID},
				Name:             converter.String(name),
				Description:      converter.String(description),
			},
		},
	}
}

// seDockerRegistryBuildAuthMap constructs the authorization computed map attribute.
func seDockerRegistryBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
