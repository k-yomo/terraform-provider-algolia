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
				),
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
