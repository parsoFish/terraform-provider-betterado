package acceptancetests

import "fmt"

// agileSystemProcessTypeId is the well-known type ID for the Agile system process
// in Azure DevOps. It is used as the parent process type when creating custom
// inherited processes in acceptance tests.
//
// Sourced from:
// https://learn.microsoft.com/en-us/rest/api/azure/devops/processes/processes/list?view=azure-devops-rest-7.1&tabs=HTTP#get-the-list-of-processes
//
// This constant is defined here (without a build tag) so that it is available to
// all workitemtrackingprocess acceptance-test files regardless of which build tags
// are active. Previously it was defined only in resource_workitemtrackingprocess_process_test.go
// (build tag: all || resource_workitemtrackingprocess_process), causing build failures when
// tests were compiled without that tag.
const agileSystemProcessTypeId = "adcc42ab-9882-485e-a3ed-7678f01f66bc"

// process returns Terraform HCL that creates a betterado_workitemtrackingprocess_process
// resource. Used as a dependency by other workitemtrackingprocess acceptance tests.
func process(name string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_process" "test" {
  name                   = "%s"
  parent_process_type_id = "%s"
}
`, name, agileSystemProcessTypeId)
}

// disabledProcess is like process but creates the process with is_enabled = false.
// Used by resource_workitemtrackingprocess_process_test.go (build tag: all).
//
//nolint:unused
func disabledProcess(name string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_process" "test" {
  name                   = "%s"
  parent_process_type_id = "%s"
  is_enabled             = false
}
`, name, agileSystemProcessTypeId)
}
