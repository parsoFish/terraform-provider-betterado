// Package identity provides Azure DevOps Identity API data sources.
// This file contains shared helper functions used by both the legacy SDKv2
// implementations and the new framework implementations.
package identity

import (
	"fmt"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/tfhelper"
)

// ---------------------------------------------------------------------------
// Identity group helpers
// ---------------------------------------------------------------------------

// getIdentityGroupsWithProjectID fetches all identity groups for the given project ID.
func getIdentityGroupsWithProjectID(clients *client.AggregatedClient, projectID string) ([]identity.Identity, error) {
	response, err := clients.IdentityClient.ListGroups(clients.Ctx, identity.ListGroupsArgs{
		ScopeIds: &projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("Getting groups: %v", err)
	}
	return *response, nil
}

// selectIdentityGroup finds the first group whose ProviderDisplayName case-insensitively
// matches groupName.
func selectIdentityGroup(groups *[]identity.Identity, groupName string) *identity.Identity {
	for _, group := range *groups {
		if strings.EqualFold(*group.ProviderDisplayName, groupName) {
			return &group
		}
	}
	return nil
}

// flattenIdentityGroups converts identity groups to a list of attribute maps.
func flattenIdentityGroups(groups *[]identity.Identity) []interface{} {
	if groups == nil {
		return []interface{}{}
	}
	results := make([]interface{}, len(*groups))
	for i, group := range *groups {
		groupMap := make(map[string]interface{})

		if group.Id != nil {
			groupID := *group.Id
			groupMap["id"] = groupID.String()
		}

		if group.ProviderDisplayName != nil {
			groupMap["name"] = *group.ProviderDisplayName
		}

		if group.Descriptor != nil {
			groupMap["descriptor"] = *group.Descriptor
		}

		if group.SubjectDescriptor != nil {
			groupMap["subject_descriptor"] = *group.SubjectDescriptor
		}

		results[i] = groupMap
	}
	return results
}

// getIdentityGroupHash returns a hash for a group map (used in schema.Set).
func getIdentityGroupHash(v interface{}) int {
	return tfhelper.HashString(v.(map[string]interface{})["id"].(string))
}

// ---------------------------------------------------------------------------
// Identity user helpers
// ---------------------------------------------------------------------------

// getIdentityUsersWithFilterValue fetches identity users matching the given
// search filter and filter value.
func getIdentityUsersWithFilterValue(clients *client.AggregatedClient, searchFilter string, filterValue string) (*[]identity.Identity, error) {
	response, err := clients.IdentityClient.ReadIdentities(clients.Ctx, identity.ReadIdentitiesArgs{
		SearchFilter: &searchFilter,
		FilterValue:  &filterValue,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

// validateIdentityUser selects the first user matching userName under the given
// searchFilter (case-insensitive).
func validateIdentityUser(users *[]identity.Identity, userName string, searchFilter string) *identity.Identity {
	for _, user := range *users {
		prop := user.Properties.(map[string]interface{})

		switch searchFilter {
		case "General":
			return &user
		case "DisplayName":
			// ProviderDisplayName holds the source-IdP display name; for most users this is
			// the human-readable name (e.g. "John Smith"). For built-in service accounts
			// (e.g. "Project Collection Build Service") ADO stores a GUID there and puts the
			// human-readable name in CustomDisplayName instead. Check both so that both
			// regular users and service accounts can be found by display name.
			providerMatch := user.ProviderDisplayName != nil &&
				strings.Contains(strings.ToLower(*user.ProviderDisplayName), strings.ToLower(userName))
			customMatch := user.CustomDisplayName != nil &&
				strings.Contains(strings.ToLower(*user.CustomDisplayName), strings.ToLower(userName))
			if providerMatch || customMatch {
				return &user
			}
		case "MailAddress":
			if v, ok := prop["Mail"]; ok && v != nil {
				mailProp := v.(map[string]interface{})
				if emailAddress, ok := mailProp["$value"].(string); ok {
					if strings.Contains(strings.ToLower(emailAddress), strings.ToLower(userName)) {
						return &user
					}
				}
			}
		case "AccountName":
			if v, ok := prop["Account"]; ok && v != nil {
				mailProp := v.(map[string]interface{})
				if emailAddress, ok := mailProp["$value"].(string); ok {
					if strings.Contains(strings.ToLower(emailAddress), strings.ToLower(userName)) {
						return &user
					}
				}
			}
		}
	}
	return nil
}
