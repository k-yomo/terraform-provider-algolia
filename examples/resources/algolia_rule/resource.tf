resource "algolia_rule" "example" {
  index_name = "example_index"
  object_id  = "example"

  conditions {
    pattern   = "{facet:category}"
    anchoring = "contains"
  }

  consequence {
    params_json = jsonencode({
      automaticFacetFilters = {
        facet       = "category"
        disjunctive = true
      }
    })
  }
}
