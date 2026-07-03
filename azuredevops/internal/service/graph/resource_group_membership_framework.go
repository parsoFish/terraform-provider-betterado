package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	schemavalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// ── Default helpers ───────────────────────────────────────────────────────────

// staticStringDefault returns a defaults.String that always sets a fixed string value.
type staticStringDefault struct{ value string }

func staticModeDefault(v string) defaults.String { return staticStringDefault{value: v} }

func (d staticStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d staticStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%s`", d.value)
}

func (d staticStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// compile-time interface checks
var (
	_ resource.Resource              = (*GroupMembershipResource)(nil)
	_ resource.ResourceWithConfigure = (*GroupMembershipResource)(nil)
)

// GroupMembershipResource is the terraform-plugin-framework implementation of
// betterado_group_membership. The SDKv2 ResourceGroupMembership() export is
// removed from provider.go and replaced by this.
//
// Note: the shared helper functions (applyMembershipUpdate, addMembers,
// removeMembers, expandGroupMembers, getGroupMemberships, getGroupMembershipSet)
// remain in resource_group_membership.go because resource_group.go also uses them.
type GroupMembershipResource struct {
	client *client.AggregatedClient
}

// NewGroupMembershipResource returns a new resource.Resource for betterado_group_membership.
func NewGroupMembershipResource() resource.Resource {
	return &GroupMembershipResource{}
}

// groupMembershipResourceModel maps the schema data to Go types.
type groupMembershipResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Group   types.String `tfsdk:"group"`
	Mode    types.String `tfsdk:"mode"`
	Members types.Set    `tfsdk:"members"`
}

// Metadata sets the type name for this resource.
func (r *GroupMembershipResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_group_membership"
}

// Schema defines the Terraform schema for betterado_group_membership.
func (r *GroupMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages group memberships in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					groupUseStateForUnknown(),
				},
			},
			"group": schema.StringAttribute{
				Required:    true,
				Description: "The descriptor of the group to manage memberships for.",
				PlanModifiers: []planmodifier.String{
					groupRequiresReplace(),
				},
				Validators: []schemavalidator.String{
					stringNotEmpty(),
				},
			},
			"mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  staticModeDefault("add"),
				Validators: []schemavalidator.String{
					stringOneOfCaseInsensitive("add", "overwrite"),
				},
				Description: "The mode to use when managing memberships. Valid values are 'add' and 'overwrite'. Defaults to 'add'.",
			},
			"members": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "A set of descriptors of subjects (groups/users) to make members of the group.",
			},
		},
	}
}

// Configure stores the AggregatedClient for use in CRUD methods.
func (r *GroupMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create implements the create CRUD operation for betterado_group_membership.
func (r *GroupMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupMembershipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := plan.Group.ValueString()
	mode := plan.Mode.ValueString()

	var membersToAdd []string
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &membersToAdd, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// In overwrite mode, compute the set of members to remove (those that exist
	// in ADO but are not in the desired set).
	membersToRemove := []string{}
	if strings.EqualFold("overwrite", mode) {
		existing, err := getGroupMemberships(r.client, group)
		if err != nil {
			resp.Diagnostics.AddError("Failed to read existing memberships", fmt.Sprintf("reading group memberships during create: %+v", err))
			return
		}
		desiredSet := toStringSet(membersToAdd)
		for _, m := range *existing {
			if !desiredSet[*m.MemberDescriptor] {
				membersToRemove = append(membersToRemove, *m.MemberDescriptor)
			}
		}
	}

	toAdd := expandGroupMembersFromList(group, membersToAdd)
	toRemove := expandGroupMembersFromList(group, membersToRemove)
	if err := applyMembershipUpdate(r.client, &toAdd, &toRemove); err != nil {
		resp.Diagnostics.AddError("Failed to apply membership update", fmt.Sprintf("adding group memberships during create: %+v", err))
		return
	}

	if err := r.waitForMembershipSync(group, membersToAdd, membersToRemove, 2); err != nil {
		resp.Diagnostics.AddError("Timeout waiting for membership sync", err.Error())
		return
	}

	// Use the group descriptor as the resource ID (memberships are tied to the group).
	plan.ID = types.StringValue(group)

	newState, diags := r.readMembershipsIntoModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Read reads the current membership state from ADO.
func (r *GroupMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, diags := r.readMembershipsIntoModel(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Update updates the membership set in ADO.
func (r *GroupMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state groupMembershipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := state.ID.ValueString()

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

	toAddList := expandGroupMembersFromList(group, toAdd)
	toRemoveList := expandGroupMembersFromList(group, toRemove)
	if err := applyMembershipUpdate(r.client, &toAddList, &toRemoveList); err != nil {
		resp.Diagnostics.AddError("Failed to update group memberships", err.Error())
		return
	}

	if err := r.waitForMembershipSync(group, toAdd, toRemove, 3); err != nil {
		resp.Diagnostics.AddError("Timeout waiting for membership sync", err.Error())
		return
	}

	plan.ID = state.ID
	newState, diags := r.readMembershipsIntoModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Delete removes the memberships tracked in state from the group.
func (r *GroupMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := state.ID.ValueString()

	var membersToRemove []string
	resp.Diagnostics.Append(state.Members.ElementsAs(ctx, &membersToRemove, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	toRemove := expandGroupMembersFromList(group, membersToRemove)
	if err := removeMembers(r.client, &toRemove); err != nil {
		resp.Diagnostics.AddError("Failed to remove group memberships", fmt.Sprintf("removing group memberships during delete: %+v", err))
		return
	}

	if err := r.waitForMembershipSync(group, nil, membersToRemove, 2); err != nil {
		resp.Diagnostics.AddError("Timeout waiting for membership sync after delete", err.Error())
		return
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// readMembershipsIntoModel reads the current memberships from ADO and populates
// a groupMembershipResourceModel, respecting the mode setting.
func (r *GroupMembershipResource) readMembershipsIntoModel(ctx context.Context, current groupMembershipResourceModel) (groupMembershipResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	group := current.Group.ValueString()
	mode := current.Mode.ValueString()

	actualMemberships, err := getGroupMemberships(r.client, group)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			// Group gone — return current state unchanged; caller can remove resource.
			return current, diags
		}
		diags.AddError("Failed to read group memberships", fmt.Sprintf("reading group memberships during read: %+v", err))
		return groupMembershipResourceModel{}, diags
	}

	// Determine which members to track in state based on mode:
	// - overwrite: track ALL members ADO returns
	// - add: track ONLY members that are in the current state (we only manage
	//   those we explicitly added)
	var stateMembers []string
	memberDiags := current.Members.ElementsAs(ctx, &stateMembers, false)
	diags.Append(memberDiags...)
	if diags.HasError() {
		return groupMembershipResourceModel{}, diags
	}
	stateMembersSet := toStringSet(stateMembers)

	members := []string{}
	for _, m := range *actualMemberships {
		if strings.EqualFold("overwrite", mode) || stateMembersSet[*m.MemberDescriptor] {
			members = append(members, *m.MemberDescriptor)
		}
	}

	memberVals := make([]attr.Value, len(members))
	for i, mbr := range members {
		memberVals[i] = types.StringValue(mbr)
	}
	membersSet, setDiags := types.SetValue(types.StringType, memberVals)
	diags.Append(setDiags...)
	if diags.HasError() {
		return groupMembershipResourceModel{}, diags
	}

	return groupMembershipResourceModel{
		ID:      current.ID,
		Group:   current.Group,
		Mode:    current.Mode,
		Members: membersSet,
	}, diags
}

// waitForMembershipSync polls ADO until the desired membership changes are reflected
// or the timeout is reached.
func (r *GroupMembershipResource) waitForMembershipSync(group string, toAdd, toRemove []string, continuousOccurrences int) error {
	toAddSet := toStringSet(toAdd)
	toRemoveSet := toStringSet(toRemove)

	stateConf := &retry.StateChangeConf{
		Pending: []string{"Waiting"},
		Target:  []string{"Synched"},
		Refresh: func() (interface{}, string, error) {
			actual, err := getGroupMemberships(r.client, group)
			if err != nil {
				return nil, "", fmt.Errorf("reading group memberships: %+v", err)
			}
			actualSet := map[string]bool{}
			for _, m := range *actual {
				actualSet[*m.MemberDescriptor] = true
			}

			syncState := "Synched"
			// If any member we want to ADD is not yet present — still waiting.
			for mbr := range toAddSet {
				if !actualSet[mbr] {
					syncState = "Waiting"
					break
				}
			}
			// If any member we want to REMOVE is still present — still waiting.
			if syncState == "Synched" {
				for mbr := range toRemoveSet {
					if actualSet[mbr] {
						syncState = "Waiting"
						break
					}
				}
			}

			return syncState, syncState, nil
		},
		Timeout:                   60 * time.Minute,
		MinTimeout:                5 * time.Second,
		Delay:                     5 * time.Second,
		ContinuousTargetOccurence: continuousOccurrences,
	}

	if _, err := stateConf.WaitForStateContext(r.client.Ctx); err != nil {
		return fmt.Errorf("error waiting for DevOps synching memberships for group [%s]: %+v", group, err)
	}
	return nil
}
