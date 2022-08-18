package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceVirtualIndex(t *testing.T) {
	// NOTE: Deleting virtual replica fails due to the same reason as the below issue.
	// https://github.com/algolia/algoliasearch-client-javascript/issues/1377
	// TODO: Remove t.Skip() once the issue is resolved.
	t.Skip()

	indexName := randResourceID(80)
	virtualIndexName := fmt.Sprintf("%s_virtual", indexName)
	indexResourceName := fmt.Sprintf("algolia_index.%s", indexName)
	virtualIndexResourceName := fmt.Sprintf("algolia_virtual_index.%s", virtualIndexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVirtualIndex(indexName, virtualIndexName),
				Check: resource.ComposeTestCheckFunc(
					// index
					resource.TestCheckResourceAttr(indexResourceName, "name", indexName),
					resource.TestCheckResourceAttr(indexResourceName, "virtual", "false"),
					// virtual index
					resource.TestCheckResourceAttr(virtualIndexResourceName, "name", virtualIndexName),
					resource.TestCheckResourceAttr(virtualIndexResourceName, "deletion_protection", "false"),
				),
			},
			{
				Config: testAccResourceVirtualIndexUpdate(indexName, virtualIndexName),
				Check: resource.ComposeTestCheckFunc(
					// index
					resource.TestCheckResourceAttr(indexResourceName, "name", indexName),
					resource.TestCheckResourceAttr(indexResourceName, "virtual", "false"),
					resource.TestCheckResourceAttr(indexResourceName, "attributes_config.0.attributes_to_retrieve.0", "*"),
					testCheckResourceListAttr(indexResourceName, "attributes_config.0.searchable_attributes", []string{"name", "description", "category_name"}),
					testCheckResourceListAttr(indexResourceName, "attributes_config.0.attributes_for_faceting", []string{"category_id"}),
					testCheckResourceListAttr(indexResourceName, "ranking_config.0.ranking", []string{"typo", "geo"}),
					resource.TestCheckResourceAttr(indexResourceName, "advanced_config.0.distinct", "2"),
					resource.TestCheckResourceAttr(indexResourceName, "advanced_config.0.attribute_for_distinct", "url"),
					// virtual index
					resource.TestCheckResourceAttr(virtualIndexResourceName, "name", virtualIndexName),
					testCheckResourceListAttr(virtualIndexResourceName, "ranking_config.0.custom_ranking", []string{"desc(likes)"}),
					testCheckResourceListAttr(virtualIndexResourceName, "advanced_config.0.response_fields", []string{"*"}),
					resource.TestCheckResourceAttr(virtualIndexResourceName, "advanced_config.0.distinct", "1"),
					resource.TestCheckResourceAttr(virtualIndexResourceName, "deletion_protection", "false"),
				),
			},
			{
				ResourceName:      virtualIndexResourceName,
				ImportStateId:     virtualIndexName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deletion_protection",
				},
			},
		},
		CheckDestroy: testAccCheckIndexDestroy,
	})
}

// nolint:unused
func testAccResourceVirtualIndex(primaryIndexName string, virtualIndexName string) string {
	return `
resource "algolia_index" "` + primaryIndexName + `" {
  name = "` + primaryIndexName + `"

  ranking_config {
    ranking = ["typo", "geo"]
  }

  deletion_protection = false
}

resource "algolia_virtual_index" "` + virtualIndexName + `" {
  name               = "` + virtualIndexName + `"
  primary_index_name = algolia_index.` + primaryIndexName + `.name

  deletion_protection = false
}
`
}

// nolint:unused
func testAccResourceVirtualIndexUpdate(primaryIndexName string, virtualIndexName string) string {
	return `
resource "algolia_index" "` + primaryIndexName + `" {
  name = "` + primaryIndexName + `"

  attributes_config {
    attributes_to_retrieve = ["*"]
    searchable_attributes = ["name", "description", "category_name"]
    attributes_for_faceting = ["category_id"]
  }

  ranking_config {
    ranking = ["typo", "geo"]
  }

  advanced_config {
    response_fields = ["*"]
    distinct = 2
    attribute_for_distinct = "url"
  }

  deletion_protection = false
}

resource "algolia_virtual_index" "` + virtualIndexName + `" {
  name               = "` + virtualIndexName + `"
  primary_index_name = algolia_index.` + primaryIndexName + `.name

  ranking_config {
    custom_ranking = ["desc(likes)"]
  }

  advanced_config {
    response_fields = ["*"]
    distinct = 1
  }

  deletion_protection = false
}
`
}
