package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

// TODO: Cover all params
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
				),
			},
			{
				Config: testAccResourceIndexUpdate(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", indexName),
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
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     indexName,
				ImportState:       true,
				ImportStateVerify: true,
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
	return fmt.Sprintf(`
resource "algolia_index" "%s" {
  name = "%s"

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

  languages_config {
    remove_stop_words_for = ["en"]
  }
}
`, name, name)
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
