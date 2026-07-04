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
	_ resource.Resource              = (*ServiceEndpointExternalTFSResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointExternalTFSResource)(nil)
)

// ── local plan modifier + default helpers ─────────────────────────────────────

type seExternalTFSRequiresReplaceModifier struct{}

func seExternalTFSRequiresReplace() planmodifier.String {
	return seExternalTFSRequiresReplaceModifier{}
}

func (m seExternalTFSRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seExternalTFSRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seExternalTFSRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seExternalTFSUseStateForUnknownModifier struct{}

func seExternalTFSUseStateForUnknown() planmodifier.String {
	return seExternalTFSUseStateForUnknownModifier{}
}

func (m seExternalTFSUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seExternalTFSUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seExternalTFSUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seExternalTFSStringDefault struct{ value string }

func seExternalTFSDefaultString(v string) defaults.String { return seExternalTFSStringDefault{v} }
func (d seExternalTFSStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seExternalTFSStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seExternalTFSStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointExternalTFSResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_externaltfs.
type ServiceEndpointExternalTFSResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointExternalTFSResource returns a new resource.Resource.
func NewServiceEndpointExternalTFSResource() resource.Resource {
	return &ServiceEndpointExternalTFSResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServiceEndpointExternalTFSResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_externaltfs"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointExternalTFSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an External TFS Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seExternalTFSUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seExternalTFSRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seExternalTFSDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"connection_url": schema.StringAttribute{
				Required:    true,
				Description: "URL of the Azure DevOps organization or the TFS Project Collection to connect to.",
			},
			"auth_personal": schema.ListNestedAttribute{
				Required:    true,
				Description: "An auth_personal block as documented below.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"personal_access_token": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "Personal access tokens are applicable only for connections targeting Azure DevOps organization or TFS 2017 (and higher).",
						},
					},
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

func (r *ServiceEndpointExternalTFSResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointExternalTFSAuthPersonalModel struct {
	PersonalAccessToken types.String `tfsdk:"personal_access_token"`
}

type serviceEndpointExternalTFSModel struct {
	ID                  types.String                                  `tfsdk:"id"`
	ProjectID           types.String                                  `tfsdk:"project_id"`
	ServiceEndpointName types.String                                  `tfsdk:"service_endpoint_name"`
	Description         types.String                                  `tfsdk:"description"`
	ConnectionURL       types.String                                  `tfsdk:"connection_url"`
	AuthPersonal        []serviceEndpointExternalTFSAuthPersonalModel `tfsdk:"auth_personal"`
	Authorization       types.Map                                     `tfsdk:"authorization"`
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointExternalTFSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointExternalTFSModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	pat := ""
	if len(plan.AuthPersonal) > 0 {
		pat = plan.AuthPersonal[0].PersonalAccessToken.ValueString()
	}

	endpoint := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("externaltfs"),
		Url:         converter.String(plan.ConnectionURL.ValueString()),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"apitoken": pat,
			},
			Scheme: converter.String("Token"),
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
	plan.Authorization = seExternalTFSBuildAuthMap(ctx, "Token")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointExternalTFSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointExternalTFSModel
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
		state.ConnectionURL = types.StringValue(*ep.Url)
	}
	// NOTE: API never returns PAT; preserve state value to avoid spurious diffs.

	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seExternalTFSBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointExternalTFSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointExternalTFSModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointExternalTFSModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	pat := ""
	if len(plan.AuthPersonal) > 0 {
		pat = plan.AuthPersonal[0].PersonalAccessToken.ValueString()
	}

	endpoint := &serviceendpoint.ServiceEndpoint{
		Id:          &endpointID,
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("externaltfs"),
		Url:         converter.String(plan.ConnectionURL.ValueString()),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"apitoken": pat,
			},
			Scheme: converter.String("Token"),
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
	plan.Authorization = seExternalTFSBuildAuthMap(ctx, "Token")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointExternalTFSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointExternalTFSModel
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

func seExternalTFSBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
