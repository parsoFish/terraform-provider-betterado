package core

// project_helpers.go — shared helpers for the core package.
//
// These were previously inline in the SDKv2 source files (resource_project.go,
// data_projects.go, resource_project_features.go, resource_team.go). They are
// moved here so the framework implementations can use them after the SDKv2
// files are deleted.  Profile rule 3b: dedup means delete, not just deregister.

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	linq "github.com/ahmetb/go-linq"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/featuremanagement"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/operations"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	securityhelper "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/permissions/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ── Project operation helpers ─────────────────────────────────────────────────

// projectBusyTimeoutDuration is used for retrying operations on projects that
// may be busy (e.g. other apply operations in progress).
var projectBusyTimeoutDuration time.Duration = 6 * time.Minute

// pollOperationResult returns a StateRefreshFunc that polls an ADO operation
// until it succeeds, fails, or is cancelled.
func pollOperationResult(clients *client.AggregatedClient, operationRef *operations.OperationReference) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ret, err := clients.OperationsClient.GetOperation(clients.Ctx, operations.GetOperationArgs{
			OperationId: operationRef.Id,
			PluginId:    operationRef.PluginId,
		})
		if err != nil {
			return nil, string(operations.OperationStatusValues.Failed), err
		}

		if *ret.Status != operations.OperationStatusValues.Succeeded {
			log.Printf("[DEBUG] Waiting for project operation success. Operation result %v", ret.DetailedMessage)
		}

		return ret, string(*ret.Status), nil
	}
}

// getProject polls the ADO API until a project with the given ID or name
// becomes available, retrying up to timeout duration.
func getProject(clients *client.AggregatedClient, projectID string, projectName string, timeout time.Duration) (*adocore.TeamProject, error) {
	identifier := projectID
	if identifier == "" {
		identifier = projectName
	}

	var project *adocore.TeamProject
	stateConf := &retry.StateChangeConf{
		ContinuousTargetOccurence: 1,
		Delay:                     5 * time.Second,
		MinTimeout:                20 * time.Second,
		Timeout:                   timeout,
		Pending:                   []string{"pending"},
		Target:                    []string{"success"},
		Refresh: func() (result interface{}, state string, err error) {
			project, err = clients.CoreClient.GetProject(clients.Ctx, adocore.GetProjectArgs{
				ProjectId:           &identifier,
				IncludeCapabilities: converter.Bool(true),
				IncludeHistory:      converter.Bool(false),
			})
			if err != nil {
				if utils.ResponseWasNotFound(err) {
					log.Printf("[INFO] Project not found. ID/Name: %s . Error: %+v", identifier, err)
				}
				return project, "pending", err
			}
			return project, "success", nil
		},
	}

	if _, err := stateConf.WaitForStateContext(clients.Ctx); err != nil {
		if utils.ResponseWasNotFound(err) {
			return nil, err
		}
		return nil, fmt.Errorf("Project not found. (ID: %s or name: %s), Error: %+v", projectID, projectName, err)
	}

	return project, nil
}

// updateProject sends an UpdateProject request and waits for the operation to
// complete, retrying on transient busy errors.
func updateProject(clients *client.AggregatedClient, project *adocore.TeamProject, timeout time.Duration) error {
	var operationRef *operations.OperationReference

	err := retry.RetryContext(clients.Ctx, projectBusyTimeoutDuration, func() *retry.RetryError {
		var updateErr error
		operationRef, updateErr = clients.CoreClient.UpdateProject(
			clients.Ctx,
			adocore.UpdateProjectArgs{
				ProjectUpdate: project,
				ProjectId:     project.Id,
			},
		)
		if updateErr != nil {
			return retry.RetryableError(updateErr)
		}
		return nil
	})
	if err != nil {
		return err
	}

	stateConf := &retry.StateChangeConf{
		ContinuousTargetOccurence: 1,
		Delay:                     10 * time.Second,
		MinTimeout:                10 * time.Second,
		Timeout:                   timeout,
		Pending: []string{
			string(operations.OperationStatusValues.InProgress),
			string(operations.OperationStatusValues.Queued),
			string(operations.OperationStatusValues.NotSet),
		},
		Target: []string{
			string(operations.OperationStatusValues.Failed),
			string(operations.OperationStatusValues.Succeeded),
			string(operations.OperationStatusValues.Cancelled),
		},
		Refresh: pollOperationResult(clients, operationRef),
	}

	if _, err := stateConf.WaitForStateContext(clients.Ctx); err != nil {
		return fmt.Errorf("waiting for project ready. %v ", err)
	}
	return nil
}

// deleteProject queues an ADO project deletion and waits for the operation to
// complete, retrying on transient busy errors.
func deleteProject(clients *client.AggregatedClient, id string, timeout time.Duration) error {
	projectUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("Invalid project UUID: %s", id)
	}

	var operationRef *operations.OperationReference

	err = retry.RetryContext(clients.Ctx, projectBusyTimeoutDuration, func() *retry.RetryError {
		var deleteErr error
		operationRef, deleteErr = clients.CoreClient.QueueDeleteProject(clients.Ctx, adocore.QueueDeleteProjectArgs{
			ProjectId: &projectUUID,
		})

		if deleteErr != nil {
			return retry.RetryableError(deleteErr)
		}
		return nil
	})
	if err != nil {
		return err
	}

	stateConf := &retry.StateChangeConf{
		ContinuousTargetOccurence: 1,
		Delay:                     10 * time.Second,
		MinTimeout:                10 * time.Second,
		Timeout:                   timeout,
		Pending: []string{
			string(operations.OperationStatusValues.InProgress),
			string(operations.OperationStatusValues.Queued),
			string(operations.OperationStatusValues.NotSet),
		},
		Target: []string{
			string(operations.OperationStatusValues.Failed),
			string(operations.OperationStatusValues.Succeeded),
			string(operations.OperationStatusValues.Cancelled),
		},
		Refresh: pollOperationResult(clients, operationRef),
	}

	if _, err := stateConf.WaitForStateContext(clients.Ctx); err != nil {
		return fmt.Errorf("waiting for project delete. %v ", err)
	}
	return nil
}

// ── Projects list helpers ─────────────────────────────────────────────────────

// getProjectsForStateAndName returns all ADO projects matching the given state
// and optionally filtered by name. It handles continuation-token pagination.
func getProjectsForStateAndName(clients *client.AggregatedClient, projectState string, projectName string) ([]adocore.TeamProjectReference, error) {
	var projects []adocore.TeamProjectReference
	var currentToken string

	for hasMore := true; hasMore; {
		newProjects, latestToken, err := getProjectsWithContinuationToken(clients, projectState, currentToken)
		currentToken = latestToken
		if err != nil {
			return nil, err
		}

		if projectName != "" {
			for _, project := range newProjects {
				if strings.EqualFold(*project.Name, projectName) {
					return []adocore.TeamProjectReference{project}, nil
				}
			}
		} else {
			projects = append(projects, newProjects...)
		}
		hasMore = currentToken != ""
	}

	return projects, nil
}

// getProjectsWithContinuationToken fetches a page of projects using the given
// continuation token, returning the next token for pagination.
func getProjectsWithContinuationToken(clients *client.AggregatedClient, projectState string, continuationToken string) ([]adocore.TeamProjectReference, string, error) {
	args := adocore.GetProjectsArgs{
		StateFilter: converter.ToPtr(adocore.ProjectState(projectState)),
	}
	if continuationToken != "" {
		token, err := strconv.Atoi(continuationToken)
		if err != nil {
			return nil, "", err
		}
		args.ContinuationToken = &token
	}

	response, err := clients.CoreClient.GetProjects(clients.Ctx, args)
	if err != nil {
		return nil, "", err
	}

	return response.Value, response.ContinuationToken, nil
}

// ── Project feature helpers ───────────────────────────────────────────────────

// ProjectFeatureType is the short name for an ADO project feature.
type ProjectFeatureType string

type projectFeatureTypeValuesType struct {
	Boards       ProjectFeatureType
	Repositories ProjectFeatureType
	Pipelines    ProjectFeatureType
	TestPlans    ProjectFeatureType
	Artifacts    ProjectFeatureType
}

// ProjectFeatureTypeValues lists the valid ADO project feature short names.
var ProjectFeatureTypeValues = projectFeatureTypeValuesType{
	Boards:       "boards",
	Repositories: "repositories",
	Pipelines:    "pipelines",
	TestPlans:    "testplans",
	Artifacts:    "artifacts",
}

// projectFeatureNameMap maps ADO feature API IDs to their short names.
var projectFeatureNameMap = map[string]ProjectFeatureType{
	"ms.vss-work.agile":           ProjectFeatureTypeValues.Boards,
	"ms.vss-code.version-control": ProjectFeatureTypeValues.Repositories,
	"ms.vss-build.pipelines":      ProjectFeatureTypeValues.Pipelines,
	"ms.vss-test-web.test":        ProjectFeatureTypeValues.TestPlans,
	"ms.azure-artifacts.feature":  ProjectFeatureTypeValues.Artifacts,
}

// projectFeatureNameMapReverse maps short names to their ADO feature API IDs.
var projectFeatureNameMapReverse = map[ProjectFeatureType]string{
	ProjectFeatureTypeValues.Boards:       "ms.vss-work.agile",
	ProjectFeatureTypeValues.Repositories: "ms.vss-code.version-control",
	ProjectFeatureTypeValues.Pipelines:    "ms.vss-build.pipelines",
	ProjectFeatureTypeValues.TestPlans:    "ms.vss-test-web.test",
	ProjectFeatureTypeValues.Artifacts:    "ms.azure-artifacts.feature",
}

// getProjectFeatureStates fetches the current enable/disable state for all
// known ADO project features on the given project.
func getProjectFeatureStates(ctx context.Context, fc featuremanagement.Client, projectID string) (*map[ProjectFeatureType]featuremanagement.ContributedFeatureEnabledValue, error) {
	states, err := fc.QueryFeatureStates(ctx, featuremanagement.QueryFeatureStatesArgs{
		Query: &featuremanagement.ContributedFeatureStateQuery{
			FeatureIds: &[]string{
				"ms.vss-work.agile",
				"ms.vss-code.version-control",
				"ms.vss-build.pipelines",
				"ms.vss-test-web.test",
				"ms.azure-artifacts.feature",
			},
			ScopeValues: &map[string]string{
				"project": projectID,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Get project features error, project: %s, error: %+v", projectID, err)
	}

	featureStates := make(map[ProjectFeatureType]featuremanagement.ContributedFeatureEnabledValue)
	for k, v := range projectFeatureNameMap {
		if state, ok := (*states.FeatureStates)[k]; ok {
			featureStates[v] = *state.State
		}
	}
	return &featureStates, nil
}

// ── Process template helpers ──────────────────────────────────────────────────

// lookupProcessTemplateID returns the UUID string of the process template with
// the given name (case-insensitive).
func lookupProcessTemplateID(clients *client.AggregatedClient, templateName string) (string, error) {
	processes, err := clients.CoreClient.GetProcesses(clients.Ctx, adocore.GetProcessesArgs{})
	if err != nil {
		return "", err
	}

	for _, p := range *processes {
		if strings.EqualFold(*p.Name, templateName) {
			return p.Id.String(), nil
		}
	}

	return "", fmt.Errorf("No process template found")
}

// lookupProcessTemplateName returns the name of the process template with the
// given UUID string.
func lookupProcessTemplateName(clients *client.AggregatedClient, templateID string) (string, error) {
	id, err := uuid.Parse(templateID)
	if err != nil {
		return "", fmt.Errorf("Error parsing Work Item Template ID, got %s: %v", templateID, err)
	}

	process, err := clients.CoreClient.GetProcessById(clients.Ctx, adocore.GetProcessByIdArgs{
		ProcessId: &id,
	})
	if err != nil {
		return "", fmt.Errorf("Error looking up template by ID: %v", err)
	}

	return *process.Name, nil
}

// ── Team member / administrator helpers ──────────────────────────────────────

// getTeamMembers returns the set of SubjectDescriptors for all members of the
// given team.
func getTeamMembers(clients *client.AggregatedClient, team *adocore.WebApiTeam) (*schema.Set, error) {
	members, err := clients.IdentityClient.ReadMembers(clients.Ctx, identity.ReadMembersArgs{
		ContainerId: converter.String(team.Id.String()),
	})
	if err != nil {
		return nil, err
	}

	return getSubjectDescriptors(clients, members)
}

// setTeamMembers reconciles the team's member list to match subjectDescriptors
// by adding and removing members as needed.
func setTeamMembers(clients *client.AggregatedClient, team *adocore.WebApiTeam, subjectDescriptors *[]string) error {
	currentMemberSet, err := getTeamMembers(clients, team)
	if err != nil {
		return err
	}
	if (subjectDescriptors == nil || len(*subjectDescriptors) == 0) && currentMemberSet.Len() == 0 {
		return nil
	}
	if subjectDescriptors == nil {
		subjectDescriptors = &[]string{}
	}

	currentMembers := currentMemberSet.List()

	if err := removeTeamMembers(clients, team, linq.From(currentMembers).Except(linq.From(*subjectDescriptors))); err != nil {
		return err
	}

	if err := addTeamMembers(clients, team, linq.From(*subjectDescriptors).Except(linq.From(currentMembers)), false); err != nil {
		return err
	}

	return nil
}

// removeTeamMembers removes the identities matching the given LINQ query from
// the team's membership.
func removeTeamMembers(clients *client.AggregatedClient, team *adocore.WebApiTeam, query linq.Query) error {
	idList, err := getIdentitiesFromSubjects(clients, query)
	if err != nil {
		return err
	}

	for _, id := range *idList {
		log.Printf("[TRACE] Removing member %s from team %s", id.Id.String(), *team.Name)

		_, err := clients.IdentityClient.RemoveMember(clients.Ctx, identity.RemoveMemberArgs{
			ContainerId: converter.String(team.Id.String()),
			MemberId:    converter.String(id.Id.String()),
		})
		if err != nil {
			return fmt.Errorf("Error removing member %s from team %s: %+v", id.Id.String(), *team.Name, err)
		}
	}
	return nil
}

// addTeamMembers adds the identities matching the given LINQ query to the
// team's membership.
func addTeamMembers(clients *client.AggregatedClient, team *adocore.WebApiTeam, query linq.Query, isAddMode bool) error {
	idList, err := getIdentitiesFromSubjects(clients, query)
	if err != nil {
		return err
	}
	if idList == nil || len(*idList) != query.Count() {
		return fmt.Errorf("Failed to load identity data for subjects")
	}

	for _, id := range *idList {
		log.Printf("[TRACE] Adding member %s to team %s", id.Id.String(), *team.Name)

		ok, err := clients.IdentityClient.AddMember(clients.Ctx, identity.AddMemberArgs{
			ContainerId: converter.String(team.Id.String()),
			MemberId:    converter.String(id.Id.String()),
		})
		if err != nil {
			return fmt.Errorf("Error adding member %s to team %s: %+v", *id.SubjectDescriptor, *team.Name, err)
		}
		if ok != nil && !*ok {
			if !isAddMode {
				return fmt.Errorf("Failed adding member %s to team %s", *id.SubjectDescriptor, *team.Name)
			}
			log.Printf("[TRACE] Member %s is already a member of team %s", *id.SubjectDescriptor, *team.Name)
		}
	}

	return nil
}

// getIdentitiesFromSubjects resolves a LINQ query of SubjectDescriptor strings
// to the corresponding Identity objects.
func getIdentitiesFromSubjects(clients *client.AggregatedClient, query linq.Query) (*[]identity.Identity, error) {
	if !query.Any() {
		return &[]identity.Identity{}, nil
	}

	descriptors := query.
		Aggregate(func(r interface{}, i interface{}) interface{} {
			if r.(string) == "" {
				return i
			}
			return r.(string) + "," + i.(string)
		}).(string)

	idlist, err := clients.IdentityClient.ReadIdentities(clients.Ctx, identity.ReadIdentitiesArgs{
		SubjectDescriptors: converter.String(descriptors),
	})
	if err != nil {
		return nil, err
	}

	return idlist, err
}

// getIdentitySecurityNamespace returns the SecurityNamespace for the given team
// in its project. d may be nil (passing nil is supported for framework callers).
func getIdentitySecurityNamespace(d *schema.ResourceData, clients *client.AggregatedClient, team *adocore.WebApiTeam) (*securityhelper.SecurityNamespace, error) {
	return securityhelper.NewSecurityNamespace(d,
		clients,
		securityhelper.SecurityNamespaceIDValues.Identity,
		func(d *schema.ResourceData, clients *client.AggregatedClient) (string, error) {
			return team.ProjectId.String() + "\\" + team.Id.String(), nil
		})
}

// getTeamAdministrators returns the current list of team administrators as a
// set of SubjectDescriptors. d may be nil for framework callers.
func getTeamAdministrators(d *schema.ResourceData, clients *client.AggregatedClient, team *adocore.WebApiTeam) (*schema.Set, error) {
	sn, err := getIdentitySecurityNamespace(d, clients, team)
	if err != nil {
		return nil, err
	}

	actionDefinitions, err := sn.GetActionDefinitions()
	if err != nil {
		return nil, err
	}

	acl, err := sn.GetAccessControlList(nil)
	if err != nil {
		return nil, err
	}

	adminDescriptorList := []string{}
	if acl != nil && acl.AcesDictionary != nil {
		bit := *(*actionDefinitions)["Read"].Bit | *(*actionDefinitions)["Write"].Bit | *(*actionDefinitions)["Delete"].Bit | *(*actionDefinitions)["ManageMembership"].Bit | *(*actionDefinitions)["CreateScope"].Bit
		for _, ace := range *acl.AcesDictionary {
			if *ace.Allow&bit == bit {
				adminDescriptorList = append(adminDescriptorList, *ace.Descriptor)
			}
		}
	}
	return getSubjectDescriptors(clients, &adminDescriptorList)
}

// updateTeamAdministrators reconciles the team's administrator list to match
// subjectDescriptors by granting/revoking the admin permission set. d may be nil.
func updateTeamAdministrators(d *schema.ResourceData, clients *client.AggregatedClient, team *adocore.WebApiTeam, subjectDescriptors *[]string) error {
	currentAdministratorSet, err := getTeamAdministrators(d, clients, team)
	if err != nil {
		return err
	}
	if (subjectDescriptors == nil || len(*subjectDescriptors) == 0) && currentAdministratorSet.Len() == 0 {
		return nil
	}

	currentAdministrators := currentAdministratorSet.List()

	log.Print("[DEBUG] updateTeamAdministrators::removing deleted administrators from team")
	err = setTeamAdministratorsPermissions(d,
		clients,
		team,
		linq.From(currentAdministrators).Except(linq.From(*subjectDescriptors)),
		securityhelper.PermissionTypeValues.NotSet)
	if err != nil {
		return err
	}

	log.Print("[DEBUG] updateTeamAdministrators::adding missing administrators to team")
	err = setTeamAdministratorsPermissions(d,
		clients,
		team,
		linq.From(*subjectDescriptors).Except(linq.From(currentAdministrators)),
		securityhelper.PermissionTypeValues.Allow)
	if err != nil {
		return err
	}

	return nil
}

// setTeamAdministratorsPermissions grants or revokes admin permissions for the
// given set of SubjectDescriptors. d may be nil for framework callers.
func setTeamAdministratorsPermissions(d *schema.ResourceData, clients *client.AggregatedClient, team *adocore.WebApiTeam, subjectDescriptors linq.Query, permission securityhelper.PermissionType) error {
	if !subjectDescriptors.Any() {
		log.Print("[DEBUG] setTeamAdministratorsPermissions::list of subject descriptors is empty")
		return nil
	}

	sn, err := getIdentitySecurityNamespace(d, clients, team)
	if err != nil {
		return err
	}

	principalPermissionCreator := func(query linq.Query, permission securityhelper.PermissionType) *[]securityhelper.SetPrincipalPermission {
		var subjectList []securityhelper.SetPrincipalPermission

		query.Select(func(item interface{}) interface{} {
			return securityhelper.SetPrincipalPermission{
				Replace: true,
				PrincipalPermission: securityhelper.PrincipalPermission{
					SubjectDescriptor: item.(string),
					Permissions: map[securityhelper.ActionName]securityhelper.PermissionType{
						"Read":             permission,
						"Write":            permission,
						"Delete":           permission,
						"ManageMembership": permission,
						"CreateScope":      permission,
					},
				},
			}
		}).ToSlice(&subjectList)
		return &subjectList
	}

	principalPermissions := principalPermissionCreator(subjectDescriptors, permission)
	err = sn.SetPrincipalPermissions(principalPermissions)
	if err != nil {
		return err
	}

	return nil
}

// getSubjectDescriptors resolves a list of member identity strings (descriptors)
// to a schema.Set of SubjectDescriptor strings, batching in groups of 20.
func getSubjectDescriptors(clients *client.AggregatedClient, members *[]string) (*schema.Set, error) {
	set := schema.NewSet(schema.HashString, nil)

	if members == nil || len(*members) == 0 {
		return set, nil
	}

	start := 0
	step := 20 // 20 descriptors per request
	var subMembers []string
	for start*step <= len(*members) {
		if (start+1)*step < len(*members) {
			subMembers = (*members)[start*step : (start+1)*step]
		} else {
			subMembers = (*members)[start*step:]
		}
		start++
		if len(subMembers) > 0 {
			descriptors := linq.From(subMembers).
				Aggregate(func(r interface{}, i interface{}) interface{} {
					if r.(string) == "" {
						return i
					}
					return r.(string) + "," + i.(string)
				}).(string)

			memberIdentities, err := clients.IdentityClient.ReadIdentities(clients.Ctx, identity.ReadIdentitiesArgs{
				Descriptors: &descriptors,
			})
			if err != nil {
				return nil, err
			}

			if memberIdentities != nil && len(*memberIdentities) > 0 {
				for _, memberIdentity := range *memberIdentities {
					set.Add(*memberIdentity.SubjectDescriptor)
				}
			}
		}
	}

	return set, nil
}
