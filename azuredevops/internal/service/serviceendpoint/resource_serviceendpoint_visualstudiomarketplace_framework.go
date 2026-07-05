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
	_ resource.Resource              = (*ServiceEndpointVisualStudioMarketplaceResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointVisualStudioMarketplaceResource)(nil)
)

type seMarketplaceRequiresReplaceModifier struct{}

func seMarketplaceRequiresReplace() planmodifier.String {
	return seMarketplaceRequiresReplaceModifier{}
}
func (m seMarketplaceRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seMarketplaceRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seMarketplaceRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seMarketplaceUseStateForUnknownModifier struct{}

func seMarketplaceUseStateForUnknown() planmodifier.String {
	return seMarketplaceUseStateForUnknownModifier{}
}
func (m seMarketplaceUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seMarketplaceUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seMarketplaceUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seMarketplaceStringDefault struct{ value string }

func seMarketplaceDefaultString(v string) defaults.String { return seMarketplaceStringDefault{v} }
func (d seMarketplaceStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seMarketplaceStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seMarketplaceStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointVisualStudioMarketplaceResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_visualstudiomarketplace.
type ServiceEndpointVisualStudioMarketplaceResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointVisualStudioMarketplaceResource returns a new resource.Resource.
func NewServiceEndpointVisualStudioMarketplaceResource() resource.Resource {
	return &ServiceEndpointVisualStudioMarketplaceResource{}
}

func (r *ServiceEndpointVisualStudioMarketplaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_visualstudiomarketplace"
}

func (r *ServiceEndpointVisualStudioMarketplaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Visual Studio Marketplace Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seMarketplaceUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seMarketplaceRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seMarketplaceDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The Visual Studio Marketplace URL.",
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
							Description: "The access token.",
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
							Description: "The username.",
						},
						"password": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The password.",
						},
					},
				},
			},
		},
	}
}

func (r *ServiceEndpointVisualStudioMarketplaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointVisualStudioMarketplaceModel struct {
	ID                  types.String                `tfsdk:"id"`
	ProjectID           types.String                `tfsdk:"project_id"`
	ServiceEndpointName types.String                `tfsdk:"service_endpoint_name"`
	Description         types.String                `tfsdk:"description"`
	URL                 types.String                `tfsdk:"url"`
	Authorization       types.Map                   `tfsdk:"authorization"`
	AuthenticationToken []seMarketplaceTokenModel   `tfsdk:"authentication_token"`
	AuthenticationBasic []seMarketplaceBasicModel   `tfsdk:"authentication_basic"`
}

type seMarketplaceTokenModel struct {
	Token types.String `tfsdk:"token"`
}

type seMarketplaceBasicModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r *ServiceEndpointVisualStudioMarketplaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointVisualStudioMarketplaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, scheme, err := seMarketplaceBuildEndpoint(plan)
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
	plan.Authorization = seMarketplaceBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointVisualStudioMarketplaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointVisualStudioMarketplaceModel
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
	state.Authorization = seMarketplaceBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServiceEndpointVisualStudioMarketplaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointVisualStudioMarketplaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointVisualStudioMarketplaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint, scheme, err := seMarketplaceBuildEndpoint(plan)
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
	plan.Authorization = seMarketplaceBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointVisualStudioMarketplaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointVisualStudioMarketplaceModel
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

// ── helpers ───────────────────────────────────────────────────────────────────

func seMarketplaceBuildEndpoint(plan serviceEndpointVisualStudioMarketplaceModel) (*serviceendpoint.ServiceEndpoint, string, error) {
	ep := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(plan.ServiceEndpointName.ValueString()),
		Owner:       converter.String("library"),
		Description: converter.String(plan.Description.ValueString()),
		Type:        converter.String("TFSMarketplacePublishing"),
		Url:         converter.String(plan.URL.ValueString()),
	}

	if len(plan.AuthenticationToken) > 0 {
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"apitoken": plan.AuthenticationToken[0].Token.ValueString(),
			},
			Scheme: converter.String("Token"),
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

func seMarketplaceBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
