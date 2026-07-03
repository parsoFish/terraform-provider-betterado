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

var (
	_ resource.Resource              = (*ServiceEndpointGcpTerraformResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointGcpTerraformResource)(nil)
)

// ── plan modifier / default helpers ──────────────────────────────────────────

type seGcpRequiresReplaceModifier struct{}

func seGcpRequiresReplace() planmodifier.String { return seGcpRequiresReplaceModifier{} }
func (m seGcpRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seGcpRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seGcpRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seGcpUseStateForUnknownModifier struct{}

func seGcpUseStateForUnknown() planmodifier.String { return seGcpUseStateForUnknownModifier{} }
func (m seGcpUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seGcpUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seGcpUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seGcpStringDefault struct{ value string }

func seGcpDefaultString(v string) defaults.String { return seGcpStringDefault{v} }
func (d seGcpStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seGcpStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seGcpStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// NewServiceEndpointGcpTerraformResource is the constructor for the framework provider.
func NewServiceEndpointGcpTerraformResource() resource.Resource {
	return &ServiceEndpointGcpTerraformResource{}
}

// ServiceEndpointGcpTerraformResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_gcp_terraform.
type ServiceEndpointGcpTerraformResource struct {
	client *client.AggregatedClient
}

type serviceEndpointGcpTerraformModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	ServiceEndpointName types.String `tfsdk:"service_endpoint_name"`
	Description         types.String `tfsdk:"description"`
	Authorization       types.Map    `tfsdk:"authorization"`
	PrivateKey          types.String `tfsdk:"private_key"`
	TokenURI            types.String `tfsdk:"token_uri"`
	GcpProjectID        types.String `tfsdk:"gcp_project_id"`
	ClientEmail         types.String `tfsdk:"client_email"`
	Scope               types.String `tfsdk:"scope"`
}

func (r *ServiceEndpointGcpTerraformResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_gcp_terraform"
}

func (r *ServiceEndpointGcpTerraformResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a GCP Terraform service endpoint within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					seGcpUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					seGcpRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  seGcpDefaultString("Managed by Terraform"),
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"private_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Private Key for connecting to the endpoint.",
			},
			"token_uri": schema.StringAttribute{
				Required:    true,
				Description: "The token uri field in the JSON key file for creating the JSON Web Token.",
			},
			"gcp_project_id": schema.StringAttribute{
				Required:    true,
				Description: "GCP project ID.",
			},
			"client_email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seGcpDefaultString(""),
				Description: "The client email field in the JSON key file for creating the JSON Web Token.",
			},
			"scope": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seGcpDefaultString(""),
				Description: "Scope to be provided.",
			},
		},
	}
}

func (r *ServiceEndpointGcpTerraformResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ServiceEndpointGcpTerraformResource) buildEndpoint(plan *serviceEndpointGcpTerraformModel) *serviceendpoint.ServiceEndpoint {
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()
	return &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("GoogleCloudServiceEndpoint"),
		Url:         converter.String("https://www.googleapis.com/"),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"Issuer":     plan.ClientEmail.ValueString(),
				"Audience":   plan.TokenURI.ValueString(),
				"Scope":      plan.Scope.ValueString(),
				"PrivateKey": plan.PrivateKey.ValueString(),
			},
			Scheme: converter.String("JWT"),
		},
		Data: &map[string]string{
			"project": plan.GcpProjectID.ValueString(),
		},
		ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
			{
				ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
				Name:             converter.String(name),
				Description:      converter.String(description),
			},
		},
	}
}

func (r *ServiceEndpointGcpTerraformResource) flattenEndpoint(ctx context.Context, ep *serviceendpoint.ServiceEndpoint, state *serviceEndpointGcpTerraformModel) {
	if ep.Name != nil {
		state.ServiceEndpointName = types.StringValue(*ep.Name)
	}
	if ep.Description != nil {
		state.Description = types.StringValue(*ep.Description)
	}
	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	state.Authorization = m

	if ep.Authorization != nil && ep.Authorization.Parameters != nil {
		params := *ep.Authorization.Parameters
		if v, ok := params["Issuer"]; ok {
			state.ClientEmail = types.StringValue(v)
		}
		if v, ok := params["Audience"]; ok {
			state.TokenURI = types.StringValue(v)
		}
		if v, ok := params["Scope"]; ok {
			state.Scope = types.StringValue(v)
		}
		// PrivateKey is sensitive — never returned by API; preserve state
	}
	if ep.Data != nil {
		if v, ok := (*ep.Data)["project"]; ok {
			state.GcpProjectID = types.StringValue(v)
		}
	}
}

func (r *ServiceEndpointGcpTerraformResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointGcpTerraformModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ep := r.buildEndpoint(&plan)
	created, err := r.client.ServiceEndpointClient.CreateServiceEndpoint(ctx, serviceendpoint.CreateServiceEndpointArgs{Endpoint: ep})
	if err != nil {
		resp.Diagnostics.AddError("Error creating service endpoint", err.Error())
		return
	}
	plan.ID = types.StringValue(created.Id.String())
	r.flattenEndpoint(ctx, created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointGcpTerraformResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointGcpTerraformModel
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
	r.flattenEndpoint(ctx, ep, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServiceEndpointGcpTerraformResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointGcpTerraformModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointGcpTerraformModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpointID := uuid.MustParse(state.ID.ValueString())
	ep := r.buildEndpoint(&plan)
	ep.Id = &endpointID
	updated, err := r.client.ServiceEndpointClient.UpdateServiceEndpoint(ctx, serviceendpoint.UpdateServiceEndpointArgs{
		Endpoint:   ep,
		EndpointId: &endpointID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating service endpoint", err.Error())
		return
	}
	plan.ID = state.ID
	r.flattenEndpoint(ctx, updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointGcpTerraformResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointGcpTerraformModel
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
