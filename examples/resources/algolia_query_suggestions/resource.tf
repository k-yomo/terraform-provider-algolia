resource "algolia_index" "example" {
  name = "example"
}

resource "algolia_index" "example_source" {
  name = "example_source"
}

resource "algolia_query_suggestions" "example" {
  index_name = algolia_index.example.name

  source_indices {
    index_name  = algolia_index.example_source.name
    min_hits    = 10
    min_letters = 3
    generate    = [["brand"], ["brand", "category"]]
  }

  languages = ["en", "ja"]
}
