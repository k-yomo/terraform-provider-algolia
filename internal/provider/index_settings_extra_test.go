package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testRenderingContentJSON = `{"facetOrdering":{"facets":{"order":["brand","color"]},"values":{"brand":{"order":["Apple","Samsung"],"sortRemainingBy":"alpha"}}}}`

func TestAccResourceIndex_extraSettings(t *testing.T) {
	indexName := randResourceID(80)
	resourceName := fmt.Sprintf("algolia_index.%s", indexName)
	dataSourceName := fmt.Sprintf("data.algolia_index.%s", indexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexExtraSettings(indexName, `
  mode                    = "keywordSearch"
  enable_re_ranking       = true
  re_ranking_apply_filter = "brand:Apple"
  rendering_content_json  = jsonencode(`+testRenderingContentHCL+`)
  user_data_json          = jsonencode({ team = "search", priority = 1 })
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mode", "keywordSearch"),
					resource.TestCheckResourceAttr(resourceName, "enable_re_ranking", "true"),
					resource.TestCheckResourceAttr(resourceName, "re_ranking_apply_filter", "brand:Apple"),
					testCheckJSONResourceAttr(resourceName, "rendering_content_json", testRenderingContentJSON),
					testCheckJSONResourceAttr(resourceName, "user_data_json", `{"team":"search","priority":1}`),
					testAccCheckIndexSettings(indexName, func(s *search.SettingsResponse) error {
						if s.GetMode() != search.MODE_KEYWORD_SEARCH {
							return fmt.Errorf("mode: expected keywordSearch, got %q", s.GetMode())
						}
						if !s.GetEnableReRanking() {
							return fmt.Errorf("enableReRanking: expected true")
						}
						return testAccCheckJSONEqual("renderingContent", s.RenderingContent, testRenderingContentJSON)
					}),
					// The data source must expose the same values.
					resource.TestCheckResourceAttr(dataSourceName, "mode", "keywordSearch"),
					resource.TestCheckResourceAttr(dataSourceName, "enable_re_ranking", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "re_ranking_apply_filter", "brand:Apple"),
					testCheckJSONResourceAttr(dataSourceName, "rendering_content_json", testRenderingContentJSON),
					testCheckJSONResourceAttr(dataSourceName, "user_data_json", `{"team":"search","priority":1}`),
				),
			},
			{
				// Clearing only the filter must still be planned and applied.
				Config: testAccResourceIndexExtraSettings(indexName, `
  mode                    = "keywordSearch"
  enable_re_ranking       = true
  re_ranking_apply_filter = ""
  rendering_content_json  = jsonencode(`+testRenderingContentHCL+`)
  user_data_json          = jsonencode({ team = "search", priority = 1 })
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "re_ranking_apply_filter", ""),
					testAccCheckIndexSettings(indexName, func(s *search.SettingsResponse) error {
						if f := s.ReRankingApplyFilter.Get(); f != nil {
							if got, _ := reRankingApplyFilterToString(f); got != "" {
								return fmt.Errorf("reRankingApplyFilter: expected empty, got %q", got)
							}
						}
						return nil
					}),
				),
			},
			{
				// Update every value, and clear the JSON settings.
				Config: testAccResourceIndexExtraSettings(indexName, `
  enable_re_ranking       = false
  re_ranking_apply_filter = "brand:Samsung"
  rendering_content_json  = jsonencode({})
  user_data_json          = jsonencode({})
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_re_ranking", "false"),
					resource.TestCheckResourceAttr(resourceName, "re_ranking_apply_filter", "brand:Samsung"),
					// Algolia omits an empty renderingContent but returns an empty userData as `{}`.
					// Both are equivalent to the configured `{}`, so no diff is planned.
					resource.TestCheckResourceAttr(resourceName, "rendering_content_json", ""),
					resource.TestCheckResourceAttr(resourceName, "user_data_json", "{}"),
					testAccCheckIndexSettings(indexName, func(s *search.SettingsResponse) error {
						if s.GetEnableReRanking() {
							return fmt.Errorf("enableReRanking: expected false")
						}
						return nil
					}),
				),
			},
			{
				// Set values again, then remove them from the configuration in the next step.
				Config: testAccResourceIndexExtraSettings(indexName, `
  enable_re_ranking      = true
  rendering_content_json = jsonencode(`+testRenderingContentHCL+`)
`),
				Check: resource.TestCheckResourceAttr(resourceName, "enable_re_ranking", "true"),
			},
			{
				// Removing the attributes must neither plan a change nor reset the values in Algolia.
				Config: testAccResourceIndexExtraSettings(indexName, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_re_ranking", "true"),
					testAccCheckIndexSettings(indexName, func(s *search.SettingsResponse) error {
						if !s.GetEnableReRanking() {
							return fmt.Errorf("enableReRanking: expected the value set outside the configuration to be kept")
						}
						return testAccCheckJSONEqual("renderingContent", s.RenderingContent, testRenderingContentJSON)
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateId:           indexName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_protection"},
			},
		},
	})
}

func TestAccResourceIndex_extraSettingsKeptOnUpgrade(t *testing.T) {
	indexName := randResourceID(80)
	resourceName := fmt.Sprintf("algolia_index.%s", indexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexExtraSettings(indexName, ""),
				// Simulate settings configured outside Terraform (e.g. from the dashboard)
				// before the provider supported them.
				Check: func(s *terraform.State) error {
					apiClient := newTestAPIClient()
					settings := search.NewEmptyIndexSettings().
						SetEnableReRanking(true).
						SetUserData(map[string]any{"owner": "dashboard"})
					res, err := apiClient.searchClient.SetSettings(apiClient.searchClient.NewApiSetSettingsRequest(indexName, settings))
					if err != nil {
						return err
					}
					_, err = apiClient.searchClient.WaitForTask(indexName, res.TaskID)
					return err
				},
			},
			{
				// An unrelated change must not reset the settings managed outside Terraform.
				Config: testAccResourceIndexExtraSettings(indexName, `
  enable_rules = false
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_re_ranking", "true"),
					testCheckJSONResourceAttr(resourceName, "user_data_json", `{"owner":"dashboard"}`),
					resource.TestCheckResourceAttr(resourceName, "enable_rules", "false"),
					testAccCheckIndexSettings(indexName, func(s *search.SettingsResponse) error {
						if !s.GetEnableReRanking() {
							return fmt.Errorf("enableReRanking was reset")
						}
						// Regression: `enable_rules = false` used to be dropped by d.GetOk.
						if s.GetEnableRules() {
							return fmt.Errorf("enableRules: expected false")
						}
						return testAccCheckJSONEqual("userData", s.UserData, `{"owner":"dashboard"}`)
					}),
				),
			},
		},
	})
}

func TestAccResourceVirtualIndex_extraSettings(t *testing.T) {
	primaryIndexName := randResourceID(60)
	virtualIndexName := fmt.Sprintf("%s_virtual", primaryIndexName)
	resourceName := fmt.Sprintf("algolia_virtual_index.%s", virtualIndexName)
	dataSourceName := fmt.Sprintf("data.algolia_virtual_index.%s", virtualIndexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "algolia_index" "` + primaryIndexName + `" {
  name                = "` + primaryIndexName + `"
  deletion_protection = false
}

resource "algolia_virtual_index" "` + virtualIndexName + `" {
  name               = "` + virtualIndexName + `"
  primary_index_name = algolia_index.` + primaryIndexName + `.name

  enable_re_ranking       = true
  re_ranking_apply_filter = "brand:Apple"
  rendering_content_json  = jsonencode(` + testRenderingContentHCL + `)
  mode                    = "keywordSearch"

  deletion_protection = false
}

data "algolia_virtual_index" "` + virtualIndexName + `" {
  name = algolia_virtual_index.` + virtualIndexName + `.name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_re_ranking", "true"),
					resource.TestCheckResourceAttr(resourceName, "re_ranking_apply_filter", "brand:Apple"),
					testCheckJSONResourceAttr(resourceName, "rendering_content_json", testRenderingContentJSON),
					resource.TestCheckResourceAttr(resourceName, "mode", "keywordSearch"),
					resource.TestCheckNoResourceAttr(resourceName, "user_data_json"),
					testAccCheckIndexSettings(virtualIndexName, func(s *search.SettingsResponse) error {
						return testAccCheckJSONEqual("renderingContent", s.RenderingContent, testRenderingContentJSON)
					}),
					resource.TestCheckResourceAttr(dataSourceName, "enable_re_ranking", "true"),
					testCheckJSONResourceAttr(dataSourceName, "rendering_content_json", testRenderingContentJSON),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateId:           virtualIndexName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_protection"},
			},
		},
	})
}

const testRenderingContentHCL = `{
    facetOrdering = {
      facets = { order = ["brand", "color"] }
      values = {
        brand = { order = ["Apple", "Samsung"], sortRemainingBy = "alpha" }
      }
    }
  }`

func testAccResourceIndexExtraSettings(indexName, extraSettings string) string {
	return `
resource "algolia_index" "` + indexName + `" {
  name                = "` + indexName + `"
  deletion_protection = false
` + extraSettings + `
}

data "algolia_index" "` + indexName + `" {
  name = algolia_index.` + indexName + `.name
}
`
}

// testAccCheckIndexSettings verifies the settings actually stored in Algolia.
func testAccCheckIndexSettings(indexName string, check func(*search.SettingsResponse) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiClient := newTestAPIClient()
		settings, err := apiClient.searchClient.GetSettings(apiClient.searchClient.NewApiGetSettingsRequest(indexName))
		if err != nil {
			return err
		}
		if err := check(settings); err != nil {
			return fmt.Errorf("index %s: %w", indexName, err)
		}
		return nil
	}
}

func testAccCheckJSONEqual(name string, actual any, expectedJSON string) error {
	b, err := json.Marshal(actual)
	if err != nil {
		return err
	}
	equal, err := jsonBytesEqual([]byte(expectedJSON), b)
	if err != nil {
		return err
	}
	if !equal {
		return fmt.Errorf("%s: expected %s, got %s", name, expectedJSON, b)
	}
	return nil
}

// TestAccResourceIndex_upgradeFromV062 applies a configuration with the last release that did
// not support the newer settings, then plans it with this provider: the plan must be empty.
// The configuration avoids `enable_rules = false`, which v0.6.2 never applied, so the new
// provider would rightly plan to fix it.
func TestAccResourceIndex_upgradeFromV062(t *testing.T) {
	primaryIndexName := randResourceID(60)
	virtualIndexName := fmt.Sprintf("%s_virtual", primaryIndexName)
	config := `
resource "algolia_index" "` + primaryIndexName + `" {
  name                = "` + primaryIndexName + `"
  deletion_protection = false

  ranking_config {
    custom_ranking = ["desc(popularity)"]
  }
  faceting_config {
    max_values_per_facet = 50
  }
}

resource "algolia_virtual_index" "` + virtualIndexName + `" {
  name                = "` + virtualIndexName + `"
  primary_index_name  = algolia_index.` + primaryIndexName + `.name
  deletion_protection = false

  ranking_config {
    custom_ranking = ["asc(price)"]
  }
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"algolia": {Source: "k-yomo/algolia", VersionConstraint: "0.6.2"},
				},
				Config: config,
				// v0.6.2 does not round-trip its own state against the current API
				// (e.g. it drops the default `ranking` on read), which this release fixes.
				ExpectNonEmptyPlan: true,
			},
			{
				ProviderFactories: providerFactories,
				Config:            config,
				PlanOnly:          true,
			},
		},
	})
}

// TestAccResourceIndex_partialBlocksAreIdempotent is a regression test for list and set
// attributes with Algolia defaults (e.g. `ranking`, `attributes_to_retrieve`): configuring their
// block without them used to plan a change on every run (#94).
func TestAccResourceIndex_partialBlocksAreIdempotent(t *testing.T) {
	primaryIndexName := randResourceID(60)
	virtualIndexName := fmt.Sprintf("%s_virtual", primaryIndexName)
	primaryResourceName := fmt.Sprintf("algolia_index.%s", primaryIndexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "algolia_index" "` + primaryIndexName + `" {
  name                = "` + primaryIndexName + `"
  deletion_protection = false

  attributes_config {
    searchable_attributes = ["title"]
  }
  ranking_config {
    custom_ranking = ["desc(popularity)"]
  }
  query_strategy_config {
    query_type = "prefixAll"
  }
  advanced_config {
    min_proximity = 2
  }
}

resource "algolia_virtual_index" "` + virtualIndexName + `" {
  name                = "` + virtualIndexName + `"
  primary_index_name  = algolia_index.` + primaryIndexName + `.name
  deletion_protection = false

  attributes_config {
    unretrievable_attributes = ["cost"]
  }
  query_strategy_config {
    query_type = "prefixAll"
  }
  advanced_config {
    min_proximity = 2
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(primaryResourceName, "ranking_config.0.ranking.#", "8"),
					testCheckResourceListAttr(primaryResourceName, "attributes_config.0.attributes_to_retrieve", []string{"*"}),
					testCheckResourceListAttr(primaryResourceName, "advanced_config.0.response_fields", []string{"*"}),
					// Unconfigured lists must not be sent as `[]`: `attributesToRetrieve: []` makes
					// search return only objectID, and `alternativesAsExact: []` disables them.
					testAccCheckIndexSettings(primaryIndexName, testAccCheckDefaultListSettingsKept),
					testAccCheckIndexSettings(virtualIndexName, testAccCheckDefaultListSettingsKept),
				),
			},
		},
	})
}

func testAccCheckDefaultListSettingsKept(s *search.SettingsResponse) error {
	if s.AttributesToRetrieve != nil && len(s.AttributesToRetrieve) == 0 {
		return fmt.Errorf("attributesToRetrieve was overridden with []")
	}
	if s.ResponseFields != nil && len(s.ResponseFields) == 0 {
		return fmt.Errorf("responseFields was overridden with []")
	}
	if s.AlternativesAsExact != nil && len(s.AlternativesAsExact) == 0 {
		return fmt.Errorf("alternativesAsExact was overridden with []")
	}
	if s.AdvancedSyntaxFeatures != nil && len(s.AdvancedSyntaxFeatures) == 0 {
		return fmt.Errorf("advancedSyntaxFeatures was overridden with []")
	}
	return nil
}
