package build

// resource_resource_authorization_framework.go — terraform-plugin-framework implementation of
// betterado_resource_authorization (deprecated).
//
// The SDKv2 ResourceResourceAuthorization() in resource_resource_authorization.go remains
// in place for state-file compatibility. The mux serves new apply/plan calls through this
// framework resource once it is registered in framework_provider.go Resources().
// The SDKv2 "betterado_resource_authorization" entry in provider.go ResourcesMap is removed
// (commented with a migration note) to avoid "Invalid Provider Server Combination".
//
// This resource is deprecated; use betterado_pipeline_authorization instead.
//
// No build tag on this production file — only _test.go files carry //go:build.
//
// Sub-packages like stringplanmodifier / stringdefault are not vendored, so we
// implement the minimal plan-modifier/default interfaces inline (same pattern as
// resource_build_folder_framework.go).

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = &ResourceAuthorizationResource{}
	_ resource.ResourceWithConfigure = &ResourceAuthorizationResource{}
)

const resourceAuthorizationDeprecationMessage = "This resource will be deprecated and removed in the future. Please use `betterado_pipeline_authorization` instead."

// ResourceAuthorizationResource is the framework resource for betterado_resource_authorization.
type ResourceAuthorizationResource struct {
	client *client.AggregatedClient
}

// NewResourceAuthorizationResource returns a new framework resource.Resource.
func NewResourceAuthorizationResource() resource.Resource {
	return &ResourceAuthorizationResource{}
}

// ── Model ────────────────────────────────────────────────────────────────────

type resourceAuthorizationFrameworkModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	ResourceID   types.String `tfsdk:"resource_id"`
	DefinitionID types.Int64  `tfsdk:"definition_id"`
	Type         types.String `tfsdk:"type"`
	Authorized   types.Bool   `tfsdk:"authorized"`
}

// ── Metadata / Schema ────────────────────────────────────────────────────────

func (r *ResourceAuthorizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_authorization"
}

func (r *ResourceAuthorizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:        resourceAuthorizationDeprecationMessage,
		DeprecationMessage: resourceAuthorizationDeprecationMessage,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the resource authorization.",
				PlanModifiers: []planmodifier.String{
					raUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					raRequiresReplaceString(),
				},
			},
			"resource_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the resource.",
				Validators: []validator.String{
					paNotWhitespaceValidator{},
				},
			},
			"definition_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The ID of the build definition.",
				Validators: []validator.Int64{
					paInt64AtLeastValidator{min: 1},
				},
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The type of the resource. Valid values: endpoint, queue, variablegroup. Default: endpoint.",
				Default:     raStaticStringDefault("endpoint"),
				Validators: []validator.String{
					raTypeValidator{},
				},
			},
			"authorized": schema.BoolAttribute{
				Required:    true,
				Description: "Indicates whether the resource is authorized for use.",
			},
		},
	}
}

// ── Validators ───────────────────────────────────────────────────────────────

type raTypeValidator struct{}

func (v raTypeValidator) Description(_ context.Context) string {
	return "type must be one of: endpoint, queue, variablegroup"
}
func (v raTypeValidator) MarkdownDescription(_ context.Context) string {
	return v.Description(context.Background())
}
func (v raTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	valid := []string{"endpoint", "queue", "variablegroup"}
	val := strings.ToLower(req.ConfigValue.ValueString())
	for _, v := range valid {
		if val == v {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Invalid type",
		fmt.Sprintf("type must be one of %v, got: %q", valid, req.ConfigValue.ValueString()))
}

// ── Inline plan-modifiers and defaults ───────────────────────────────────────
// Mirrors the pattern in resource_build_folder_framework.go.
// Sub-packages (stringplanmodifier, stringdefault) are not vendored.

type raUseStateForUnknownImpl struct{}

func raUseStateForUnknown() planmodifier.String { return raUseStateForUnknownImpl{} }

func (raUseStateForUnknownImpl) Description(_ context.Context) string {
	return "use prior state value for unknown"
}
func (raUseStateForUnknownImpl) MarkdownDescription(_ context.Context) string {
	return "use prior state value for unknown"
}
func (raUseStateForUnknownImpl) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type raRequiresReplaceStringImpl struct{}

func raRequiresReplaceString() planmodifier.String { return raRequiresReplaceStringImpl{} }

func (raRequiresReplaceStringImpl) Description(_ context.Context) string {
	return "requires replacement if changed"
}
func (raRequiresReplaceStringImpl) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}
func (raRequiresReplaceStringImpl) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type raStaticStringDefaultImpl struct{ value string }

func raStaticStringDefault(v string) defaults.String { return raStaticStringDefaultImpl{value: v} }

func (d raStaticStringDefaultImpl) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d raStaticStringDefaultImpl) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%q`", d.value)
}
func (d raStaticStringDefaultImpl) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Configure ────────────────────────────────────────────────────────────────

func (r *ResourceAuthorizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData))
		return
	}
	r.client = agg
}

// ── CRUD ─────────────────────────────────────────────────────────────────────

func (r *ResourceAuthorizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceAuthorizationFrameworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceRef, projectID, definitionID := raExpandAuthorizedResource(plan)
	if err := raSendAuthorizedResourceToAPI(ctx, r.client, resourceRef, projectID, definitionID); err != nil {
		resp.Diagnostics.AddError("Error creating resource authorization", fmt.Sprintf("creating authorized resource: %+v", err))
		return
	}

	plan.ID = types.StringValue(*resourceRef.Id)
	r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ResourceAuthorizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceAuthorizationFrameworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readIntoModel(ctx, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceAuthorizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceAuthorizationFrameworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceRef, projectID, definitionID := raExpandAuthorizedResource(plan)
	if err := raSendAuthorizedResourceToAPI(ctx, r.client, resourceRef, projectID, definitionID); err != nil {
		resp.Diagnostics.AddError("Error updating resource authorization", fmt.Sprintf("updating authorized resource: %+v", err))
		return
	}

	r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ResourceAuthorizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceAuthorizationFrameworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Deletion works by setting authorized to false.
	state.Authorized = types.BoolValue(false)
	resourceRef, projectID, definitionID := raExpandAuthorizedResource(state)
	if err := raSendAuthorizedResourceToAPI(ctx, r.client, resourceRef, projectID, definitionID); err != nil {
		resp.Diagnostics.AddError("Error deleting resource authorization", fmt.Sprintf("deleting authorized resource: %+v", err))
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func raExpandAuthorizedResource(m resourceAuthorizationFrameworkModel) (*build.DefinitionResourceReference, string, int) {
	resourceRef := build.DefinitionResourceReference{
		Authorized: converter.Bool(m.Authorized.ValueBool()),
		Id:         converter.String(m.ResourceID.ValueString()),
		Name:       nil,
		Type:       converter.String(m.Type.ValueString()),
	}
	defID := 0
	if !m.DefinitionID.IsNull() && !m.DefinitionID.IsUnknown() {
		defID = int(m.DefinitionID.ValueInt64())
	}
	return &resourceRef, m.ProjectID.ValueString(), defID
}

func raSendAuthorizedResourceToAPI(ctx context.Context, clients *client.AggregatedClient, resourceRef *build.DefinitionResourceReference, projectID string, definitionID int) error {
	var err error
	if definitionID == 0 {
		_, err = clients.BuildClient.AuthorizeProjectResources(ctx, build.AuthorizeProjectResourcesArgs{
			Resources: &[]build.DefinitionResourceReference{*resourceRef},
			Project:   &projectID,
		})
	} else {
		_, err = clients.BuildClient.AuthorizeDefinitionResources(ctx, build.AuthorizeDefinitionResourcesArgs{
			Resources:    &[]build.DefinitionResourceReference{*resourceRef},
			Project:      &projectID,
			DefinitionId: &definitionID,
		})
	}
	return err
}

func (r *ResourceAuthorizationResource) readIntoModel(ctx context.Context, m *resourceAuthorizationFrameworkModel) {
	projectID := m.ProjectID.ValueString()
	definitionID := 0
	if !m.DefinitionID.IsNull() && !m.DefinitionID.IsUnknown() {
		definitionID = int(m.DefinitionID.ValueInt64())
	}
	authorized := m.Authorized.ValueBool()
	resourceRef := build.DefinitionResourceReference{
		Authorized: converter.Bool(authorized),
		Id:         converter.String(m.ResourceID.ValueString()),
		Name:       nil,
		Type:       converter.String(m.Type.ValueString()),
	}

	if definitionID == 0 {
		if authorized {
			resourceRefs, err := r.client.BuildClient.GetProjectResources(ctx, build.GetProjectResourcesArgs{
				Project: &projectID,
				Type:    resourceRef.Type,
				Id:      resourceRef.Id,
			})
			if err != nil || resourceRefs == nil || len(*resourceRefs) == 0 {
				m.ID = types.StringValue("")
				return
			}
			ref := (*resourceRefs)[0]
			m.ResourceID = types.StringValue(converter.ToString(ref.Id, ""))
			m.Type = types.StringValue(converter.ToString(ref.Type, ""))
			if ref.Authorized != nil {
				m.Authorized = types.BoolValue(*ref.Authorized)
			}
			m.ID = types.StringValue(converter.ToString(ref.Id, ""))
		}
		// authorized=false: flatten from user-provided plan (no server read).
	} else {
		resourceRefs, err := r.client.BuildClient.GetDefinitionResources(ctx, build.GetDefinitionResourcesArgs{
			Project:      &projectID,
			DefinitionId: &definitionID,
		})
		if err != nil || resourceRefs == nil || len(*resourceRefs) == 0 {
			m.ID = types.StringValue("")
			return
		}
		for _, ref := range *resourceRefs {
			if ref.Id != nil && resourceRef.Id != nil && *ref.Id == *resourceRef.Id {
				m.ResourceID = types.StringValue(converter.ToString(ref.Id, ""))
				m.Type = types.StringValue(converter.ToString(ref.Type, ""))
				if ref.Authorized != nil {
					m.Authorized = types.BoolValue(*ref.Authorized)
				}
				m.ID = types.StringValue(converter.ToString(ref.Id, ""))
				return
			}
		}
	}
}
