//go:build (all || data_task_group) && !exclude_data_task_group
// +build all data_task_group
// +build !exclude_data_task_group

package taskagent

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestDataTaskGroup_Read_Populates verifies that dataTaskGroupRead sets
// the expected state attributes when the mock returns a populated TaskGroup.
func TestDataTaskGroup_Read_Populates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskAgentClient := azdosdkmocks.NewMockTaskagentClient(ctrl)
	clients := &client.AggregatedClient{
		TaskAgentClient: taskAgentClient,
		Ctx:             context.Background(),
	}

	taskAgentClient.
		EXPECT().
		GetTaskGroups(clients.Ctx, gomock.Any()).
		Return(&[]taskagent.TaskGroup{testTaskGroup}, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataTaskGroup().Schema, map[string]interface{}{
		"project_id": testTaskGroupProjectID.String(),
		"id":         testTaskGroupID.String(),
	})

	err := dataTaskGroupRead(resourceData, clients)
	require.NoError(t, err)
	require.Equal(t, testTaskGroupID.String(), resourceData.Id())
	require.Equal(t, "MyTaskGroup", resourceData.Get("name"))
}

// TestDataTaskGroup_Read_NotFound verifies that dataTaskGroupRead returns an
// error containing "not found" when the mock returns a 404 WrappedError.
func TestDataTaskGroup_Read_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskAgentClient := azdosdkmocks.NewMockTaskagentClient(ctrl)
	clients := &client.AggregatedClient{
		TaskAgentClient: taskAgentClient,
		Ctx:             context.Background(),
	}

	notFoundStatusCode := http.StatusNotFound
	notFoundErr := azuredevops.WrappedError{
		StatusCode: &notFoundStatusCode,
	}

	taskAgentClient.
		EXPECT().
		GetTaskGroups(clients.Ctx, gomock.Any()).
		Return(nil, notFoundErr).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataTaskGroup().Schema, map[string]interface{}{
		"project_id": testTaskGroupProjectID.String(),
		"id":         testTaskGroupID.String(),
	})

	err := dataTaskGroupRead(resourceData, clients)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}
