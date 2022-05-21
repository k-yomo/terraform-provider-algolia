resource "algolia_index" "example" {
  name = "example"
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
      "body",
      "likes"
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

  replicas = ["virtual(example_replica)"]
}

resource "algolia_virtual_index" "example_virtual_replica" {
  name = "example_replica"

  attributes_config {
    unretrievable_attributes = [
      "author_email"
    ]
    attributes_to_retrieve = [
      "title",
      "category",
      "tag",
      "likes"
    ]
  }

  ranking_config {
    custom_ranking = ["desc(likes)"]
  }

  faceting_config {
    max_values_per_facet = 50
    sort_facet_values_by = "alpha"
  }

  languages_config {
    remove_stop_words_for = ["en"]
  }
}
