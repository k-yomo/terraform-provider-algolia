package provider

import (
	"context"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"app_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ALGOLIA_APP_ID", nil),
					Description: "The ID of the application",
				},
				"api_key": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ALGOLIA_API_KEY", nil),
					Description: "The API key to access algolia resources",
				},
			},
		}
		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	algoliaClient *search.Client
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		config := search.Configuration{
			AppID:          d.Get("app_id").(string),
			APIKey:         d.Get("api_key").(string),
			ExtraUserAgent: p.UserAgent("terraform-provider-algolia", version),
		}
		algoliaClient := search.NewClientWithConfig(config)

		return &apiClient{algoliaClient: algoliaClient}, nil
	}
}
