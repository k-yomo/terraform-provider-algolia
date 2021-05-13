package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceIndex(t *testing.T) {
	indexName := randStringStartWithAlpha(100)
	dataSourceName := fmt.Sprintf("data.algolia_index.%s", indexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIndex(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", indexName),
					testCheckResourceListAttr(dataSourceName, "attributes_config.0.searchable_attributes", []string{"title", "category,tag", "unordered(description)"}),
					testCheckResourceListAttr(dataSourceName, "attributes_config.0.attributes_for_faceting", []string{"category"}),
					testCheckResourceListAttr(dataSourceName, "attributes_config.0.unretrievable_attributes", []string{"author_email"}),
					testCheckResourceListAttr(dataSourceName, "attributes_config.0.attributes_to_retrieve", []string{"body", "category", "description", "tag", "title"}),
					testCheckResourceListAttr(dataSourceName, "ranking_config.0.ranking", []string{"words", "proximity"}),
					resource.TestCheckNoResourceAttr(dataSourceName, "ranking_config.0.replicas.0"),
					resource.TestCheckResourceAttr(dataSourceName, "faceting_config.0.max_values_per_facet", "50"),
					resource.TestCheckResourceAttr(dataSourceName, "faceting_config.0.sort_facet_values_by", "alpha"),
				),
			},
		},
	})
}

func testAccDatasourceIndex(name string) string {
	return `
resource "algolia_index" "` + name + `" {
  name = "` + name + `"
  attributes_config {
    searchable_attributes = [
      "title",
      "category,tag",
      "unordered(description)",
    ]
    attributes_for_faceting = [
      "category"
    ]
    unretrievable_attributes = [
      "author_email"
    ]
    attributes_to_retrieve = [
      "title",
      "category",
      "tag",
      "description",
      "body"
    ]
  }
  ranking_config {
    ranking = [
      "words",
      "proximity"
    ]
  }
  faceting_config {
    max_values_per_facet = 50
    sort_facet_values_by = "alpha"
  }

  languages_config {
    remove_stop_words_for = ["en"]
  }
}

data "algolia_index" "` + name + `" {
  name = algolia_index.` + name + `.name
}
`
}
