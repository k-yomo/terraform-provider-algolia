package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
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
					testCheckJSONResourceAttr(resourceName, "consequence.0.params_json", `{"automaticFacetFilters":[{"facet":"category","disjunctive":true,"score":0}]}`),
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
					testCheckJSONResourceAttr(resourceName, "consequence.0.params_json", `{"query":{"edits":[{"type":"remove","delete":"tag"}]},"automaticFacetFilters":[{"facet":"tag","disjunctive":true,"score":0}]}`),
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

		_, err := apiClient.searchClient.GetRule(
			apiClient.searchClient.NewApiGetRuleRequest(rs.Primary.Attributes["index_name"], rs.Primary.ID),
		)
		if err == nil {
			return fmt.Errorf("rule '%s' still exists", rs.Primary.ID)
		}
		if !algoliautil.IsNotFoundError(err) {
			return err
		}
	}

	return nil
}

// TestAccResourceRuleWithDeprecatedParams covers the deprecated `consequence.params` block,
// which must keep working (and round-trip without a diff) until it is removed.
func TestAccResourceRuleWithDeprecatedParams(t *testing.T) {
	indexName := randResourceID(100)
	objectID := randResourceID(64)
	resourceName := fmt.Sprintf("algolia_rule.%s", objectID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRuleWithDeprecatedParams(indexName, objectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleParams(indexName, objectID, `{"query":{"edits":[{"type":"replace","delete":"cheap","insert":"budget"}]},"automaticFacetFilters":[{"facet":"brand","score":2,"disjunctive":true}],"automaticOptionalFacetFilters":[{"facet":"color","score":1,"disjunctive":false}]}`),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.object_query.0.type", "replace"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.object_query.0.delete", "cheap"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.object_query.0.insert", "budget"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.automatic_facet_filters.0.facet", "brand"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.automatic_facet_filters.0.score", "2"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.automatic_facet_filters.0.disjunctive", "true"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.automatic_optional_facet_filters.0.facet", "color"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params_json", ""),
				),
			},
			{
				Config: testAccResourceRuleWithDeprecatedParamsUpdate(indexName, objectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleParams(indexName, objectID, `{"query":"budget shoes"}`),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.query", "budget shoes"),
					resource.TestCheckResourceAttr(resourceName, "consequence.0.params.0.object_query.#", "0"),
				),
			},
		},
		CheckDestroy: testAccCheckRuleDestroy,
	})
}

func TestAccResourceRule_disappears(t *testing.T) {
	indexName := randResourceID(100)
	objectID := randResourceID(64)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRule(indexName, objectID),
				Check: func(s *terraform.State) error {
					apiClient := newTestAPIClient()
					res, err := apiClient.searchClient.DeleteRule(apiClient.searchClient.NewApiDeleteRuleRequest(indexName, objectID))
					if err != nil {
						return err
					}
					_, err = apiClient.searchClient.WaitForTask(indexName, res.TaskID)
					return err
				},
				// The rule was deleted outside Terraform, so the refresh must drop it from state
				// and the plan must propose re-creating it.
				ExpectNonEmptyPlan: true,
			},
		},
		CheckDestroy: testAccCheckRuleDestroy,
	})
}

// testAccCheckRuleParams verifies the params actually stored in Algolia, not just Terraform state.
func testAccCheckRuleParams(indexName, objectID, expectedJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiClient := newTestAPIClient()
		rule, err := apiClient.searchClient.GetRule(apiClient.searchClient.NewApiGetRuleRequest(indexName, objectID))
		if err != nil {
			return err
		}
		if rule.Consequence.Params == nil {
			return fmt.Errorf("rule %s has no consequence params in Algolia", objectID)
		}
		actual, err := json.Marshal(rule.Consequence.Params)
		if err != nil {
			return err
		}
		equal, err := jsonBytesEqual([]byte(expectedJSON), actual)
		if err != nil {
			return err
		}
		if !equal {
			return fmt.Errorf("rule %s params: expected %s, got %s", objectID, expectedJSON, actual)
		}
		return nil
	}
}

func testAccResourceRuleWithDeprecatedParams(indexName, objectID string) string {
	return `
resource "algolia_index" "` + indexName + `" {
  name = "` + indexName + `"
  deletion_protection = false
}

resource "algolia_rule" "` + objectID + `" {
  index_name = algolia_index.` + indexName + `.name
  object_id  = "` + objectID + `"

  conditions {
    pattern   = "{facet:brand} {facet:color} cheap"
    anchoring = "contains"
  }

  consequence {
    params {
      object_query {
        type   = "replace"
        delete = "cheap"
        insert = "budget"
      }
      automatic_facet_filters {
        facet       = "brand"
        score       = 2
        disjunctive = true
      }
      automatic_optional_facet_filters {
        facet = "color"
      }
    }
  }
}
`
}

func testAccResourceRuleWithDeprecatedParamsUpdate(indexName, objectID string) string {
	return `
resource "algolia_index" "` + indexName + `" {
  name = "` + indexName + `"
  deletion_protection = false
}

resource "algolia_rule" "` + objectID + `" {
  index_name = algolia_index.` + indexName + `.name
  object_id  = "` + objectID + `"

  conditions {
    pattern   = "cheap shoes"
    anchoring = "is"
  }

  consequence {
    params {
      query = "budget shoes"
    }
  }
}
`
}
