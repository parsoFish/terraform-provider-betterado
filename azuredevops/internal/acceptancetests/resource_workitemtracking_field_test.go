package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// getFieldDirectClient builds an AggregatedClient directly from AZDO env vars.
// ProtoV6ProviderFactories does not expose Meta(), so CheckDestroy functions must
// build their own client.
func getFieldDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// captureFieldEvidence performs a live API GET of the created field and persists
// the response as forge demo live-evidence. Best-effort: failure never fails the test.
func captureFieldEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		referenceName := res.Primary.ID
		clients, clientErr := getFieldDirectClient()
		if clientErr != nil {
			return nil //nolint:nilerr // best-effort: evidence capture never fails the test
		}
		field, fieldErr := clients.WorkItemTrackingClient.GetWorkItemField(clients.Ctx, workitemtracking.GetWorkItemFieldArgs{
			FieldNameOrRefName: &referenceName,
		})
		if fieldErr != nil || field == nil {
			return nil //nolint:nilerr // best-effort: evidence capture never fails the test
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		if len(orgURL) > 0 && orgURL[len(orgURL)-1] == '/' {
			orgURL = orgURL[:len(orgURL)-1]
		}
		url := fmt.Sprintf("%s/_apis/wit/fields/%s?api-version=7.1", orgURL, referenceName)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemtracking-field", url, field)
		return nil
	}
}

func TestAccWorkItemTrackingField_Basic(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldBasic(fieldName),
				Check: resource.ComposeTestCheckFunc(
					// Computed attributes
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttrSet(tfNode, "supported_operations.#"),
					// Default values
					resource.TestCheckResourceAttr(tfNode, "usage", "workItem"),
					resource.TestCheckResourceAttr(tfNode, "read_only", "false"),
					resource.TestCheckResourceAttr(tfNode, "is_picklist_suggested", "false"),
					resource.TestCheckResourceAttr(tfNode, "is_locked", "false"),
					// Computed attributes that must be populated
					resource.TestCheckResourceAttrSet(tfNode, "can_sort_by"),
					resource.TestCheckResourceAttrSet(tfNode, "is_queryable"),
					resource.TestCheckResourceAttrSet(tfNode, "is_identity"),
					resource.TestCheckResourceAttrSet(tfNode, "is_picklist"),
					// Live evidence capture
					captureFieldEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_Complete(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldComplete(fieldName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_Boolean(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldWithType(fieldName, "boolean"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_Html(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldWithType(fieldName, "html"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_Integer(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldWithType(fieldName, "integer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_DateTime(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldWithType(fieldName, "dateTime"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_PlainText(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldWithType(fieldName, "plainText"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_Double(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldWithType(fieldName, "double"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_Identity(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldWithType(fieldName, "identity"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_TreePath(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldWithType(fieldName, "treePath"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_History(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldWithType(fieldName, "history"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_Guid(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldWithType(fieldName, "guid"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_Lock(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldBasic(fieldName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: lockField(fieldName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_Restore(t *testing.T) {
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldBasic(fieldName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: "# empty config to delete the field",
			},
			{
				Config: restoreField(fieldName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:            tfNode,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"restore"},
			},
		},
	})
}

func TestAccWorkItemTrackingField_Picklist(t *testing.T) {
	fieldName := generateFieldName()
	listName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldAndListDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldPicklist(fieldName, listName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrPair(tfNode, "picklist_id", "betterado_workitemtrackingprocess_list.test", "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_PicklistSuggested(t *testing.T) {
	fieldName := generateFieldName()
	listName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldAndListDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldPicklistSuggested(fieldName, listName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrPair(tfNode, "picklist_id", "betterado_workitemtrackingprocess_list.test", "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkItemTrackingField_PicklistInteger(t *testing.T) {
	fieldName := generateFieldName()
	listName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtracking_field.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFieldAndListDestroyed,
		Steps: []resource.TestStep{
			{
				Config: fieldPicklistInteger(fieldName, listName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrPair(tfNode, "picklist_id", "betterado_workitemtrackingprocess_list.test", "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func fieldBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtracking_field" "test" {
  name           = "%s"
  reference_name = "Custom.%s"
  type           = "string"
}
`, name, name)
}

func fieldWithType(name string, fieldType string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtracking_field" "test" {
  name           = "%s"
  reference_name = "Custom.%s"
  type           = "%s"
}
`, name, name, fieldType)
}

func fieldComplete(name string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtracking_field" "test" {
  name                  = "%s"
  reference_name        = "Custom.%s"
  type                  = "string"
  description           = "Test field description"
  usage                 = "workItem"
  read_only             = false
  is_picklist_suggested = false
  is_locked             = false
}
`, name, name)
}

func lockField(name string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtracking_field" "test" {
  name           = "%s"
  reference_name = "Custom.%s"
  type           = "string"
  is_locked      = true
}
`, name, name)
}

func restoreField(name string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtracking_field" "test" {
  name           = "%s"
  reference_name = "Custom.%s"
  type           = "string"
  restore        = true
}
`, name, name)
}

func fieldPicklist(fieldName, listName string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_list" "test" {
  name  = "%s"
  items = ["Red", "Green", "Blue"]
}

resource "betterado_workitemtracking_field" "test" {
  name           = "%s"
  reference_name = "Custom.%s"
  type           = "string"
  picklist_id    = betterado_workitemtrackingprocess_list.test.id
}
`, listName, fieldName, fieldName)
}

func fieldPicklistSuggested(fieldName, listName string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_list" "test" {
  name         = "%s"
  items        = ["Option1", "Option2", "Option3"]
  is_suggested = true
}

resource "betterado_workitemtracking_field" "test" {
  name           = "%s"
  reference_name = "Custom.%s"
  type           = "string"
  picklist_id    = betterado_workitemtrackingprocess_list.test.id
}
`, listName, fieldName, fieldName)
}

func fieldPicklistInteger(fieldName, listName string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_list" "test" {
  name  = "%s"
  type  = "integer"
  items = ["1", "2", "3", "5", "8"]
}

resource "betterado_workitemtracking_field" "test" {
  name           = "%s"
  reference_name = "Custom.%s"
  type           = "integer"
  picklist_id    = betterado_workitemtrackingprocess_list.test.id
}
`, listName, fieldName, fieldName)
}

// generateFieldName generates a valid field name without hyphens or other invalid characters
func generateFieldName() string {
	return strings.ReplaceAll(testutils.GenerateResourceName(), "-", "")
}

func checkFieldAndListDestroyed(s *terraform.State) error {
	if err := checkFieldDestroyed(s); err != nil {
		return err
	}
	return checkListDestroyed(s)
}

// checkFieldDestroyed verifies that all fields referenced in the state are destroyed. This will be invoked
// *after* terraform destroys the resource but *before* the state is wiped clean.
func checkFieldDestroyed(s *terraform.State) error {
	// ProtoV6ProviderFactories does not expose Meta(), so we build a direct client.
	clients, err := getFieldDirectClient()
	if err != nil {
		return fmt.Errorf("checkFieldDestroyed: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_workitemtracking_field" {
			continue
		}

		referenceName := res.Primary.ID

		_, err := clients.WorkItemTrackingClient.GetWorkItemField(clients.Ctx, workitemtracking.GetWorkItemFieldArgs{
			FieldNameOrRefName: &referenceName,
		})
		if utils.ResponseWasNotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}
