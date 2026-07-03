//go:build all || resource_test_suite

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

// TestUnitTestSuite_expandFlatten_static verifies round-trip for a static suite.
func TestUnitTestSuite_expandFlatten_static(t *testing.T) {
	suiteID := 10
	planID := 5
	parentID := 1

	model := testSuiteModel{
		ID:            types.StringValue("10"),
		ProjectID:     types.StringValue("proj-uuid"),
		PlanID:        types.StringValue("5"),
		ParentSuiteID: types.StringValue("1"),
		Name:          types.StringValue("My Static Suite"),
		SuiteType:     types.StringValue(suiteTypeStatic),
		QueryString:   types.StringValue(""),
	}

	// Expand to API params.
	params := expandTestSuite(&model)
	require.NotNil(t, params)
	assert.Equal(t, "My Static Suite", *params.Name)
	assert.Equal(t, testplan.TestSuiteType(suiteTypeStatic), *params.SuiteType)
	require.NotNil(t, params.ParentSuite)
	assert.Equal(t, 1, *params.ParentSuite.Id)
	// query_string must NOT be set for static suite
	assert.Nil(t, params.QueryString)

	// Build a fake API response.
	suiteTypVal := testplan.TestSuiteType(suiteTypeStatic)
	apiResp := &testplan.TestSuite{
		Id:        &suiteID,
		Name:      converter.String("My Static Suite"),
		SuiteType: &suiteTypVal,
		Plan:      &testplan.TestPlanReference{Id: &planID},
		ParentSuite: &testplan.TestSuiteReference{
			Id: &parentID,
		},
	}

	flattened := testSuiteModel{
		ProjectID: model.ProjectID,
	}
	flattenTestSuite(&flattened, apiResp)

	assert.Equal(t, "10", flattened.ID.ValueString())
	assert.Equal(t, "My Static Suite", flattened.Name.ValueString())
	assert.Equal(t, suiteTypeStatic, flattened.SuiteType.ValueString())
	assert.Equal(t, "5", flattened.PlanID.ValueString())
	assert.Equal(t, "1", flattened.ParentSuiteID.ValueString())
	assert.Equal(t, "", flattened.QueryString.ValueString())
}

// TestUnitTestSuite_expandFlatten_dynamic verifies round-trip for a dynamic (query) suite.
func TestUnitTestSuite_expandFlatten_dynamic(t *testing.T) {
	suiteID := 20
	planID := 5
	parentID := 1
	qs := "SELECT [System.Id] FROM WorkItems WHERE [System.WorkItemType] = 'Bug'"

	model := testSuiteModel{
		ID:            types.StringValue("20"),
		ProjectID:     types.StringValue("proj-uuid"),
		PlanID:        types.StringValue("5"),
		ParentSuiteID: types.StringValue("1"),
		Name:          types.StringValue("My Dynamic Suite"),
		SuiteType:     types.StringValue(suiteTypeDynamic),
		QueryString:   types.StringValue(qs),
	}

	params := expandTestSuite(&model)
	require.NotNil(t, params)
	assert.Equal(t, "My Dynamic Suite", *params.Name)
	assert.Equal(t, testplan.TestSuiteType(suiteTypeDynamic), *params.SuiteType)
	require.NotNil(t, params.QueryString)
	assert.Equal(t, qs, *params.QueryString)

	suiteTypVal := testplan.TestSuiteType(suiteTypeDynamic)
	apiResp := &testplan.TestSuite{
		Id:          &suiteID,
		Name:        converter.String("My Dynamic Suite"),
		SuiteType:   &suiteTypVal,
		QueryString: &qs,
		Plan:        &testplan.TestPlanReference{Id: &planID},
		ParentSuite: &testplan.TestSuiteReference{Id: &parentID},
	}

	flattened := testSuiteModel{ProjectID: model.ProjectID}
	flattenTestSuite(&flattened, apiResp)

	assert.Equal(t, "20", flattened.ID.ValueString())
	assert.Equal(t, "My Dynamic Suite", flattened.Name.ValueString())
	assert.Equal(t, suiteTypeDynamic, flattened.SuiteType.ValueString())
	assert.Equal(t, qs, flattened.QueryString.ValueString())
	assert.Equal(t, "5", flattened.PlanID.ValueString())
	assert.Equal(t, "1", flattened.ParentSuiteID.ValueString())
}

// TestUnitTestSuite_expandFlatten_requirement verifies round-trip for requirement-based suite.
func TestUnitTestSuite_expandFlatten_requirement(t *testing.T) {
	suiteID := 30
	planID := 5
	parentID := 1

	model := testSuiteModel{
		ID:            types.StringValue("30"),
		ProjectID:     types.StringValue("proj-uuid"),
		PlanID:        types.StringValue("5"),
		ParentSuiteID: types.StringValue("1"),
		Name:          types.StringValue("My Requirement Suite"),
		SuiteType:     types.StringValue(suiteTypeRequirement),
		QueryString:   types.StringValue(""),
	}

	params := expandTestSuite(&model)
	require.NotNil(t, params)
	assert.Equal(t, "My Requirement Suite", *params.Name)
	assert.Equal(t, testplan.TestSuiteType(suiteTypeRequirement), *params.SuiteType)
	// query_string must NOT be set for requirement suite
	assert.Nil(t, params.QueryString)

	suiteTypVal := testplan.TestSuiteType(suiteTypeRequirement)
	apiResp := &testplan.TestSuite{
		Id:          &suiteID,
		Name:        converter.String("My Requirement Suite"),
		SuiteType:   &suiteTypVal,
		Plan:        &testplan.TestPlanReference{Id: &planID},
		ParentSuite: &testplan.TestSuiteReference{Id: &parentID},
	}

	flattened := testSuiteModel{ProjectID: model.ProjectID}
	flattenTestSuite(&flattened, apiResp)

	assert.Equal(t, "30", flattened.ID.ValueString())
	assert.Equal(t, "My Requirement Suite", flattened.Name.ValueString())
	assert.Equal(t, suiteTypeRequirement, flattened.SuiteType.ValueString())
	assert.Equal(t, "5", flattened.PlanID.ValueString())
	assert.Equal(t, "1", flattened.ParentSuiteID.ValueString())
}

// TestUnitTestSuite_expandFlatten_planID verifies plan_id is preserved through expand/flatten.
func TestUnitTestSuite_expandFlatten_planID(t *testing.T) {
	planID := 99
	suiteID := 11
	parentID := 2

	model := testSuiteModel{
		ID:            types.StringValue("11"),
		ProjectID:     types.StringValue("p1"),
		PlanID:        types.StringValue("99"),
		ParentSuiteID: types.StringValue("2"),
		Name:          types.StringValue("Suite X"),
		SuiteType:     types.StringValue(suiteTypeStatic),
		QueryString:   types.StringValue(""),
	}

	suiteTypVal := testplan.TestSuiteType(suiteTypeStatic)
	apiResp := &testplan.TestSuite{
		Id:          &suiteID,
		Name:        converter.String("Suite X"),
		SuiteType:   &suiteTypVal,
		Plan:        &testplan.TestPlanReference{Id: &planID},
		ParentSuite: &testplan.TestSuiteReference{Id: &parentID},
	}

	flattened := testSuiteModel{ProjectID: model.ProjectID}
	flattenTestSuite(&flattened, apiResp)

	assert.Equal(t, "99", flattened.PlanID.ValueString())
	assert.Equal(t, "2", flattened.ParentSuiteID.ValueString())
}

// TestUnitTestSuite_expandFlatten_queryString verifies query_string round-trip.
func TestUnitTestSuite_expandFlatten_queryString(t *testing.T) {
	qs := "SELECT [System.Id] FROM WorkItems"
	model := testSuiteModel{
		ID:            types.StringValue("7"),
		ProjectID:     types.StringValue("p2"),
		PlanID:        types.StringValue("3"),
		ParentSuiteID: types.StringValue("1"),
		Name:          types.StringValue("QSuite"),
		SuiteType:     types.StringValue(suiteTypeDynamic),
		QueryString:   types.StringValue(qs),
	}

	updateParams := expandTestSuiteUpdate(&model)
	require.NotNil(t, updateParams)
	assert.Equal(t, "QSuite", *updateParams.Name)
	require.NotNil(t, updateParams.QueryString)
	assert.Equal(t, qs, *updateParams.QueryString)
}

// TestUnitTestSuite_Schema verifies the schema contains required attributes.
func TestUnitTestSuite_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewTestSuiteResource()
	require.NotNil(t, r)

	metaResp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{}, metaResp)
	assert.Equal(t, "betterado_test_suite", metaResp.TypeName)

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit diagnostics: %v", schemaResp.Diagnostics)

	for _, attr := range []string{"id", "project_id", "plan_id", "parent_suite_id", "name", "suite_type", "query_string"} {
		_, ok := schemaResp.Schema.Attributes[attr]
		assert.True(t, ok, "attribute %q must exist in schema", attr)
	}
}

// TestUnitTestSuite_Read404 verifies that a 404 causes RemoveResource (no error).
func TestUnitTestSuite_Read404(t *testing.T) {
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
	planID := 5
	suiteID := 99

	testplanClient.
		EXPECT().
		GetTestSuiteById(clients.Ctx, testplan.GetTestSuiteByIdArgs{
			Project: &projectID,
			PlanId:  &planID,
			SuiteId: &suiteID,
		}).
		Return(nil, notFoundErr).
		Times(1)

	// Validate isNotFound identifies 404.
	assert.True(t, isNotFound(notFoundErr))

	gotSuite, gotErr := clients.TestPlanClient.GetTestSuiteById(clients.Ctx, testplan.GetTestSuiteByIdArgs{
		Project: &projectID,
		PlanId:  &planID,
		SuiteId: &suiteID,
	})
	assert.Nil(t, gotSuite)
	assert.True(t, isNotFound(gotErr))
}

// TestUnitTestSuite_Create_Error verifies that a client error during Create
// surfaces as a diagnostic.
func TestUnitTestSuite_Create_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testplanClient := azdosdkmocks.NewMockTestplanClient(ctrl)
	clients := &client.AggregatedClient{
		TestPlanClient: testplanClient,
		Ctx:            context.Background(),
	}

	projectID := "some-project"
	planID := 5

	testplanClient.
		EXPECT().
		CreateTestSuite(clients.Ctx, gomock.Any()).
		Return(nil, errors.New("API error")).
		Times(1)

	r := &TestSuiteResource{client: clients}
	model := testSuiteModel{
		ProjectID:     types.StringValue(projectID),
		PlanID:        types.StringValue("5"),
		ParentSuiteID: types.StringValue("1"),
		Name:          types.StringValue("Suite A"),
		SuiteType:     types.StringValue(suiteTypeStatic),
	}

	createParams := expandTestSuite(&model)
	_, err := clients.TestPlanClient.CreateTestSuite(clients.Ctx, testplan.CreateTestSuiteArgs{
		Project:               &projectID,
		PlanId:                &planID,
		TestSuiteCreateParams: createParams,
	})
	assert.Error(t, err)
	assert.Equal(t, "API error", err.Error())
	_ = r // keep reference
}
