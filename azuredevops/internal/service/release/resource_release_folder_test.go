//go:build (all || resource_release_folder) && !exclude_resource_release_folder
// +build all resource_release_folder
// +build !exclude_resource_release_folder

package release

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/google/uuid"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── Package-level fixtures ─────────────────────────────────────────────────

var testReleaseFolderProjectID = uuid.New()
var testReleaseFolderPath = `\Production\Web`
var testReleaseFolderDescription = "Web production release folder"

var testReleaseFolder = releaseapi.Folder{
	Path:        converter.String(testReleaseFolderPath),
	Description: converter.String(testReleaseFolderDescription),
}

// ── 1. Expand/Flatten roundtrip ───────────────────────────────────────────

// TestReleaseFolder_ExpandFlatten_Roundtrip verifies that expandReleaseFolder
// followed by flattenReleaseFolder preserves path and description losslessly.
func TestReleaseFolder_ExpandFlatten_Roundtrip(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id":  testReleaseFolderProjectID.String(),
		"path":        testReleaseFolderPath,
		"description": testReleaseFolderDescription,
	})

	folder, projectID := expandReleaseFolder(resourceData)
	require.NotNil(t, folder)
	require.Equal(t, testReleaseFolderProjectID.String(), projectID)
	require.Equal(t, testReleaseFolderPath, converter.ToString(folder.Path, ""))
	require.Equal(t, testReleaseFolderDescription, converter.ToString(folder.Description, ""))

	// Now flatten back and verify
	flattenReleaseFolder(resourceData, folder, projectID)
	require.Equal(t, testReleaseFolderProjectID.String(), resourceData.Get("project_id").(string))
	require.Equal(t, testReleaseFolderPath, resourceData.Get("path").(string))
	require.Equal(t, testReleaseFolderDescription, resourceData.Get("description").(string))
}

// ── 2. Create success ─────────────────────────────────────────────────────

// TestReleaseFolder_Create_Success verifies that resourceReleaseFolderCreate
// calls CreateFolder and sets the resource ID to the composite project_id/path.
func TestReleaseFolder_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	releaseClient.
		EXPECT().
		CreateFolder(clients.Ctx, gomock.Any()).
		Return(&testReleaseFolder, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id":  testReleaseFolderProjectID.String(),
		"path":        testReleaseFolderPath,
		"description": testReleaseFolderDescription,
	})

	diags := resourceReleaseFolderCreate(context.Background(), resourceData, clients)
	require.Empty(t, diags)

	expectedID := releaseFolderID(testReleaseFolderProjectID.String(), testReleaseFolderPath)
	require.Equal(t, expectedID, resourceData.Id())
}

// ── 3. Create error propagation ───────────────────────────────────────────

// TestReleaseFolder_Create_Error verifies that an error from CreateFolder
// surfaces as non-empty Diagnostics.
func TestReleaseFolder_Create_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	releaseClient.
		EXPECT().
		CreateFolder(clients.Ctx, gomock.Any()).
		Return(nil, errors.New("CreateFolder() Failed")).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id":  testReleaseFolderProjectID.String(),
		"path":        testReleaseFolderPath,
		"description": "",
	})

	diags := resourceReleaseFolderCreate(context.Background(), resourceData, clients)
	require.NotEmpty(t, diags)
	require.True(t, diags.HasError())
}

// ── 4. Read: 404 clears ID ────────────────────────────────────────────────

// TestReleaseFolder_Read_NotFound verifies that when GetFolders returns a 404
// WrappedError, resourceReleaseFolderRead clears the resource ID and returns
// no diagnostics (graceful drift detection).
func TestReleaseFolder_Read_NotFound(t *testing.T) {
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

	releaseClient.
		EXPECT().
		GetFolders(clients.Ctx, gomock.Any()).
		Return(nil, notFoundErr).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id":  testReleaseFolderProjectID.String(),
		"path":        testReleaseFolderPath,
		"description": "",
	})
	resourceData.SetId(releaseFolderID(testReleaseFolderProjectID.String(), testReleaseFolderPath))

	diags := resourceReleaseFolderRead(context.Background(), resourceData, clients)
	require.Empty(t, diags)
	require.Equal(t, "", resourceData.Id())
}

// ── 5. Delete surfaces error (including non-empty folder) ─────────────────

// TestReleaseFolder_Delete_Error verifies that when DeleteFolder returns an error
// (including a "contains release definitions" error), resourceReleaseFolderDelete
// surfaces it as non-empty Diagnostics with the human-readable message.
func TestReleaseFolder_Delete_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	// Simulate ADO error indicating folder contains release definitions
	releaseClient.
		EXPECT().
		DeleteFolder(clients.Ctx, gomock.Any()).
		Return(errors.New("folder contains release definitions and cannot be deleted")).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id":  testReleaseFolderProjectID.String(),
		"path":        testReleaseFolderPath,
		"description": "",
	})
	resourceData.SetId(releaseFolderID(testReleaseFolderProjectID.String(), testReleaseFolderPath))

	diags := resourceReleaseFolderDelete(context.Background(), resourceData, clients)
	require.NotEmpty(t, diags)
	require.True(t, diags.HasError())
	// AC2: must contain the human-readable message
	require.Contains(t, diags[0].Summary, "Cannot delete folder")
	require.Contains(t, diags[0].Summary, "contains release definitions")
}
