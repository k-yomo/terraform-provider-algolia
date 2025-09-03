package provider

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/errs"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceAPIKey(t *testing.T) {
	name := randResourceID(100)
	resourceName := fmt.Sprintf("algolia_api_key.%s", name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAPIKey(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "key", regexp.MustCompile("^.{1,}$")),
					testCheckResourceListAttr(resourceName, "acl", []string{"search"}),
					resource.TestCheckNoResourceAttr(resourceName, "expires_at"),
					resource.TestCheckResourceAttr(resourceName, "max_hits_per_query", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_queries_per_ip_per_hour", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "indexes.0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccResourceAPIKeyUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "key", regexp.MustCompile("^.{1,}$")),
					testCheckResourceListAttr(resourceName, "acl", []string{"browse", "search"}),
					resource.TestCheckResourceAttr(resourceName, "expires_at", "2030-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr(resourceName, "max_hits_per_query", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_queries_per_ip_per_hour", "10000"),
					testCheckResourceListAttr(resourceName, "indexes", []string{"dev_*"}),
					testCheckResourceListAttr(resourceName, "referers", []string{"https://algolia.com/\\*"}),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a test api key"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.Modules[0].Resources[resourceName].Primary.Attributes["key"], nil
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"expires_at"},
			},
		},
		CheckDestroy: testAccCheckApiKeyDestroy,
	})
}

func testAccResourceAPIKey(name string) string {
	return fmt.Sprintf(`
resource "algolia_api_key" "%s" {
  acl = ["search"]
}`, name)
}

func testAccResourceAPIKeyUpdate(name string) string {
	return fmt.Sprintf(`
resource "algolia_api_key" "%s" {
  acl                         = ["browse", "search"]
  expires_at                  = "2030-01-01T00:00:00Z"
  max_hits_per_query          = 100
  max_queries_per_ip_per_hour = 10000
  indexes                     = ["dev_*"]
  referers                    = ["https://algolia.com/\\*"]
  description                 = "This is a test api key"
}`, name)
}

func testAccCheckApiKeyDestroy(s *terraform.State) error {
	apiClient := newTestAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "algolia_api_key" {
			continue
		}

		_, err := apiClient.searchClient.GetAPIKey(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("api key '%s' still exists", rs.Primary.ID)
		}
		if _, ok := errs.IsAlgoliaErrWithCode(err, http.StatusNotFound); !ok {
			return err
		}
	}

	return nil
}
