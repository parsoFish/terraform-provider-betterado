//go:build (all || resource_serviceendpoint_jfrog_artifactory_v2 || resource_serviceendpoint_jfrog_distribution_v2 || resource_serviceendpoint_jfrog_platform_v2 || resource_serviceendpoint_jfrog_xray_v2) && !exclude_serviceendpoints
// +build all resource_serviceendpoint_jfrog_artifactory_v2 resource_serviceendpoint_jfrog_distribution_v2 resource_serviceendpoint_jfrog_platform_v2 resource_serviceendpoint_jfrog_xray_v2
// +build !exclude_serviceendpoints

package serviceendpoint

import "github.com/google/uuid"

// Shared JFrog v2 test variables referenced across all four JFrog v2 test files.
var (
	artifactoryRandomServiceEndpointProjectIDpassword = uuid.New()
	artifactoryRandomServiceEndpointProjectID         = uuid.New()
)
