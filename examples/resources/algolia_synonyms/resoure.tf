resource "algolia_index" "example" {
  name = "example"
}
resource "algolia_synonyms" "example" {
  index_name = algolia_index.example.name

  synonyms {
    object_id = "test_1"
    type      = "synonym"
    synonyms  = ["smartphone", "mobile phone", "cell phone"]
  }
  synonyms {
    object_id = "test_2"
    type      = "oneWaySynonym"
    input     = "smartphone"
    synonyms  = ["iPhone", "Pixel"]
  }
  synonyms {
    object_id   = "test_3"
    type        = "altCorrection1"
    word        = "tablet"
    corrections = ["ipad"]
  }
  synonyms {
    object_id   = "test_4"
    type        = "altCorrection2"
    word        = "tablet"
    corrections = ["iphone"]
  }
  synonyms {
    object_id    = "test_5"
    type         = "placeholder"
    placeholder  = "<model>"
    replacements = ["6", "7", "8"]
  }
}