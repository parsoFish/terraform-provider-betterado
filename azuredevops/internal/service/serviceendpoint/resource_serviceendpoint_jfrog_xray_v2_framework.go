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
	_ resource.Resource              = (*ServiceEndpointJFrogXRayV2Resource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointJFrogXRayV2Resource)(nil)
)

type seJFrogXRayV2RequiresReplaceModifier struct{}

func seJFrogXRayV2RequiresReplace() planmodifier.String {
	return seJFrogXRayV2RequiresReplaceModifier{}
}

func (m seJFrogXRayV2RequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seJFrogXRayV2RequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seJFrogXRayV2RequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seJFrogXRayV2UseStateForUnknownModifier struct{}

func seJFrogXRayV2UseStateForUnknown() planmodifier.String {
	return seJFrogXRayV2UseStateForUnknownModifier{}
}

func (m seJFrogXRayV2UseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seJFrogXRayV2UseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seJFrogXRayV2UseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seJFrogXRayV2StringDefault struct{ value string }

func seJFrogXRayV2DefaultString(v string) defaults.String {
	return seJFrogXRayV2StringDefault{v}
}

func (d seJFrogXRayV2StringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seJFrogXRayV2StringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seJFrogXRayV2StringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ServiceEndpointJFrogXRayV2Resource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_jfrog_xray_v2.
type ServiceEndpointJFrogXRayV2Resource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointJFrogXRayV2Resource returns a new resource.Resource.
func NewServiceEndpointJFrogXRayV2Resource() resource.Resource {
	return &ServiceEndpointJFrogXRayV2Resource{}
}

func (r *ServiceEndpointJFrogXRayV2Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_jfrog_xray_v2"
}

func (r *ServiceEndpointJFrogXRayV2Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JFrog XRay V2 Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seJFrogXRayV2UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seJFrogXRayV2RequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seJFrogXRayV2DefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "Url for the JFrog XRay Server.",
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
							Description: "The JFrog XRay access token.",
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
							Description: "The JFrog XRay user name.",
						},
						"password": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The JFrog XRay password.",
						},
					},
				},
			},
		},
	}
}

func (r *ServiceEndpointJFrogXRayV2Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointJFrogXRayV2Model struct {
	ID                  types.String                   `tfsdk:"id"`
	ProjectID           types.String                   `tfsdk:"project_id"`
	ServiceEndpointName types.String                   `tfsdk:"service_endpoint_name"`
	Description         types.String                   `tfsdk:"description"`
	URL                 types.String                   `tfsdk:"url"`
	Authorization       types.Map                      `tfsdk:"authorization"`
	AuthenticationToken []seJFrogXRayV2TokenModel      `tfsdk:"authentication_token"`
	AuthenticationBasic []seJFrogXRayV2BasicModel      `tfsdk:"authentication_basic"`
}

type seJFrogXRayV2TokenModel struct {
	Token types.String `tfsdk:"token"`
}

type seJFrogXRayV2BasicModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r *ServiceEndpointJFrogXRayV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointJFrogXRayV2Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, scheme, err := seJFrogXRayV2BuildEndpoint(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building service endpoint", err.Error())
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	endpoint.ServiceEndpointProjectReferences = &[]serviceendpoint.ServiceEndpointProjectReference{
		{
			ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
			Name:             converter.String(plan.ServiceEndpointName.ValueString()),
			Description:      converter.String(plan.Description.ValueString()),
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
	plan.Authorization = seJFrogXRayV2BuildAuthMap(ctx, scheme)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointJFrogXRayV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointJFrogXRayV2Model
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
	state.Authorization = seJFrogXRayV2BuildAuthMap(ctx, scheme)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServiceEndpointJFrogXRayV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointJFrogXRayV2Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointJFrogXRayV2Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())

	endpoint, scheme, err := seJFrogXRayV2BuildEndpoint(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building service endpoint", err.Error())
		return
	}

	endpoint.Id = &endpointID
	endpoint.ServiceEndpointProjectReferences = &[]serviceendpoint.ServiceEndpointProjectReference{
		{
			ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
			Name:             converter.String(plan.ServiceEndpointName.ValueString()),
			Description:      converter.String(plan.Description.ValueString()),
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
	plan.Authorization = seJFrogXRayV2BuildAuthMap(ctx, scheme)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointJFrogXRayV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointJFrogXRayV2Model
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

func seJFrogXRayV2BuildEndpoint(plan serviceEndpointJFrogXRayV2Model) (*serviceendpoint.ServiceEndpoint, string, error) {
	ep := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(plan.ServiceEndpointName.ValueString()),
		Owner:       converter.String("library"),
		Description: converter.String(plan.Description.ValueString()),
		Type:        converter.String("jfrogXrayService"),
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

func seJFrogXRayV2BuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
