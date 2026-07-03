package serviceendpoint

import (
	"context"
	"fmt"
	"strconv"

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
	_ resource.Resource              = (*ServiceEndpointAwsResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointAwsResource)(nil)
)

// ── plan modifier / default helpers ──────────────────────────────────────────

type seAwsRequiresReplaceModifier struct{}

func seAwsRequiresReplace() planmodifier.String { return seAwsRequiresReplaceModifier{} }
func (m seAwsRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seAwsRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seAwsRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seAwsUseStateForUnknownModifier struct{}

func seAwsUseStateForUnknown() planmodifier.String { return seAwsUseStateForUnknownModifier{} }
func (m seAwsUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seAwsUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seAwsUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seAwsStringDefault struct{ value string }

func seAwsDefaultString(v string) defaults.String { return seAwsStringDefault{v} }
func (d seAwsStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seAwsStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seAwsStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

type seAwsBoolDefault struct{ value bool }

func seAwsDefaultBool(v bool) defaults.Bool { return seAwsBoolDefault{v} }
func (d seAwsBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d seAwsBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d seAwsBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// NewServiceEndpointAwsResource is the constructor for the framework provider.
func NewServiceEndpointAwsResource() resource.Resource {
	return &ServiceEndpointAwsResource{}
}

// ServiceEndpointAwsResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_aws.
type ServiceEndpointAwsResource struct {
	client *client.AggregatedClient
}

type serviceEndpointAwsModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	ServiceEndpointName types.String `tfsdk:"service_endpoint_name"`
	Description         types.String `tfsdk:"description"`
	Authorization       types.Map    `tfsdk:"authorization"`
	AccessKeyID         types.String `tfsdk:"access_key_id"`
	SecretAccessKey     types.String `tfsdk:"secret_access_key"`
	SessionToken        types.String `tfsdk:"session_token"`
	RoleToAssume        types.String `tfsdk:"role_to_assume"`
	RoleSessionName     types.String `tfsdk:"role_session_name"`
	ExternalID          types.String `tfsdk:"external_id"`
	UseOIDC             types.Bool   `tfsdk:"use_oidc"`
}

func (r *ServiceEndpointAwsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_aws"
}

func (r *ServiceEndpointAwsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an AWS service endpoint within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					seAwsUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					seAwsRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  seAwsDefaultString("Managed by Terraform"),
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"access_key_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAwsDefaultString(""),
				Description: "The AWS access key ID for signing programmatic requests.",
			},
			"secret_access_key": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Default:     seAwsDefaultString(""),
				Description: "The AWS secret access key for signing programmatic requests.",
			},
			"session_token": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Default:     seAwsDefaultString(""),
				Description: "The AWS session token for signing programmatic requests.",
			},
			"role_to_assume": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAwsDefaultString(""),
				Description: "The Amazon Resource Name (ARN) of the role to assume.",
			},
			"role_session_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAwsDefaultString(""),
				Description: "Optional identifier for the assumed role session.",
			},
			"external_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAwsDefaultString(""),
				Description: "A unique identifier that is used by third parties when assuming roles in their customers' accounts.",
			},
			"use_oidc": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seAwsDefaultBool(false),
				Description: "Enable this to attempt getting credentials with OIDC token from Azure Devops.",
			},
		},
	}
}

func (r *ServiceEndpointAwsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceEndpointAwsResource) buildEndpoint(plan *serviceEndpointAwsModel) *serviceendpoint.ServiceEndpoint {
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()
	return &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("aws"),
		Url:         converter.String("https://aws.amazon.com/"),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"username":        plan.AccessKeyID.ValueString(),
				"password":        plan.SecretAccessKey.ValueString(),
				"sessionToken":    plan.SessionToken.ValueString(),
				"assumeRoleArn":   plan.RoleToAssume.ValueString(),
				"roleSessionName": plan.RoleSessionName.ValueString(),
				"externalId":      plan.ExternalID.ValueString(),
				"useOIDC":         strconv.FormatBool(plan.UseOIDC.ValueBool()),
			},
			Scheme: converter.String("UsernamePassword"),
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

func (r *ServiceEndpointAwsResource) flattenEndpoint(ctx context.Context, ep *serviceendpoint.ServiceEndpoint, state *serviceEndpointAwsModel) {
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
		if v, ok := params["username"]; ok {
			state.AccessKeyID = types.StringValue(v)
		}
		if v, ok := params["assumeRoleArn"]; ok {
			state.RoleToAssume = types.StringValue(v)
		}
		if v, ok := params["roleSessionName"]; ok {
			state.RoleSessionName = types.StringValue(v)
		}
		if v, ok := params["externalId"]; ok {
			state.ExternalID = types.StringValue(v)
		}
		if v, ok := params["useOIDC"]; ok && v != "" {
			b, err := strconv.ParseBool(v)
			if err == nil {
				state.UseOIDC = types.BoolValue(b)
			}
		}
		// secret_access_key / session_token never returned by API — preserve state
	}
}

func (r *ServiceEndpointAwsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointAwsModel
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

func (r *ServiceEndpointAwsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointAwsModel
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

func (r *ServiceEndpointAwsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointAwsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointAwsModel
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

func (r *ServiceEndpointAwsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointAwsModel
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
