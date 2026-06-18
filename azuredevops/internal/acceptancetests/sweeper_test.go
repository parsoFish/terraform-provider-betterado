package acceptancetests

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// TestMain wires the Terraform SDK sweeper framework into this package. With the
// `-sweep` flag, `go test` runs the registered sweepers (reaping leaked fixtures)
// instead of the test functions; without it, tests run exactly as before.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// fixtureProjectNameRE matches the names testutils.GenerateResourceName produces
// ("test-acc-" + 10 lowercase-alphanum). Acceptance fixtures also surface under an
// "AccTest" prefix, so the sweeper reaps both. keepProjects is a belt-and-suspenders
// allowlist of real projects that must never be touched even if a name pattern ever
// widens to overlap one.
var fixtureProjectNameRE = regexp.MustCompile(`^test-acc-[a-z0-9]{10}$`)

var keepProjects = map[string]bool{
	"DPLife":                  true,
	"Ohana":                   true,
	"PublicProjects":          true,
	"betterado-standing-demo": true,
}

// isFixtureProject reports whether a project name is an acceptance-test leftover
// safe to delete: a fixture-pattern or AccTest-prefixed name that is not a known
// real project.
func isFixtureProject(name string) bool {
	if keepProjects[name] {
		return false
	}
	return fixtureProjectNameRE.MatchString(name) || strings.HasPrefix(name, "AccTest")
}

func init() {
	resource.AddTestSweepers("betterado_project", &resource.Sweeper{
		Name: "betterado_project",
		F:    sweepFixtureProjects,
	})
}

// sweepFixtureProjects reaps ADO projects left behind by acceptance runs that were
// killed, timed out, or panicked before their per-test t.Cleanup / CheckDestroy
// could fire. The normal path is still handled per-test; this is the backstop for
// the abnormal one (and the reason orphans accumulated during the capstone).
//
// Run:
//
//	TF_ACC=1 go test -tags all -timeout 30m -sweep=ado \
//	  -sweep-run=betterado_project ./azuredevops/internal/acceptancetests/...
func sweepFixtureProjects(string) error {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	if orgURL == "" || pat == "" {
		return fmt.Errorf("sweep requires AZDO_ORG_SERVICE_URL and AZDO_PERSONAL_ACCESS_TOKEN")
	}

	clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
	if err != nil {
		return fmt.Errorf("sweep: build client: %w", err)
	}

	leaked, err := listFixtureProjects(clients)
	if err != nil {
		return err
	}
	log.Printf("[INFO] betterado_project sweep: %d leaked fixture project(s) to delete", len(leaked))

	var swept, failed int
	for _, p := range leaked {
		id := p.Id.String()
		if err := deleteFixtureProject(clients, id); err != nil {
			log.Printf("[ERROR] sweep: delete %s (%s): %v", *p.Name, id, err)
			failed++
			continue
		}
		log.Printf("[INFO] sweep: deleted leaked project %s (%s)", *p.Name, id)
		swept++
	}
	log.Printf("[INFO] betterado_project sweep complete: %d deleted, %d failed", swept, failed)
	if failed > 0 {
		return fmt.Errorf("sweep: %d project(s) failed to delete", failed)
	}
	return nil
}

// listFixtureProjects returns every WellFormed project whose name matches the
// acceptance-fixture pattern, paging through all results.
func listFixtureProjects(clients *client.AggregatedClient) ([]core.TeamProjectReference, error) {
	var out []core.TeamProjectReference
	token := ""
	for {
		args := core.GetProjectsArgs{StateFilter: converter.ToPtr(core.ProjectStateValues.WellFormed)}
		if token != "" {
			n, err := strconv.Atoi(token)
			if err != nil {
				return nil, fmt.Errorf("sweep: bad continuation token %q: %w", token, err)
			}
			args.ContinuationToken = &n
		}
		resp, err := clients.CoreClient.GetProjects(clients.Ctx, args)
		if err != nil {
			return nil, fmt.Errorf("sweep: list projects: %w", err)
		}
		for i := range resp.Value {
			if p := resp.Value[i]; p.Name != nil && p.Id != nil && isFixtureProject(*p.Name) {
				out = append(out, p)
			}
		}
		if resp.ContinuationToken == "" {
			return out, nil
		}
		token = resp.ContinuationToken
	}
}
