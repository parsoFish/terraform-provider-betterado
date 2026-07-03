package taskagent

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*EnvironmentResourceKubernetes)(nil)
	_ resource.ResourceWithConfigure = (*EnvironmentResourceKubernetes)(nil)
)

// requiresReplaceSetModifier forces replacement when a Set attribute changes.
type requiresReplaceSetModifier struct{}

func requiresReplaceSet() planmodifier.Set { return requiresReplaceSetModifier{} }
func (requiresReplaceSetModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (requiresReplaceSetModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (requiresReplaceSetModifier) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// EnvironmentResourceKubernetes is the terraform-plugin-framework implementation of
// betterado_environment_resource_kubernetes.
type EnvironmentResourceKubernetes struct {
	client *client.AggregatedClient
}

// NewEnvironmentResourceKubernetesResource returns a new resource.Resource for
// betterado_environment_resource_kubernetes.
func NewEnvironmentResourceKubernetesResource() resource.Resource {
	return &EnvironmentResourceKubernetes{}
}

// environmentResourceKubernetesModel is the Terraform state model.
type environmentResourceKubernetesModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.String `tfsdk:"project_id"`
	EnvironmentID     types.Int64  `tfsdk:"environment_id"`
	ServiceEndpointID types.String `tfsdk:"service_endpoint_id"`
	Name              types.String `tfsdk:"name"`
	Namespace         types.String `tfsdk:"namespace"`
	ClusterName       types.String `tfsdk:"cluster_name"`
	Tags              types.Set    `tfsdk:"tags"`
}

func (r *EnvironmentResourceKubernetes) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_environment_resource_kubernetes"
}

func (r *EnvironmentResourceKubernetes) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Kubernetes resource within an Azure DevOps Environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"environment_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the environment.",
				PlanModifiers: []planmodifier.Int64{
					requiresReplaceInt64(),
				},
			},
			"service_endpoint_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the service endpoint (Kubernetes service connection).",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Kubernetes resource.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"namespace": schema.StringAttribute{
				Required:    true,
				Description: "The Kubernetes namespace.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"cluster_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultString(""),
				Description: "The Kubernetes cluster name.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Tags to assign to the Kubernetes resource.",
				PlanModifiers: []planmodifier.Set{
					requiresReplaceSet(),
				},
			},
		},
	}
}

func (r *EnvironmentResourceKubernetes) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentResourceKubernetes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model environmentResourceKubernetesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpointID, err := uuid.Parse(model.ServiceEndpointID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid service_endpoint_id",
			fmt.Sprintf("parse service_endpoint_id as UUID: %+v", err))
		return
	}

	tags, diags := expandStringSetAttr(ctx, model.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	environmentID := int(model.EnvironmentID.ValueInt64())

	created, err := r.client.TaskAgentClient.AddKubernetesResourcExistingEndpoint(r.client.Ctx,
		taskagent.AddKubernetesResourceArgsExistingEndpoint{
			CreateParameters: &taskagent.KubernetesResourceCreateParametersExistingEndpoint{
				ClusterName:       converter.String(model.ClusterName.ValueString()),
				Name:              converter.String(model.Name.ValueString()),
				Namespace:         converter.String(model.Namespace.ValueString()),
				Tags:              &tags,
				ServiceEndpointId: &serviceEndpointID,
			},
			Project:       &projectID,
			EnvironmentId: &environmentID,
		})
	if err != nil {
		resp.Diagnostics.AddError("Error creating Kubernetes environment resource",
			fmt.Sprintf("creating Kubernetes resource in Azure DevOps: %+v", err))
		return
	}

	model.ID = types.StringValue(strconv.Itoa(*created.Id))

	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentResourceKubernetes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model environmentResourceKubernetesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update is not used — all attributes are ForceNew/RequiresReplace, so any change destroys and recreates.
func (r *EnvironmentResourceKubernetes) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported",
		"betterado_environment_resource_kubernetes does not support in-place updates; all attributes require resource replacement.")
}

func (r *EnvironmentResourceKubernetes) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model environmentResourceKubernetesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid resource ID",
			fmt.Sprintf("parse resource ID: %+v", err))
		return
	}

	projectID := model.ProjectID.ValueString()
	environmentID := int(model.EnvironmentID.ValueInt64())

	if err := r.client.TaskAgentClient.DeleteKubernetesResource(r.client.Ctx,
		taskagent.DeleteKubernetesResourceArgs{
			Project:       &projectID,
			EnvironmentId: &environmentID,
			ResourceId:    &resourceID,
		}); err != nil {
		resp.Diagnostics.AddError("Error deleting Kubernetes environment resource",
			fmt.Sprintf("deleting Kubernetes resource: %+v", err))
	}
}

// readIntoModel fetches the Kubernetes resource by ID and populates model fields.
// If the resource is not found (404), model.ID is set to null (triggers removal).
func (r *EnvironmentResourceKubernetes) readIntoModel(ctx context.Context, model *environmentResourceKubernetesModel) diag.Diagnostics {
	var diags diag.Diagnostics

	resourceID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid resource ID", fmt.Sprintf("parse resource ID: %+v", err))
		return diags
	}

	projectID := model.ProjectID.ValueString()
	environmentID := int(model.EnvironmentID.ValueInt64())

	fetched, err := r.client.TaskAgentClient.GetKubernetesResource(r.client.Ctx,
		taskagent.GetKubernetesResourceArgs{
			Project:       &projectID,
			EnvironmentId: &environmentID,
			ResourceId:    &resourceID,
		})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return diags
		}
		diags.AddError("Error reading Kubernetes environment resource",
			fmt.Sprintf("looking up Kubernetes resource with ID %d in environment %d project %s: %v",
				resourceID, environmentID, projectID, err))
		return diags
	}

	if fetched.Name != nil {
		model.Name = types.StringValue(*fetched.Name)
	}
	if fetched.Namespace != nil {
		model.Namespace = types.StringValue(*fetched.Namespace)
	}
	model.ClusterName = types.StringValue(converter.ToString(fetched.ClusterName, ""))

	if fetched.ServiceEndpointId != nil {
		model.ServiceEndpointID = types.StringValue(fetched.ServiceEndpointId.String())
	}

	if fetched.EnvironmentReference != nil && fetched.EnvironmentReference.Id != nil {
		model.EnvironmentID = types.Int64Value(int64(*fetched.EnvironmentReference.Id))
	}

	if fetched.Tags != nil {
		tagVals := make([]string, len(*fetched.Tags))
		copy(tagVals, *fetched.Tags)
		tagSet, d := flattenStringSetAttr(ctx, tagVals)
		diags.Append(d...)
		model.Tags = tagSet
	} else {
		model.Tags = types.SetValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

// expandStringSetAttr converts a types.Set of strings to a []string.
func expandStringSetAttr(ctx context.Context, s types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if s.IsNull() || s.IsUnknown() {
		return []string{}, diags
	}
	var elems []string
	diags.Append(s.ElementsAs(ctx, &elems, false)...)
	return elems, diags
}

// flattenStringSetAttr converts a []string to a types.Set.
func flattenStringSetAttr(_ context.Context, vals []string) (types.Set, diag.Diagnostics) {
	elems := make([]attr.Value, len(vals))
	for i, v := range vals {
		elems[i] = types.StringValue(v)
	}
	return types.SetValue(types.StringType, elems)
}
