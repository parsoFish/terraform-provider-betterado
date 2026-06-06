package environmenttemplates

import (
	"github.com/google/uuid"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
)

// ReleaseDefinitionEnvironmentTemplate is re-exported from the vendored release SDK.
// The struct is present in the vendor but the release.Client has no methods for it.
type ReleaseDefinitionEnvironmentTemplate = releaseapi.ReleaseDefinitionEnvironmentTemplate

// GetEnvironmentTemplatesArgs holds the arguments for GetEnvironmentTemplates.
type GetEnvironmentTemplatesArgs struct {
	// Project name or ID (required).
	Project *string
}

// GetEnvironmentTemplateArgs holds the arguments for GetEnvironmentTemplate.
type GetEnvironmentTemplateArgs struct {
	// Project name or ID (required).
	Project *string
	// ID of the environment template (required).
	TemplateId *uuid.UUID
}

// SaveEnvironmentTemplateArgs holds the arguments for SaveEnvironmentTemplate.
type SaveEnvironmentTemplateArgs struct {
	// Project name or ID (required).
	Project *string
	// Template to create or update (required).
	Template *ReleaseDefinitionEnvironmentTemplate
}

// DeleteEnvironmentTemplateArgs holds the arguments for DeleteEnvironmentTemplate.
type DeleteEnvironmentTemplateArgs struct {
	// Project name or ID (required).
	Project *string
	// ID of the environment template to delete (required).
	TemplateId *uuid.UUID
}
