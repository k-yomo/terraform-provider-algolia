package provider

import (
	"context"
	"fmt"

	suggestions "github.com/algolia/algoliasearch-client-go/v4/algolia/query-suggestions"
	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/algolia/algoliasearch-client-go/v4/algolia/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
	"github.com/hashicorp/terraform-provider-algolia/internal/mutex"
)

// Global Key/Value Mutex
var mutexKV = mutex.NewKV()

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
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("ALGOLIA_API_KEY", nil),
					Description: "The API key to access algolia resources. Defaults to the env variable `ALGOLIA_API_KEY`.",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"algolia_index":             resourceIndex(),
				"algolia_virtual_index":     resourceVirtualIndex(),
				"algolia_api_key":           resourceAPIKey(),
				"algolia_rule":              resourceRule(),
				"algolia_synonyms":          resourceSynonyms(),
				"algolia_query_suggestions": resourceQuerySuggestions(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"algolia_index":         dataSourceIndex(),
				"algolia_virtual_index": dataSourceVirtualIndex(),
			},
		}
		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	userAgent string
	appID     string
	apiKey    string
	requester transport.Requester

	searchClient *search.APIClient
}

func (a *apiClient) newSuggestionsClient(region suggestions.Region) (*suggestions.APIClient, error) {
	cfg := suggestions.QuerySuggestionsConfiguration{
		Configuration: transport.Configuration{
			AppID:     a.appID,
			ApiKey:    a.apiKey,
			UserAgent: a.userAgent,
			Requester: a.requester,
		},
		Region: region,
	}
	return suggestions.NewClientWithConfig(cfg)
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		userAgent := p.UserAgent("terraform-provider-algolia", version)
		client, err := newAPIClient(d.Get("app_id").(string), d.Get("api_key").(string), userAgent)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return client, nil
	}
}

func newAPIClient(appID, apiKey, userAgent string) (*apiClient, error) {
	var algoliaRequester transport.Requester
	if logging.IsDebugOrHigher() {
		algoliaRequester = algoliautil.NewDebugRequester()
	}

	searchConfig := search.SearchConfiguration{
		Configuration: transport.Configuration{
			AppID:     appID,
			ApiKey:    apiKey,
			UserAgent: userAgent,
			Requester: algoliaRequester,
		},
	}
	searchClient, err := search.NewClientWithConfig(searchConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Algolia search client: %w", err)
	}

	return &apiClient{
		appID:        appID,
		apiKey:       apiKey,
		userAgent:    userAgent,
		requester:    algoliaRequester,
		searchClient: searchClient,
	}, nil
}
