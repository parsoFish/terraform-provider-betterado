// Package graph provides Azure DevOps Graph API resources and data sources.
// This file contains shared helper functions used by both the legacy SDKv2
// implementations and the new framework implementations.
package graph

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ---------------------------------------------------------------------------
// Group lookup helpers
// ---------------------------------------------------------------------------

// getProjectDescriptor returns the graph descriptor for a project by its UUID.
// If projectID is empty, an empty string is returned (indicating collection scope).
func getProjectDescriptor(clients *client.AggregatedClient, projectID string) (string, error) {
	if projectID == "" {
		return "", nil
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return "", err
	}

	descriptor, err := clients.GraphClient.GetDescriptor(clients.Ctx, graph.GetDescriptorArgs{StorageKey: &projectUUID})
	if err != nil {
		return "", err
	}

	return *descriptor.Value, nil
}

// getGroupsForDescriptor fetches all groups under the given project descriptor
// (or all collection-level groups when projectDescriptor is empty).
func getGroupsForDescriptor(clients *client.AggregatedClient, projectDescriptor string) (*[]graph.GraphGroup, error) {
	var groups []graph.GraphGroup
	var currentToken string

	for hasMore := true; hasMore; {
		newGroups, latestToken, err := getGroupsWithContinuationToken(clients, projectDescriptor, currentToken)
		currentToken = latestToken
		if err != nil {
			return nil, err
		}

		if newGroups != nil && len(*newGroups) > 0 {
			var filteredGroups []graph.GraphGroup

			if projectDescriptor == "" {
				// filter on collection groups
				for _, grp := range *newGroups {
					if grp.Domain == nil {
						continue
					}

					domain := strings.ToLower(*grp.Domain)
					if strings.HasPrefix(domain, "vstfs:///framework/identitydomain") ||
						strings.HasPrefix(domain, "vstfs:///framework/generic") {
						filteredGroups = append(filteredGroups, grp)
					}
				}
				groups = append(groups, filteredGroups...)
			} else {
				for _, grp := range *newGroups {
					domain := strings.ToLower(*grp.Domain)
					if grp.Domain != nil && strings.HasPrefix(domain, "vstfs:///classification/teamproject/") {
						filteredGroups = append(filteredGroups, grp)
					}
				}
				groups = append(groups, filteredGroups...)
			}
		}
		hasMore = currentToken != ""
	}

	return &groups, nil
}

// getGroupsWithContinuationToken fetches one page of groups from the API.
func getGroupsWithContinuationToken(clients *client.AggregatedClient, projectDescriptor string, continuationToken string) (*[]graph.GraphGroup, string, error) {
	args := graph.ListGroupsArgs{}
	if projectDescriptor != "" {
		args.ScopeDescriptor = &projectDescriptor
	}
	if continuationToken != "" {
		args.ContinuationToken = &continuationToken
	}

	response, err := clients.GraphClient.ListGroups(clients.Ctx, args)
	if err != nil {
		return nil, "", err
	}

	if response.ContinuationToken != nil && len(*response.ContinuationToken) > 1 {
		return nil, "", fmt.Errorf("Expected at most 1 continuation token, but found %d", len(*response.ContinuationToken))
	}

	var newToken string
	if response.ContinuationToken != nil && len(*response.ContinuationToken) > 0 {
		newToken = (*response.ContinuationToken)[0]
	}

	return response.GraphGroups, newToken, nil
}

// selectGroup finds the first group whose display name case-insensitively matches groupName.
func selectGroup(groups *[]graph.GraphGroup, groupName string) *graph.GraphGroup {
	for _, group := range *groups {
		if strings.EqualFold(*group.DisplayName, groupName) {
			return &group
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Group membership helpers
// ---------------------------------------------------------------------------

// applyMembershipUpdate applies a membership delta: removes memberships first,
// then adds new ones.
func applyMembershipUpdate(clients *client.AggregatedClient, toAdd *[]graph.GraphMembership, toRemove *[]graph.GraphMembership) error {
	if toRemove != nil && len(*toRemove) > 0 {
		err := removeMembers(clients, toRemove)
		if err != nil {
			return fmt.Errorf("Removing group memberships during update: %+v", err)
		}
	}

	if toAdd != nil && len(*toAdd) > 0 {
		err := addMembers(clients, toAdd)
		if err != nil {
			return fmt.Errorf("Adding group memberships during update: %+v", err)
		}
	}
	return nil
}

// addMembers adds each membership to a group via the AzDO REST API.
func addMembers(clients *client.AggregatedClient, memberships *[]graph.GraphMembership) error {
	if memberships != nil {
		for _, membership := range *memberships {
			_, err := clients.GraphClient.AddMembership(clients.Ctx, graph.AddMembershipArgs{
				SubjectDescriptor:   membership.MemberDescriptor,
				ContainerDescriptor: membership.ContainerDescriptor,
			})
			if err != nil {
				return fmt.Errorf("Error adding member %s to group %s: %+v",
					converter.ToString(membership.MemberDescriptor, "nil"),
					converter.ToString(membership.ContainerDescriptor, "nil"),
					err)
			}
		}
	}
	return nil
}

// removeMembers removes each membership from a group via the AzDO REST API.
func removeMembers(clients *client.AggregatedClient, memberships *[]graph.GraphMembership) error {
	if memberships != nil {
		for _, membership := range *memberships {
			err := clients.GraphClient.RemoveMembership(clients.Ctx, graph.RemoveMembershipArgs{
				SubjectDescriptor:   membership.MemberDescriptor,
				ContainerDescriptor: membership.ContainerDescriptor,
			})
			if err != nil {
				return fmt.Errorf("Error removing member from group: %+v", err)
			}
		}
	}
	return nil
}

// getGroupMemberships returns the current direct memberships of a group.
func getGroupMemberships(clients *client.AggregatedClient, groupDescriptor string) (*[]graph.GraphMembership, error) {
	return clients.GraphClient.ListMemberships(clients.Ctx, graph.ListMembershipsArgs{
		SubjectDescriptor: &groupDescriptor,
		Direction:         &graph.GraphTraversalDirectionValues.Down,
		Depth:             converter.Int(1),
	})
}

// ---------------------------------------------------------------------------
// Group resource helpers
// ---------------------------------------------------------------------------

// groupReadMembers returns the current direct members of a group, returned as
// GraphMembership entries with ContainerDescriptor set.
func groupReadMembers(groupDescriptor string, clients *client.AggregatedClient) (*[]graph.GraphMembership, error) {
	actualMembers, err := clients.GraphClient.ListMemberships(clients.Ctx, graph.ListMembershipsArgs{
		SubjectDescriptor: &groupDescriptor,
		Direction:         &graph.GraphTraversalDirectionValues.Down,
		Depth:             converter.Int(1),
	})
	if err != nil {
		return nil, fmt.Errorf("Reading group memberships: %+v", err)
	}

	members := make([]graph.GraphMembership, len(*actualMembers))
	for i, membership := range *actualMembers {
		members[i] = graph.GraphMembership{
			ContainerDescriptor: &groupDescriptor,
			MemberDescriptor:    membership.MemberDescriptor,
		}
	}

	return &members, nil
}

// domain2ProjectID extracts the project UUID from an AzDO domain string.
// Returns empty string if the domain is not a TeamProject domain.
func domain2ProjectID(domain string) (projectID string) {
	if domain == "" {
		return ""
	}
	if !strings.HasPrefix(domain, "vstfs:///Classification/TeamProject") {
		return ""
	}
	return domain[36:]
}

// ---------------------------------------------------------------------------
// Service principal helpers
// ---------------------------------------------------------------------------

// getIdentityServicePrincipalsWithFilterValue fetches identity service principals
// matching the given search filter and filter value.
func getIdentityServicePrincipalsWithFilterValue(clients *client.AggregatedClient, searchFilter *string, filterValue string) (*[]identity.Identity, error) {
	return clients.IdentityClient.ReadIdentities(clients.Ctx, identity.ReadIdentitiesArgs{
		SearchFilter: searchFilter,
		FilterValue:  &filterValue,
	})
}

// flattenIdentityServicePrincipals reduces identity results to the fields needed
// for service principal selection.
func flattenIdentityServicePrincipals(servicePrincipals *[]identity.Identity) (*[]identity.Identity, error) {
	if servicePrincipals == nil || len(*servicePrincipals) == 0 {
		return nil, fmt.Errorf("Input Service Principals Parameter is nil")
	}
	results := make([]identity.Identity, 0)
	for _, servicePrincipal := range *servicePrincipals {
		if servicePrincipal.Descriptor == nil {
			return nil, fmt.Errorf("User Object does not contain an id")
		}
		results = append(results, identity.Identity{
			Id:                  servicePrincipal.Id,
			Descriptor:          servicePrincipal.Descriptor,
			ProviderDisplayName: servicePrincipal.ProviderDisplayName,
			SubjectDescriptor:   servicePrincipal.SubjectDescriptor,
		})
	}
	return &results, nil
}

// validateIdentityServicePrincipal selects the first service principal whose
// ProviderDisplayName contains displayName (case-insensitive).
func validateIdentityServicePrincipal(servicePrincipals *[]identity.Identity, displayName string) *identity.Identity {
	if servicePrincipals == nil || len(*servicePrincipals) == 0 {
		return nil
	}
	for _, servicePrincipal := range *servicePrincipals {
		if strings.Contains(strings.ToLower(*servicePrincipal.ProviderDisplayName), strings.ToLower(displayName)) {
			return &servicePrincipal
		}
	}
	return nil
}

// getServicePrincipal fetches a graph service principal by its descriptor.
func getServicePrincipal(clients *client.AggregatedClient, servicePrincipalDescriptor *string) (*graph.GraphServicePrincipal, error) {
	return clients.GraphClient.GetServicePrincipal(clients.Ctx, graph.GetServicePrincipalArgs{
		ServicePrincipalDescriptor: servicePrincipalDescriptor,
	})
}

// ---------------------------------------------------------------------------
// User helpers
// ---------------------------------------------------------------------------

// getUsersWithContinuationToken fetches one page of graph users, returning
// the users and the next continuation token (empty string when exhausted).
func getUsersWithContinuationToken(clients *client.AggregatedClient, subjectTypes *[]string, continuationToken string) ([]graph.GraphUser, string, error) {
	args := graph.ListUsersArgs{
		SubjectTypes: subjectTypes,
	}
	if continuationToken != "" {
		args.ContinuationToken = &continuationToken
	}
	response, err := clients.GraphClient.ListUsers(clients.Ctx, args)
	if err != nil {
		return nil, "", fmt.Errorf("Listing users: %q", err)
	}

	continuationToken = ""
	if response.ContinuationToken != nil && (*response.ContinuationToken)[0] != "" {
		continuationToken = (*response.ContinuationToken)[0]
	}

	return *response.GraphUsers, continuationToken, nil
}
