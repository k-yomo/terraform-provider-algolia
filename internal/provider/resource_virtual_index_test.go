package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceVirtualIndex(t *testing.T) {
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

func testAccResourceVirtualIndex(primaryIndexName string, virtualIndexName string) string {
	return `
resource "algolia_index" "` + primaryIndexName + `" {
  name = "` + primaryIndexName + `"

  ranking_config {
    ranking = ["typo", "geo"]
  }

  languages_config {
	decompounded_attributes {
		language = "de"
		attributes = ["name"]
	}
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

// Regression test for #258: settings such as custom_ranking must be applied when the virtual
// index is created (not only on a later update). Verified against Algolia, not just state.
func TestAccResourceVirtualIndex_settingsOnCreate(t *testing.T) {
	indexName := randResourceID(80)
	virtualIndexName := fmt.Sprintf("%s_price_asc", indexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "algolia_index" "` + indexName + `" {
  name = "` + indexName + `"

  attributes_config {
    searchable_attributes  = ["title", "description"]
    attributes_to_retrieve = ["title", "price"]
  }
  ranking_config {
    ranking = ["typo", "geo", "words", "filters", "proximity", "attribute", "exact", "custom"]
  }

  deletion_protection = false
}

resource "algolia_virtual_index" "` + virtualIndexName + `" {
  name               = "` + virtualIndexName + `"
  primary_index_name = algolia_index.` + indexName + `.name

  attributes_config {
    attributes_to_retrieve = ["title", "price"]
  }
  ranking_config {
    custom_ranking = ["asc(price)"]
  }

  deletion_protection = false
}
`,
				Check: func(s *terraform.State) error {
					apiClient := newTestAPIClient()
					settings, err := apiClient.searchClient.GetSettings(apiClient.searchClient.NewApiGetSettingsRequest(virtualIndexName))
					if err != nil {
						return err
					}
					if got := settings.GetCustomRanking(); len(got) != 1 || got[0] != "asc(price)" {
						return fmt.Errorf("customRanking of %s in Algolia = %v, want [asc(price)]", virtualIndexName, got)
					}
					if got := settings.GetAttributesToRetrieve(); len(got) != 2 {
						return fmt.Errorf("attributesToRetrieve of %s in Algolia = %v, want [title price]", virtualIndexName, got)
					}
					return nil
				},
			},
		},
	})
}
