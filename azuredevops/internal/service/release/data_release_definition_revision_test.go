//go:build (all || resource_release_definition) && !exclude_resource_release_definition
// +build all resource_release_definition
// +build !exclude_resource_release_definition

package release

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestDataReleaseDefinitionRevision_Read_ReturnsJSON verifies that when
// GetDefinitionRevision returns a valid ReadCloser the data source drains it
// into json_content and returns no error.
func TestDataReleaseDefinitionRevision_Read_ReturnsJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	projectID := testReleaseDefinitionProjectID.String()
	defID := 42
	rev := 3
	jsonBody := `{"id":1}`

	releaseClient.
		EXPECT().
		GetDefinitionRevision(clients.Ctx, releaseapi.GetDefinitionRevisionArgs{
			Project:      &projectID,
			DefinitionId: &defID,
			Revision:     &rev,
		}).
		Return(io.NopCloser(bytes.NewBufferString(jsonBody)), nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinitionRevision().Schema, map[string]interface{}{
		"project_id":            projectID,
		"release_definition_id": defID,
		"revision":              rev,
	})

	err := dataReleaseDefinitionRevisionRead(resourceData, clients)
	require.NoError(t, err)
	require.Equal(t, jsonBody, resourceData.Get("json_content").(string))
	require.Equal(t, fmt.Sprintf("%s-%d-%d", projectID, defID, rev), resourceData.Id())
}

// TestDataReleaseDefinitionRevision_Read_Error verifies that when
// GetDefinitionRevision returns an error the data source propagates it and
// clears the resource ID.
func TestDataReleaseDefinitionRevision_Read_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	projectID := testReleaseDefinitionProjectID.String()
	defID := 42
	rev := 3

	releaseClient.
		EXPECT().
		GetDefinitionRevision(clients.Ctx, releaseapi.GetDefinitionRevisionArgs{
			Project:      &projectID,
			DefinitionId: &defID,
			Revision:     &rev,
		}).
		Return(nil, fmt.Errorf("api error")).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinitionRevision().Schema, map[string]interface{}{
		"project_id":            projectID,
		"release_definition_id": defID,
		"revision":              rev,
	})
	resourceData.SetId("some-prior-id")

	err := dataReleaseDefinitionRevisionRead(resourceData, clients)
	require.Error(t, err)
	require.Equal(t, "", resourceData.Id())
}
