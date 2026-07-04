package taskagent

// resource_variable_group.go — KV-search helpers retained for use by
// resource_variable_group_framework.go and data_variable_group_framework.go.
// The superseded SDKv2 resource (ResourceVariableGroup + CRUD + expand/flatten)
// has been deleted; only the still-referenced helpers below remain.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// updateVariableGroup is a shared helper used by resource_variable_group_framework.go and
// resource_variable_group_variable_framework.go to patch a variable group via the ADO API.
func updateVariableGroup(clients *client.AggregatedClient, parameters *taskagent.VariableGroupParameters, variableGroupID *int) (*taskagent.VariableGroup, error) {
	return clients.TaskAgentClient.UpdateVariableGroup(
		clients.Ctx,
		taskagent.UpdateVariableGroupArgs{
			GroupId:                 variableGroupID,
			VariableGroupParameters: parameters,
		},
	)
}

const (
	azureKeyVaultType = "AzureKeyVault"
)

// KeyVaultSecretAttributes holds metadata returned by the Azure Key Vault secrets proxy.
type KeyVaultSecretAttributes struct {
	Enabled       *bool   `json:"enabled,omitempty"`
	Created       *int64  `json:"created,omitempty"`
	Updated       *int64  `json:"updated,omitempty"`
	Expire        *int64  `json:"exp,omitempty"`
	RecoveryLevel *string `json:"recoveryLevel,omitempty"`
}

// KeyVaultSecret is a single secret entry returned by the KV proxy API.
type KeyVaultSecret struct {
	ContentType              *string `json:"contentType,omitempty"`
	ID                       *string `json:"id,omitempty"`
	KeyVaultSecretAttributes `json:"attributes,omitempty"`
}

// KeyVaultSecretResult is the paginated response from the KV proxy API.
type KeyVaultSecretResult struct {
	Value    *[]KeyVaultSecret `json:"value,omitempty"`
	NextLink *string           `json:"nextLink,omitempty"`
}

// isKeyVaultVariableGroupType returns true when the variable group type is AzureKeyVault.
func isKeyVaultVariableGroupType(variableGrouptype *string) bool {
	return variableGrouptype != nil && *variableGrouptype == azureKeyVaultType
}

// searchAzureKVSecrets fetches secrets from an Azure Key Vault via the ADO service-endpoint proxy
// and returns the subset whose names match the requested variable list.
func searchAzureKVSecrets(clients *client.AggregatedClient, projectID, kvName, serviceEndpointID string, variables []interface{}, depth int) (kvSecrets map[string]interface{}, invalidSecrets []string, err error) {
	var azkvSecretsRaw *KeyVaultSecretResult
	token, loop := "", 0
	kvSecrets = make(map[string]interface{})
	invalidSecrets = make([]string, 0)

	secretNames := make(map[string]string)
	for _, val := range variables {
		name := val.(map[string]interface{})["name"].(string)
		secretNames[name] = name
	}
	for {
		kvSecretsMap := make(map[string]taskagent.AzureKeyVaultVariableValue)
		if azKVSecrets, err := getKVSecretServiceEndpointProxy(clients, kvName, projectID, serviceEndpointID, token); err == nil {
			azkvSecretsRaw, token, err = parseKVSecretResp(azKVSecrets)
			if err != nil {
				return nil, nil, err
			}
			for _, secret := range *azkvSecretsRaw.Value {
				name := getSecretName(*secret.ID)
				kvVariable := taskagent.AzureKeyVaultVariableValue{
					Value:       nil,
					ContentType: secret.ContentType,
					IsSecret:    converter.Bool(true),
					Enabled:     secret.Enabled,
				}

				if secret.ContentType == nil {
					kvVariable.ContentType = converter.String("")
				}
				if secret.Expire != nil {
					kvVariable.Expires = &azuredevops.Time{
						Time: time.Unix(*secret.Expire, 0),
					}
				}
				kvSecretsMap[name] = kvVariable
			}

			// search secret
			for name, secret := range kvSecretsMap {
				if len(secretNames) == 0 {
					break
				}
				if _, ok := secretNames[name]; ok && *secret.Enabled {
					kvSecrets[name] = secret
					delete(secretNames, name)
				}
			}

			// stop search
			if token == "" || loop == depth || len(secretNames) == 0 {
				for k := range secretNames {
					invalidSecrets = append(invalidSecrets, k)
				}
				return kvSecrets, invalidSecrets, err
			}
			loop++
		} else {
			return nil, nil, fmt.Errorf("Failed to get the Azure Key vault secrets. %+v ", err)
		}
	}
}

func parseKVSecretResp(azKVSecrets *serviceendpoint.ServiceEndpointRequestResult) (*KeyVaultSecretResult, string, error) {
	if azKVSecrets != nil && *azKVSecrets.StatusCode == "ok" {
		var kvSecrets KeyVaultSecretResult
		secretJson := azKVSecrets.Result.([]interface{})[0].(string)
		if err := json.Unmarshal([]byte(secretJson), &kvSecrets); err != nil {
			return nil, "", fmt.Errorf("Failed to parse the Azure Key value secrets. Service response: %s . %+v ", secretJson, err)
		}

		token, err := getSkipToken(kvSecrets.NextLink)
		if err != nil {
			return nil, "", fmt.Errorf("falied to get skip token, error: %+v", err)
		}
		return &kvSecrets, token, nil
	}
	return nil, "", fmt.Errorf("Failed to get the Azure Key value. Error: ( code: %s, messge: %s )", *azKVSecrets.StatusCode, *azKVSecrets.ErrorMessage)
}

func getKVSecretServiceEndpointProxy(clients *client.AggregatedClient, kvName string, projectID string, serviceEndpointID string, token string) (*serviceendpoint.ServiceEndpointRequestResult, error) {
	reqArgs := serviceendpoint.ExecuteServiceEndpointRequestArgs{
		ServiceEndpointRequest: &serviceendpoint.ServiceEndpointRequest{
			DataSourceDetails: &serviceendpoint.DataSourceDetails{
				DataSourceName: converter.String("AzureKeyVaultSecrets"),
				Parameters: &map[string]string{
					"KeyVaultName": kvName,
				},
			},
			ResultTransformationDetails: &serviceendpoint.ResultTransformationDetails{},
		},
		Project:    &projectID,
		EndpointId: &serviceEndpointID,
	}
	if token != "" {
		(*reqArgs.ServiceEndpointRequest.DataSourceDetails.Parameters)["SkipToken"] = token
		reqArgs.ServiceEndpointRequest.DataSourceDetails.DataSourceName = converter.String("AzureKeyVaultSecretsWithSkipToken")
	}
	return clients.ServiceEndpointClient.ExecuteServiceEndpointRequest(clients.Ctx, reqArgs)
}

func getSecretName(secretID string) (secret string) {
	if len(secretID) == 0 {
		return ""
	}
	secretURL := strings.Split(secretID, "/")
	return secretURL[len(secretURL)-1]
}

func getSkipToken(link *string) (string, error) {
	if link == nil || len(*link) == 0 {
		return "", nil
	}
	linkUrl, err := url.Parse(*link)
	if err != nil {
		return "", err
	}

	params, err := url.ParseQuery(linkUrl.RawQuery)
	if err != nil {
		return "", err
	}

	token := params["$skiptoken"]
	if len(token) > 0 {
		return token[0], nil
	}
	// if skip token not found, just return "" as the skip token
	return "", nil
}
