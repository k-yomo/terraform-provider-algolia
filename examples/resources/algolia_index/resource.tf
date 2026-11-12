resource "algolia_index" "primary" {
  name = "primary-index"
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

  languages_config {
    remove_stop_words_for = ["en"]
  }

  // Order facets and facet values in the search UI.
  rendering_content_json = jsonencode({
    facetOrdering = {
      facets = { order = ["category"] }
      values = {
        category = { order = ["Books", "Movies"], sortRemainingBy = "count" }
      }
    }
  })
}

resource "algolia_index" "replica" {
  name               = "replica-index"
  primary_index_name = algolia_index.primary.name
  // Configure settings for this replica
}
