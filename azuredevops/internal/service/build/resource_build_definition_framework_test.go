//go:build all || resource_build_definition_framework

package build

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildDefinitionFramework_Schema calls Schema() on the framework resource and
// asserts that the schema declares all required attributes with no error diagnostics.
//
// This is the quality-gate test; the gate command is:
//
//	go test -tags all -run TestBuildDefinitionFramework_Schema ./azuredevops/internal/service/build/...
func TestBuildDefinitionFramework_Schema(t *testing.T) {
	t.Parallel()

	r := &BuildDefinitionResource{}

	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}

	r.Schema(context.Background(), schemaReq, schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not return error diagnostics: %s", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes

	// Top-level required attributes per AC1 and AC2.
	requiredAttrs := []string{
		"name",
		"project_id",
		"revision",
		"path",
		"agent_pool_name",
		"repository",
		"variable",
		"ci_trigger",
		"pull_request_trigger",
		"agent_specification",
		"job_authorization_scope",
		"queue_status",
		"skip_first_run",
	}
	for _, attr := range requiredAttrs {
		_, ok := attrs[attr]
		assert.True(t, ok, "schema must declare %q attribute", attr)
	}

	// Spot-check that nested attribute blocks exist and are non-nil.
	for _, nestedAttrName := range []string{"repository", "variable", "ci_trigger", "pull_request_trigger"} {
		a, exists := attrs[nestedAttrName]
		assert.True(t, exists, "schema must declare %q attribute", nestedAttrName)
		assert.NotNil(t, a, "schema attribute %q must not be nil", nestedAttrName)
	}

	// AC1: ci_trigger must have override sub-block.
	ciTriggerAttr, ok := attrs["ci_trigger"]
	require.True(t, ok, "schema must have ci_trigger")
	ciList, ok := ciTriggerAttr.(schema.ListNestedAttribute)
	require.True(t, ok, "ci_trigger must be ListNestedAttribute")
	_, hasOverride := ciList.NestedObject.Attributes["override"]
	assert.True(t, hasOverride, "ci_trigger must have override sub-block")

	// AC1: pull_request_trigger must have override and forks (Required) sub-blocks.
	prTriggerAttr, ok := attrs["pull_request_trigger"]
	require.True(t, ok, "schema must have pull_request_trigger")
	prList, ok := prTriggerAttr.(schema.ListNestedAttribute)
	require.True(t, ok, "pull_request_trigger must be ListNestedAttribute")
	_, hasForks := prList.NestedObject.Attributes["forks"]
	assert.True(t, hasForks, "pull_request_trigger must have forks sub-block (Required for SDKv2 parity)")
	_, hasOverridePR := prList.NestedObject.Attributes["override"]
	assert.True(t, hasOverridePR, "pull_request_trigger must have override sub-block")

	// AC3: validators must be present on key attributes.
	// name must have at least one validator (StringIsNotWhiteSpace parity).
	nameAttr, ok := attrs["name"]
	require.True(t, ok, "schema must have name attribute")
	nameStr, ok := nameAttr.(schema.StringAttribute)
	require.True(t, ok, "name must be StringAttribute")
	assert.NotEmpty(t, nameStr.Validators, "name attribute must have validators (StringIsNotWhiteSpace parity)")

	// path must have at least one validator (path format parity).
	pathAttr, ok := attrs["path"]
	require.True(t, ok, "schema must have path attribute")
	pathStr, ok := pathAttr.(schema.StringAttribute)
	require.True(t, ok, "path must be StringAttribute")
	assert.NotEmpty(t, pathStr.Validators, "path attribute must have validators (path format parity)")

	// job_authorization_scope must have enum validator.
	jasAttr, ok := attrs["job_authorization_scope"]
	require.True(t, ok, "schema must have job_authorization_scope attribute")
	jasStr, ok := jasAttr.(schema.StringAttribute)
	require.True(t, ok, "job_authorization_scope must be StringAttribute")
	assert.NotEmpty(t, jasStr.Validators, "job_authorization_scope must have enum validators")

	// queue_status must have enum validator.
	qsAttr, ok := attrs["queue_status"]
	require.True(t, ok, "schema must have queue_status attribute")
	qsStr, ok := qsAttr.(schema.StringAttribute)
	require.True(t, ok, "queue_status must be StringAttribute")
	assert.NotEmpty(t, qsStr.Validators, "queue_status must have enum validators")
}

// TestBuildDefinitionFramework_PathValidator verifies path validator logic directly.
func TestBuildDefinitionFramework_PathValidator(t *testing.T) {
	t.Parallel()

	v := bdFwPathValidator{}
	desc := v.Description(context.Background())
	assert.Contains(t, desc, "\\", "Description must mention backslash requirement")
}

// TestBuildDefinitionFramework_StringNotWhitespace verifies not-whitespace validator.
func TestBuildDefinitionFramework_StringNotWhitespace(t *testing.T) {
	t.Parallel()

	v := bdFwStringNotWhitespace{}
	desc := v.Description(context.Background())
	assert.Contains(t, desc, "whitespace")
}

// TestBuildFolderFramework_PathValidator verifies path validator is wired on build_folder schema.
func TestBuildFolderFramework_PathValidator(t *testing.T) {
	t.Parallel()

	r := &BuildFolderResource{}

	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), schemaReq, schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not return error diagnostics: %s", schemaResp.Diagnostics)

	pathAttr, ok := schemaResp.Schema.Attributes["path"]
	require.True(t, ok, "schema must have path attribute")
	pathStr, ok := pathAttr.(schema.StringAttribute)
	require.True(t, ok, "path must be StringAttribute")
	assert.NotEmpty(t, pathStr.Validators, "path attribute must have validators (path format parity for AC3)")
}
