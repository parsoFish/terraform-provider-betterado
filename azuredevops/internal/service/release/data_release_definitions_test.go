//go:build (all || resource_release_definition) && !exclude_resource_release_definition
// +build all resource_release_definition
// +build !exclude_resource_release_definition

package release

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestDataReleaseDefinitions_ReadList_PopulatesAll verifies that when
// GetReleaseDefinitions returns 2 definitions, the data source sets
// release_definitions with length 2 and the first item's name is correct.
func TestDataReleaseDefinitions_ReadList_PopulatesAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	projectID := testReleaseDefinitionProjectID.String()
	defID1 := 1
	defID2 := 2

	releaseClient.
		EXPECT().
		GetReleaseDefinitions(clients.Ctx, releaseapi.GetReleaseDefinitionsArgs{
			Project: &projectID,
		}).
		Return(&releaseapi.GetReleaseDefinitionsResponseValue{
			Value: []releaseapi.ReleaseDefinition{
				{
					Id:   &defID1,
					Name: converter.String("Alpha"),
					Path: converter.String("\\"),
				},
				{
					Id:   &defID2,
					Name: converter.String("Beta"),
					Path: converter.String("\\staging"),
				},
			},
		}, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinitions().Schema, map[string]interface{}{
		"project_id": projectID,
		"path":       "",
		"name":       "",
	})

	diags := dataReleaseDefinitionsRead(context.Background(), resourceData, clients)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	rawList := resourceData.Get("release_definitions").([]interface{})
	require.Len(t, rawList, 2)

	first := rawList[0].(map[string]interface{})
	require.Equal(t, "Alpha", first["name"])
	require.Equal(t, "1", first["id"])
	require.Equal(t, "\\", first["path"])
}

// TestDataReleaseDefinitions_ReadList_WithPathFilter verifies that when a
// non-empty path filter is provided, GetReleaseDefinitions is called with that
// path and only the returned definitions appear in the list.
func TestDataReleaseDefinitions_ReadList_WithPathFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	projectID := testReleaseDefinitionProjectID.String()
	filterPath := "\\staging"
	defID := 42

	releaseClient.
		EXPECT().
		GetReleaseDefinitions(clients.Ctx, releaseapi.GetReleaseDefinitionsArgs{
			Project: &projectID,
			Path:    &filterPath,
		}).
		Return(&releaseapi.GetReleaseDefinitionsResponseValue{
			Value: []releaseapi.ReleaseDefinition{
				{
					Id:   &defID,
					Name: converter.String("StagingDef"),
					Path: converter.String(filterPath),
				},
			},
		}, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinitions().Schema, map[string]interface{}{
		"project_id": projectID,
		"path":       filterPath,
		"name":       "",
	})

	diags := dataReleaseDefinitionsRead(context.Background(), resourceData, clients)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	rawList := resourceData.Get("release_definitions").([]interface{})
	require.Len(t, rawList, 1)
	item := rawList[0].(map[string]interface{})
	require.Equal(t, "StagingDef", item["name"])
}

// TestDataReleaseDefinitions_ReadList_APIErrorSurfaces verifies that when
// GetReleaseDefinitions returns a generic API error, the function returns a
// non-empty Diagnostics (i.e. the error is surfaced to the caller).
func TestDataReleaseDefinitions_ReadList_APIErrorSurfaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	projectID := testReleaseDefinitionProjectID.String()

	releaseClient.
		EXPECT().
		GetReleaseDefinitions(clients.Ctx, releaseapi.GetReleaseDefinitionsArgs{
			Project: &projectID,
		}).
		Return(nil, errors.New("internal server error")).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinitions().Schema, map[string]interface{}{
		"project_id": projectID,
		"path":       "",
		"name":       "",
	})

	diags := dataReleaseDefinitionsRead(context.Background(), resourceData, clients)
	require.True(t, diags.HasError(), "expected diagnostics to have an error")
	require.Contains(t, diags[0].Summary, "internal server error")
}
