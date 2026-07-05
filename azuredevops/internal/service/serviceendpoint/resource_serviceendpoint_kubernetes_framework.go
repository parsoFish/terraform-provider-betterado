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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*ServiceEndpointKubernetesResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointKubernetesResource)(nil)
)

type seKubernetesRequiresReplaceModifier struct{}

func seKubernetesRequiresReplace() planmodifier.String {
	return seKubernetesRequiresReplaceModifier{}
}
func (m seKubernetesRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seKubernetesRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m seKubernetesRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seKubernetesUseStateForUnknownModifier struct{}

func seKubernetesUseStateForUnknown() planmodifier.String {
	return seKubernetesUseStateForUnknownModifier{}
}
func (m seKubernetesUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seKubernetesUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m seKubernetesUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seKubernetesStringDefault struct{ value string }

func seKubernetesDefaultString(v string) defaults.String { return seKubernetesStringDefault{v} }
func (d seKubernetesStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seKubernetesStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d seKubernetesStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointKubernetesResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_kubernetes.
type ServiceEndpointKubernetesResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointKubernetesResource returns a new resource.Resource.
func NewServiceEndpointKubernetesResource() resource.Resource {
	return &ServiceEndpointKubernetesResource{}
}

func (r *ServiceEndpointKubernetesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_kubernetes"
}

func (r *ServiceEndpointKubernetesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Kubernetes Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seKubernetesUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seKubernetesRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seKubernetesDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"apiserver_url": schema.StringAttribute{
				Required:    true,
				Description: "URL to Kubernetes API-Server.",
			},
			"authorization_type": schema.StringAttribute{
				Required:    true,
				Description: "Type of credentials to use (AzureSubscription, Kubeconfig, ServiceAccount).",
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Specifies the authorization scheme and parameters.",
			},
		},
		Blocks: map[string]schema.Block{
			"azure_subscription": schema.SetNestedBlock{
				Description: "AzureSubscription-type of configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"azure_environment": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("AzureCloud"),
							Description: "Type of azure cloud.",
						},
						"cluster_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of AKS resource.",
						},
						"subscription_id": schema.StringAttribute{
							Required:    true,
							Description: "ID of azure subscription.",
						},
						"subscription_name": schema.StringAttribute{
							Required:    true,
							Description: "Name of azure subscription.",
						},
						"tenant_id": schema.StringAttribute{
							Required:    true,
							Description: "ID of AAD tenant.",
						},
						"resourcegroup_id": schema.StringAttribute{
							Required:    true,
							Description: "ID of resource group.",
						},
						"namespace": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("default"),
							Description: "Accessed namespace.",
						},
						"cluster_admin": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Enable Cluster Admin.",
						},
					},
				},
			},
			"kubeconfig": schema.ListNestedBlock{
				Description: "Kubeconfig-type of configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"kube_config": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "Content of the kubeconfig file.",
						},
						"cluster_context": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Context of the cluster.",
						},
						"accept_untrusted_certs": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
							Description: "Enable this if your authentication uses untrusted certificates.",
						},
					},
				},
			},
			"service_account": schema.ListNestedBlock{
				Description: "ServiceAccount-type of configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"ca_cert": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "Secret cert.",
						},
						"token": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "Secret token.",
						},
						"accept_untrusted_certs": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Enable this if authentication uses untrusted certificates.",
						},
					},
				},
			},
		},
	}
}

func (r *ServiceEndpointKubernetesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointKubernetesModel struct {
	ID                  types.String                      `tfsdk:"id"`
	ProjectID           types.String                      `tfsdk:"project_id"`
	ServiceEndpointName types.String                      `tfsdk:"service_endpoint_name"`
	Description         types.String                      `tfsdk:"description"`
	ApiserverURL        types.String                      `tfsdk:"apiserver_url"`
	AuthorizationType   types.String                      `tfsdk:"authorization_type"`
	Authorization       types.Map                         `tfsdk:"authorization"`
	AzureSubscription   []seKubernetesAzureSubModel       `tfsdk:"azure_subscription"`
	Kubeconfig          []seKubernetesKubeconfigModel     `tfsdk:"kubeconfig"`
	ServiceAccount      []seKubernetesServiceAccountModel `tfsdk:"service_account"`
}

type seKubernetesAzureSubModel struct {
	AzureEnvironment types.String `tfsdk:"azure_environment"`
	ClusterName      types.String `tfsdk:"cluster_name"`
	SubscriptionID   types.String `tfsdk:"subscription_id"`
	SubscriptionName types.String `tfsdk:"subscription_name"`
	TenantID         types.String `tfsdk:"tenant_id"`
	ResourceGroupID  types.String `tfsdk:"resourcegroup_id"`
	Namespace        types.String `tfsdk:"namespace"`
	ClusterAdmin     types.Bool   `tfsdk:"cluster_admin"`
}

type seKubernetesKubeconfigModel struct {
	KubeConfig           types.String `tfsdk:"kube_config"`
	ClusterContext       types.String `tfsdk:"cluster_context"`
	AcceptUntrustedCerts types.Bool   `tfsdk:"accept_untrusted_certs"`
}

type seKubernetesServiceAccountModel struct {
	CACert               types.String `tfsdk:"ca_cert"`
	Token                types.String `tfsdk:"token"`
	AcceptUntrustedCerts types.Bool   `tfsdk:"accept_untrusted_certs"`
}

func (r *ServiceEndpointKubernetesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointKubernetesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, err := seKubernetesBuildEndpoint(plan)
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
	plan.Authorization = seKubernetesBuildAuthMap(ctx, "Kubernetes")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointKubernetesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointKubernetesModel
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
		state.ApiserverURL = types.StringValue(*ep.Url)
	}
	if ep.Data != nil {
		if v, ok := (*ep.Data)["authorizationType"]; ok {
			state.AuthorizationType = types.StringValue(v)
		}
	}

	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seKubernetesBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServiceEndpointKubernetesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointKubernetesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointKubernetesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint, err := seKubernetesBuildEndpoint(plan)
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
	plan.Authorization = seKubernetesBuildAuthMap(ctx, "Kubernetes")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceEndpointKubernetesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointKubernetesModel
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

func seKubernetesBuildEndpoint(plan serviceEndpointKubernetesModel) (*serviceendpoint.ServiceEndpoint, error) {
	ep := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(plan.ServiceEndpointName.ValueString()),
		Owner:       converter.String("library"),
		Description: converter.String(plan.Description.ValueString()),
		Type:        converter.String("kubernetes"),
		Url:         converter.String(plan.ApiserverURL.ValueString()),
	}

	switch plan.AuthorizationType.ValueString() {
	case "AzureSubscription":
		if len(plan.AzureSubscription) == 0 {
			return nil, fmt.Errorf("azure_subscription block is required when authorization_type is AzureSubscription")
		}
		sub := plan.AzureSubscription[0]
		clusterID := fmt.Sprintf(
			"/subscriptions/%s/resourcegroups/%s/providers/Microsoft.ContainerService/managedClusters/%s",
			sub.SubscriptionID.ValueString(),
			sub.ResourceGroupID.ValueString(),
			sub.ClusterName.ValueString(),
		)
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"azureEnvironment": sub.AzureEnvironment.ValueString(),
				"azureTenantId":    sub.TenantID.ValueString(),
			},
			Scheme: converter.String("Kubernetes"),
		}
		ep.Data = &map[string]string{
			"authorizationType":     "AzureSubscription",
			"azureSubscriptionId":   sub.SubscriptionID.ValueString(),
			"azureSubscriptionName": sub.SubscriptionName.ValueString(),
			"clusterId":             clusterID,
			"namespace":             sub.Namespace.ValueString(),
			"clusterAdmin":          strconv.FormatBool(sub.ClusterAdmin.ValueBool()),
		}

	case "Kubeconfig":
		if len(plan.Kubeconfig) == 0 {
			return nil, fmt.Errorf("kubeconfig block is required when authorization_type is Kubeconfig")
		}
		kc := plan.Kubeconfig[0]
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"clusterContext": kc.ClusterContext.ValueString(),
				"kubeconfig":     kc.KubeConfig.ValueString(),
			},
			Scheme: converter.String("Kubernetes"),
		}
		ep.Data = &map[string]string{
			"authorizationType":    "Kubeconfig",
			"acceptUntrustedCerts": strconv.FormatBool(kc.AcceptUntrustedCerts.ValueBool()),
		}

	case "ServiceAccount":
		if len(plan.ServiceAccount) == 0 {
			return nil, fmt.Errorf("service_account block is required when authorization_type is ServiceAccount")
		}
		sa := plan.ServiceAccount[0]
		ep.Authorization = &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"apiToken":                  sa.Token.ValueString(),
				"serviceAccountCertificate": sa.CACert.ValueString(),
			},
			Scheme: converter.String("Token"),
		}
		ep.Data = &map[string]string{
			"acceptUntrustedCerts": strconv.FormatBool(sa.AcceptUntrustedCerts.ValueBool()),
			"authorizationType":    "ServiceAccount",
		}

	default:
		return nil, fmt.Errorf("unsupported authorization_type: %s", plan.AuthorizationType.ValueString())
	}

	return ep, nil
}

func seKubernetesBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
