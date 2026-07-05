package serviceendpoint

import (
	"context"
	"fmt"

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
	_ resource.Resource              = (*ServiceEndpointOpenshiftResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointOpenshiftResource)(nil)
)

type seOpenshiftRequiresReplaceModifier struct{}

func seOpenshiftRequiresReplace() planmodifier.String { return seOpenshiftRequiresReplaceModifier{} }
func (m seOpenshiftRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seOpenshiftRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seOpenshiftRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seOpenshiftUseStateForUnknownModifier struct{}

func seOpenshiftUseStateForUnknown() planmodifier.String {
	return seOpenshiftUseStateForUnknownModifier{}
}
func (m seOpenshiftUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seOpenshiftUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seOpenshiftUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seOpenshiftStringDefault struct{ value string }

func seOpenshiftDefaultString(v string) defaults.String { return seOpenshiftStringDefault{v} }
func (d seOpenshiftStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seOpenshiftStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seOpenshiftStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointOpenshiftResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_openshift.
type ServiceEndpointOpenshiftResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointOpenshiftResource returns a new resource.Resource.
func NewServiceEndpointOpenshiftResource() resource.Resource {
	return &ServiceEndpointOpenshiftResource{}
}

func (r *ServiceEndpointOpenshiftResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_openshift"
}

func (r *ServiceEndpointOpenshiftResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OpenShift Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seOpenshiftUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seOpenshiftRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seOpenshiftDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"server_url": schema.StringAttribute{
				Optional:    true,
				Description: "The OpenShift server URL.",
			},
			"accept_untrusted_certs": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to accept untrusted certificates.",
			},
			"certificate_authority_file": schema.StringAttribute{
				Optional:    true,
				Description: "The certificate authority file path.",
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Specifies the authorization scheme and parameters.",
			},
		},
		Blocks: map[string]schema.Block{
			"auth_basic": schema.ListNestedBlock{
				Description: "Basic authentication block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
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
			"auth_token": schema.ListNestedBlock{
				Description: "Token authentication block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Required:    true,
							Description: "The token.",
						},
					},
				},
			},
			"auth_none": schema.ListNestedBlock{
				Description: "No authentication block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"kube_config": schema.StringAttribute{
							Optional:    true,
							Description: "The kubeconfig.",
						},
					},
				},
			},
		},
	}
}

func (r *ServiceEndpointOpenshiftResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointOpenshiftModel struct {
	ID                       types.String              `tfsdk:"id"`
	ProjectID                types.String              `tfsdk:"project_id"`
	ServiceEndpointName      types.String              `tfsdk:"service_endpoint_name"`
	Description              types.String              `tfsdk:"description"`
	ServerURL                types.String              `tfsdk:"server_url"`
	AcceptUntrustedCerts     types.Bool                `tfsdk:"accept_untrusted_certs"`
	CertificateAuthorityFile types.String              `tfsdk:"certificate_authority_file"`
	Authorization            types.Map                 `tfsdk:"authorization"`
	AuthBasic                []seOpenshiftBasicModel   `tfsdk:"auth_basic"`
	AuthToken                []seOpenshiftTokenModel   `tfsdk:"auth_token"`
	AuthNone                 []seOpenshiftNoneModel    `tfsdk:"auth_none"`
}

type seOpenshiftBasicModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type seOpenshiftTokenModel struct {
	Token types.String `tfsdk:"token"`
}

type seOpenshiftNoneModel struct {
	KubeConfig types.String `tfsdk:"kube_config"`
}

func (r *ServiceEndpointOpenshiftResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointOpenshiftModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, scheme := seOpenshiftBuildEndpoint(plan)
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
	plan.Authorization = seOpenshiftBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointOpenshiftResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointOpenshiftModel
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
		state.ServerURL = types.StringValue(*ep.Url)
	}

	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seOpenshiftBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServiceEndpointOpenshiftResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointOpenshiftModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointOpenshiftModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint, scheme := seOpenshiftBuildEndpoint(plan)
	endpoint.Id = &endpointID
	endpoint.ServiceEndpointProjectReferences = &[]serviceendpoint.ServiceEndpointProjectReference{
		{
			ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
			Name:             converter.String(name),
			Description:      converter.String(description),
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
	plan.Authorization = seOpenshiftBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointOpenshiftResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointOpenshiftModel
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

func seOpenshiftBuildEndpoint(plan serviceEndpointOpenshiftModel) (*serviceendpoint.ServiceEndpoint, string) {
	ep := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(plan.ServiceEndpointName.ValueString()),
		Owner:       converter.String("library"),
		Description: converter.String(plan.Description.ValueString()),
		Type:        converter.String("openshift"),
		Url:         converter.String(plan.ServerURL.ValueString()),
	}

	if len(plan.AuthBasic) > 0 {
		params := map[string]string{
			"username":             plan.AuthBasic[0].Username.ValueString(),
			"password":             plan.AuthBasic[0].Password.ValueString(),
			"acceptUntrustedCerts": fmt.Sprintf("%v", plan.AcceptUntrustedCerts.ValueBool()),
		}
		if !plan.CertificateAuthorityFile.IsNull() && plan.CertificateAuthorityFile.ValueString() != "" {
			params["certificateAuthorityFile"] = plan.CertificateAuthorityFile.ValueString()
		}
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &params,
			Scheme:     converter.String("UsernamePassword"),
		}
		return ep, "UsernamePassword"
	}

	if len(plan.AuthToken) > 0 {
		params := map[string]string{
			"apitoken":             plan.AuthToken[0].Token.ValueString(),
			"acceptUntrustedCerts": fmt.Sprintf("%v", plan.AcceptUntrustedCerts.ValueBool()),
		}
		if !plan.CertificateAuthorityFile.IsNull() && plan.CertificateAuthorityFile.ValueString() != "" {
			params["certificateAuthorityFile"] = plan.CertificateAuthorityFile.ValueString()
		}
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &params,
			Scheme:     converter.String("Token"),
		}
		return ep, "Token"
	}

	if len(plan.AuthNone) > 0 {
		params := map[string]string{}
		if !plan.AuthNone[0].KubeConfig.IsNull() {
			params["kubeConfig"] = plan.AuthNone[0].KubeConfig.ValueString()
		}
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &params,
			Scheme:     converter.String("None"),
		}
		return ep, "None"
	}

	ep.Authorization = &serviceendpoint.EndpointAuthorization{
		Parameters: &map[string]string{},
		Scheme:     converter.String("None"),
	}
	return ep, "None"
}

func seOpenshiftBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
