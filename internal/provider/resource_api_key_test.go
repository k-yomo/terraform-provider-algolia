package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
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
					resource.TestCheckResourceAttr(resourceName, "query_parameters", ""),
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
					resource.TestCheckResourceAttr(resourceName, "query_parameters", "typoTolerance=strict&ignorePlurals=false"),
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
			{
				// Removing query_parameters from the config must clear it on the key.
				Config: testAccResourceAPIKey(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "query_parameters", ""),
				),
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
  query_parameters            = "typoTolerance=strict&ignorePlurals=false"
}`, name)
}

func TestQueryParametersEqual(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"", "", true},
		{"typoTolerance=strict&ignorePlurals=false", "typoTolerance=strict&ignorePlurals=false", true},
		{"typoTolerance=strict&ignorePlurals=false", "ignorePlurals=false&typoTolerance=strict", true},
		{"filters=brand%3AApple", "filters=brand:Apple", true},
		{"typoTolerance=strict", "typoTolerance=min", false},
		{"typoTolerance=strict", "", false},
		{"a=1&a=2", "a=2&a=1", false},
	}
	for _, tt := range tests {
		if got := queryParametersEqual(tt.a, tt.b); got != tt.want {
			t.Errorf("queryParametersEqual(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func testAccCheckApiKeyDestroy(s *terraform.State) error {
	apiClient := newTestAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "algolia_api_key" {
			continue
		}

		_, err := apiClient.searchClient.GetApiKey(
			apiClient.searchClient.NewApiGetApiKeyRequest(rs.Primary.ID),
		)
		if err == nil {
			return fmt.Errorf("api key '%s' still exists", rs.Primary.ID)
		}
		if !algoliautil.IsNotFoundError(err) {
			return err
		}
	}

	return nil
}

func TestAccResourceAPIKey_disappears(t *testing.T) {
	name := randResourceID(100)
	resourceName := fmt.Sprintf("algolia_api_key.%s", name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAPIKey(name),
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return fmt.Errorf("not found: %s", resourceName)
					}
					key := rs.Primary.Attributes["key"]
					apiClient := newTestAPIClient()
					if _, err := apiClient.searchClient.DeleteApiKey(apiClient.searchClient.NewApiDeleteApiKeyRequest(key)); err != nil {
						return err
					}
					_, err := apiClient.searchClient.WaitForApiKey(key, search.API_KEY_OPERATION_DELETE)
					return err
				},
				// Deleted outside Terraform: refresh must drop it from state and plan a re-create.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Regression test for #278 (adding wildcard index patterns on update hung forever) and
// #158 (ACL-only update hung forever).
func TestAccResourceAPIKey_updateIndexesAndACL(t *testing.T) {
	indexName := randResourceID(60)
	name := randResourceID(100)
	resourceName := fmt.Sprintf("algolia_api_key.%s", name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAPIKeyWithIndexes(name, indexName, `["addObject", "deleteIndex"]`, false),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceListAttr(resourceName, "indexes", []string{indexName + "_events", indexName + "_feeds"}),
				),
			},
			{
				// #278: add wildcard patterns next to the existing index names.
				Config: testAccResourceAPIKeyWithIndexes(name, indexName, `["addObject", "deleteIndex"]`, true),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceListAttr(resourceName, "indexes", []string{
						indexName + "_events", indexName + "_events_tmp_*",
						indexName + "_feeds", indexName + "_feeds_tmp_*",
					}),
				),
			},
			{
				// #158: change only the ACL.
				Config: testAccResourceAPIKeyWithIndexes(name, indexName, `["addObject", "browse", "deleteIndex"]`, true),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceListAttr(resourceName, "acl", []string{"addObject", "browse", "deleteIndex"}),
				),
			},
		},
	})
}

func testAccResourceAPIKeyWithIndexes(name, indexName, acl string, withWildcards bool) string {
	indexes := fmt.Sprintf(`["%[1]s_events", "%[1]s_feeds"]`, indexName)
	if withWildcards {
		indexes = fmt.Sprintf(`["%[1]s_events", "%[1]s_feeds", "%[1]s_events_tmp_*", "%[1]s_feeds_tmp_*"]`, indexName)
	}
	return fmt.Sprintf(`
resource "algolia_api_key" "%s" {
  acl         = %s
  indexes     = %s
  description = "test key"
}`, name, acl, indexes)
}
