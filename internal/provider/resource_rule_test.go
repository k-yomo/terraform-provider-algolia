package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccResourceRule(t *testing.T) {
	t.Parallel()

	indexName := acctest.RandStringFromCharSet(100, acctest.CharSetAlpha)
	objectID := acctest.RandStringFromCharSet(64, acctest.CharSetAlpha)
	resourceName := fmt.Sprintf("algolia_rule.%s", objectID)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRule(indexName, objectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "index_name", indexName),
					resource.TestCheckResourceAttr(resourceName, "object_id", objectID),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.pattern", "{facet:category}"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.anchoring", "contains"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.automatic_facet_filters.0.facet", "category"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.automatic_facet_filters.0.disjunctive", "true"),
					//testCheckResourceListAttr(resourceName, "consequence.0.promote.0.object_ids", []string{"promote-12345"}),
					//resource.TestCheckResourceAttr(resourceName, "consequence.0.promote.0.position", "0"),
					//testCheckResourceListAttr(resourceName, "consequence.0.hide", []string{"hide-12345"}),
				),
			},
		},
	})
}

func testAccResourceRule(indexName, objectID string) string {
	return `
resource "algolia_index" "` + indexName + `" {
  name = "` + indexName + `"
}

resource "algolia_rule" "` + objectID + `" {
  index_name = algolia_index.` + indexName + `.name
  object_id = "` + objectID + `"

  conditions {
    pattern   = "{facet:category}"
    anchoring = "contains"
  }

  consequence {
    params {
      automatic_facet_filters {
        facet       = "category"
        disjunctive = true
      }
    }
    // specifing id cause 404 error
    //promote {
    //  object_ids  = ["promote-12345"]
    //  position = 0
    //}
    //hide = ["hide-12345"]
  }
}
`
}
