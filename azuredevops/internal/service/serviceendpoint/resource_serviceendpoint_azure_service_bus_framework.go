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
	_ resource.Resource              = (*ServiceEndpointAzureServiceBusResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointAzureServiceBusResource)(nil)
)

// ── plan modifier / default helpers ──────────────────────────────────────────

type seASBRequiresReplaceModifier struct{}

func seASBRequiresReplace() planmodifier.String { return seASBRequiresReplaceModifier{} }
func (m seASBRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seASBRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seASBRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seASBUseStateForUnknownModifier struct{}

func seASBUseStateForUnknown() planmodifier.String { return seASBUseStateForUnknownModifier{} }
func (m seASBUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seASBUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seASBUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seASBStringDefault struct{ value string }

func seASBDefaultString(v string) defaults.String { return seASBStringDefault{v} }
func (d seASBStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seASBStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seASBStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// NewServiceEndpointAzureServiceBusResource is the constructor for the framework provider.
func NewServiceEndpointAzureServiceBusResource() resource.Resource {
	return &ServiceEndpointAzureServiceBusResource{}
}

// ServiceEndpointAzureServiceBusResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_azure_service_bus.
type ServiceEndpointAzureServiceBusResource struct {
	client *client.AggregatedClient
}

type serviceEndpointAzureServiceBusModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	ServiceEndpointName types.String `tfsdk:"service_endpoint_name"`
	Description         types.String `tfsdk:"description"`
	Authorization       types.Map    `tfsdk:"authorization"`
	ConnectionString    types.String `tfsdk:"connection_string"`
	QueueName           types.String `tfsdk:"queue_name"`
}

func (r *ServiceEndpointAzureServiceBusResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_azure_service_bus"
}

func (r *ServiceEndpointAzureServiceBusResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure Service Bus service endpoint within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					seASBUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					seASBRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  seASBDefaultString("Managed by Terraform"),
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"connection_string": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"queue_name": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *ServiceEndpointAzureServiceBusResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceEndpointAzureServiceBusResource) buildEndpoint(plan *serviceEndpointAzureServiceBusModel) *serviceendpoint.ServiceEndpoint {
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()
	return &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("AzureServiceBus"),
		Url:         converter.String("https://management.core.windows.net/"),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"serviceBusConnectionString": plan.ConnectionString.ValueString(),
			},
			Scheme: converter.String("None"),
		},
		Data: &map[string]string{
			"serviceBusQueueName": plan.QueueName.ValueString(),
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

func (r *ServiceEndpointAzureServiceBusResource) flattenEndpoint(ctx context.Context, ep *serviceendpoint.ServiceEndpoint, state *serviceEndpointAzureServiceBusModel) {
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

	if ep.Data != nil {
		if v, ok := (*ep.Data)["serviceBusQueueName"]; ok {
			state.QueueName = types.StringValue(v)
		}
	}
	// connection_string is sensitive and never returned by the API — preserve state value
}

func (r *ServiceEndpointAzureServiceBusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointAzureServiceBusModel
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

func (r *ServiceEndpointAzureServiceBusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointAzureServiceBusModel
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

func (r *ServiceEndpointAzureServiceBusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointAzureServiceBusModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointAzureServiceBusModel
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

func (r *ServiceEndpointAzureServiceBusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointAzureServiceBusModel
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
