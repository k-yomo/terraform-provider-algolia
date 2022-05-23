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

func TestAccResourceIndexWithReplica(t *testing.T) {
	// NOTE: Deleting replica fails due to the same reason as the below issue.
	// https://github.com/algolia/algoliasearch-client-javascript/issues/1377
	// TODO: Remove t.Skip() once the issue is resolved.
	t.Skip()

	primaryIndexName := randStringStartWithAlpha(80)
	replicaIndexName := fmt.Sprintf("%s_replica", primaryIndexName)
	primaryIndexResourceName := fmt.Sprintf("algolia_index.%s", primaryIndexName)
	replicaIndexResourceName := fmt.Sprintf("algolia_index.%s", replicaIndexName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexWithReplica(primaryIndexName, replicaIndexName),
				Check: resource.ComposeTestCheckFunc(
					// primary index
					resource.TestCheckResourceAttr(primaryIndexResourceName, "name", primaryIndexName),
					// replica index
					resource.TestCheckResourceAttr(replicaIndexResourceName, "name", replicaIndexName),
					resource.TestCheckResourceAttr(replicaIndexResourceName, "primary_index_name", primaryIndexName),
				),
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

// nolint:unused
func testAccResourceIndexWithReplica(name string, replicaName string) string {
	return `
resource "algolia_index" "` + name + `" {
  name = "` + name + `"

  deletion_protection = false
}

resource "algolia_index" "` + replicaName + `" {
  name               =  "` + replicaName + `"
  primary_index_name = algolia_index.` + name + `.name

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
