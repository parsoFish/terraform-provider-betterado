package serviceendpoint

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource               = (*ServiceEndpointAzureCRResource)(nil)
	_ resource.ResourceWithConfigure  = (*ServiceEndpointAzureCRResource)(nil)
	_ resource.ResourceWithModifyPlan = (*ServiceEndpointAzureCRResource)(nil)
)

// ── plan modifier helpers ─────────────────────────────────────────────────────

type seAzureCRRequiresReplaceModifier struct{}

func seAzureCRRequiresReplace() planmodifier.String { return seAzureCRRequiresReplaceModifier{} }
func (m seAzureCRRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seAzureCRRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seAzureCRRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seAzureCRUseStateForUnknownModifier struct{}

func seAzureCRUseStateForUnknown() planmodifier.String { return seAzureCRUseStateForUnknownModifier{} }

func (m seAzureCRUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seAzureCRUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seAzureCRUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── default helpers ───────────────────────────────────────────────────────────

type seAzureCRStringDefault struct{ value string }

func seAzureCRDefaultString(v string) defaults.String { return seAzureCRStringDefault{v} }
func (d seAzureCRStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seAzureCRStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seAzureCRStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointAzureCRResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_azurecr.
type ServiceEndpointAzureCRResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointAzureCRResource returns a new resource.Resource.
func NewServiceEndpointAzureCRResource() resource.Resource {
	return &ServiceEndpointAzureCRResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServiceEndpointAzureCRResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_azurecr"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointAzureCRResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure Container Registry Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seAzureCRUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seAzureCRRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureCRDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"azurecr_spn_tenantid": schema.StringAttribute{
				Required:    true,
				Description: "The service principal tenant id which should be used.",
			},
			"azurecr_subscription_id": schema.StringAttribute{
				Required:    true,
				Description: "The Azure subscription Id which should be used.",
			},
			"azurecr_subscription_name": schema.StringAttribute{
				Required:    true,
				Description: "The Azure subscription name which should be used.",
			},
			"azurecr_name": schema.StringAttribute{
				Required:    true,
				Description: "The AzureContainerRegistry registry which should be used.",
				PlanModifiers: []planmodifier.String{
					seAzureCRRequiresReplace(),
				},
			},
			"resource_group": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureCRDefaultString(""),
				Description: "Scope Resource Group.",
				PlanModifiers: []planmodifier.String{
					seAzureCRRequiresReplace(),
				},
			},
			"service_endpoint_authentication_scheme": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureCRDefaultString("ServicePrincipal"),
				Description: "The AzureCR Service Endpoint Authentication Scheme: 'WorkloadIdentityFederation', 'ManagedServiceIdentity' or 'ServicePrincipal'.",
				PlanModifiers: []planmodifier.String{
					seAzureCRRequiresReplace(),
				},
			},
			"workload_identity_federation_issuer": schema.StringAttribute{
				Computed:    true,
				Description: "The issuer of the workload identity federation service principal.",
				PlanModifiers: []planmodifier.String{
					seAzureCRUseStateForUnknown(),
				},
			},
			"workload_identity_federation_subject": schema.StringAttribute{
				Computed:    true,
				Description: "The subject of the workload identity federation service principal.",
				PlanModifiers: []planmodifier.String{
					seAzureCRUseStateForUnknown(),
				},
			},
			"app_object_id": schema.StringAttribute{
				Computed:    true,
				Description: "The object ID of the service principal's associated application.",
				PlanModifiers: []planmodifier.String{
					seAzureCRUseStateForUnknown(),
				},
			},
			"spn_object_id": schema.StringAttribute{
				Computed:    true,
				Description: "The object ID of the service principal.",
				PlanModifiers: []planmodifier.String{
					seAzureCRUseStateForUnknown(),
				},
			},
			"az_spn_role_assignment_id": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal role assignment ID.",
				PlanModifiers: []planmodifier.String{
					seAzureCRUseStateForUnknown(),
				},
			},
			"az_spn_role_permissions": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal role permissions.",
				PlanModifiers: []planmodifier.String{
					seAzureCRUseStateForUnknown(),
				},
			},
			"service_principal_id": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal id.",
				PlanModifiers: []planmodifier.String{
					seAzureCRUseStateForUnknown(),
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

func (r *ServiceEndpointAzureCRResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointAzureCRModel struct {
	ID                                  types.String `tfsdk:"id"`
	ProjectID                           types.String `tfsdk:"project_id"`
	ServiceEndpointName                 types.String `tfsdk:"service_endpoint_name"`
	Description                         types.String `tfsdk:"description"`
	AzureCRSpnTenantID                  types.String `tfsdk:"azurecr_spn_tenantid"`
	AzureCRSubscriptionID               types.String `tfsdk:"azurecr_subscription_id"`
	AzureCRSubscriptionName             types.String `tfsdk:"azurecr_subscription_name"`
	AzureCRName                         types.String `tfsdk:"azurecr_name"`
	ResourceGroup                       types.String `tfsdk:"resource_group"`
	ServiceEndpointAuthenticationScheme types.String `tfsdk:"service_endpoint_authentication_scheme"`
	WorkloadIdentityFederationIssuer    types.String `tfsdk:"workload_identity_federation_issuer"`
	WorkloadIdentityFederationSubject   types.String `tfsdk:"workload_identity_federation_subject"`
	AppObjectID                         types.String `tfsdk:"app_object_id"`
	SpnObjectID                         types.String `tfsdk:"spn_object_id"`
	AzSpnRoleAssignmentID               types.String `tfsdk:"az_spn_role_assignment_id"`
	AzSpnRolePermissions                types.String `tfsdk:"az_spn_role_permissions"`
	ServicePrincipalID                  types.String `tfsdk:"service_principal_id"`
	Authorization                       types.Map    `tfsdk:"authorization"`
}

// ── ModifyPlan ────────────────────────────────────────────────────────────────

// ModifyPlan resolves computed-only fields to known prior-state values when
// the plan would otherwise show them as unknown (causing non-empty re-plans).
func (r *ServiceEndpointAzureCRResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var state, plan serviceEndpointAzureCRModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve computed fields from prior state when plan shows unknown.
	if plan.ServicePrincipalID.IsUnknown() && !state.ServicePrincipalID.IsNull() {
		plan.ServicePrincipalID = state.ServicePrincipalID
	}
	if plan.WorkloadIdentityFederationIssuer.IsUnknown() && !state.WorkloadIdentityFederationIssuer.IsNull() {
		plan.WorkloadIdentityFederationIssuer = state.WorkloadIdentityFederationIssuer
	}
	if plan.WorkloadIdentityFederationSubject.IsUnknown() && !state.WorkloadIdentityFederationSubject.IsNull() {
		plan.WorkloadIdentityFederationSubject = state.WorkloadIdentityFederationSubject
	}
	if plan.AppObjectID.IsUnknown() && !state.AppObjectID.IsNull() {
		plan.AppObjectID = state.AppObjectID
	}
	if plan.SpnObjectID.IsUnknown() && !state.SpnObjectID.IsNull() {
		plan.SpnObjectID = state.SpnObjectID
	}
	if plan.AzSpnRoleAssignmentID.IsUnknown() && !state.AzSpnRoleAssignmentID.IsNull() {
		plan.AzSpnRoleAssignmentID = state.AzSpnRoleAssignmentID
	}
	if plan.AzSpnRolePermissions.IsUnknown() && !state.AzSpnRolePermissions.IsNull() {
		plan.AzSpnRolePermissions = state.AzSpnRolePermissions
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointAzureCRResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointAzureCRModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint, diags := seAzureCRBuildEndpoint(ctx, &plan, &projectID, name, description)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.ServiceEndpointClient.CreateServiceEndpoint(ctx, serviceendpoint.CreateServiceEndpointArgs{
		Endpoint: endpoint,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating AzureCR service endpoint", err.Error())
		return
	}

	plan.ID = types.StringValue(created.Id.String())
	// Set computed fields from the response
	seAzureCRFlattenResponse(ctx, created, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointAzureCRResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointAzureCRModel
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
		resp.Diagnostics.AddError("Error reading AzureCR service endpoint", err.Error())
		return
	}
	if ep == nil || ep.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	seAzureCRFlattenResponse(ctx, ep, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointAzureCRResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointAzureCRModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointAzureCRModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	// Carry forward computed fields from state during update
	if plan.ServicePrincipalID.IsUnknown() || plan.ServicePrincipalID.IsNull() {
		plan.ServicePrincipalID = state.ServicePrincipalID
	}
	if plan.AppObjectID.IsUnknown() || plan.AppObjectID.IsNull() {
		plan.AppObjectID = state.AppObjectID
	}
	if plan.SpnObjectID.IsUnknown() || plan.SpnObjectID.IsNull() {
		plan.SpnObjectID = state.SpnObjectID
	}
	if plan.AzSpnRoleAssignmentID.IsUnknown() || plan.AzSpnRoleAssignmentID.IsNull() {
		plan.AzSpnRoleAssignmentID = state.AzSpnRoleAssignmentID
	}
	if plan.AzSpnRolePermissions.IsUnknown() || plan.AzSpnRolePermissions.IsNull() {
		plan.AzSpnRolePermissions = state.AzSpnRolePermissions
	}

	endpoint, diags := seAzureCRBuildEndpoint(ctx, &plan, &projectID, name, description)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint.Id = &endpointID

	updated, err := r.client.ServiceEndpointClient.UpdateServiceEndpoint(ctx, serviceendpoint.UpdateServiceEndpointArgs{
		Endpoint:   endpoint,
		EndpointId: &endpointID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating AzureCR service endpoint", err.Error())
		return
	}

	plan.ID = state.ID
	seAzureCRFlattenResponse(ctx, updated, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointAzureCRResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointAzureCRModel
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
		resp.Diagnostics.AddError("Error deleting AzureCR service endpoint", err.Error())
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func seAzureCRBuildEndpoint(_ context.Context, plan *serviceEndpointAzureCRModel, projectID *uuid.UUID, name, description string) (*serviceendpoint.ServiceEndpoint, diag.Diagnostics) {
	var diags diag.Diagnostics

	scheme := plan.ServiceEndpointAuthenticationScheme.ValueString()
	subscriptionID := plan.AzureCRSubscriptionID.ValueString()
	resourceGroup := plan.ResourceGroup.ValueString()
	acrName := plan.AzureCRName.ValueString()

	scope := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerRegistry/registries/%s",
		subscriptionID, resourceGroup, acrName,
	)
	loginServer := fmt.Sprintf("%s.azurecr.io", strings.ToLower(acrName))
	azureContainerRegistryURL := fmt.Sprintf("https://%s", loginServer)

	ep := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("dockerregistry"),
		Url:         converter.String(azureContainerRegistryURL),
		ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
			{
				ProjectReference: &serviceendpoint.ProjectReference{Id: projectID},
				Name:             converter.String(name),
				Description:      converter.String(description),
			},
		},
	}

	switch EndpointAuthenticationScheme(scheme) {
	case ServicePrincipal:
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"authenticationType": "spnKey",
				"tenantId":           plan.AzureCRSpnTenantID.ValueString(),
				"loginServer":        loginServer,
				"scope":              scope,
				"serviceprincipalid": plan.ServicePrincipalID.ValueString(),
			},
			Scheme: converter.String(scheme),
		}
		ep.Data = &map[string]string{
			"registryId":               scope,
			"subscriptionId":           subscriptionID,
			"subscriptionName":         plan.AzureCRSubscriptionName.ValueString(),
			"registrytype":             "ACR",
			"appObjectId":              plan.AppObjectID.ValueString(),
			"spnObjectId":              plan.SpnObjectID.ValueString(),
			"azureSpnPermissions":      plan.AzSpnRolePermissions.ValueString(),
			"azureSpnRoleAssignmentId": plan.AzSpnRoleAssignmentID.ValueString(),
		}
	case WorkloadIdentityFederation:
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"serviceprincipalid": plan.ServicePrincipalID.ValueString(),
				"tenantId":           plan.AzureCRSpnTenantID.ValueString(),
				"loginServer":        loginServer,
				"scope":              scope,
			},
			Scheme: converter.String(scheme),
		}
		ep.Data = &map[string]string{
			"registryId":               scope,
			"registrytype":             "ACR",
			"subscriptionId":           subscriptionID,
			"subscriptionName":         plan.AzureCRSubscriptionName.ValueString(),
			"creationMode":             "Automatic",
			"appObjectId":              plan.AppObjectID.ValueString(),
			"spnObjectId":              plan.SpnObjectID.ValueString(),
			"azureSpnPermissions":      plan.AzSpnRolePermissions.ValueString(),
			"azureSpnRoleAssignmentId": plan.AzSpnRoleAssignmentID.ValueString(),
		}
	default:
		diags.AddError(
			"Unsupported authentication scheme",
			fmt.Sprintf("Authentication scheme %q is not supported. Use 'ServicePrincipal' or 'WorkloadIdentityFederation'.", scheme),
		)
	}

	return ep, diags
}

func seAzureCRFlattenResponse(ctx context.Context, ep *serviceendpoint.ServiceEndpoint, state *serviceEndpointAzureCRModel) {
	if ep.Name != nil {
		state.ServiceEndpointName = types.StringValue(*ep.Name)
	}
	if ep.Description != nil {
		state.Description = types.StringValue(*ep.Description)
	}

	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
		state.ServiceEndpointAuthenticationScheme = types.StringValue(scheme)
	}

	if ep.Authorization != nil && ep.Authorization.Parameters != nil {
		params := *ep.Authorization.Parameters
		if v, ok := params["tenantId"]; ok {
			state.AzureCRSpnTenantID = types.StringValue(v)
		}
		if v, ok := params["serviceprincipalid"]; ok {
			state.ServicePrincipalID = types.StringValue(v)
		}
		if v, ok := params["workloadIdentityFederationIssuer"]; ok {
			state.WorkloadIdentityFederationIssuer = types.StringValue(v)
		} else {
			state.WorkloadIdentityFederationIssuer = types.StringValue("")
		}
		if v, ok := params["workloadIdentityFederationSubject"]; ok {
			state.WorkloadIdentityFederationSubject = types.StringValue(v)
		} else {
			state.WorkloadIdentityFederationSubject = types.StringValue("")
		}
		if scopeVal, ok := params["scope"]; ok {
			s := strings.Split(scopeVal, "/")
			if len(s) >= 9 {
				state.ResourceGroup = types.StringValue(s[4])
				state.AzureCRName = types.StringValue(s[8])
			}
		}
	}

	if ep.Data != nil {
		data := *ep.Data
		if v, ok := data["subscriptionId"]; ok {
			state.AzureCRSubscriptionID = types.StringValue(v)
		}
		if v, ok := data["subscriptionName"]; ok {
			state.AzureCRSubscriptionName = types.StringValue(v)
		}
		if v, ok := data["appObjectId"]; ok {
			state.AppObjectID = types.StringValue(v)
		} else {
			state.AppObjectID = types.StringValue("")
		}
		if v, ok := data["spnObjectId"]; ok {
			state.SpnObjectID = types.StringValue(v)
		} else {
			state.SpnObjectID = types.StringValue("")
		}
		if v, ok := data["azureSpnPermissions"]; ok {
			state.AzSpnRolePermissions = types.StringValue(v)
		} else {
			state.AzSpnRolePermissions = types.StringValue("")
		}
		if v, ok := data["azureSpnRoleAssignmentId"]; ok {
			state.AzSpnRoleAssignmentID = types.StringValue(v)
		} else {
			state.AzSpnRoleAssignmentID = types.StringValue("")
		}
	}

	authMap := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, authMap)
	state.Authorization = m
}
