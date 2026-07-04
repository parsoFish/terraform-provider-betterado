//go:build all || resource_test_configuration

// Package testplan — test variable unit tests.
// The gate command is:
//
//	go test -tags all -run TestUnitTestConfiguration ./azuredevops/internal/service/testplan/
//
// Variable-specific tests live in resource_test_configuration_framework_test.go
// and use the TestUnitTestConfiguration_ prefix so the gate picks them up.
// This file is kept as a companion to resource_test_variable_framework.go per
// the files_in_scope declaration.
package testplan
