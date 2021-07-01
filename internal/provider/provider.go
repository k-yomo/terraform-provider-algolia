package provider

import (
	"context"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/suggestions"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// nolint: gochecknoinits
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
					Description: "The ID of the application. Defaults to the env variable `ALGOLIA_APP_ID`.",
				},
				"api_key": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ALGOLIA_API_KEY", nil),
					Description: "The API key to access algolia resources. Defaults to the env variable `ALGOLIA_API_KEY`.",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"algolia_index":             resourceIndex(),
				"algolia_api_key":           resourceAPIKey(),
				"algolia_rule":              resourceRule(),
				"algolia_synonyms":          resourceSynonyms(),
				"algolia_query_suggestions": resourceQuerySuggestions(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"algolia_index": dataSourceIndex(),
			},
		}
		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	searchClient      *search.Client
	suggestionsClient *suggestions.Client
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		userAgent := p.UserAgent("terraform-provider-algolia", version)
		return newAPIClient(d.Get("app_id").(string), d.Get("api_key").(string), userAgent), nil
	}
}

func newAPIClient(appID, apiKey, userAgent string) *apiClient {
	searchConfig := search.Configuration{
		AppID:          appID,
		APIKey:         apiKey,
		ExtraUserAgent: userAgent,
	}
	searchClient := search.NewClientWithConfig(searchConfig)

	suggestionsConfig := suggestions.Configuration{
		AppID:          appID,
		APIKey:         apiKey,
		ExtraUserAgent: userAgent,
	}
	suggestionsClient := suggestions.NewClientWithConfig(suggestionsConfig)

	return &apiClient{
		searchClient:      searchClient,
		suggestionsClient: suggestionsClient,
	}
}
