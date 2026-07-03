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
	_ resource.Resource              = (*ServiceEndpointIncomingWebhookResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointIncomingWebhookResource)(nil)
)

// ── local plan modifier + default helpers ─────────────────────────────────────

type seIncomingWebhookRequiresReplaceModifier struct{}

func seIncomingWebhookRequiresReplace() planmodifier.String {
	return seIncomingWebhookRequiresReplaceModifier{}
}

func (m seIncomingWebhookRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seIncomingWebhookRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seIncomingWebhookRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seIncomingWebhookUseStateForUnknownModifier struct{}

func seIncomingWebhookUseStateForUnknown() planmodifier.String {
	return seIncomingWebhookUseStateForUnknownModifier{}
}

func (m seIncomingWebhookUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seIncomingWebhookUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seIncomingWebhookUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seIncomingWebhookStringDefault struct{ value string }

func seIncomingWebhookDefaultString(v string) defaults.String {
	return seIncomingWebhookStringDefault{v}
}

func (d seIncomingWebhookStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seIncomingWebhookStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seIncomingWebhookStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointIncomingWebhookResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_incomingwebhook.
type ServiceEndpointIncomingWebhookResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointIncomingWebhookResource returns a new resource.Resource.
func NewServiceEndpointIncomingWebhookResource() resource.Resource {
	return &ServiceEndpointIncomingWebhookResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServiceEndpointIncomingWebhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_incomingwebhook"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointIncomingWebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Incoming WebHook Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seIncomingWebhookUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seIncomingWebhookRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seIncomingWebhookDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"webhook_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the WebHook.",
			},
			"secret": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Default:     seIncomingWebhookDefaultString(""),
				Description: "Optional secret for the webhook. WebHook service will use this secret to calculate the payload checksum.",
			},
			"http_header": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seIncomingWebhookDefaultString(""),
				Description: "Optional http header name on which checksum will be sent.",
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

func (r *ServiceEndpointIncomingWebhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointIncomingWebhookModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	ServiceEndpointName types.String `tfsdk:"service_endpoint_name"`
	Description         types.String `tfsdk:"description"`
	WebhookName         types.String `tfsdk:"webhook_name"`
	Secret              types.String `tfsdk:"secret"`
	HTTPHeader          types.String `tfsdk:"http_header"`
	Authorization       types.Map    `tfsdk:"authorization"`
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointIncomingWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointIncomingWebhookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("incomingwebhook"),
		Url:         converter.String("https://dev.azure.com"),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"webhookname": plan.WebhookName.ValueString(),
				"secret":      plan.Secret.ValueString(),
				"header":      plan.HTTPHeader.ValueString(),
			},
			Scheme: converter.String("None"),
		},
		ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
			{
				ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
				Name:             converter.String(name),
				Description:      converter.String(description),
			},
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
	plan.Authorization = seIncomingWebhookBuildAuthMap(ctx, "None")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointIncomingWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointIncomingWebhookModel
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
	if ep.Authorization != nil && ep.Authorization.Parameters != nil {
		if v, ok := (*ep.Authorization.Parameters)["webhookname"]; ok {
			state.WebhookName = types.StringValue(v)
		}
		if v, ok := (*ep.Authorization.Parameters)["header"]; ok {
			state.HTTPHeader = types.StringValue(v)
		}
		// NOTE: API never returns secret; preserve state value to avoid spurious diffs.
	}

	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seIncomingWebhookBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointIncomingWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointIncomingWebhookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointIncomingWebhookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint := &serviceendpoint.ServiceEndpoint{
		Id:          &endpointID,
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("incomingwebhook"),
		Url:         converter.String("https://dev.azure.com"),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"webhookname": plan.WebhookName.ValueString(),
				"secret":      plan.Secret.ValueString(),
				"header":      plan.HTTPHeader.ValueString(),
			},
			Scheme: converter.String("None"),
		},
		ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
			{
				ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
				Name:             converter.String(name),
				Description:      converter.String(description),
			},
		},
	}

	_, err := r.client.ServiceEndpointClient.UpdateServiceEndpoint(ctx, serviceendpoint.UpdateServiceEndpointArgs{
		Endpoint:   endpoint,
		EndpointId: &endpointID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating service endpoint", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Authorization = seIncomingWebhookBuildAuthMap(ctx, "None")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointIncomingWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointIncomingWebhookModel
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

func seIncomingWebhookBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
