package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TODO: Cover all fields
func TestAccResourceIndex(t *testing.T) {
	indexName := randStringStartWithAlpha(100)
	resourceName := fmt.Sprintf("algolia_index.%s", indexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndex(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", indexName),
					resource.TestCheckResourceAttr(resourceName, "virtual", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "attributes_config.0.searchable_attributes.0"),
					resource.TestCheckNoResourceAttr(resourceName, "attributes_config.0.attributes_for_faceting.0"),
					resource.TestCheckNoResourceAttr(resourceName, "attributes_config.0.unretrievable_attributes.0"),
					resource.TestCheckResourceAttr(resourceName, "attributes_config.0.attributes_to_retrieve.0", "*"),
					testCheckResourceListAttr(resourceName, "ranking_config.0.ranking", []string{"typo", "geo", "words", "filters", "proximity", "attribute", "exact", "custom"}),
					resource.TestCheckNoResourceAttr(resourceName, "ranking_config.0.replicas.0"),
					testCheckResourceListAttr(resourceName, "highlight_and_snippet_config.0.attributes_to_highlight", []string{}),
					testCheckResourceListAttr(resourceName, "highlight_and_snippet_config.0.attributes_to_snippet", []string{}),
					resource.TestCheckResourceAttr(resourceName, "highlight_and_snippet_config.0.highlight_pre_tag", "<em>"),
					resource.TestCheckResourceAttr(resourceName, "highlight_and_snippet_config.0.highlight_post_tag", "</em>"),
					resource.TestCheckResourceAttr(resourceName, "highlight_and_snippet_config.0.snippet_ellipsis_text", ""),
					resource.TestCheckResourceAttr(resourceName, "highlight_and_snippet_config.0.restrict_highlight_and_snippet_arrays", "false"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				Config:       testAccResourceIndex(indexName),
				ResourceName: resourceName,
				Destroy:      true,
				ExpectError:  regexp.MustCompile("cannot destroy index without setting deletion_protection=false and running `terraform apply`"),
			},
			{
				Config: testAccResourceIndexUpdate(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", indexName),
					resource.TestCheckResourceAttr(resourceName, "virtual", "false"),
					testCheckResourceListAttr(resourceName, "attributes_config.0.searchable_attributes", []string{"title", "category,tag", "unordered(description)"}),
					testCheckResourceListAttr(resourceName, "attributes_config.0.attributes_for_faceting", []string{"category"}),
					testCheckResourceListAttr(resourceName, "attributes_config.0.unretrievable_attributes", []string{"author_email"}),
					testCheckResourceListAttr(resourceName, "attributes_config.0.attributes_to_retrieve", []string{"body", "category", "description", "tag", "title"}),
					testCheckResourceListAttr(resourceName, "ranking_config.0.ranking", []string{"words", "proximity"}),
					resource.TestCheckNoResourceAttr(resourceName, "ranking_config.0.replicas.0"),
					resource.TestCheckResourceAttr(resourceName, "faceting_config.0.max_values_per_facet", "50"),
					resource.TestCheckResourceAttr(resourceName, "faceting_config.0.sort_facet_values_by", "alpha"),
					testCheckResourceListAttr(resourceName, "highlight_and_snippet_config.0.attributes_to_highlight", []string{"title"}),
					testCheckResourceListAttr(resourceName, "highlight_and_snippet_config.0.attributes_to_snippet", []string{"description:100"}),
					resource.TestCheckResourceAttr(resourceName, "highlight_and_snippet_config.0.highlight_pre_tag", "<b>"),
					resource.TestCheckResourceAttr(resourceName, "highlight_and_snippet_config.0.highlight_post_tag", "</b>"),
					resource.TestCheckResourceAttr(resourceName, "highlight_and_snippet_config.0.snippet_ellipsis_text", "..."),
					resource.TestCheckResourceAttr(resourceName, "highlight_and_snippet_config.0.restrict_highlight_and_snippet_arrays", "true"),
					resource.TestCheckResourceAttr(resourceName, "pagination_config.0.hits_per_page", "100"),
					resource.TestCheckResourceAttr(resourceName, "pagination_config.0.pagination_limited_to", "500"),
					resource.TestCheckResourceAttr(resourceName, "typos_config.0.min_word_size_for_1_typo", "3"),
					resource.TestCheckResourceAttr(resourceName, "typos_config.0.min_word_size_for_2_typos", "6"),
					resource.TestCheckResourceAttr(resourceName, "typos_config.0.typo_tolerance", "strict"),
					resource.TestCheckResourceAttr(resourceName, "typos_config.0.allow_typos_on_numeric_tokens", "false"),
					testCheckResourceListAttr(resourceName, "typos_config.0.disable_typo_tolerance_on_attributes", []string{"model"}),
					testCheckResourceListAttr(resourceName, "typos_config.0.disable_typo_tolerance_on_words", []string{"test"}),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
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
		CheckDestroy: testAccCheckIndexDestroy,
	})
}

func TestAccResourceVirtualIndex(t *testing.T) {
	indexName := randStringStartWithAlpha(80)
	virtualIndexName := fmt.Sprintf("%s_virtual", indexName)
	indexResourceName := fmt.Sprintf("algolia_index.%s", indexName)
	virtualIndexResourceName := fmt.Sprintf("algolia_index.%s", virtualIndexName)

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
					resource.TestCheckResourceAttr(indexResourceName, "attributes_config.0.attributes_to_retrieve.0", "*"),
					testCheckResourceListAttr(indexResourceName, "attributes_config.0.searchable_attributes", []string{"name", "description", "category_name"}),
					testCheckResourceListAttr(indexResourceName, "attributes_config.0.attributes_for_faceting", []string{"category_id"}),
					testCheckResourceListAttr(indexResourceName, "ranking_config.0.ranking", []string{"typo", "geo"}),
					testCheckResourceListAttr(indexResourceName, "ranking_config.0.replicas", []string{fmt.Sprintf("virtual(%s)", virtualIndexName)}),
					resource.TestCheckResourceAttr(indexResourceName, "advanced_config.0.distinct", "2"),
					resource.TestCheckResourceAttr(indexResourceName, "advanced_config.0.attribute_for_distinct", "url"),
					// virtual index
					resource.TestCheckResourceAttr(virtualIndexResourceName, "name", virtualIndexName),
					resource.TestCheckResourceAttr(virtualIndexResourceName, "virtual", "true"),
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
					"virtual",
					"attributes_config",
					"ranking_config",
					"advanced_config",
					"performance_config",
					"deletion_protection",
				},
			},
			{
				// NOTE: Removing from replica, then deleting virtual index should work, but it fails.
				// https://www.algolia.com/doc/guides/managing-results/refine-results/sorting/how-to/deleting-replicas/?client=go#using-the-api
				// So deleting primary index first, then deleting virtual index here
				Config: testAccResourceVirtualIndexOnly(virtualIndexName),
			},
		},
		CheckDestroy: testAccCheckIndexDestroy,
	})
}

func testAccResourceIndex(name string) string {
	return fmt.Sprintf(`
resource "algolia_index" "%s" {
  name = "%s"
}
`, name, name)
}

func testAccResourceIndexUpdate(name string) string {
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

  highlight_and_snippet_config {
    attributes_to_highlight = ["title"]
    attributes_to_snippet = ["description:100"]
    highlight_pre_tag = "<b>"
    highlight_post_tag = "</b>"
    snippet_ellipsis_text = "..."
    restrict_highlight_and_snippet_arrays = true
  }

  pagination_config {
    hits_per_page = 100
    pagination_limited_to = 500
  }

  typos_config {
    min_word_size_for_1_typo = 3
    min_word_size_for_2_typos = 6
    typo_tolerance = "strict"
    allow_typos_on_numeric_tokens = false
    disable_typo_tolerance_on_attributes = ["model"]
    disable_typo_tolerance_on_words = ["test"]
  }

  languages_config {
    remove_stop_words_for = ["en"]
  }

  deletion_protection = false
}
`
}

func testAccResourceVirtualIndex(name string, virtualIndexName string) string {
	return `
resource "algolia_index" "` + name + `" {
  name = "` + name + `"

  attributes_config {
    attributes_to_retrieve = ["*"]
    searchable_attributes = ["name", "description", "category_name"]
    attributes_for_faceting = ["category_id"]
  }

  ranking_config {
    ranking = ["typo", "geo"]
    replicas = ["virtual(` + virtualIndexName + `)"]
  }

  advanced_config {
    response_fields = ["*"]
    distinct = 2
    attribute_for_distinct = "url"
  }

  deletion_protection = false
}

resource "algolia_index" "` + virtualIndexName + `" {
  name    = "` + virtualIndexName + `"
  virtual = true

  ranking_config {
    custom_ranking = ["desc(likes)"]
  }

  advanced_config {
    response_fields = ["*"]
    distinct = 1
  }

  deletion_protection = false

  depends_on = [algolia_index.` + name + `] 
}
`
}

func testAccResourceVirtualIndexOnly(virtualIndexName string) string {
	return `
resource "algolia_index" "` + virtualIndexName + `" {
  name    = "` + virtualIndexName + `"
  virtual = true

  ranking_config {
    custom_ranking = ["desc(likes)"]
  }

  deletion_protection = false
}
`
}

func testAccCheckIndexDestroy(s *terraform.State) error {
	apiClient := newTestAPIClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "algolia_index" {
			continue
		}

		exists, err := apiClient.searchClient.InitIndex(rs.Primary.ID).Exists()
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("index '%s' still exists", rs.Primary.ID)
		}
	}

	return nil
}
