package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccResourceQuerySuggestions(t *testing.T) {
	indexName := randStringStartWithAlpha(100)
	sourceIndexName := randStringStartWithAlpha(100)
	resourceName := fmt.Sprintf("algolia_query_suggestions.%s", indexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceQuerySuggestions(indexName, sourceIndexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "index_name", indexName),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.index_name", sourceIndexName),
					testCheckResourceListAttr(resourceName, "source_indices.0.analytics_tags", []string{}),
					testCheckResourceListAttr(resourceName, "source_indices.0.facets", []string{}),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.min_hits", "5"),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.min_letters", "4"),
					testCheckResourceListAttr(resourceName, "source_indices.0.generate", []string{}),
					testCheckResourceListAttr(resourceName, "source_indices.0.external", []string{}),
					testCheckResourceListAttr(resourceName, "languages", []string{"en"}),
					resource.TestCheckNoResourceAttr(resourceName, "exclude"),
				),
			},
			{
				Config: testAccResourceQuerySuggestionsUpdate(indexName, sourceIndexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "index_name", indexName),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.index_name", sourceIndexName),
					testCheckResourceListAttr(resourceName, "source_indices.0.analytics_tags", []string{}),
					testCheckResourceListAttr(resourceName, "source_indices.0.facets", []string{}),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.min_hits", "10"),
					resource.TestCheckResourceAttr(resourceName, "source_indices.0.min_letters", "3"),
					testCheckResourceListAttr(resourceName, "source_indices.0.generate.0", []string{"brand"}),
					testCheckResourceListAttr(resourceName, "source_indices.0.generate.1", []string{"brand", "category"}),
					testCheckResourceListAttr(resourceName, "source_indices.0.external", []string{}),
					testCheckResourceListAttr(resourceName, "languages", []string{"en", "ja"}),
					resource.TestCheckNoResourceAttr(resourceName, "exclude"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     indexName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceQuerySuggestions(indexName, sourceIndexName string) string {
	return `
resource "algolia_index" "` + indexName + `" {
  name = "` + indexName + `"
}

resource "algolia_index" "` + sourceIndexName + `" {
  name = "` + sourceIndexName + `"
}

resource "algolia_query_suggestions" "` + indexName + `" {
  index_name = algolia_index.` + indexName + `.name

  source_indices {
    index_name = algolia_index.` + sourceIndexName + `.name
  }

  languages = ["en"]
}
`
}

// TODO: add missing attributes
func testAccResourceQuerySuggestionsUpdate(indexName, sourceIndexName string) string {
	return `
resource "algolia_index" "` + sourceIndexName + `" {
  name = "` + sourceIndexName + `"
}

resource "algolia_query_suggestions" "` + indexName + `" {
  index_name = "` + indexName + `"

  source_indices {
    index_name  = algolia_index.` + sourceIndexName + `.name
    min_hits    = 10
    min_letters = 3
    generate    = [["brand"], ["brand", "category"]]
  }

  languages = ["en", "ja"]
}
`
}
