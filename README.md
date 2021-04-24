# Terraform Provider Algolia

![License: MIT](https://img.shields.io/badge/License-MPL2.0-blue.svg)
![Test Workflow](https://github.com/k-yomo/terraform-provider-algolia/workflows/Tests/badge.svg)
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
- create api key

```terraform
terraform {
  required_providers {
    stripe = {
      source = "k-yomo/algolia"
      version = "0.0.2" # use the latest released version
    }
  }
}

provider "algolia" {
  app_id = "XXXXXXXXXX"
}

resource "algolia_api_key" "example" {
  acl                         = ["search", "browse"]
  expires_at                  = 2524608000 # 01 Jan 2050 00:00:00 GMT
  max_hits_per_query          = 100
  max_queries_per_ip_per_hour = 10000
  description                 = "This is a example api key"
  indexes                     = ["dev_*"]
  referers                    = ["https://algolia.com/\\*"]
}
```

## Contributing

I appreciate your help!

To contribute, please read the [Contributing to Terraform - Algolia Provider](./CONTRIBUTING.md)
