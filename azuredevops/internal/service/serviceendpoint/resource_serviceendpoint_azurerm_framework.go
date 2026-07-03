package serviceendpoint

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource              = (*ServiceEndpointAzureRMResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointAzureRMResource)(nil)
)

// ── plan modifier helpers ─────────────────────────────────────────────────────

type seAzureRMRequiresReplaceModifier struct{}

func seAzureRMRequiresReplace() planmodifier.String { return seAzureRMRequiresReplaceModifier{} }
func (m seAzureRMRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seAzureRMRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seAzureRMRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seAzureRMUseStateForUnknownModifier struct{}

func seAzureRMUseStateForUnknown() planmodifier.String { return seAzureRMUseStateForUnknownModifier{} }
func (m seAzureRMUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seAzureRMUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seAzureRMUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── default helpers ───────────────────────────────────────────────────────────

type seAzureRMStringDefault struct{ value string }

func seAzureRMDefaultString(v string) defaults.String { return seAzureRMStringDefault{v} }
func (d seAzureRMStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seAzureRMStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seAzureRMStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

type seAzureRMBoolDefault struct{ value bool }

func seAzureRMDefaultBool(v bool) defaults.Bool { return seAzureRMBoolDefault{v} }
func (d seAzureRMBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}
func (d seAzureRMBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}
func (d seAzureRMBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// NewServiceEndpointAzureRMResource is the constructor for the framework provider.
func NewServiceEndpointAzureRMResource() resource.Resource {
	return &ServiceEndpointAzureRMResource{}
}

// ServiceEndpointAzureRMResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_azurerm.
type ServiceEndpointAzureRMResource struct {
	client *client.AggregatedClient
}

// serviceEndpointAzureRMCredentialsModel maps the credentials block.
type serviceEndpointAzureRMCredentialsModel struct {
	ServicePrincipalID          types.String `tfsdk:"serviceprincipalid"`
	ServicePrincipalKey         types.String `tfsdk:"serviceprincipalkey"`
	ServicePrincipalCertificate types.String `tfsdk:"serviceprincipalcertificate"`
}

// serviceEndpointAzureRMFeaturesModel maps the features block.
type serviceEndpointAzureRMFeaturesModel struct {
	Validate types.Bool `tfsdk:"validate"`
}

// serviceEndpointAzureRMModel is the resource state model.
type serviceEndpointAzureRMModel struct {
	ID                                  types.String                             `tfsdk:"id"`
	ProjectID                           types.String                             `tfsdk:"project_id"`
	ServiceEndpointName                 types.String                             `tfsdk:"service_endpoint_name"`
	Description                         types.String                             `tfsdk:"description"`
	Authorization                       types.Map                                `tfsdk:"authorization"`
	AzureRMSpnTenantID                  types.String                             `tfsdk:"azurerm_spn_tenantid"`
	AzureRMSubscriptionID               types.String                             `tfsdk:"azurerm_subscription_id"`
	AzureRMSubscriptionName             types.String                             `tfsdk:"azurerm_subscription_name"`
	AzureRMManagementGroupID            types.String                             `tfsdk:"azurerm_management_group_id"`
	AzureRMManagementGroupName          types.String                             `tfsdk:"azurerm_management_group_name"`
	ResourceGroup                       types.String                             `tfsdk:"resource_group"`
	Environment                         types.String                             `tfsdk:"environment"`
	ServiceEndpointAuthenticationScheme types.String                             `tfsdk:"service_endpoint_authentication_scheme"`
	ServerURL                           types.String                             `tfsdk:"server_url"`
	WorkloadIdentityFederationIssuer    types.String                             `tfsdk:"workload_identity_federation_issuer"`
	WorkloadIdentityFederationSubject   types.String                             `tfsdk:"workload_identity_federation_subject"`
	ServicePrincipalID                  types.String                             `tfsdk:"service_principal_id"`
	Credentials                         []serviceEndpointAzureRMCredentialsModel `tfsdk:"credentials"`
	Features                            []serviceEndpointAzureRMFeaturesModel    `tfsdk:"features"`
}

func (r *ServiceEndpointAzureRMResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_azurerm"
}

func (r *ServiceEndpointAzureRMResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure RM service endpoint within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					seAzureRMUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The project ID or project name.",
				PlanModifiers: []planmodifier.String{
					seAzureRMRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The service endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureRMDefaultString("Managed by Terraform"),
				Description: "The service endpoint description.",
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"azurerm_spn_tenantid": schema.StringAttribute{
				Required:    true,
				Description: "The service principal tenant id which should be used.",
			},
			"azurerm_subscription_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureRMDefaultString(""),
				Description: "The Azure subscription Id which should be used.",
			},
			"azurerm_subscription_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureRMDefaultString(""),
				Description: "The Azure subscription name which should be used.",
			},
			"azurerm_management_group_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureRMDefaultString(""),
				Description: "The Azure managementGroup Id which should be used.",
			},
			"azurerm_management_group_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureRMDefaultString(""),
				Description: "The Azure managementGroup name which should be used.",
			},
			"resource_group": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureRMDefaultString(""),
				Description: "Scope Resource Group",
				PlanModifiers: []planmodifier.String{
					seAzureRMRequiresReplace(),
				},
			},
			"environment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureRMDefaultString("AzureCloud"),
				Description: "Environment (Azure Cloud type)",
				PlanModifiers: []planmodifier.String{
					seAzureRMRequiresReplace(),
				},
			},
			"service_endpoint_authentication_scheme": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAzureRMDefaultString("ServicePrincipal"),
				Description: "The AzureRM Service Endpoint Authentication Scheme.",
				PlanModifiers: []planmodifier.String{
					seAzureRMRequiresReplace(),
				},
			},
			"server_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A URL to the server for AzureStack environments.",
				PlanModifiers: []planmodifier.String{
					seAzureRMRequiresReplace(),
					seAzureRMUseStateForUnknown(),
				},
			},
			"workload_identity_federation_issuer": schema.StringAttribute{
				Computed:    true,
				Description: "The issuer of the workload identity federation service principal.",
				PlanModifiers: []planmodifier.String{
					seAzureRMUseStateForUnknown(),
				},
			},
			"workload_identity_federation_subject": schema.StringAttribute{
				Computed:    true,
				Description: "The subject of the workload identity federation service principal.",
				PlanModifiers: []planmodifier.String{
					seAzureRMUseStateForUnknown(),
				},
			},
			"service_principal_id": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal ID.",
				PlanModifiers: []planmodifier.String{
					seAzureRMUseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"credentials": schema.ListNestedBlock{
				Description: "Service principal credentials.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"serviceprincipalid": schema.StringAttribute{
							Required:    true,
							Description: "The service principal id which should be used.",
						},
						"serviceprincipalkey": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Sensitive:   true,
							Default:     seAzureRMDefaultString(""),
							Description: "The service principal key.",
						},
						"serviceprincipalcertificate": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Sensitive:   true,
							Default:     seAzureRMDefaultString(""),
							Description: "The service principal certificate.",
						},
					},
				},
			},
			"features": schema.ListNestedBlock{
				Description: "Feature flags.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"validate": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     seAzureRMDefaultBool(false),
							Description: "Whether or not to validate connection with Azure after create or update operations.",
						},
					},
				},
			},
		},
	}
}

func (r *ServiceEndpointAzureRMResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── helpers ───────────────────────────────────────────────────────────────────

func azureRMEndpointURL(environment, serverURL string) (string, error) {
	switch environment {
	case "AzureCloud":
		return "https://management.azure.com/", nil
	case "AzureChinaCloud":
		return "https://management.chinacloudapi.cn/", nil
	case "AzureUSGovernment":
		return "https://management.usgovcloudapi.net/", nil
	case "AzureGermanCloud":
		return "https://management.microsoftazure.de", nil
	case "AzureStack":
		if serverURL == "" {
			return "", fmt.Errorf("`server_url` is required when `environment` is `AzureStack`")
		}
		return serverURL, nil
	default:
		return "https://management.azure.com/", nil
	}
}

func azureRMBuildAuth(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}

func (r *ServiceEndpointAzureRMResource) expandToServiceEndpoint(plan *serviceEndpointAzureRMModel) (*serviceendpoint.ServiceEndpoint, error) {
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()
	scheme := EndpointAuthenticationScheme(plan.ServiceEndpointAuthenticationScheme.ValueString())
	environment := plan.Environment.ValueString()
	subID := plan.AzureRMSubscriptionID.ValueString()
	subName := plan.AzureRMSubscriptionName.ValueString()
	mgmtGrpID := plan.AzureRMManagementGroupID.ValueString()
	mgmtGrpName := plan.AzureRMManagementGroupName.ValueString()
	tenantID := plan.AzureRMSpnTenantID.ValueString()
	resourceGroup := plan.ResourceGroup.ValueString()
	serverURL := plan.ServerURL.ValueString()

	// Validate scope: at least one of subscription or managementGroup must be set.
	scopeMap := map[string][]string{
		"subscription":    {subID, subName},
		"managementGroup": {mgmtGrpID, mgmtGrpName},
	}
	if err := validateScopeLevel(scopeMap); err != nil {
		return nil, err
	}

	endpointURL, err := azureRMEndpointURL(environment, serverURL)
	if err != nil {
		return nil, err
	}

	hasCredentials := len(plan.Credentials) > 0
	schemeHasCreationMode := scheme == ServicePrincipal || scheme == WorkloadIdentityFederation
	var creationMode EndpointCreationMode
	if schemeHasCreationMode {
		if hasCredentials {
			creationMode = Manual
		} else {
			creationMode = Automatic
		}
	}

	var scopeLevel string
	var scope string
	if subID != "" {
		scope = fmt.Sprintf("/subscriptions/%s", subID)
		scopeLevel = "Subscription"
		if schemeHasCreationMode && resourceGroup != "" {
			scope += fmt.Sprintf("/resourcegroups/%s", resourceGroup)
			scopeLevel = "ResourceGroup"
		}
	}

	var authParams *map[string]string
	var data *map[string]string

	switch scheme {
	case ServicePrincipal:
		if creationMode == Automatic {
			authParams = &map[string]string{
				"authenticationType":  "spnKey",
				"serviceprincipalid":  "",
				"serviceprincipalkey": "",
				"tenantid":            tenantID,
			}
		} else {
			creds := plan.Credentials[0]
			params := map[string]string{
				"serviceprincipalid": creds.ServicePrincipalID.ValueString(),
				"tenantid":           tenantID,
			}
			if spnKey := creds.ServicePrincipalKey.ValueString(); spnKey != "" {
				params["authenticationType"] = "spnKey"
				params["serviceprincipalkey"] = spnKey
			}
			if spnCert := creds.ServicePrincipalCertificate.ValueString(); spnCert != "" {
				params["authenticationType"] = "spnCertificate"
				params["servicePrincipalCertificate"] = spnCert
			}
			authParams = &params
		}
		data = &map[string]string{
			"creationMode": string(creationMode),
			"environment":  environment,
		}

	case ManagedServiceIdentity:
		authParams = &map[string]string{"tenantid": tenantID}
		data = &map[string]string{"environment": environment}

	case WorkloadIdentityFederation:
		if creationMode == Automatic {
			authParams = &map[string]string{
				"serviceprincipalid": "",
				"tenantid":           tenantID,
			}
		} else {
			creds := plan.Credentials[0]
			spID := creds.ServicePrincipalID.ValueString()
			if spID == "" {
				return nil, fmt.Errorf("serviceprincipalid is required for WorkloadIdentityFederation")
			}
			authParams = &map[string]string{
				"serviceprincipalid": spID,
				"tenantid":           tenantID,
			}
		}
		data = &map[string]string{
			"creationMode": string(creationMode),
			"environment":  environment,
		}
	}

	if scopeLevel == "Subscription" || scopeLevel == "ResourceGroup" {
		(*data)["scopeLevel"] = "Subscription"
		(*data)["subscriptionId"] = subID
		(*data)["subscriptionName"] = subName
	}
	if scopeLevel == "ResourceGroup" {
		(*authParams)["scope"] = scope
	}
	if mgmtGrpID != "" {
		(*data)["scopeLevel"] = "ManagementGroup"
		(*data)["managementGroupId"] = mgmtGrpID
		(*data)["managementGroupName"] = mgmtGrpName
	}

	return &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("azurerm"),
		Url:         converter.String(endpointURL),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: authParams,
			Scheme:     converter.String(string(scheme)),
		},
		Data: data,
		ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
			{
				ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
				Name:             converter.String(name),
				Description:      converter.String(description),
			},
		},
	}, nil
}

func (r *ServiceEndpointAzureRMResource) flattenFromServiceEndpoint(ctx context.Context, ep *serviceendpoint.ServiceEndpoint, state *serviceEndpointAzureRMModel) {
	if ep.Name != nil {
		state.ServiceEndpointName = types.StringValue(*ep.Name)
	}
	if ep.Description != nil {
		state.Description = types.StringValue(*ep.Description)
	}

	authScheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		authScheme = *ep.Authorization.Scheme
	}
	state.Authorization = azureRMBuildAuth(ctx, authScheme)
	scheme := EndpointAuthenticationScheme(authScheme)

	// WIF fields must always be set to a known value after apply.
	// Initialise to empty string; overwrite below if API returns actual values.
	// This prevents Terraform from seeing "unknown after apply" for computed-only fields.
	if state.WorkloadIdentityFederationIssuer.IsUnknown() {
		state.WorkloadIdentityFederationIssuer = types.StringValue("")
	}
	if state.WorkloadIdentityFederationSubject.IsUnknown() {
		state.WorkloadIdentityFederationSubject = types.StringValue("")
	}
	if state.ServicePrincipalID.IsUnknown() {
		state.ServicePrincipalID = types.StringValue("")
	}

	if ep.Authorization != nil && ep.Authorization.Parameters != nil {
		params := *ep.Authorization.Parameters
		if v, ok := params["tenantid"]; ok {
			state.AzureRMSpnTenantID = types.StringValue(v)
		}
		if v, ok := params["serviceprincipalid"]; ok {
			state.ServicePrincipalID = types.StringValue(v)
		}
		// WIF fields: for WIF scheme, populate from API; for other schemes they stay as
		// empty strings (already set above), which is the correct known value.
		if scheme == WorkloadIdentityFederation {
			if v, ok := params["workloadIdentityFederationIssuer"]; ok {
				state.WorkloadIdentityFederationIssuer = types.StringValue(v)
			}
			if v, ok := params["workloadIdentityFederationSubject"]; ok {
				state.WorkloadIdentityFederationSubject = types.StringValue(v)
			}
		}
		// scope -> resource_group
		if scopeVal, ok := params["scope"]; ok {
			parts := strings.Split(scopeVal, "/")
			if len(parts) == 5 {
				state.ResourceGroup = types.StringValue(parts[4])
			}
		}
		// Preserve credentials.serviceprincipalid from API if manual mode
		if len(state.Credentials) > 0 {
			if spID, ok := params["serviceprincipalid"]; ok {
				state.Credentials[0].ServicePrincipalID = types.StringValue(spID)
			}
			// Note: sensitive fields (key, cert) are never returned by the API — preserve state values.
		}
	}

	if ep.Data != nil {
		data := *ep.Data
		if v, ok := data["environment"]; ok {
			state.Environment = types.StringValue(v)
		}
		if v, ok := data["subscriptionId"]; ok {
			state.AzureRMSubscriptionID = types.StringValue(v)
		}
		if v, ok := data["subscriptionName"]; ok {
			state.AzureRMSubscriptionName = types.StringValue(v)
		}
		if v, ok := data["managementGroupId"]; ok {
			state.AzureRMManagementGroupID = types.StringValue(v)
		}
		if v, ok := data["managementGroupName"]; ok {
			state.AzureRMManagementGroupName = types.StringValue(v)
		}
	}

	if ep.Url != nil {
		state.ServerURL = types.StringValue(*ep.Url)
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointAzureRMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointAzureRMModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ep, err := r.expandToServiceEndpoint(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building service endpoint", err.Error())
		return
	}

	created, err := r.client.ServiceEndpointClient.CreateServiceEndpoint(ctx, serviceendpoint.CreateServiceEndpointArgs{
		Endpoint: ep,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating service endpoint", err.Error())
		return
	}

	plan.ID = types.StringValue(created.Id.String())
	r.flattenFromServiceEndpoint(ctx, created, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointAzureRMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointAzureRMModel
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

	r.flattenFromServiceEndpoint(ctx, ep, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointAzureRMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointAzureRMModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointAzureRMModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())

	ep, err := r.expandToServiceEndpoint(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building service endpoint", err.Error())
		return
	}
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
	r.flattenFromServiceEndpoint(ctx, updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointAzureRMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointAzureRMModel
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
