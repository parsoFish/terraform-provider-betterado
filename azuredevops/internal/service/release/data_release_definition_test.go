//go:build (all || resource_release_definition) && !exclude_resource_release_definition
// +build all resource_release_definition
// +build !exclude_resource_release_definition

package release

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestDataReleaseDefinition_ReadById_Populates verifies that when GetReleaseDefinition
// returns a definition, the data source sets Id and key attribute values in state.
func TestDataReleaseDefinition_ReadById_Populates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	defID := 10
	projectID := testReleaseDefinitionProjectID.String()
	fixture := &releaseapi.ReleaseDefinition{
		Id:                &defID,
		Name:              converter.String("MyDataDef"),
		Path:              converter.String("\\staging"),
		Description:       converter.String("A description"),
		ReleaseNameFormat: converter.String("Release-$(rev:r)"),
	}

	releaseClient.
		EXPECT().
		GetReleaseDefinition(clients.Ctx, releaseapi.GetReleaseDefinitionArgs{
			Project:      &projectID,
			DefinitionId: &defID,
		}).
		Return(fixture, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinition().Schema, map[string]interface{}{
		"project_id":            projectID,
		"release_definition_id": defID,
		"name":                  "",
	})

	err := dataReleaseDefinitionRead(resourceData, clients)
	require.NoError(t, err)
	require.Equal(t, "10", resourceData.Id())
	require.Equal(t, "MyDataDef", resourceData.Get("name").(string))
	require.Equal(t, "\\staging", resourceData.Get("path").(string))
	require.Equal(t, "A description", resourceData.Get("description").(string))
	require.Equal(t, "Release-$(rev:r)", resourceData.Get("release_name_format").(string))
}

// TestDataReleaseDefinition_ReadByName_Populates verifies that when only name is
// provided the data source calls GetReleaseDefinitions with a name filter and
// populates state from the single matching result.
func TestDataReleaseDefinition_ReadByName_Populates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	defID := 20
	projectID := testReleaseDefinitionProjectID.String()
	defName := "NamedDef"

	fixture := releaseapi.ReleaseDefinition{
		Id:                &defID,
		Name:              converter.String(defName),
		Path:              converter.String("\\"),
		Description:       converter.String(""),
		ReleaseNameFormat: converter.String("Release-$(rev:r)"),
	}

	releaseClient.
		EXPECT().
		GetReleaseDefinitions(clients.Ctx, releaseapi.GetReleaseDefinitionsArgs{
			Project:    &projectID,
			SearchText: converter.String(defName),
		}).
		Return(&releaseapi.GetReleaseDefinitionsResponseValue{
			Value: []releaseapi.ReleaseDefinition{fixture},
		}, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinition().Schema, map[string]interface{}{
		"project_id":            projectID,
		"release_definition_id": 0,
		"name":                  defName,
	})

	err := dataReleaseDefinitionRead(resourceData, clients)
	require.NoError(t, err)
	require.Equal(t, "20", resourceData.Id())
	require.Equal(t, defName, resourceData.Get("name").(string))
	require.Equal(t, "\\", resourceData.Get("path").(string))
	require.Equal(t, "Release-$(rev:r)", resourceData.Get("release_name_format").(string))
}

// TestDataReleaseDefinition_Read_404ReturnsError verifies that when GetReleaseDefinition
// returns a 404 WrappedError, the data source returns a non-nil error and clears the ID.
func TestDataReleaseDefinition_Read_404ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	notFoundStatusCode := http.StatusNotFound
	notFoundErr := azuredevops.WrappedError{
		StatusCode: &notFoundStatusCode,
	}

	defID := 99
	projectID := testReleaseDefinitionProjectID.String()

	releaseClient.
		EXPECT().
		GetReleaseDefinition(clients.Ctx, releaseapi.GetReleaseDefinitionArgs{
			Project:      &projectID,
			DefinitionId: &defID,
		}).
		Return(nil, notFoundErr).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinition().Schema, map[string]interface{}{
		"project_id":            projectID,
		"release_definition_id": defID,
		"name":                  "",
	})
	resourceData.SetId("99")

	err := dataReleaseDefinitionRead(resourceData, clients)
	require.Error(t, err)
	require.Equal(t, "", resourceData.Id())
}
