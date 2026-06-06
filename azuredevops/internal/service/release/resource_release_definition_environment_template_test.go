//go:build (all || resource_release_definition_environment_template) && !exclude_resource_release_definition_environment_template
// +build all resource_release_definition_environment_template
// +build !exclude_resource_release_definition_environment_template

package release

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/environmenttemplates"
)

// TestReleaseDefinitionEnvironmentTemplate_Expand verifies that expandEnvironmentTemplate
// correctly converts Terraform resource data into a ReleaseDefinitionEnvironmentTemplate.
func TestReleaseDefinitionEnvironmentTemplate_Expand(t *testing.T) {
	r := ResourceReleaseDefinitionEnvironmentTemplate()

	d := schema.InternalMap(r.Schema).InternalValidate(schemaMap(r.Schema))
	require.Nil(t, d)

	resourceData := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"project_id":  "00000000-0000-0000-0000-000000000001",
		"name":        "My Template",
		"description": "A test template",
		"category":    "Deploy",
		"icon_task_id": "00000000-0000-0000-0000-000000000002",
		"environment": "",
	})

	template := expandEnvironmentTemplate(resourceData)

	require.NotNil(t, template)
	assert.Equal(t, "My Template", *template.Name)
	assert.Equal(t, "A test template", *template.Description)
	assert.Equal(t, "Deploy", *template.Category)

	expectedIconTaskID, err := uuid.Parse("00000000-0000-0000-0000-000000000002")
	require.NoError(t, err)
	require.NotNil(t, template.IconTaskId)
	assert.Equal(t, expectedIconTaskID, *template.IconTaskId)
}

// TestReleaseDefinitionEnvironmentTemplate_Flatten verifies that flattenEnvironmentTemplate
// correctly populates Terraform resource state from a ReleaseDefinitionEnvironmentTemplate.
func TestReleaseDefinitionEnvironmentTemplate_Flatten(t *testing.T) {
	r := ResourceReleaseDefinitionEnvironmentTemplate()

	d := schema.InternalMap(r.Schema).InternalValidate(schemaMap(r.Schema))
	require.Nil(t, d)

	templateID, err := uuid.Parse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	require.NoError(t, err)

	canDelete := true
	iconURI := "https://example.com/icon.png"

	template := &environmenttemplates.ReleaseDefinitionEnvironmentTemplate{
		Id:          &templateID,
		Name:        strPtr("My Flatten Template"),
		Description: strPtr("Flatten description"),
		Category:    strPtr("Build"),
		CanDelete:   &canDelete,
		IconUri:     &iconURI,
	}

	resourceData := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"project_id": "00000000-0000-0000-0000-000000000001",
	})

	flattenEnvironmentTemplate(template, resourceData)

	assert.Equal(t, "My Flatten Template", resourceData.Get("name"))
	assert.Equal(t, "Flatten description", resourceData.Get("description"))
	assert.Equal(t, "Build", resourceData.Get("category"))
	assert.Equal(t, true, resourceData.Get("can_delete"))
	assert.Equal(t, "https://example.com/icon.png", resourceData.Get("icon_uri"))
}

// schemaMap is a local alias used by InternalValidate; keeps the test self-contained.
type schemaMap = map[string]*schema.Schema

// strPtr returns a pointer to the given string value.
func strPtr(s string) *string {
	return &s
}
