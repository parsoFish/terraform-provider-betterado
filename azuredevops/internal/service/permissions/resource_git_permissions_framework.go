package permissions

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	securityhelper "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/permissions/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

var _ resource.Resource = &gitPermissionsFrameworkResource{}

type gitPermissionsFrameworkResource struct {
	client *client.AggregatedClient
}

// NewGitPermissionsResource returns a new framework resource for betterado_git_permissions.
func NewGitPermissionsResource() resource.Resource {
	return &gitPermissionsFrameworkResource{}
}

type gitPermissionsModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	RepositoryID types.String `tfsdk:"repository_id"`
	BranchName   types.String `tfsdk:"branch_name"`
	Principal    types.String `tfsdk:"principal"`
	Permissions  types.Map    `tfsdk:"permissions"`
	Replace      types.Bool   `tfsdk:"replace"`
}

func (r *gitPermissionsFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_permissions"
}

func (r *gitPermissionsFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Git repository ACL permissions in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					ppUseStateForUnknownString(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					ppRequiresReplaceString(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid UUID",
					),
				},
			},
			"repository_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					ppRequiresReplaceString(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid UUID",
					),
				},
			},
			"branch_name": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					ppRequiresReplaceString(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^$|\S`),
						"must not consist entirely of whitespace",
					),
				},
			},
			"principal": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					ppRequiresReplaceString(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`\S`),
						"must not be empty or consist entirely of whitespace",
					),
				},
			},
			"permissions": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"replace": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  ppStaticBool(true),
			},
		},
	}
}

func (r *gitPermissionsFrameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *gitPermissionsFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model gitPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyPermissions(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Create error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *gitPermissionsFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model gitPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sn, err := r.newSecurityNamespace(&model)
	if err != nil {
		resp.Diagnostics.AddError("Read error", err.Error())
		return
	}
	principal := model.Principal.ValueString()
	principalPermissions, err := sn.GetPrincipalPermissions(&[]string{principal})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("getting principal permissions: %s", err))
		return
	}
	if principalPermissions == nil {
		log.Printf("[INFO] betterado_git_permissions: ACL token %q not found; removing from state", sn.GetToken())
		resp.State.RemoveResource(ctx)
		return
	}
	configPerms := map[string]string{}
	resp.Diagnostics.Append(model.Permissions.ElementsAs(ctx, &configPerms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	filtered := map[string]string{}
	for _, pp := range *principalPermissions {
		for action, perm := range pp.Permissions {
			if _, ok := configPerms[string(action)]; ok {
				filtered[string(action)] = string(perm)
			}
		}
	}
	permMap, diags := types.MapValueFrom(ctx, types.StringType, filtered)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Permissions = permMap
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *gitPermissionsFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model gitPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyPermissions(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Update error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *gitPermissionsFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model gitPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sn, err := r.newSecurityNamespace(&model)
	if err != nil {
		resp.Diagnostics.AddError("Delete error", err.Error())
		return
	}
	principal := model.Principal.ValueString()
	configPerms := map[string]string{}
	resp.Diagnostics.Append(model.Permissions.ElementsAs(ctx, &configPerms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	notsetMap := make(map[securityhelper.ActionName]securityhelper.PermissionType, len(configPerms))
	for action := range configPerms {
		notsetMap[securityhelper.ActionName(action)] = securityhelper.PermissionTypeValues.NotSet
	}
	setPerm := []securityhelper.SetPrincipalPermission{{
		Replace: true,
		PrincipalPermission: securityhelper.PrincipalPermission{
			SubjectDescriptor: principal,
			Permissions:       notsetMap,
		},
	}}
	if err := sn.SetPrincipalPermissions(&setPerm); err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("removing principal permissions: %s", err))
	}
}

func (r *gitPermissionsFrameworkResource) newSecurityNamespace(model *gitPermissionsModel) (*securityhelper.SecurityNamespace, error) {
	projectID := model.ProjectID.ValueString()
	repositoryID := ""
	if !model.RepositoryID.IsNull() && !model.RepositoryID.IsUnknown() {
		repositoryID = model.RepositoryID.ValueString()
	}
	branchName := ""
	if !model.BranchName.IsNull() && !model.BranchName.IsUnknown() {
		branchName = model.BranchName.ValueString()
	}
	return securityhelper.NewSecurityNamespace(
		nil,
		r.client,
		securityhelper.SecurityNamespaceIDValues.GitRepositories,
		func(_ *sdkschema.ResourceData, _ *client.AggregatedClient) (string, error) {
			return buildGitToken(projectID, repositoryID, branchName)
		},
	)
}

// buildGitToken mirrors the SDKv2 createGitToken logic without ResourceData.
//
//	repoV2/#ProjectID#
//	repoV2/#ProjectID#/#RepositoryID#
//	repoV2/#ProjectID#/#RepositoryID#/refs/heads
//	repoV2/#ProjectID#/#RepositoryID#/refs/heads/#BranchID#
func buildGitToken(projectID, repositoryID, branchName string) (string, error) {
	aclToken := "repoV2/" + projectID
	if repositoryID != "" {
		aclToken += "/" + repositoryID
	}
	if branchName != "" {
		if repositoryID == "" {
			return "", fmt.Errorf("unable to create ACL token for branch %q: no repository_id specified", branchName)
		}
		re := regexp.MustCompile(`(/?refs/heads/)?(.*)+`)
		branchPath := re.FindStringSubmatch(branchName)
		paths := strings.Split(branchPath[len(branchPath)-1], "/")
		encodedPaths := make([]string, len(paths))
		for i, subBranchPath := range paths {
			encoded, err := converter.EncodeUtf16HexString(subBranchPath)
			if err != nil {
				return "", err
			}
			encodedPaths[i] = encoded
		}
		aclToken += "/refs/heads/" + strings.Join(encodedPaths, "/")
	}
	return aclToken, nil
}

func (r *gitPermissionsFrameworkResource) applyPermissions(ctx context.Context, model *gitPermissionsModel) error {
	sn, err := r.newSecurityNamespace(model)
	if err != nil {
		return fmt.Errorf("creating security namespace: %w", err)
	}
	principal := model.Principal.ValueString()
	configPerms := map[string]string{}
	if diags := model.Permissions.ElementsAs(ctx, &configPerms, false); diags.HasError() {
		return fmt.Errorf("reading permissions from plan")
	}
	replace := model.Replace.ValueBool()
	permMap := make(map[securityhelper.ActionName]securityhelper.PermissionType, len(configPerms))
	for k, v := range configPerms {
		permMap[securityhelper.ActionName(k)] = securityhelper.PermissionType(strings.ToLower(v))
	}
	setPerm := []securityhelper.SetPrincipalPermission{{
		Replace: replace,
		PrincipalPermission: securityhelper.PrincipalPermission{
			SubjectDescriptor: principal,
			Permissions:       permMap,
		},
	}}
	if err := sn.SetPrincipalPermissions(&setPerm); err != nil {
		return fmt.Errorf("setting principal permissions: %w", err)
	}
	deadline := time.Now().Add(60 * time.Minute)
	for time.Now().Before(deadline) {
		currentPerms, err := sn.GetPrincipalPermissions(&[]string{principal})
		if err != nil {
			return fmt.Errorf("polling permissions: %w", err)
		}
		if currentPerms != nil && len(*currentPerms) == 1 {
			synced := true
			for key, want := range permMap {
				got, ok := (*currentPerms)[0].Permissions[key]
				if !ok || !strings.EqualFold(string(want), string(got)) {
					synced = false
					break
				}
			}
			if synced {
				break
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	token := sn.GetToken()
	model.ID = types.StringValue(fmt.Sprintf("%s/%s", token, principal))
	return nil
}
