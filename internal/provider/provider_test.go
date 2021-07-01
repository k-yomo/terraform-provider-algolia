package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"algolia": func() (*schema.Provider, error) {
		return newTestAlgoliaProvider(), nil
	},
}

func TestProvider(t *testing.T) {
	if err := newTestAlgoliaProvider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func newTestAlgoliaProvider() *schema.Provider {
	return New("dev")()
}

func newTestAPIClient() *apiClient {
	return newAPIClient(os.Getenv("ALGOLIA_APP_ID"), os.Getenv("ALGOLIA_API_KEY"), "test")
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("ALGOLIA_APP_ID") == "" {
		t.Fatal("env variable 'ALGOLIA_APP_ID' is not set")
	}
	if os.Getenv("ALGOLIA_API_KEY") == "" {
		t.Fatal("env variable 'ALGOLIA_API_KEY' is not set")
	}
}
