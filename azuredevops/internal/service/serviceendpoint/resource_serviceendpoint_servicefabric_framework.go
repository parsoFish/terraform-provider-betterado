package serviceendpoint

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*ServiceEndpointServiceFabricResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointServiceFabricResource)(nil)
)

type seServiceFabricRequiresReplaceModifier struct{}

func seServiceFabricRequiresReplace() planmodifier.String {
	return seServiceFabricRequiresReplaceModifier{}
}
func (m seServiceFabricRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seServiceFabricRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seServiceFabricRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seServiceFabricUseStateForUnknownModifier struct{}

func seServiceFabricUseStateForUnknown() planmodifier.String {
	return seServiceFabricUseStateForUnknownModifier{}
}
func (m seServiceFabricUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seServiceFabricUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seServiceFabricUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seServiceFabricStringDefault struct{ value string }

func seServiceFabricDefaultString(v string) defaults.String { return seServiceFabricStringDefault{v} }
func (d seServiceFabricStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seServiceFabricStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seServiceFabricStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointServiceFabricResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_servicefabric.
type ServiceEndpointServiceFabricResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointServiceFabricResource returns a new resource.Resource.
func NewServiceEndpointServiceFabricResource() resource.Resource {
	return &ServiceEndpointServiceFabricResource{}
}

func (r *ServiceEndpointServiceFabricResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_servicefabric"
}

func (r *ServiceEndpointServiceFabricResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Service Fabric Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seServiceFabricUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seServiceFabricRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seServiceFabricDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"cluster_endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Client connection endpoint for the cluster (prefix with 'tcp://').",
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Specifies the authorization scheme and parameters.",
			},
		},
		Blocks: map[string]schema.Block{
			"certificate": schema.ListNestedBlock{
				Description: "Certificate authentication block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"server_certificate_lookup": schema.StringAttribute{
							Required:    true,
							Description: "Server certificate lookup method (Thumbprint or CommonName).",
						},
						"server_certificate_thumbprint": schema.StringAttribute{
							Optional:    true,
							Description: "The thumbprint(s) of the cluster's certificate(s).",
						},
						"server_certificate_common_name": schema.StringAttribute{
							Optional:    true,
							Description: "The common name(s) of the cluster's certificate(s).",
						},
						"client_certificate": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "Base64 encoding of the cluster's client certificate file.",
						},
						"client_certificate_password": schema.StringAttribute{
							Optional:    true,
							Sensitive:   true,
							Description: "Password for the certificate.",
						},
					},
				},
			},
			"azure_active_directory": schema.ListNestedBlock{
				Description: "Azure Active Directory authentication block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"server_certificate_lookup": schema.StringAttribute{
							Required:    true,
							Description: "Server certificate lookup method (Thumbprint or CommonName).",
						},
						"server_certificate_thumbprint": schema.StringAttribute{
							Optional:    true,
							Description: "The thumbprint(s) of the cluster's certificate(s).",
						},
						"server_certificate_common_name": schema.StringAttribute{
							Optional:    true,
							Description: "The common name(s) of the cluster's certificate(s).",
						},
						"username": schema.StringAttribute{
							Required:    true,
							Description: "Specify an Azure Active Directory account.",
						},
						"password": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "Password for the Azure Active Directory account.",
						},
					},
				},
			},
			"none": schema.ListNestedBlock{
				Description: "No-authentication block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"unsecured": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Skip using windows security for authentication.",
						},
						"cluster_spn": schema.StringAttribute{
							Optional:    true,
							Description: "Fully qualified domain SPN for gMSA account.",
						},
					},
				},
			},
		},
	}
}

func (r *ServiceEndpointServiceFabricResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointServiceFabricModel struct {
	ID                  types.String                        `tfsdk:"id"`
	ProjectID           types.String                        `tfsdk:"project_id"`
	ServiceEndpointName types.String                        `tfsdk:"service_endpoint_name"`
	Description         types.String                        `tfsdk:"description"`
	ClusterEndpoint     types.String                        `tfsdk:"cluster_endpoint"`
	Authorization       types.Map                           `tfsdk:"authorization"`
	Certificate         []seServiceFabricCertModel          `tfsdk:"certificate"`
	AzureActiveDirectory []seServiceFabricAADModel          `tfsdk:"azure_active_directory"`
	None                []seServiceFabricNoneModel          `tfsdk:"none"`
}

type seServiceFabricCertModel struct {
	ServerCertificateLookup      types.String `tfsdk:"server_certificate_lookup"`
	ServerCertificateThumbprint  types.String `tfsdk:"server_certificate_thumbprint"`
	ServerCertificateCommonName  types.String `tfsdk:"server_certificate_common_name"`
	ClientCertificate            types.String `tfsdk:"client_certificate"`
	ClientCertificatePassword    types.String `tfsdk:"client_certificate_password"`
}

type seServiceFabricAADModel struct {
	ServerCertificateLookup     types.String `tfsdk:"server_certificate_lookup"`
	ServerCertificateThumbprint types.String `tfsdk:"server_certificate_thumbprint"`
	ServerCertificateCommonName types.String `tfsdk:"server_certificate_common_name"`
	Username                    types.String `tfsdk:"username"`
	Password                    types.String `tfsdk:"password"`
}

type seServiceFabricNoneModel struct {
	Unsecured  types.Bool   `tfsdk:"unsecured"`
	ClusterSPN types.String `tfsdk:"cluster_spn"`
}

func (r *ServiceEndpointServiceFabricResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointServiceFabricModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, scheme, err := seServiceFabricBuildEndpoint(plan)
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
	plan.Authorization = seServiceFabricBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointServiceFabricResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointServiceFabricModel
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
		state.ClusterEndpoint = types.StringValue(*ep.Url)
	}

	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seServiceFabricBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServiceEndpointServiceFabricResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointServiceFabricModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointServiceFabricModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint, scheme, err := seServiceFabricBuildEndpoint(plan)
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
	plan.Authorization = seServiceFabricBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointServiceFabricResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointServiceFabricModel
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

func seServiceFabricBuildCertLookupParams(lookup, thumbprint, commonName string) map[string]string {
	params := map[string]string{"certLookup": lookup}
	switch lookup {
	case "Thumbprint":
		params["servercertthumbprint"] = thumbprint
	case "CommonName":
		params["servercertcommonname"] = commonName
	}
	return params
}

func seServiceFabricBuildEndpoint(plan serviceEndpointServiceFabricModel) (*serviceendpoint.ServiceEndpoint, string, error) {
	ep := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(plan.ServiceEndpointName.ValueString()),
		Owner:       converter.String("library"),
		Description: converter.String(plan.Description.ValueString()),
		Type:        converter.String("servicefabric"),
		Url:         converter.String(plan.ClusterEndpoint.ValueString()),
	}

	if len(plan.Certificate) > 0 {
		cert := plan.Certificate[0]
		params := seServiceFabricBuildCertLookupParams(
			cert.ServerCertificateLookup.ValueString(),
			cert.ServerCertificateThumbprint.ValueString(),
			cert.ServerCertificateCommonName.ValueString(),
		)
		params["certificate"] = cert.ClientCertificate.ValueString()
		params["certificatepassword"] = cert.ClientCertificatePassword.ValueString()
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &params,
			Scheme:     converter.String("Certificate"),
		}
		return ep, "Certificate", nil
	}

	if len(plan.AzureActiveDirectory) > 0 {
		aad := plan.AzureActiveDirectory[0]
		params := seServiceFabricBuildCertLookupParams(
			aad.ServerCertificateLookup.ValueString(),
			aad.ServerCertificateThumbprint.ValueString(),
			aad.ServerCertificateCommonName.ValueString(),
		)
		params["username"] = aad.Username.ValueString()
		params["password"] = aad.Password.ValueString()
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &params,
			Scheme:     converter.String("UsernamePassword"),
		}
		return ep, "UsernamePassword", nil
	}

	if len(plan.None) > 0 {
		n := plan.None[0]
		params := map[string]string{
			"Unsecured":  strconv.FormatBool(n.Unsecured.ValueBool()),
			"ClusterSpn": n.ClusterSPN.ValueString(),
		}
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &params,
			Scheme:     converter.String("None"),
		}
		return ep, "None", nil
	}

	return nil, "", fmt.Errorf("one of certificate, azure_active_directory, or none blocks must be specified")
}

func seServiceFabricBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
