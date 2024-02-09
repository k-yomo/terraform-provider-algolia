# Terraform Provider Algolia

[![License: MPL-2.0](https://img.shields.io/badge/License-MPL2.0-blue.svg)](./LICENSE)
[![Tests Workflow](https://github.com/k-yomo/terraform-provider-algolia/workflows/Tests/badge.svg)](https://github.com/k-yomo/terraform-provider-algolia/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/k-yomo/terraform-provider-algolia/branch/main/graph/badge.svg)](https://codecov.io/gh/k-yomo/terraform-provider-algolia)
[![Go Report Card](https://goreportcard.com/badge/k-yomo/terraform-provider-algolia)](https://goreportcard.com/report/k-yomo/terraform-provider-algolia)

Terraform Provider for [Algolia](https://www.algolia.com).

## Documentation

Full, comprehensive documentation is available on the Terraform website:

[https://registry.terraform.io/providers/k-yomo/algolia/latest/docs](https://registry.terraform.io/providers/k-yomo/algolia/latest/docs)

## Using the provider

Set an environment variable `ALGOLIA_API_KEY` to store your Algolia API key.

```sh
$ export ALGOLIA_API_KEY=<your api key>
```

The example below demonstrates the following operations:

- create index
- create rule for the index
- create api key to search the index

```terraform
terraform {
  required_providers {
    algolia = {
      source = "k-yomo/algolia"
      version = ">= 0.1.0, < 1.0.0"
    }
  }
}

provider "algolia" {
  app_id = "XXXXXXXXXX"
}

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
}

resource "algolia_rule" "example" {
  index_name = algolia_index.example.name
  object_id = "example-rule"

  conditions {
    pattern = "{facet:category}"
    anchoring = "contains"
  }

  consequence {
    params {
      automatic_facet_filters {
        facet = "category"
        disjunctive = true
      }
    }
  }
}


resource "algolia_api_key" "example" {
  acl = [
    "search",
    "browse"
  ]
  expires_at = "2030-01-01T00:00:00Z"
  max_hits_per_query = 100
  max_queries_per_ip_per_hour = 10000
  description = "This is a example api key"
  indexes = [algolia_index.example.name]
  referers = ["https://algolia.com/\\*"]
  query_parameters = ""
}
```

## Supported resources

- [x] [Settings](https://www.algolia.com/doc/api-client/methods/settings/)
- [x] [Rule](https://www.algolia.com/doc/api-client/methods/rules/)
- [x] [Api Keys](https://www.algolia.com/doc/api-client/methods/api-keys/)
- [x] [Synonym](https://www.algolia.com/doc/api-client/methods/synonyms/)
- [x] [Query Suggestions](https://www.algolia.com/doc/rest-api/query-suggestions/)
- [ ] [A/B Test](https://www.algolia.com/doc/api-client/methods/ab-test/)
- [ ] [Dictionaries](https://www.algolia.com/doc/api-client/methods/dictionaries/)
- [ ] [Personalization](https://www.algolia.com/doc/api-client/methods/personalization/)

## Contributing

I appreciate your help!

To contribute, please read the [Contributing to Terraform - Algolia Provider](./CONTRIBUTING.md)

## Sponsor
<a href="https://www.manomano.com">
  <img src="https://avatars.githubusercontent.com/u/25135828?s=200&v=4" width="64">
  <br>
  ManoMano
</a>
