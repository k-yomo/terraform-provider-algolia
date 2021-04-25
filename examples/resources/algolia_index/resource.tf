resource "algolia_index" "example" {
  name = "example"
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
  ranking = [
    "words",
    "proximity"
  ]
  replicas = [
    "replica_test1",
    "replica_test2",
  ]
  max_values_per_facet = 50
}