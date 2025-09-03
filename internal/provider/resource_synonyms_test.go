package provider

import (
	"fmt"
	"io"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSynonyms(t *testing.T) {
	indexName := randResourceID(100)
	resourceName := fmt.Sprintf("algolia_synonyms.%s", indexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSynonyms(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "index_name", indexName),
					resource.TestCheckResourceAttr(resourceName, "synonyms.0.object_id", "test_1"),
					resource.TestCheckResourceAttr(resourceName, "synonyms.0.type", "synonym"),
					testCheckResourceListAttr(resourceName, "synonyms.0.synonyms", []string{"cell phone", "mobile phone", "smartphone"}),
				),
			},
			{
				Config: testAccResourceSynonymsUpdate(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "index_name", indexName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     indexName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testAccCheckSynonymsDestroy,
	})
}

func testAccResourceSynonyms(indexName string) string {
	return `
resource "algolia_index" "` + indexName + `" {
  name = "` + indexName + `"
  deletion_protection = false
}

resource "algolia_synonyms" "` + indexName + `" {
  index_name = algolia_index.` + indexName + `.name

  synonyms {
    object_id = "test_1"
    type      = "synonym"
    synonyms  = ["smartphone", "mobile phone", "cell phone"]
  }
}
`
}

func testAccResourceSynonymsUpdate(indexName string) string {
	return `
resource "algolia_synonyms" "` + indexName + `" {
  index_name = "` + indexName + `"

  synonyms {
    object_id = "test_1"
    type      = "synonym"
    synonyms  = ["smartphone", "mobile phone", "cell phone"]
  }
  synonyms {
    object_id = "test_2"
    type      = "oneWaySynonym"
    input     = "smartphone"
    synonyms  = ["iPhone", "Pixel"]
  }
  synonyms {
    object_id   = "test_3"
    type        = "altCorrection1"
    word        = "tablet"
    corrections = ["ipad"]
  }
  synonyms {
    object_id   = "test_4"
    type        = "altCorrection2"
    word        = "tablet"
    corrections = ["iphone"]
  }
  synonyms {
    object_id    = "test_5"
    type         = "placeholder"
    placeholder  = "<model>"
    replacements = ["6", "7", "8"]
  }
}
`
}

func testAccCheckSynonymsDestroy(s *terraform.State) error {
	apiClient := newTestAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "algolia_synonyms" {
			continue
		}

		synonymsIter, err := apiClient.searchClient.InitIndex(rs.Primary.ID).BrowseSynonyms()
		if err != nil {
			return err
		}
		if _, err := synonymsIter.Next(); err != io.EOF {
			return fmt.Errorf("synonyms for index '%s' still exists", rs.Primary.ID)
		}
	}

	return nil
}
