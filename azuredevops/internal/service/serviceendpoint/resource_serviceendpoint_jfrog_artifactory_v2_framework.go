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
	_ resource.Resource              = (*ServiceEndpointJFrogArtifactoryV2Resource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointJFrogArtifactoryV2Resource)(nil)
)

// ── local plan modifier + default helpers ───────────────��─────────────────────

type seJFrogArtifactoryV2RequiresReplaceModifier struct{}

func seJFrogArtifactoryV2RequiresReplace() planmodifier.String {
	return seJFrogArtifactoryV2RequiresReplaceModifier{}
}

func (m seJFrogArtifactoryV2RequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seJFrogArtifactoryV2RequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seJFrogArtifactoryV2RequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seJFrogArtifactoryV2UseStateForUnknownModifier struct{}

func seJFrogArtifactoryV2UseStateForUnknown() planmodifier.String {
	return seJFrogArtifactoryV2UseStateForUnknownModifier{}
}

func (m seJFrogArtifactoryV2UseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seJFrogArtifactoryV2UseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seJFrogArtifactoryV2UseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seJFrogArtifactoryV2StringDefault struct{ value string }

func seJFrogArtifactoryV2DefaultString(v string) defaults.String {
	return seJFrogArtifactoryV2StringDefault{v}
}

func (d seJFrogArtifactoryV2StringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seJFrogArtifactoryV2StringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seJFrogArtifactoryV2StringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────���───────────────────────────────────────────

// ServiceEndpointJFrogArtifactoryV2Resource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_jfrog_artifactory_v2.
type ServiceEndpointJFrogArtifactoryV2Resource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointJFrogArtifactoryV2Resource returns a new resource.Resource.
func NewServiceEndpointJFrogArtifactoryV2Resource() resource.Resource {
	return &ServiceEndpointJFrogArtifactoryV2Resource{}
}

// ── Metadata ─────────────────────────────────────────��────────────────────────

func (r *ServiceEndpointJFrogArtifactoryV2Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_jfrog_artifactory_v2"
}

// ── Schema ──────────────────────────────────────────────────────���─────────────

func (r *ServiceEndpointJFrogArtifactoryV2Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JFrog Artifactory V2 Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seJFrogArtifactoryV2UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seJFrogArtifactoryV2RequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seJFrogArtifactoryV2DefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "Url for the JFrog Artifactory Server.",
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Specifies the authorization scheme and parameters.",
			},
		},
		Blocks: map[string]schema.Block{
			"authentication_token": schema.ListNestedBlock{
				Description: "Token authentication block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The JFrog Artifactory access token.",
						},
					},
				},
			},
			"authentication_basic": schema.ListNestedBlock{
				Description: "Username/password authentication block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The JFrog Artifactory user name.",
						},
						"password": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The JFrog Artifactory password.",
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────��─────────────────────────────���─────────────

func (r *ServiceEndpointJFrogArtifactoryV2Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── State models ────────────��─────────────────────────────────��───────────────

type serviceEndpointJFrogArtifactoryV2Model struct {
	ID                  types.String                     `tfsdk:"id"`
	ProjectID           types.String                     `tfsdk:"project_id"`
	ServiceEndpointName types.String                     `tfsdk:"service_endpoint_name"`
	Description         types.String                     `tfsdk:"description"`
	URL                 types.String                     `tfsdk:"url"`
	Authorization       types.Map                        `tfsdk:"authorization"`
	AuthenticationToken []seJFrogArtifactoryV2TokenModel `tfsdk:"authentication_token"`
	AuthenticationBasic []seJFrogArtifactoryV2BasicModel `tfsdk:"authentication_basic"`
}

type seJFrogArtifactoryV2TokenModel struct {
	Token types.String `tfsdk:"token"`
}

type seJFrogArtifactoryV2BasicModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// ── Create ──────────────────────��───────────────────────────────���─────────────

func (r *ServiceEndpointJFrogArtifactoryV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointJFrogArtifactoryV2Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, scheme, err := seJFrogArtifactoryV2BuildEndpoint(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building service endpoint", err.Error())
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint.ServiceEndpointProjectReferences = &[]serviceendpoint.ServiceEndpointProjectReference{
		{
			ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
			Name:             converter.String(name),
			Description:      converter.String(description),
		},
	}

	created, err := r.client.ServiceEndpointClient.CreateServiceEndpoint(ctx, serviceendpoint.CreateServiceEndpointArgs{
		Endpoint: endpoint,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating service endpoint", err.Error())
		return
	}

	plan.ID = types.StringValue(created.Id.String())
	plan.Authorization = seJFrogArtifactoryV2BuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ────────────────────���─────────────────────────────��───────────────────

func (r *ServiceEndpointJFrogArtifactoryV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointJFrogArtifactoryV2Model
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
		resp.Diagnostics.AddError("Error reading service endpoint", err.Error())
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
	if ep.Url != nil {
		state.URL = types.StringValue(*ep.Url)
	}

	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seJFrogArtifactoryV2BuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ─────────────────────────────────────────────────────��──────────────

func (r *ServiceEndpointJFrogArtifactoryV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointJFrogArtifactoryV2Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointJFrogArtifactoryV2Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint, scheme, err := seJFrogArtifactoryV2BuildEndpoint(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building service endpoint", err.Error())
		return
	}

	endpoint.Id = &endpointID
	endpoint.ServiceEndpointProjectReferences = &[]serviceendpoint.ServiceEndpointProjectReference{
		{
			ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
			Name:             converter.String(name),
			Description:      converter.String(description),
		},
	}

	_, err = r.client.ServiceEndpointClient.UpdateServiceEndpoint(ctx, serviceendpoint.UpdateServiceEndpointArgs{
		Endpoint:   endpoint,
		EndpointId: &endpointID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating service endpoint", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Authorization = seJFrogArtifactoryV2BuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────���───────────────────────────────────────────────

func (r *ServiceEndpointJFrogArtifactoryV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointJFrogArtifactoryV2Model
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
		resp.Diagnostics.AddError("Error deleting service endpoint", err.Error())
	}
}

// ── helpers ─────────────────────��──────────────────────────���──────────────────

func seJFrogArtifactoryV2BuildEndpoint(plan serviceEndpointJFrogArtifactoryV2Model) (*serviceendpoint.ServiceEndpoint, string, error) {
	ep := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(plan.ServiceEndpointName.ValueString()),
		Owner:       converter.String("library"),
		Description: converter.String(plan.Description.ValueString()),
		Type:        converter.String("jfrogArtifactoryService"),
		Url:         converter.String(plan.URL.ValueString()),
	}

	if len(plan.AuthenticationToken) > 0 {
		authParams := map[string]string{
			"apitoken": plan.AuthenticationToken[0].Token.ValueString(),
		}
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &authParams,
			Scheme:     converter.String("Token"),
		}
		return ep, "Token", nil
	}

	if len(plan.AuthenticationBasic) > 0 {
		authParams := map[string]string{
			"username": plan.AuthenticationBasic[0].Username.ValueString(),
			"password": plan.AuthenticationBasic[0].Password.ValueString(),
		}
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &authParams,
			Scheme:     converter.String("UsernamePassword"),
		}
		return ep, "UsernamePassword", nil
	}

	return nil, "", fmt.Errorf("one of authentication_token or authentication_basic must be specified")
}

func seJFrogArtifactoryV2BuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
