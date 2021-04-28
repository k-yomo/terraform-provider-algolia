# algolia Provider
The Algolia provider is used to configure your Algolia settings.

## Setting API Key
In order to make requests against the Algolia API, you need to set an API key.
You set the API key to Terraform using the environment variable `ALGOLIA_API_KEY`.
```sh
$ export ALGOLIA_API_KEY={{my-api-key}}
```

## Example Usage
Then typical provider configuration will look something like:
```terraform
provider "algolia" {
  app_id  = "my-app-id"
}
```

## Schema
### Optional
- `api_key` (String) The API key to access algolia resources. Defaults to the env variable `ALGOLIA_API_KEY`.
- `app_id` (String) The ID of the application. Defaults to the env variable `ALGOLIA_APP_ID`.

## Contributing
If you'd like to help extend the Algolia provider, that's more than welcome! Our full contribution guide is available at [CONTRIBUTING.md](https://github.com/k-yomo/terraform-provider-algolia/blob/main/CONTRIBUTING.md)

Pull requests can be made against [Provider Repository](https://github.com/k-yomo/terraform-provider-algolia/).
