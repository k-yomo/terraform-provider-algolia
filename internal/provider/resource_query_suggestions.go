package provider

import (
	"context"
	"log"
	"time"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/region"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/suggestions"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				Default:      region.US,
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
	suggestionsClient := newSuggestionsClient(d, m)

	indexName := d.Get("index_name").(string)
	err := suggestionsClient.CreateConfig(mapToQuerySuggestionsIndexConfig(d), ctx)
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
	suggestionsClient := newSuggestionsClient(d, m)

	indexName := d.Get("index_name").(string)
	err := suggestionsClient.UpdateConfig(mapToQuerySuggestionsIndexConfig(d), ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(indexName)

	return resourceQuerySuggestionsRead(ctx, d, m)
}

func resourceQuerySuggestionsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	suggestionsClient := newSuggestionsClient(d, m)

	indexName := d.Get("index_name").(string)
	err := suggestionsClient.DeleteConfig(indexName, ctx)
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
	suggestionsClient := newSuggestionsClient(d, m)

	indexName := d.Id()

	var querySuggestionsIndexConfig *suggestions.IndexConfiguration
	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		querySuggestionsIndexConfig, err = suggestionsClient.GetConfig(indexName, ctx)

		if d.IsNewResource() && algoliautil.IsRetryableError(err) {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		if algoliautil.IsNotFoundError(err) {
			log.Printf("[WARN] query suggestions index (%s) not found, removing from state", d.Id())
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
				"attribute": f["attribute"],
				"amount":    f["amount"],
			})
		}
		sourceIndices = append(sourceIndices, map[string]interface{}{
			"index_name":     sourceIndex.IndexName,
			"analytics_tags": sourceIndex.AnalyticsTags,
			"facets":         facets,
			"min_hits":       sourceIndex.MinHits,
			"min_letters":    sourceIndex.MinLetters,
			"generate":       sourceIndex.Generate,
			"external":       sourceIndex.External,
		})
	}

	values := map[string]interface{}{
		"index_name":     querySuggestionsIndexConfig.IndexName,
		"source_indices": sourceIndices,
		"languages":      querySuggestionsIndexConfig.Languages.StringArray,
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
		indexConfig.Languages = suggestions.BoolOrStringArray{StringArray: castStringSet(v)}
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

func newSuggestionsClient(d *schema.ResourceData, m interface{}) *suggestions.Client {
	apiClient := m.(*apiClient)
	r := region.Region(d.Get("region").(string))
	return apiClient.newSuggestionsClient(r)
}
