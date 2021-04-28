package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccResourceSynonyms(t *testing.T) {
	t.Parallel()

	indexName := acctest.RandStringFromCharSet(100, acctest.CharSetAlpha)
	resourceName := fmt.Sprintf("algolia_synonyms.%s", indexName)

	resource.UnitTest(t, resource.TestCase{
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
				Config:            testAccResourceSynonyms(indexName),
				ResourceName:      resourceName,
				ImportStateId:     indexName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceSynonyms(indexName string) string {
	return `
resource "algolia_index" "` + indexName + `" {
  name = "` + indexName + `"
}

resource "algolia_synonyms" "` + indexName + `" {
  index_name = "` + indexName + `"

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
