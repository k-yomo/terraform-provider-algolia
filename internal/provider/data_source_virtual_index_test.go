package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVirtualIndex(t *testing.T) {
	indexName := randResourceID(80)
	virtualIndexName := fmt.Sprintf("%s_virtual", indexName)
	dataSourceName := fmt.Sprintf("data.algolia_virtual_index.%s", virtualIndexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVirtualIndex(indexName, virtualIndexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", virtualIndexName),
					resource.TestCheckResourceAttr(dataSourceName, "primary_index_name", indexName),
					testCheckResourceListAttr(dataSourceName, "ranking_config.0.custom_ranking", []string{"desc(likes)"}),
					resource.TestCheckResourceAttr(dataSourceName, "faceting_config.0.max_values_per_facet", "50"),
					resource.TestCheckResourceAttr(dataSourceName, "pagination_config.0.hits_per_page", "30"),
					resource.TestCheckResourceAttr(dataSourceName, "advanced_config.0.distinct", "1"),
				),
			},
		},
	})
}

func testAccDataSourceVirtualIndex(primaryIndexName, virtualIndexName string) string {
	return `
resource "algolia_index" "` + primaryIndexName + `" {
  name = "` + primaryIndexName + `"

  attributes_config {
    attributes_for_faceting = ["category"]
  }

  advanced_config {
    attribute_for_distinct = "url"
    distinct               = 1
  }

  deletion_protection = false
}

resource "algolia_virtual_index" "` + virtualIndexName + `" {
  name               = "` + virtualIndexName + `"
  primary_index_name = algolia_index.` + primaryIndexName + `.name

  ranking_config {
    custom_ranking = ["desc(likes)"]
  }
  faceting_config {
    max_values_per_facet = 50
  }
  pagination_config {
    hits_per_page = 30
  }
  advanced_config {
    distinct = 1
  }

  deletion_protection = false
}

data "algolia_virtual_index" "` + virtualIndexName + `" {
  name = algolia_virtual_index.` + virtualIndexName + `.name
}
`
}
