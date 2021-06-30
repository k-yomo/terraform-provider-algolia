package provider

import (
	"context"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/suggestions"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
							Elem:     &schema.Schema{Type: schema.TypeMap},
							DefaultFunc: func() (interface{}, error) {
								return []map[string]interface{}{}, nil
							},
							Description: "A list of facets to define as categories for the query suggestions.",
						},
						"min_hits": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     5, // This is the API's behaviour which is different from the doc https://www.algolia.com/doc/rest-api/query-suggestions/#method-param-minhits.
							Description: "Minimum number of hits (e.g., matching records in the source index) to generate a suggestions.",
						},
						"min_letters": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     4, // This is the API's behaviour which is different from the doc https://www.algolia.com/doc/rest-api/query-suggestions/#method-param-minletters
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
				Type: schema.TypeSet,
				Elem: &schema.Schema{Type: schema.TypeString},
				Set:  schema.HashString,
				// For now, we get deserialization error if `languages` is not set.
				// https://github.com/algolia/algoliasearch-client-go/issues/671
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
	apiClient := m.(*apiClient)

	indexName := d.Get("index_name").(string)
	err := apiClient.suggestionsClient.CreateConfig(mapToQuerySuggestionsIndexConfig(d), ctx)
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
	apiClient := m.(*apiClient)

	indexName := d.Get("index_name").(string)
	err := apiClient.suggestionsClient.UpdateConfig(mapToQuerySuggestionsIndexConfig(d), ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(indexName)

	return resourceQuerySuggestionsRead(ctx, d, m)
}

func resourceQuerySuggestionsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	indexName := d.Get("index_name").(string)
	err := apiClient.suggestionsClient.DeleteConfig(indexName, ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceQuerySuggestionsStateContext(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := d.Set("index_name", d.Id()); err != nil {
		return nil, err
	}
	if err := refreshQuerySuggestionsState(ctx, d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func refreshQuerySuggestionsState(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*apiClient)

	indexName := d.Get("index_name").(string)

	querySuggestionsIndexConfig, err := apiClient.suggestionsClient.GetConfig(indexName, ctx)
	if err != nil {
		d.SetId("")
		return err
	}

	var sourceIndices []interface{}
	for _, sourceIndex := range querySuggestionsIndexConfig.SourceIndices {
		sourceIndices = append(sourceIndices, map[string]interface{}{
			"index_name":     sourceIndex.IndexName,
			"analytics_tags": sourceIndex.AnalyticsTags,
			"facets":         sourceIndex.Facets,
			"min_hits":       sourceIndex.MinHits,
			"min_letters":    sourceIndex.MinLetters,
			"generate":       sourceIndex.Generate,
			"external":       sourceIndex.External,
		})
	}

	values := map[string]interface{}{
		"index_name":     querySuggestionsIndexConfig.IndexName,
		"source_indices": sourceIndices,
		"languages":      querySuggestionsIndexConfig.Languages,
		"exclude":        querySuggestionsIndexConfig.Exclude,
	}
	if err := setValues(d, values); err != nil {
		return err
	}

	return nil
}

func mapToQuerySuggestionsIndexConfig(d *schema.ResourceData) suggestions.IndexConfiguration {
	indexConfig := suggestions.IndexConfiguration{
		IndexName: d.Get("index_name").(string),
	}

	if v, ok := d.GetOk("source_indices"); ok {
		unmarshalSourceIndices(v, &indexConfig)
	}

	if v, ok := d.GetOk("languages"); ok {
		indexConfig.Languages = castStringSet(v)
	}

	if v, ok := d.GetOk("exclude"); ok {
		indexConfig.Exclude = castStringSet(v)
	}

	return indexConfig
}

func unmarshalSourceIndices(configured interface{}, indexConfig *suggestions.IndexConfiguration) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
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
			var facets []map[string]interface{}
			for _, facet := range v.([]interface{}) {
				facets = append(facets, castInterfaceMap(facet))
			}
			sourceIndex.Facets = facets
		}
		if v, ok := sourceIndexMap["min_hits"]; ok {
			minHits := v.(int)
			sourceIndex.MinHits = &minHits
		}
		if v, ok := sourceIndexMap["min_letters"]; ok {
			minLetters := v.(int)
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
	indexConfig.SourceIndices = sourceIndices
}
