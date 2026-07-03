package graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	schemavalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// groupRequiresReplaceModifier marks a string attribute for replacement when
// its value changes (equivalent to ForceNew in SDKv2).
type groupRequiresReplaceModifier struct{}

func groupRequiresReplace() planmodifier.String { return groupRequiresReplaceModifier{} }
func (m groupRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m groupRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m groupRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// groupUseStateForUnknownModifier copies the prior state string value when the
// plan value is unknown (prevents perpetual diffs for computed-only attributes).
type groupUseStateForUnknownModifier struct{}

func groupUseStateForUnknown() planmodifier.String { return groupUseStateForUnknownModifier{} }
func (m groupUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m groupUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m groupUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// groupSetUseStateForUnknownModifier copies the prior state set value when the
// plan value is unknown.
type groupSetUseStateForUnknownModifier struct{}

func groupSetUseStateForUnknown() planmodifier.Set { return groupSetUseStateForUnknownModifier{} }
func (m groupSetUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m groupSetUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m groupSetUseStateForUnknownModifier) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── Resource struct ───────────────────────────────────────────────────────────

// compile-time interface checks
var (
	_ resource.Resource                     = (*GroupResource)(nil)
	_ resource.ResourceWithConfigure        = (*GroupResource)(nil)
	_ resource.ResourceWithConfigValidators = (*GroupResource)(nil)
)

// GroupResource is the terraform-plugin-framework implementation of betterado_group.
type GroupResource struct {
	client *client.AggregatedClient
}

// NewGroupResource returns a new resource.Resource for betterado_group.
func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

// groupResourceModel maps the schema data to Go types.
type groupResourceModel struct {
	// ID (descriptor) is used as the Terraform resource ID.
	ID            types.String `tfsdk:"id"`
	Scope         types.String `tfsdk:"scope"`
	OriginID      types.String `tfsdk:"origin_id"`
	Mail          types.String `tfsdk:"mail"`
	DisplayName   types.String `tfsdk:"display_name"`
	Description   types.String `tfsdk:"description"`
	Members       types.Set    `tfsdk:"members"`
	URL           types.String `tfsdk:"url"`
	Origin        types.String `tfsdk:"origin"`
	SubjectKind   types.String `tfsdk:"subject_kind"`
	Domain        types.String `tfsdk:"domain"`
	PrincipalName types.String `tfsdk:"principal_name"`
	Descriptor    types.String `tfsdk:"descriptor"`
	GroupID       types.String `tfsdk:"group_id"`
}

// Metadata sets the type name for the resource.
func (r *GroupResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_group"
}

// Schema defines the terraform schema for the resource.
func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					groupUseStateForUnknown(),
				},
			},
			"scope": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					groupRequiresReplace(),
				},
				Validators: []schemavalidator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The GUID (UUID) of the project that the group is scoped to. If not set, the group is global.",
			},
			"origin_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					groupRequiresReplace(),
					groupUseStateForUnknown(),
				},
				Validators: []schemavalidator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The OriginID for the group when creating from an external source. Conflicts with mail, display_name, and scope.",
			},
			"mail": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					groupRequiresReplace(),
					groupUseStateForUnknown(),
				},
				Validators: []schemavalidator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The mail address for the group when creating from an email address. Conflicts with origin_id, display_name, and scope.",
			},
			"display_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []schemavalidator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The display name of the group. Conflicts with origin_id and mail.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the group.",
			},
			"members": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					groupSetUseStateForUnknown(),
				},
				Description: "A set of descriptors of subjects (groups/users) to add as members.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "URL to the group entity.",
			},
			"origin": schema.StringAttribute{
				Computed:    true,
				Description: "The identity store that is the origin for the group.",
			},
			"subject_kind": schema.StringAttribute{
				Computed:    true,
				Description: "Subject kind of the group.",
			},
			"domain": schema.StringAttribute{
				Computed:    true,
				Description: "The domain of the group.",
			},
			"principal_name": schema.StringAttribute{
				Computed:    true,
				Description: "The principal name of the group.",
			},
			"descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor is the primary way to reference the graph subject while the system is running.",
			},
			"group_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the group (storage key UUID).",
			},
		},
	}
}

// ConfigValidators returns resource-level validators that enforce SDKv2-parity
// ConflictsWith constraints on origin_id, mail, and display_name:
//   - origin_id, mail, and display_name all conflict with each other
//   - origin_id conflicts with scope
//   - mail conflicts with scope
func (r *GroupResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		// origin_id, mail, and display_name all conflict with each other.
		resourcevalidator.Conflicting(
			path.MatchRoot("origin_id"),
			path.MatchRoot("mail"),
			path.MatchRoot("display_name"),
		),
		// origin_id and mail additionally conflict with scope.
		resourcevalidator.Conflicting(
			path.MatchRoot("origin_id"),
			path.MatchRoot("scope"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("mail"),
			path.MatchRoot("scope"),
		),
	}
}

// Configure stores the AggregatedClient for use in CRUD methods.
func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = agg
}

// Create creates a new ADO group.
func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve scope -> scopeDescriptor
	var scopeDescriptor *string
	if !plan.Scope.IsNull() && !plan.Scope.IsUnknown() && plan.Scope.ValueString() != "" {
		scopeUid, err := uuid.Parse(plan.Scope.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid scope UUID", fmt.Sprintf("Parsing scope: %+v", err))
			return
		}
		desc, err := r.client.GraphClient.GetDescriptor(r.client.Ctx, graph.GetDescriptorArgs{
			StorageKey: &scopeUid,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to get scope descriptor", err.Error())
			return
		}
		scopeDescriptor = desc.Value
	}

	var grp *graph.GraphGroup

	switch {
	case !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() && plan.DisplayName.ValueString() != "":
		description := ""
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			description = plan.Description.ValueString()
		}
		g, err := r.client.GraphClient.CreateGroupVsts(r.client.Ctx, graph.CreateGroupVstsArgs{
			CreationContext: &graph.GraphGroupVstsCreationContext{
				DisplayName: converter.String(plan.DisplayName.ValueString()),
				Description: converter.String(description),
			},
			ScopeDescriptor: scopeDescriptor,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create group by display name", err.Error())
			return
		}
		grp = g

	case !plan.OriginID.IsNull() && !plan.OriginID.IsUnknown() && plan.OriginID.ValueString() != "":
		g, err := r.client.GraphClient.CreateGroupOriginId(r.client.Ctx, graph.CreateGroupOriginIdArgs{
			CreationContext: &graph.GraphGroupOriginIdCreationContext{
				OriginId: converter.String(plan.OriginID.ValueString()),
			},
			ScopeDescriptor: scopeDescriptor,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create group by origin_id", err.Error())
			return
		}
		grp = g

	case !plan.Mail.IsNull() && !plan.Mail.IsUnknown() && plan.Mail.ValueString() != "":
		g, err := r.client.GraphClient.CreateGroupMailAddress(r.client.Ctx, graph.CreateGroupMailAddressArgs{
			CreationContext: &graph.GraphGroupMailAddressCreationContext{
				MailAddress: converter.String(plan.Mail.ValueString()),
			},
			ScopeDescriptor: scopeDescriptor,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create group by mail", err.Error())
			return
		}
		grp = g

	default:
		resp.Diagnostics.AddError("Invalid group config", "One of display_name, origin_id, or mail must be set")
		return
	}

	// Add members if specified.
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		var memberList []string
		resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &memberList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(memberList) > 0 {
			members := expandGroupMembersFromList(*grp.Descriptor, memberList)
			if err := addMembers(r.client, &members); err != nil {
				resp.Diagnostics.AddError("Failed to add members", fmt.Sprintf("adding group memberships during create: %+v", err))
				return
			}
		}
	}

	// Read back the group and set state.
	newState, stateDiags := r.readGroupIntoModel(*grp.Descriptor, plan)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Read reads the current state of the group from ADO.
func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptor := state.ID.ValueString()

	grp, err := r.client.GraphClient.GetGroup(r.client.Ctx, graph.GetGroupArgs{
		GroupDescriptor: converter.String(descriptor),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read group", err.Error())
		return
	}

	if grp.IsDeleted != nil && *grp.IsDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	newState, stateDiags := r.readGroupIntoModel(descriptor, state)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Update updates the group in ADO.
func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptor := state.ID.ValueString()

	// Update display_name / description if changed.
	if !plan.DisplayName.Equal(state.DisplayName) || !plan.Description.Equal(state.Description) {
		if err := r.patchGroup(descriptor, plan); err != nil {
			resp.Diagnostics.AddError("Failed to update group", err.Error())
			return
		}
	}

	// Update members if changed.
	if !plan.Members.Equal(state.Members) {
		var planMembers, stateMembers []string
		resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &planMembers, false)...)
		resp.Diagnostics.Append(state.Members.ElementsAs(ctx, &stateMembers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		planSet := toStringSet(planMembers)
		stateSet := toStringSet(stateMembers)

		var toAdd, toRemove []string
		for mbr := range planSet {
			if !stateSet[mbr] {
				toAdd = append(toAdd, mbr)
			}
		}
		for mbr := range stateSet {
			if !planSet[mbr] {
				toRemove = append(toRemove, mbr)
			}
		}

		toAddList := expandGroupMembersFromList(descriptor, toAdd)
		toRemoveList := expandGroupMembersFromList(descriptor, toRemove)
		if err := applyMembershipUpdate(r.client, &toAddList, &toRemoveList); err != nil {
			resp.Diagnostics.AddError("Failed to update members", err.Error())
			return
		}
	}

	// Read back and set state.
	newState, stateDiags := r.readGroupIntoModel(descriptor, plan)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Delete removes the group from ADO.
func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptor := state.ID.ValueString()

	err := r.client.GraphClient.DeleteGroup(r.client.Ctx, graph.DeleteGroupArgs{
		GroupDescriptor: converter.String(descriptor),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			// already deleted -- treat as success per AC3
			return
		}
		resp.Diagnostics.AddError("Failed to delete group", err.Error())
		return
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// readGroupIntoModel reads the group from ADO and populates a groupResourceModel.
// It preserves user-supplied values from current where appropriate.
func (r *GroupResource) readGroupIntoModel(descriptor string, current groupResourceModel) (groupResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	grp, err := r.client.GraphClient.GetGroup(r.client.Ctx, graph.GetGroupArgs{
		GroupDescriptor: converter.String(descriptor),
	})
	if err != nil {
		diags.AddError("Failed to read group", err.Error())
		return groupResourceModel{}, diags
	}

	// Read members.
	members, err := groupReadMembers(descriptor, r.client)
	if err != nil {
		diags.AddError("Failed to read group members", err.Error())
		return groupResourceModel{}, diags
	}

	memberVals := make([]attr.Value, len(*members))
	for i, mbr := range *members {
		memberVals[i] = types.StringValue(*mbr.MemberDescriptor)
	}
	membersSet, setDiags := types.SetValue(types.StringType, memberVals)
	diags.Append(setDiags...)
	if diags.HasError() {
		return groupResourceModel{}, diags
	}

	// Fetch storage key for group_id.
	var groupID types.String
	storageKey, skErr := r.client.GraphClient.GetStorageKey(r.client.Ctx, graph.GetStorageKeyArgs{
		SubjectDescriptor: grp.Descriptor,
	})
	if skErr == nil && storageKey.Value != nil {
		groupID = types.StringValue(storageKey.Value.String())
	} else {
		groupID = types.StringNull()
	}

	// Derive scope from domain.
	var scope types.String
	var domainPID string
	if grp.Domain != nil {
		domainPID = domain2ProjectID(*grp.Domain)
	}
	switch {
	case domainPID != "":
		scope = types.StringValue(domainPID)
	case !current.Scope.IsNull() && !current.Scope.IsUnknown():
		scope = current.Scope
	default:
		scope = types.StringNull()
	}

	m := groupResourceModel{
		ID:         types.StringValue(descriptor),
		Descriptor: types.StringValue(descriptor),
		Scope:      scope,
		GroupID:    groupID,
		Members:    membersSet,
	}

	if grp.DisplayName != nil {
		m.DisplayName = types.StringValue(*grp.DisplayName)
	} else {
		m.DisplayName = types.StringNull()
	}
	// description is Optional (not Computed): if ADO returns "" and the user
	// did not set description, preserve null to avoid a plan-vs-state inconsistency.
	switch {
	case grp.Description != nil && *grp.Description != "":
		m.Description = types.StringValue(*grp.Description)
	case !current.Description.IsNull() && !current.Description.IsUnknown():
		// User explicitly set description (possibly to ""); keep their value.
		m.Description = current.Description
	default:
		m.Description = types.StringNull()
	}
	if grp.Url != nil {
		m.URL = types.StringValue(*grp.Url)
	} else {
		m.URL = types.StringNull()
	}
	if grp.Origin != nil {
		m.Origin = types.StringValue(*grp.Origin)
	} else {
		m.Origin = types.StringNull()
	}
	switch {
	case grp.OriginId != nil:
		m.OriginID = types.StringValue(*grp.OriginId)
	case !current.OriginID.IsNull() && !current.OriginID.IsUnknown():
		m.OriginID = current.OriginID
	default:
		m.OriginID = types.StringNull()
	}
	if grp.SubjectKind != nil {
		m.SubjectKind = types.StringValue(*grp.SubjectKind)
	} else {
		m.SubjectKind = types.StringNull()
	}
	if grp.Domain != nil {
		m.Domain = types.StringValue(*grp.Domain)
	} else {
		m.Domain = types.StringNull()
	}
	switch {
	case grp.MailAddress != nil:
		m.Mail = types.StringValue(*grp.MailAddress)
	case !current.Mail.IsNull() && !current.Mail.IsUnknown():
		m.Mail = current.Mail
	default:
		m.Mail = types.StringNull()
	}
	if grp.PrincipalName != nil {
		m.PrincipalName = types.StringValue(*grp.PrincipalName)
	} else {
		m.PrincipalName = types.StringNull()
	}

	return m, diags
}

// patchGroup sends a JSON-patch update to change display_name / description.
func (r *GroupResource) patchGroup(descriptor string, plan groupResourceModel) error {
	var patchOps []webapi.JsonPatchOperation
	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		patchOps = append(patchOps, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Replace,
			From:  nil,
			Path:  converter.String("/displayName"),
			Value: plan.DisplayName.ValueString(),
		})
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		patchOps = append(patchOps, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Replace,
			From:  nil,
			Path:  converter.String("/description"),
			Value: plan.Description.ValueString(),
		})
	}
	if len(patchOps) == 0 {
		return nil
	}

	_, err := r.client.GraphClient.UpdateGroup(r.client.Ctx, graph.UpdateGroupArgs{
		GroupDescriptor: converter.String(descriptor),
		PatchDocument:   &patchOps,
	})
	return err
}

// toStringSet converts a slice to a set map for diffing.
func toStringSet(s []string) map[string]bool {
	result := make(map[string]bool, len(s))
	for _, v := range s {
		result[v] = true
	}
	return result
}

// expandGroupMembersFromList converts a list of member descriptors to GraphMembership entries.
func expandGroupMembersFromList(groupDescriptor string, members []string) []graph.GraphMembership {
	result := make([]graph.GraphMembership, len(members))
	for i, mbr := range members {
		mbr := mbr
		result[i] = graph.GraphMembership{
			ContainerDescriptor: &groupDescriptor,
			MemberDescriptor:    &mbr,
		}
	}
	return result
}
