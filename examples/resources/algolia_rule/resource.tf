resource "algolia_rule" "example" {
  index_name = "example_index"
  object_id  = "example"

  conditions {
    pattern   = "{facet:category}"
    anchoring = "contains"
  }

  consequence {
    params {
      automatic_facet_filters {
        facet       = "category"
        disjunctive = true
      }
    }
  }
}
