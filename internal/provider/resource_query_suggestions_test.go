package provider

import (
	"fmt"
	"testing"

	suggestions "github.com/algolia/algoliasearch-client-go/v4/algolia/query-suggestions"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
)

func TestAccResourceQuerySuggestions(t *testing.T) {
	indexName := randResourceID(100)
	resourceName := "algolia_query_suggestions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckQuerySuggestionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceQuerySuggestions(indexName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "index_name", "suggestions-"+indexName),
					resource.TestCheckResourceAttr(resourceName, "region", "us"),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.index_name", indexName),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.min_hits", "5"),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.min_letters", "4"),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.facets.#", "0"),
				),
			},
			{
				Config: testAccResourceQuerySuggestionsUpdate(indexName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuerySuggestionsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.min_hits", "10"),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.min_letters", "3"),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.facets.0.attribute", "brand"),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.facets.0.amount", "2"),
					testCheckResourceListAttr(resourceName, "source_indices.0.generate.0", []string{"brand"}),
					testCheckResourceListAttr(resourceName, "source_indices.0.generate.1", []string{"brand", "category"}),
					testCheckResourceListAttr(resourceName, "languages", []string{"en", "ja"}),
					testCheckResourceListAttr(resourceName, "exclude", []string{"test"}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     "us/suggestions-" + indexName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceQuerySuggestions_disappears(t *testing.T) {
	indexName := randResourceID(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckQuerySuggestionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceQuerySuggestions(indexName),
				Check: func(s *terraform.State) error {
					suggestionsClient, err := newTestAPIClient().newSuggestionsClient(suggestions.US)
					if err != nil {
						return err
					}
					_, err = suggestionsClient.DeleteConfig(suggestionsClient.NewApiDeleteConfigRequest("suggestions-" + indexName))
					return err
				},
				// Deleted outside Terraform: refresh must drop it from state and plan a re-create.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQuerySuggestionsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		return nil
	}
}

func testAccResourceQuerySuggestions(indexName string) string {
	return fmt.Sprintf(`
resource "algolia_index" "test" {
  name = "%s"
  deletion_protection = false
}

resource "algolia_query_suggestions" "test" {
  index_name = "suggestions-${algolia_index.test.name}"
  region = "us"

  source_indices {
    index_name = algolia_index.test.name
  }
}
`, indexName)
}

func testAccResourceQuerySuggestionsUpdate(indexName string) string {
	return fmt.Sprintf(`
resource "algolia_index" "test" {
  name = "%s"
  deletion_protection = false

  attributes_config {
    attributes_for_faceting = ["brand", "category"]
  }
}

resource "algolia_query_suggestions" "test" {
  index_name = "suggestions-${algolia_index.test.name}"
  region = "us"

  source_indices {
    index_name  = algolia_index.test.name
    min_hits    = 10
    min_letters = 3
    facets {
      attribute = "brand"
      amount    = 2
    }
    generate = [["brand"], ["brand", "category"]]
  }

  languages = ["en", "ja"]
  exclude   = ["test"]
}
`, indexName)
}

func testAccCheckQuerySuggestionsDestroy(s *terraform.State) error {
	apiClient := newTestAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "algolia_query_suggestions" {
			continue
		}

		suggestionsClient, err := apiClient.newSuggestionsClient(suggestions.US)
		if err != nil {
			return err
		}
		_, err = suggestionsClient.GetConfig(
			suggestionsClient.NewApiGetConfigRequest(rs.Primary.ID),
		)
		if err == nil {
			return fmt.Errorf("query suggestions '%s' still exists", rs.Primary.ID)
		}
		if !algoliautil.IsNotFoundError(err) {
			return err
		}
	}

	return nil
}
