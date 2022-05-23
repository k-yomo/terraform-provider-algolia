resource "algolia_index" "example" {
  name = "example"

  attributes_config {
    searchable_attributes = [
      "title",
      "description",
    ]
    attributes_to_retrieve = ["*"]
  }

  replicas = ["virtual(example_replica)"]
}

resource "algolia_virtual_index" "example_virtual_replica" {
  name               = "example_replica"
  primary_index_name = "example"

  attributes_config {
    unretrievable_attributes = [
      "author_email"
    ]
    attributes_to_retrieve = ["*"]
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
