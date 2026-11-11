package provider

import (
	"context"
	"fmt"
	"time"

	suggestions "github.com/algolia/algoliasearch-client-go/v4/algolia/query-suggestions"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
)

func resourceQuerySuggestions() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceQuerySuggestionsCreate,
		ReadContext:   resourceQuerySuggestionsRead,
		UpdateContext: resourceQuerySuggestionsUpdate,
		DeleteContext: resourceQuerySuggestionsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceQuerySuggestionsStateContext,
		},
		Description: "A configuration that lies behind your Query Suggestions index.",
		// https://www.algolia.com/doc/rest-api/query-suggestions/#create-a-configuration
		Schema: map[string]*schema.Schema{
			"index_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Index name to target.",
			},
			"region": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "us",
				ValidateFunc: validation.StringInSlice(algoliautil.ValidRegionStrings, false),
				Description:  `Region to create the index in. "us", "eu", "de" are supported. Defaults to "us" when not specified.`,
			},
			"source_indices": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "A list of source indices used to generate a Query Suggestions index.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Index name to target.",
						},
						"analytics_tags": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Optional: true,
							DefaultFunc: func() (interface{}, error) {
								return []string{}, nil
							},
							Description: "A list of analytics tags to filter the popular searches per tag.",
						},
						"facets": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attribute": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Category attribute in your index",
									},
									"amount": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "How many of the top categories to show",
									},
								},
							},
							DefaultFunc: func() (interface{}, error) {
								return []map[string]interface{}{}, nil
							},
							Description: "A list of facets to define as categories for the query suggestions.",
						},
						"min_hits": {
							Type:        schema.TypeInt,
							Computed:    true,
							Optional:    true,
							Description: "Minimum number of hits (e.g., matching records in the source index) to generate a suggestions.",
						},
						"min_letters": {
							Type:        schema.TypeInt,
							Computed:    true,
							Optional:    true,
							Description: "Minimum number of required letters for a suggestion to remain.",
						},
						"generate": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeList,
								Elem: &schema.Schema{Type: schema.TypeString},
							},
							Description: `List of facet attributes used to generate Query Suggestions. The resulting suggestions are every combination of the facets in the nested list
(e.g., (facetA and facetB) and facetC).
` + "```" + `
[
  ["facetA", "facetB"],
  ["facetC"]
]
` + "```",
						},
						"external": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Optional: true,
							DefaultFunc: func() (interface{}, error) {
								return []string{}, nil
							},
							Description: "A list of external indices to use to generate custom Query Suggestions.",
						},
					},
				},
			},
			"languages": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "A list of languages used to de-duplicate singular and plural suggestions.",
			},
			"exclude": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "A list of words and patterns to exclude from the Query Suggestions index.",
			},
		},
	}
}

func resourceQuerySuggestionsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	suggestionsClient, err := newSuggestionsClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	indexName := d.Get("index_name").(string)
	configWithIndex := mapToQuerySuggestionsConfigWithIndex(d)
	_, err = suggestionsClient.CreateConfig(suggestionsClient.NewApiCreateConfigRequest(&configWithIndex), suggestions.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(indexName)

	return resourceQuerySuggestionsRead(ctx, d, m)
}

func resourceQuerySuggestionsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := refreshQuerySuggestionsState(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceQuerySuggestionsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	suggestionsClient, err := newSuggestionsClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	indexName := d.Get("index_name").(string)
	config := mapToQuerySuggestionsConfig(d)
	_, err = suggestionsClient.UpdateConfig(suggestionsClient.NewApiUpdateConfigRequest(indexName, &config), suggestions.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(indexName)

	return resourceQuerySuggestionsRead(ctx, d, m)
}

func resourceQuerySuggestionsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	suggestionsClient, err := newSuggestionsClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	indexName := d.Get("index_name").(string)
	_, err = suggestionsClient.DeleteConfig(suggestionsClient.NewApiDeleteConfigRequest(indexName), suggestions.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceQuerySuggestionsStateContext(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	r, id, err := parseImportRegionAndId(d.Id())
	if err != nil {
		return nil, err
	}
	if r != "" {
		if err := d.Set("region", string(r)); err != nil {
			return nil, err
		}
	}
	d.SetId(id)
	if err := refreshQuerySuggestionsState(ctx, d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func refreshQuerySuggestionsState(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	suggestionsClient, err := newSuggestionsClient(d, m)
	if err != nil {
		return err
	}

	indexName := d.Id()

	var querySuggestionsIndexConfig *suggestions.ConfigurationResponse
	err = retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		var err error
		querySuggestionsIndexConfig, err = suggestionsClient.GetConfig(suggestionsClient.NewApiGetConfigRequest(indexName), suggestions.WithContext(ctx))

		if d.IsNewResource() && algoliautil.IsRetryableError(err) {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		if algoliautil.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("query suggestions index (%s) not found, removing from state", d.Id()))
			d.SetId("")
			return nil
		}
		return err
	}

	var sourceIndices []interface{}
	for _, sourceIndex := range querySuggestionsIndexConfig.SourceIndices {
		var facets []map[string]interface{}
		for _, f := range sourceIndex.Facets {
			facets = append(facets, map[string]interface{}{
				"attribute": f.GetAttribute(),
				"amount":    int(f.GetAmount()),
			})
		}
		minHits := 0
		if sourceIndex.MinHits != nil {
			minHits = int(*sourceIndex.MinHits)
		}
		minLetters := 0
		if sourceIndex.MinLetters != nil {
			minLetters = int(*sourceIndex.MinLetters)
		}
		sourceIndices = append(sourceIndices, map[string]interface{}{
			"index_name":     sourceIndex.IndexName,
			"analytics_tags": sourceIndex.AnalyticsTags,
			"facets":         facets,
			"min_hits":       minHits,
			"min_letters":    minLetters,
			"generate":       sourceIndex.Generate,
			"external":       sourceIndex.External,
		})
	}

	var languages []string
	if querySuggestionsIndexConfig.Languages.ArrayOfString != nil {
		languages = *querySuggestionsIndexConfig.Languages.ArrayOfString
	}

	values := map[string]interface{}{
		"index_name":     querySuggestionsIndexConfig.IndexName,
		"source_indices": sourceIndices,
		"languages":      languages,
		"exclude":        querySuggestionsIndexConfig.Exclude,
	}
	if err := setValues(d, values); err != nil {
		return err
	}

	return nil
}

func mapToQuerySuggestionsConfigWithIndex(d *schema.ResourceData) suggestions.ConfigurationWithIndex {
	config := suggestions.ConfigurationWithIndex{
		IndexName: d.Get("index_name").(string),
	}

	if v, ok := d.GetOk("source_indices"); ok {
		config.SourceIndices = unmarshalSourceIndices(v)
	}

	if v, ok := d.GetOk("languages"); ok {
		config.Languages = suggestions.ArrayOfStringAsLanguages(castStringSet(v))
	}

	if v, ok := d.GetOk("exclude"); ok {
		config.Exclude = castStringSet(v)
	}

	return config
}

func mapToQuerySuggestionsConfig(d *schema.ResourceData) suggestions.Configuration {
	config := suggestions.Configuration{}

	if v, ok := d.GetOk("source_indices"); ok {
		config.SourceIndices = unmarshalSourceIndices(v)
	}

	if v, ok := d.GetOk("languages"); ok {
		config.Languages = suggestions.ArrayOfStringAsLanguages(castStringSet(v))
	}

	if v, ok := d.GetOk("exclude"); ok {
		config.Exclude = castStringSet(v)
	}

	return config
}

func unmarshalSourceIndices(configured interface{}) []suggestions.SourceIndex {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	sourceIndices := make([]suggestions.SourceIndex, 0, len(l))
	for _, v := range l {
		sourceIndexMap := v.(map[string]interface{})
		sourceIndex := suggestions.SourceIndex{
			IndexName: sourceIndexMap["index_name"].(string),
		}
		if v, ok := sourceIndexMap["analytics_tags"]; ok {
			sourceIndex.AnalyticsTags = castStringSet(v)
		}
		if v, ok := sourceIndexMap["facets"]; ok {
			var facets []suggestions.Facet
			for _, facet := range v.([]interface{}) {
				facetMap := castInterfaceMap(facet)
				attr := facetMap["attribute"].(string)
				amt := int32(facetMap["amount"].(int))
				facets = append(facets, suggestions.Facet{
					Attribute: &attr,
					Amount:    &amt,
				})
			}
			sourceIndex.Facets = facets
		}
		if v, ok := sourceIndexMap["min_hits"]; ok {
			minHits := int32(v.(int))
			sourceIndex.MinHits = &minHits
		}
		if v, ok := sourceIndexMap["min_letters"]; ok {
			minLetters := int32(v.(int))
			sourceIndex.MinLetters = &minLetters
		}
		if v, ok := sourceIndexMap["generate"]; ok {
			var generate [][]string
			for _, facet := range v.([]interface{}) {
				generate = append(generate, castStringList(facet))
			}
			sourceIndex.Generate = generate
		}
		if v, ok := sourceIndexMap["external"]; ok {
			sourceIndex.External = castStringSet(v)
		}
		sourceIndices = append(sourceIndices, sourceIndex)
	}
	return sourceIndices
}

func newSuggestionsClient(d *schema.ResourceData, m interface{}) (*suggestions.APIClient, error) {
	apiClient := m.(*apiClient)
	r := suggestions.Region(d.Get("region").(string))
	return apiClient.newSuggestionsClient(r)
}
