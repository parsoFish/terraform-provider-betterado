package graph

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance.
var _ datasource.DataSource = &usersDataSource{}

// usersDataSource is the terraform-plugin-framework data source for betterado_users.
type usersDataSource struct {
	client *client.AggregatedClient
}

// NewUsersDataSource returns a new framework data source for betterado_users.
func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

// usersDataSourceModel is the tfsdk model for the data source.
type usersDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	PrincipalName types.String `tfsdk:"principal_name"`
	Origin        types.String `tfsdk:"origin"`
	OriginID      types.String `tfsdk:"origin_id"`
	SubjectTypes  types.Set    `tfsdk:"subject_types"`
	Users         types.Set    `tfsdk:"users"`
}

// userItemModel represents one user entry in the users set.
type userItemModel struct {
	ID            types.String `tfsdk:"id"`
	Descriptor    types.String `tfsdk:"descriptor"`
	PrincipalName types.String `tfsdk:"principal_name"`
	Origin        types.String `tfsdk:"origin"`
	OriginID      types.String `tfsdk:"origin_id"`
	DisplayName   types.String `tfsdk:"display_name"`
	MailAddress   types.String `tfsdk:"mail_address"`
}

// userItemAttrTypes returns the attr.Type map for userItemModel.
var userItemAttrTypes = map[string]attr.Type{
	"id":             types.StringType,
	"descriptor":     types.StringType,
	"principal_name": types.StringType,
	"origin":         types.StringType,
	"origin_id":      types.StringType,
	"display_name":   types.StringType,
	"mail_address":   types.StringType,
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up Azure DevOps users with optional filters.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A hash of the returned user descriptors.",
			},
			"principal_name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by principal name (conflicts with origin).",
			},
			"origin": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by origin (e.g. aad, msa). Conflicts with principal_name.",
			},
			"origin_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by origin ID. Conflicts with principal_name.",
			},
			"subject_types": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Optional set of subject types to filter (e.g. [\"aad\"]).",
			},
			"users": schema.SetNestedAttribute{
				Computed:    true,
				Description: "The set of matching users.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The storage key UUID of the user.",
						},
						"descriptor": schema.StringAttribute{
							Computed:    true,
							Description: "The descriptor of the user.",
						},
						"principal_name": schema.StringAttribute{
							Computed:    true,
							Description: "The principal name of the user.",
						},
						"origin": schema.StringAttribute{
							Computed:    true,
							Description: "The origin of the user.",
						},
						"origin_id": schema.StringAttribute{
							Computed:    true,
							Description: "The origin-specific identifier of the user.",
						},
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "The display name of the user.",
						},
						"mail_address": schema.StringAttribute{
							Computed:    true,
							Description: "The mail address of the user.",
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_users data source Read: provider client not configured")
		return
	}

	var model usersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	principalName := model.PrincipalName.ValueString()
	origin := model.Origin.ValueString()
	originID := model.OriginID.ValueString()

	// Build subject types slice.
	var subjectTypes []string
	if !model.SubjectTypes.IsNull() && !model.SubjectTypes.IsUnknown() {
		var strs []types.String
		resp.Diagnostics.Append(model.SubjectTypes.ElementsAs(ctx, &strs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, s := range strs {
			subjectTypes = append(subjectTypes, s.ValueString())
		}
	}

	// Paginate through all users.
	var allUsers []graph.GraphUser
	var currentToken string
	for hasMore := true; hasMore; {
		newUsers, latestToken, err := getUsersWithContinuationToken(d.client, &subjectTypes, currentToken)
		if err != nil {
			resp.Diagnostics.AddError("Read error", fmt.Sprintf("Listing users: %s", err))
			return
		}
		currentToken = latestToken
		hasMore = currentToken != ""

		// Apply filters.
		for _, usr := range newUsers {
			match := true
			if principalName != "" {
				match = usr.PrincipalName != nil && strings.EqualFold(*usr.PrincipalName, principalName)
			}
			if match && origin != "" {
				match = usr.Origin != nil && strings.EqualFold(*usr.Origin, origin)
			}
			if match && originID != "" {
				match = usr.OriginId != nil && strings.EqualFold(*usr.OriginId, originID)
			}
			if match {
				allUsers = append(allUsers, usr)
			}
		}
	}

	// Build user items with parallel GetStorageKey lookups (same concurrency
	// pattern as the SDKv2 addStorageKeyAsId, defaulting to 10 workers to
	// avoid sequential N+1 API calls that cause test timeouts in large orgs).
	const numWorkers = 10
	userItems := make([]userItemModel, len(allUsers))

	// Initialise all items from the Graph user fields first.
	for i, usr := range allUsers {
		userItems[i] = userItemModel{
			Descriptor:    strPtrToStringValue(usr.Descriptor),
			PrincipalName: strPtrToStringValue(usr.PrincipalName),
			Origin:        strPtrToStringValue(usr.Origin),
			OriginID:      strPtrToStringValue(usr.OriginId),
			DisplayName:   strPtrToStringValue(usr.DisplayName),
			MailAddress:   strPtrToStringValue(usr.MailAddress),
			ID:            types.StringValue(""),
		}
	}

	// Fan-out GetStorageKey calls over a bounded worker pool.
	// workResult carries the descriptor (used as a correlation key) back to the
	// collector so results can be written into the pre-allocated userItems slice.
	type workResult struct {
		descriptor string
		id         string
		err        error
	}

	// Buffer the channels at len(allUsers) so no goroutine blocks if the
	// consumer is slow — each slot maps to exactly one work item.
	bufSize := len(allUsers)
	if bufSize == 0 {
		bufSize = 1
	}
	workCh := make(chan string, bufSize)
	resultCh := make(chan workResult, bufSize)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for descriptor := range workCh {
				sk, err := d.client.GraphClient.GetStorageKey(d.client.Ctx, graph.GetStorageKeyArgs{
					SubjectDescriptor: converter.String(descriptor),
				})
				if err != nil {
					resultCh <- workResult{descriptor: descriptor, err: fmt.Errorf("getting storage key for user %q: %w", descriptor, err)}
					continue
				}
				id := ""
				if sk != nil && sk.Value != nil {
					id = sk.Value.String()
				}
				resultCh <- workResult{descriptor: descriptor, id: id}
			}
		}()
	}

	// Enqueue work items.
	descriptorToIdx := make(map[string]int, len(allUsers))
	for i, usr := range allUsers {
		if usr.Descriptor != nil && *usr.Descriptor != "" {
			workCh <- *usr.Descriptor
			descriptorToIdx[*usr.Descriptor] = i
		}
	}
	close(workCh)

	// Wait for workers to finish, then close result channel.
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results.
	var storageKeyErr error
	for res := range resultCh {
		if res.err != nil {
			storageKeyErr = res.err
			continue
		}
		if idx, ok := descriptorToIdx[res.descriptor]; ok {
			userItems[idx].ID = types.StringValue(res.id)
		}
	}
	if storageKeyErr != nil {
		resp.Diagnostics.AddError("Read error", storageKeyErr.Error())
		return
	}

	var descriptors []string
	for _, usr := range allUsers {
		if usr.Descriptor != nil {
			descriptors = append(descriptors, *usr.Descriptor)
		}
	}

	// Build set value.
	usersSet, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: userItemAttrTypes}, userItems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Compute stable ID from descriptor list.
	h := sha1.New()
	_, _ = h.Write([]byte(strings.Join(descriptors, "-")))
	model.ID = types.StringValue("users#" + base64.URLEncoding.EncodeToString(h.Sum(nil)))
	model.Users = usersSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
