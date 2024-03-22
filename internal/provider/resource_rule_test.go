package provider

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/errs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceRule(t *testing.T) {
	indexName := randResourceID(100)
	objectID := randResourceID(64)
	resourceName := fmt.Sprintf("algolia_rule.%s", objectID)

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params_json", `{"automaticFacetFilters":[{"facet":"category","disjunctive":true,"score":0}]}`),
					// testCheckResourceListAttr(resourceName, "consequence.0.promote.0.object_ids", []string{"promote-12345"}),
					// resource.TestCheckResourceAttr(resourceName, "consequence.0.promote.0.position", "0"),
					// testCheckResourceListAttr(resourceName, "consequence.0.hide", []string{"hide-12345"}),
				),
			},
			{
				Config: testAccResourceRuleUpdate(indexName, objectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "index_name", indexName),
					resource.TestCheckResourceAttr(resourceName, "object_id", objectID),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.pattern", "{facet:tag}"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.anchoring", "is"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params_json", `{"query":{"edits":[{"type":"remove","delete":"tag"}]},"automaticFacetFilters":[{"facet":"tag","disjunctive":true,"score":0}]}`),
					resource.TestCheckResourceAttr(resourceName, "validity.0.from", "2030-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr(resourceName, "validity.0.until", "2030-03-31T23:59:59Z"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", indexName, objectID),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testAccCheckRuleDestroy,
	})
}

func testAccResourceRule(indexName, objectID string) string {
	return `
resource "algolia_index" "` + indexName + `" {
  name = "` + indexName + `"
  deletion_protection = false
}

resource "algolia_rule" "` + objectID + `" {
  index_name = algolia_index.` + indexName + `.name
  object_id = "` + objectID + `"

  conditions {
    pattern   = "{facet:category}"
    anchoring = "contains"
    alternatives = true
  }

  consequence {
    params_json = jsonencode({
      automaticFacetFilters = [{
        facet       = "category"
        disjunctive = true
        score 	    = 0
      }]
    })
    // specifying id cause 404 error
    //promote {
    //  object_ids  = ["promote-12345"]
    //  position = 0
    //}
    //hide = ["hide-12345"]
  }
}
`
}

func testAccResourceRuleUpdate(indexName, objectID string) string {
	return `
resource "algolia_rule" "` + objectID + `" {
  index_name = "` + indexName + `"
  object_id = "` + objectID + `"
  description = "This is a test rule"

  conditions {
    pattern   = "{facet:tag}"
    anchoring = "is"
  }

  consequence {
    params_json = jsonencode({
      query = {
		edits = [{
          type = "remove"
          delete = "tag"
		}]
      }
      automaticFacetFilters = [{
        facet       = "tag"
        disjunctive = true
        score 	    = 0
      }]
    })
  }

  validity {
    from = "2030-01-01T00:00:00Z"
    until = "2030-03-31T23:59:59Z"
  }
}
`
}

func testAccCheckRuleDestroy(s *terraform.State) error {
	apiClient := newTestAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "algolia_rule" {
			continue
		}

		_, err := apiClient.searchClient.InitIndex(rs.Primary.Attributes["index_name"]).GetRule(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("rule '%s' still exists", rs.Primary.ID)
		}
		if _, ok := errs.IsAlgoliaErrWithCode(err, http.StatusNotFound); !ok {
			return err
		}
	}

	return nil
}
