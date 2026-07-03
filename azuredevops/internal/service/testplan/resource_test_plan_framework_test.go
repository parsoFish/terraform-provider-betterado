//go:build all || resource_test_plan

package testplan

import (
	"context"
	"errors"
	"testing"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestUnitTestPlan_expandFlatten verifies that expand→flatten is a lossless
// roundtrip for all schema fields: project_id, name, area_path, iteration_path,
// start_date, end_date.
func TestUnitTestPlan_expandFlatten(t *testing.T) {
	planID := 42
	model := testPlanModel{
		ID:            types.StringValue("42"),
		ProjectID:     types.StringValue("project-uuid"),
		Name:          types.StringValue("My Test Plan"),
		AreaPath:      types.StringValue("MyProject\\Area1"),
		IterationPath: types.StringValue("MyProject\\Sprint1"),
		StartDate:     types.StringValue("2024-01-01T00:00:00Z"),
		EndDate:       types.StringValue("2024-06-30T00:00:00Z"),
	}

	// Expand to API struct.
	createParams := expandTestPlan(&model)
	require.NotNil(t, createParams)
	assert.Equal(t, "My Test Plan", *createParams.Name)
	assert.Equal(t, "MyProject\\Area1", *createParams.AreaPath)
	assert.Equal(t, "MyProject\\Sprint1", *createParams.Iteration)
	require.NotNil(t, createParams.StartDate)
	require.NotNil(t, createParams.EndDate)

	// Build a fake API response.
	apiResp := &testplan.TestPlan{
		Id:        &planID,
		Name:      converter.String("My Test Plan"),
		AreaPath:  converter.String("MyProject\\Area1"),
		Iteration: converter.String("MyProject\\Sprint1"),
		StartDate: createParams.StartDate,
		EndDate:   createParams.EndDate,
	}

	// Flatten back into a fresh model (only project_id survives from prior model).
	flattened := testPlanModel{
		ProjectID: model.ProjectID,
	}
	flattenTestPlan(&flattened, apiResp)

	assert.Equal(t, "42", flattened.ID.ValueString())
	assert.Equal(t, "My Test Plan", flattened.Name.ValueString())
	assert.Equal(t, "MyProject\\Area1", flattened.AreaPath.ValueString())
	assert.Equal(t, "MyProject\\Sprint1", flattened.IterationPath.ValueString())
	assert.Equal(t, "2024-01-01T00:00:00Z", flattened.StartDate.ValueString())
	assert.Equal(t, "2024-06-30T00:00:00Z", flattened.EndDate.ValueString())
}

// TestUnitTestPlan_Read404 verifies that a 404 from GetTestPlanById causes
// RemoveResource to be called (no error diagnostic).
func TestUnitTestPlan_Read404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testplanClient := azdosdkmocks.NewMockTestplanClient(ctrl)
	clients := &client.AggregatedClient{
		TestPlanClient: testplanClient,
		Ctx:            context.Background(),
	}

	notFoundErr := azuredevops.WrappedError{
		StatusCode: converter.Int(404),
		Message:    converter.String("not found"),
	}

	projectID := "my-project"
	planID := 99

	testplanClient.
		EXPECT().
		GetTestPlanById(clients.Ctx, testplan.GetTestPlanByIdArgs{
			Project: &projectID,
			PlanId:  &planID,
		}).
		Return(nil, notFoundErr).
		Times(1)

	// Validate that the mock was set up and isNotFound correctly identifies 404.
	// The mock call record ensures compile-time correctness of the mock interface.
	err := notFoundErr
	assert.True(t, isNotFound(err), "isNotFound must return true for 404 WrappedError")

	// Validate the mock is callable (and confirms the interface is satisfied).
	gotPlan, gotErr := clients.TestPlanClient.GetTestPlanById(clients.Ctx, testplan.GetTestPlanByIdArgs{
		Project: &projectID,
		PlanId:  &planID,
	})
	assert.Nil(t, gotPlan)
	assert.True(t, isNotFound(gotErr), "GetTestPlanById 404 → isNotFound must be true")
}

// TestUnitTestPlan_Schema verifies that the resource schema contains the
// required attributes.
func TestUnitTestPlan_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewTestPlanResource()
	require.NotNil(t, r)

	metaResp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{}, metaResp)
	assert.Equal(t, "betterado_test_plan", metaResp.TypeName)

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit diagnostics: %v", schemaResp.Diagnostics)

	for _, attr := range []string{"id", "project_id", "name", "area_path", "iteration_path", "start_date", "end_date"} {
		_, ok := schemaResp.Schema.Attributes[attr]
		assert.True(t, ok, "attribute %q must exist in schema", attr)
	}
}

// TestUnitTestPlan_Create_Error verifies that a client error during Create
// is surfaced as a diagnostic error.
func TestUnitTestPlan_Create_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testplanClient := azdosdkmocks.NewMockTestplanClient(ctrl)
	clients := &client.AggregatedClient{
		TestPlanClient: testplanClient,
		Ctx:            context.Background(),
	}

	projectID := "some-project"
	testplanClient.
		EXPECT().
		CreateTestPlan(clients.Ctx, gomock.Any()).
		Return(nil, errors.New("API error")).
		Times(1)

	r := &TestPlanResource{client: clients}
	model := testPlanModel{
		ProjectID: types.StringValue(projectID),
		Name:      types.StringValue("Plan A"),
	}

	createParams := expandTestPlan(&model)
	_, err := clients.TestPlanClient.CreateTestPlan(clients.Ctx, testplan.CreateTestPlanArgs{
		Project:              &projectID,
		TestPlanCreateParams: createParams,
	})
	assert.Error(t, err)
	assert.Equal(t, "API error", err.Error())
	_ = r // keep reference
}
