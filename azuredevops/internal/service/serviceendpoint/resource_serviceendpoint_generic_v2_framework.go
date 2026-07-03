package serviceendpoint

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*ServiceEndpointGenericV2Resource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointGenericV2Resource)(nil)
)

// ── local plan modifiers ──────────────────────────────────────────────────────

type seGenericV2RequiresReplaceModifier struct{}

func seGenericV2RequiresReplace() planmodifier.String { return seGenericV2RequiresReplaceModifier{} }

func (m seGenericV2RequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seGenericV2RequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seGenericV2RequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seGenericV2UseStateForUnknownModifier struct{}

func seGenericV2UseStateForUnknown() planmodifier.String {
	return seGenericV2UseStateForUnknownModifier{}
}

func (m seGenericV2UseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seGenericV2UseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seGenericV2UseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── local defaults ────────────────────────────────────────────────────────────

type seGenericV2StringDefault struct{ value string }

func seGenericV2DefaultString(v string) defaults.String { return seGenericV2StringDefault{v} }

func (d seGenericV2StringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seGenericV2StringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seGenericV2StringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointGenericV2Resource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_generic_v2.
type ServiceEndpointGenericV2Resource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointGenericV2Resource returns a new resource.Resource.
func NewServiceEndpointGenericV2Resource() resource.Resource {
	return &ServiceEndpointGenericV2Resource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServiceEndpointGenericV2Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_generic_v2"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointGenericV2Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Generic (v2) Service Connection within Azure DevOps, supporting arbitrary endpoint types and authorization schemes.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seGenericV2UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seGenericV2RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid UUID",
					),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the service endpoint.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seGenericV2DefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the service endpoint (e.g. 'generic', 'github', etc.).",
				PlanModifiers: []planmodifier.String{
					seGenericV2RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"server_url": schema.StringAttribute{
				Required:    true,
				Description: "The server URL of the service connection.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^https?://`),
						"must be a valid URL beginning with http:// or https://",
					),
				},
			},
			"authorization_scheme": schema.StringAttribute{
				Required:    true,
				Description: "The authorization scheme used for the service connection.",
				PlanModifiers: []planmodifier.String{
					seGenericV2RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"authorization_parameters": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				ElementType: types.StringType,
				Description: "A map of authorization parameters for the service connection.",
			},
			"parameters": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "A map of endpoint-specific data parameters.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ServiceEndpointGenericV2Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointGenericV2Model struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Type                types.String `tfsdk:"type"`
	ServerURL           types.String `tfsdk:"server_url"`
	AuthorizationScheme types.String `tfsdk:"authorization_scheme"`
	AuthorizationParams types.Map    `tfsdk:"authorization_parameters"`
	Parameters          types.Map    `tfsdk:"parameters"`
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointGenericV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointGenericV2Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.Name.ValueString()
	description := plan.Description.ValueString()
	endpointType := plan.Type.ValueString()
	serverURL := plan.ServerURL.ValueString()
	authScheme := plan.AuthorizationScheme.ValueString()

	authParams := seGenericV2MapToStringMap(ctx, plan.AuthorizationParams)
	dataParams := seGenericV2MapToStringMap(ctx, plan.Parameters)

	endpoint := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String(endpointType),
		Url:         converter.String(serverURL),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Scheme:     converter.String(authScheme),
			Parameters: &authParams,
		},
		Data: &dataParams,
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

	// Read back the created endpoint to populate computed fields
	readBack, readErr := r.readEndpoint(ctx, created.Id.String(), plan.ProjectID.ValueString())
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading back created service endpoint", readErr.Error())
		return
	}
	seGenericV2ApplyReadBack(ctx, readBack, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointGenericV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointGenericV2Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ep, err := r.readEndpoint(ctx, state.ID.ValueString(), state.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading service endpoint", err.Error())
		return
	}
	if ep == nil || ep.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	seGenericV2ApplyReadBack(ctx, ep, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointGenericV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointGenericV2Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointGenericV2Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.Name.ValueString()
	description := plan.Description.ValueString()
	endpointType := plan.Type.ValueString()
	serverURL := plan.ServerURL.ValueString()
	authScheme := plan.AuthorizationScheme.ValueString()

	authParams := seGenericV2MapToStringMap(ctx, plan.AuthorizationParams)
	dataParams := seGenericV2MapToStringMap(ctx, plan.Parameters)

	endpoint := &serviceendpoint.ServiceEndpoint{
		Id:          &endpointID,
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String(endpointType),
		Url:         converter.String(serverURL),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Scheme:     converter.String(authScheme),
			Parameters: &authParams,
		},
		Data: &dataParams,
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

	// Read back to populate computed fields
	readBack, readErr := r.readEndpoint(ctx, state.ID.ValueString(), plan.ProjectID.ValueString())
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading back updated service endpoint", readErr.Error())
		return
	}
	seGenericV2ApplyReadBack(ctx, readBack, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointGenericV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointGenericV2Model
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

func (r *ServiceEndpointGenericV2Resource) readEndpoint(ctx context.Context, id, projectID string) (*serviceendpoint.ServiceEndpoint, error) {
	endpointID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid service endpoint ID %q: %w", id, err)
	}
	return r.client.ServiceEndpointClient.GetServiceEndpointDetails(ctx, serviceendpoint.GetServiceEndpointDetailsArgs{
		EndpointId: &endpointID,
		Project:    &projectID,
	})
}

// seGenericV2MapToStringMap converts a types.Map to map[string]string.
func seGenericV2MapToStringMap(ctx context.Context, m types.Map) map[string]string {
	result := make(map[string]string)
	if m.IsNull() || m.IsUnknown() {
		return result
	}
	_ = m.ElementsAs(ctx, &result, false)
	return result
}

// seGenericV2ApplyReadBack updates model fields from a live API response.
func seGenericV2ApplyReadBack(ctx context.Context, ep *serviceendpoint.ServiceEndpoint, model *serviceEndpointGenericV2Model) {
	if ep == nil {
		return
	}
	if ep.Name != nil {
		model.Name = types.StringValue(*ep.Name)
	}
	if ep.Description != nil {
		model.Description = types.StringValue(*ep.Description)
	}
	if ep.Url != nil {
		model.ServerURL = types.StringValue(*ep.Url)
	}
	if ep.Type != nil {
		model.Type = types.StringValue(*ep.Type)
	}
	if ep.Authorization != nil {
		if ep.Authorization.Scheme != nil {
			model.AuthorizationScheme = types.StringValue(*ep.Authorization.Scheme)
		}
		// NOTE: authorization_parameters — API strips sensitive values (passwords).
		// Preserve whatever is already in state for sensitive fields; merge non-sensitive.
		// We keep the plan value for authorization_parameters to avoid spurious diffs
		// on sensitive fields the API scrubs.
	}
	if ep.Data != nil && len(*ep.Data) > 0 {
		dataMap, diags := types.MapValueFrom(ctx, types.StringType, *ep.Data)
		if !diags.HasError() {
			model.Parameters = dataMap
		}
	}
}
