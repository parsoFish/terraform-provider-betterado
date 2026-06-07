//go:build (all || data_release_folder) && !exclude_data_release_folder
// +build all data_release_folder
// +build !exclude_data_release_folder

package release

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestDataReleaseFolder_Read_Populates verifies that when GetFolders returns a matching
// folder, dataReleaseFolderRead sets Id to path and populates description with no error.
func TestDataReleaseFolder_Read_Populates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRelease := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: mockRelease,
		Ctx:           context.Background(),
	}

	projectID := testReleaseFolderProjectID.String()
	folders := &[]releaseapi.Folder{
		{
			Path:        converter.String(testReleaseFolderPath),
			Description: converter.String(testReleaseFolderDescription),
		},
	}

	mockRelease.EXPECT().
		GetFolders(gomock.Any(), releaseapi.GetFoldersArgs{
			Project: &projectID,
			Path:    converter.String(testReleaseFolderPath),
		}).
		Return(folders, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseFolder().Schema, map[string]interface{}{
		"project_id": projectID,
		"path":       testReleaseFolderPath,
	})

	err := dataReleaseFolderRead(resourceData, clients)
	require.NoError(t, err)
	require.Equal(t, testReleaseFolderPath, resourceData.Id())
	require.Equal(t, testReleaseFolderDescription, resourceData.Get("description").(string))
}

// TestDataReleaseFolder_Read_NotFound verifies that when GetFolders returns an empty slice,
// dataReleaseFolderRead returns an error containing the path.
func TestDataReleaseFolder_Read_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRelease := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: mockRelease,
		Ctx:           context.Background(),
	}

	projectID := testReleaseFolderProjectID.String()
	emptyFolders := &[]releaseapi.Folder{}

	mockRelease.EXPECT().
		GetFolders(gomock.Any(), gomock.Any()).
		Return(emptyFolders, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseFolder().Schema, map[string]interface{}{
		"project_id": projectID,
		"path":       testReleaseFolderPath,
	})

	err := dataReleaseFolderRead(resourceData, clients)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
	require.Contains(t, err.Error(), testReleaseFolderPath)
}
