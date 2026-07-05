package testutils

import (
	"context"
	"os"
	"strings"
	"testing"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/extensionmanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// EnsureExtensionInstalled installs the given extension in the Azure DevOps organisation
// if it is not already installed. It is idempotent — calling it when the extension is
// already present is a no-op. Intended for use in acceptance-test PreCheck functions so
// that tests that need a specific extension pre-installed do not need to manage the
// extension as a Terraform resource (which causes flakiness when a previous run left the
// extension installed and the Terraform resource create fails with TF1590010).
func EnsureExtensionInstalled(t *testing.T, publisherId, extensionId string) {
	t.Helper()
	loadSecretsEnv()

	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
	if err != nil {
		t.Fatalf("EnsureExtensionInstalled: build client: %v", err)
	}

	ctx := context.Background()
	_, err = clients.ExtensionManagementClient.GetInstalledExtensionByName(ctx, extensionmanagement.GetInstalledExtensionByNameArgs{
		PublisherName: &publisherId,
		ExtensionName: &extensionId,
	})
	if err == nil {
		// Already installed — idempotent success.
		return
	}
	if !strings.Contains(err.Error(), "ExtensionDoesNotExist") && !strings.Contains(err.Error(), "404") {
		// Unexpected error checking installation status.
		t.Logf("EnsureExtensionInstalled: unexpected error checking extension %s.%s: %v (will attempt install)", publisherId, extensionId, err)
	}

	// Not installed — install it now.
	_, installErr := clients.ExtensionManagementClient.InstallExtensionByName(ctx, extensionmanagement.InstallExtensionByNameArgs{
		PublisherName: &publisherId,
		ExtensionName: &extensionId,
	})
	if installErr != nil {
		// If it's "already installed" someone beat us to it — that's fine.
		if strings.Contains(installErr.Error(), "TF1590010") || strings.Contains(installErr.Error(), "already installed") {
			return
		}
		t.Fatalf("EnsureExtensionInstalled: install extension %s.%s: %v", publisherId, extensionId, installErr)
	}
}

// EnsureExtensionUninstalled uninstalls the given extension from the Azure DevOps
// organisation if it is currently installed. It is idempotent — calling it when the
// extension is already absent is a no-op. Intended for use in acceptance-test
// CheckDestroy or cleanup functions that follow tests using EnsureExtensionInstalled.
func EnsureExtensionUninstalled(t *testing.T, publisherId, extensionId string) {
	t.Helper()
	loadSecretsEnv()

	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
	if err != nil {
		t.Logf("EnsureExtensionUninstalled: build client: %v (skipping uninstall)", err)
		return
	}

	ctx := context.Background()
	uninstallErr := clients.ExtensionManagementClient.UninstallExtensionByName(ctx, extensionmanagement.UninstallExtensionByNameArgs{
		PublisherName: &publisherId,
		ExtensionName: &extensionId,
	})
	if uninstallErr != nil {
		// If it's "not found" or "does not exist" — already gone, idempotent success.
		if strings.Contains(uninstallErr.Error(), "ExtensionDoesNotExist") ||
			strings.Contains(uninstallErr.Error(), "404") ||
			strings.Contains(uninstallErr.Error(), "not found") {
			return
		}
		t.Logf("EnsureExtensionUninstalled: uninstall extension %s.%s: %v (non-fatal)", publisherId, extensionId, uninstallErr)
	}
}
