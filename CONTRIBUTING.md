# Contributing to Terraform - Algolia Provider

## Prerequisites

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.18

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the command below:
```sh
$ make install
```
4. Use the installed provider as following
```terraform
terraform {
  required_providers {
    algolia = {
      source  = "k-yomo/algolia"
      version = "999.999.999"
    }
  }
}
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```sh
go get -u github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Prerequisites](#Prerequisites) above).

When you changed schema or `examples` directory, you need to regenerate documents.
To generate or update documentation, run `make generate`.

### Test
Set required env variables.
```sh
$ echo ALGOLIA_APP_ID={APP_ID} >> .env
$ echo ALGOLIA_API_KEY={API_KEY} >> .env
$ direnv allow
```

Run `make testacc` to run the full suite of Acceptance tests.
```sh
$ make testacc
```
*Note:* Acceptance tests create real resources, and might cost money to run.

