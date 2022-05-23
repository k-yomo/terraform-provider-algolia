package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceVirtualIndex() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVirtualIndexCreate,
		ReadContext:          resourceVirtualIndexRead,
		UpdateWithoutTimeout: resourceVirtualIndexUpdate,
		DeleteContext:        resourceVirtualIndexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVirtualIndexStateContext,
		},
		Description: "A configuration for a virtual index.",
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(1 * time.Hour),
		},
		// https://www.algolia.com/doc/api-reference/settings-api-parameters/
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the virtual index. Its name should NOT be surrounded with `virtual()`.",
			},
			"primary_index_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the existing primary index name.",
			},
			"attributes_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "The configuration for attributes.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"searchable_attributes": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "The complete list of attributes used for searching.",
						},
						"attributes_for_faceting": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "The complete list of attributes that will be used for faceting.",
						},
						"unretrievable_attributes": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							Description: "List of attributes that cannot be retrieved at query time.",
						},
						"attributes_to_retrieve": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Optional: true,
							DefaultFunc: func() (interface{}, error) {
								return []string{"*"}, nil
							},
							Description: "List of attributes to be retrieved at query time.",
						},
					},
				},
			},
			"ranking_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "The configuration for ranking.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ranking": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of ranking criteria.",
						},
						"custom_ranking": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
							Description: "List of attributes for custom ranking criterion.",
						},
						"relevancy_strictness": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      100,
							ValidateFunc: validation.IntBetween(0, 100),
							Description:  "Relevancy threshold below which less relevant results aren’t included in the results",
						},
					},
				},
			},
			"faceting_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "The configuration for faceting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_values_per_facet": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      100,
							ValidateFunc: validation.IntAtMost(1000),
							Description:  "Maximum number of facet values to return for each facet during a regular search.",
						},
						"sort_facet_values_by": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "count",
							ValidateFunc: validation.StringInSlice([]string{"alpha", "count"}, false),
							Description:  "Parameter to controls how the facet values are sorted within each faceted attribute.",
						},
					},
				},
			},
			"highlight_and_snippet_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "The configuration for highlight / snippet in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attributes_to_highlight": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							Computed:    true,
							Description: "List of attributes to highlight.",
						},
						"attributes_to_snippet": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							Computed:    true,
							Description: "List of attributes to snippet, with an optional maximum number of words to snippet.",
						},
						"highlight_pre_tag": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "<em>",
							Description: "The HTML string to insert before the highlighted parts in all highlight and snippet results.",
						},
						"highlight_post_tag": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "</em>",
							Description: "The HTML string to insert after the highlighted parts in all highlight and snippet results.",
						},
						"snippet_ellipsis_text": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "String used as an ellipsis indicator when a snippet is truncated.",
						},
						"restrict_highlight_and_snippet_arrays": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Restrict highlighting and snippeting to items that matched the query.",
						},
					},
				},
			},
			"pagination_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "The configuration for pagination in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hits_per_page": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      200,
							ValidateFunc: validation.IntAtMost(1000),
							Description:  "The number of hits per page.",
						},
						"pagination_limited_to": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1000,
							Description: "The maximum number of hits accessible via pagination",
						},
					},
				},
			},
			"typos_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "The configuration for typos in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_word_size_for_1_typo": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      4,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "Minimum number of characters a word in the query string must contain to accept matches with 1 typo.",
						},
						"min_word_size_for_2_typos": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      8,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "Minimum number of characters a word in the query string must contain to accept matches with 2 typos.",
						},
						"typo_tolerance": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "true",
							ValidateFunc: validation.StringInSlice([]string{"true", "false", "min", "strict"}, false),
							Description:  "Whether typo tolerance is enabled and how it is applied",
						},
						"allow_typos_on_numeric_tokens": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether to allow typos on numbers (“numeric tokens”) in the query str",
						},
						"disable_typo_tolerance_on_attributes": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of attributes on which you want to disable typo tolerance.",
						},
						"disable_typo_tolerance_on_words": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of words on which typo tolerance will be disabled.",
						},
						"separators_to_index": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Separators (punctuation characters) to index. By default, separators are not indexed.",
						},
					},
				},
			},
			"languages_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "The configuration for languages in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ignore_plurals": {
							Type:          schema.TypeBool,
							Optional:      true,
							Default:       false,
							ConflictsWith: []string{"languages_config.0.ignore_plurals_for"},
							Description:   "Whether to treat singular, plurals, and other forms of declensions as matching terms.",
						},
						"ignore_plurals_for": {
							Type:          schema.TypeSet,
							Elem:          &schema.Schema{Type: schema.TypeString},
							Set:           schema.HashString,
							Optional:      true,
							ConflictsWith: []string{"languages_config.0.ignore_plurals"},
							Description: `Whether to treat singular, plurals, and other forms of declensions as matching terms in target languages.
List of supported languages are listed on http://nhttps//www.algolia.com/doc/api-reference/api-parameters/ignorePlurals/#usage-notes`,
						},
						"attributes_to_transliterate": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							Computed:    true,
							Description: "List of attributes to apply transliteration",
						},
						"remove_stop_words": {
							Type:          schema.TypeBool,
							Optional:      true,
							Default:       false,
							ConflictsWith: []string{"languages_config.0.remove_stop_words_for"},
							Description:   "Whether to removes stop (common) words from the query before executing it.",
						},
						"remove_stop_words_for": {
							Type:          schema.TypeSet,
							Elem:          &schema.Schema{Type: schema.TypeString},
							Set:           schema.HashString,
							Optional:      true,
							ConflictsWith: []string{"languages_config.0.remove_stop_words"},
							Description:   "List of languages to removes stop (common) words from the query before executing it.",
						},
						"camel_case_attributes": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of attributes on which to do a decomposition of camel case words.",
						},
						"decompounded_attributes": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of attributes to apply word segmentation, also known as decompounding.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"language": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"attributes": {
										Type:     schema.TypeSet,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
										Computed: true,
									},
								},
							},
						},
						"keep_diacritics_on_characters": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "List of characters that the engine shouldn’t automatically normalize.",
						},
						"custom_normalization": {
							Type:        schema.TypeMap,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "Custom normalization which overrides the engine’s default normalization",
						},
						"query_languages": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							Description: "List of languages to be used by language-specific settings and functionalities such as ignorePlurals, removeStopWords, and CJK word-detection.",
						},
						"index_languages": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of languages at the index level for language-specific processing such as tokenization and normalization.",
						},
						"decompound_query": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether to split compound words into their composing atoms in the query.",
						},
					},
				},
			},
			"enable_rules": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether Rules should be globally enabled.",
			},
			"enable_personalization": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to enable the Personalization feature.",
			},
			"query_strategy_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "The configuration for query strategy in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "prefixLast",
							ValidateFunc: validation.StringInSlice([]string{"prefixLast", "prefixAll", "prefixNone"}, false),
							Description:  "Query type to control if and how query words are interpreted as prefixes.",
						},
						"remove_words_if_no_results": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "none",
							ValidateFunc: validation.StringInSlice([]string{"none", "lastWords", "firstWords", "allOptional"}, false),
							Description:  "Strategy to remove words from the query when it doesn’t match any hits.",
						},
						"advanced_syntax": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to enable the advanced query syntax.",
						},
						"optional_words": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "A list of words that should be considered as optional when found in the query.",
						},
						"disable_prefix_on_attributes": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of attributes on which you want to disable prefix matching.",
						},
						"disable_exact_on_attributes": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of attributes on which you want to disable the exact ranking criterion.",
						},
						"exact_on_single_word_query": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "attribute",
							ValidateFunc: validation.StringInSlice([]string{"none", "lastWords", "firstWords", "allOptional"}, false),
							Description:  "Controls how the exact ranking criterion is computed when the query contains only one word.",
						},
						"alternatives_as_exact": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Optional: true,
							DefaultFunc: func() (interface{}, error) {
								return []string{"ignorePlurals", "singleWordSynonym"}, nil
							},
							Description: "List of alternatives that should be considered an exact match by the exact ranking criterion.",
						},
						"advanced_syntax_features": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Optional: true,
							DefaultFunc: func() (interface{}, error) {
								return []string{"exactPhrase", "excludeWords"}, nil
							},
							Description: "Advanced syntax features to be activated when ‘advancedSyntax’ is enabled",
						},
					},
				},
			},
			"performance_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The configuration for performance in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"numeric_attributes_for_filtering": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of numeric attributes that can be used as numerical filters.",
						},
						"allow_compression_of_integer_array": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to enable compression of large integer arrays.",
						},
					},
				},
			},
			"advanced_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "The configuration for advanced features in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_for_distinct": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the de-duplication attribute to be used with the `distinct` feature.",
						},
						"distinct": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
							Description: `Whether to enable de-duplication or grouping of results.
- When set to ` + "`0`" + `, you disable de-duplication and grouping.
- When set to ` + "`1`" + `, you enable **de-duplication**, in which only the most relevant result is returned for all records that have the same value in the distinct attribute. This is similar to the SQL ` + "`distinct`" + ` keyword.
if ` + "`distinct`" + ` is set to 1 (de-duplication):
- When set to ` + "`N (where N > 1)`" + `, you enable grouping, in which most N hits will be returned with the same value for the distinct attribute.
then the N most relevant episodes for every show are kept, with similar consequences.
`,
						},
						"replace_synonyms_in_highlight": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to highlight and snippet the original word that matches the synonym or the synonym itself.",
						},
						"min_proximity": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1,
							Description: "Precision of the `proximity` ranking criterion.",
						},
						"response_fields": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Optional: true,
							DefaultFunc: func() (interface{}, error) {
								return []string{"*"}, nil
							},
							Description: `The fields the response will contain. Applies to search and browse queries.
This parameter is mainly intended to **limit the response size.** For example, in complex queries, echoing of request parameters in the response’s params field can be undesirable.`,
						},
						"max_facet_hits": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     10,
							Description: "Maximum number of facet hits to return during a search for facet values.",
						},
						"attribute_criteria_computed_by_min_proximity": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "When attribute is ranked above proximity in your ranking formula, proximity is used to select which searchable attribute is matched in the **attribute ranking stage**.",
						},
					},
				},
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to allow Terraform to destroy the index.  Unless this field is set to false in Terraform state, a terraform destroy or terraform apply command that deletes the instance will fail.",
			},
		},
	}
}

func resourceVirtualIndexCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	indexName := d.Get("name").(string)

	primaryIndexName := d.Get("primary_index_name").(string)
	primaryIndex := apiClient.searchClient.InitIndex(primaryIndexName)
	primaryIndexSettings, err := primaryIndex.GetSettings(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	replicas := primaryIndexSettings.Replicas.Get()
	if !algoliautil.IndexExistsInReplicas(replicas, indexName, true) {
		// Modifying the primary's replica setting on primary can cause problems if other replicas
		// are modifying it at the same time. Lock the primary until we're done in order to prevent that.
		mutexKV.Lock(algoliaIndexMutexKey(apiClient.appID, primaryIndexName))
		defer mutexKV.Unlock(algoliaIndexMutexKey(apiClient.appID, primaryIndexName))

		newReplicas := append(primaryIndexSettings.Replicas.Get(), fmt.Sprintf("virtual(%s)", indexName))
		res, err := primaryIndex.SetSettings(search.Settings{
			Replicas: opt.Replicas(newReplicas...),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if err := res.Wait(); err != nil {
			return diag.FromErr(err)
		}
	}

	index := apiClient.searchClient.InitIndex(indexName)
	res, err := index.SetSettings(mapToVirtualIndexSettings(d))
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(indexName)

	return resourceVirtualIndexRead(ctx, d, m)
}

func resourceVirtualIndexRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := refreshVirtualIndexState(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVirtualIndexUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	index := apiClient.searchClient.InitIndex(d.Id())
	res, err := index.SetSettings(mapToVirtualIndexSettings(d))
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	return resourceVirtualIndexRead(ctx, d, m)
}

func resourceVirtualIndexDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.Get("deletion_protection").(bool) {
		return diag.Errorf("cannot destroy index without setting deletion_protection=false and running `terraform apply`")
	}

	apiClient := m.(*apiClient)
	indexName := d.Id()

	primaryIndexName := d.Get("primary_index_name").(string)
	// Modifying the primary's replica setting on primary can cause problems if other replicas
	// are modifying it at the same time. Lock the primary until we're done in order to prevent that.
	mutexKV.Lock(algoliaIndexMutexKey(apiClient.appID, primaryIndexName))
	defer mutexKV.Unlock(algoliaIndexMutexKey(apiClient.appID, primaryIndexName))

	primaryIndex := apiClient.searchClient.InitIndex(primaryIndexName)
	primaryIndexSettings, err := primaryIndex.GetSettings(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if algoliautil.IndexExistsInReplicas(primaryIndexSettings.Replicas.Get(), indexName, true) {
		newReplicas := algoliautil.RemoveIndexFromReplicas(primaryIndexSettings.Replicas.Get(), indexName, true)
		updateReplicasRes, err := primaryIndex.SetSettings(search.Settings{
			Replicas: opt.Replicas(newReplicas...),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if err := updateReplicasRes.Wait(); err != nil {
			return diag.FromErr(err)
		}
	}
	index := apiClient.searchClient.InitIndex(indexName)
	deleteIndexRes, err := index.Delete(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := deleteIndexRes.Wait(ctx); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVirtualIndexStateContext(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := refreshVirtualIndexState(ctx, d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func refreshVirtualIndexState(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*apiClient)

	index := apiClient.searchClient.InitIndex(d.Id())
	settings, err := index.GetSettings(ctx)
	if err != nil {
		if algoliautil.IsNotFoundError(err) {
			log.Printf("[WARN] virtual index (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	var typoTolerance string
	if b, s := settings.TypoTolerance.Get(); s != "" {
		typoTolerance = s
	} else {
		typoTolerance = strconv.FormatBool(b)
	}

	var ignorePlurals, ignorePluralsFor interface{}
	if ignore, languages := settings.IgnorePlurals.Get(); len(languages) > 0 {
		ignorePluralsFor = languages
	} else {
		ignorePlurals = ignore
	}

	var removeStopWords, removeStopWordsFor interface{}
	if remove, languages := settings.RemoveStopWords.Get(); len(languages) > 0 {
		removeStopWordsFor = languages
	} else {
		removeStopWords = remove
	}

	var decompoundedAttributes []interface{}
	for language, attributes := range settings.DecompoundedAttributes.Get() {
		decompoundedAttributes = append(decompoundedAttributes, map[string]interface{}{
			"language":   language,
			"attributes": attributes,
		})
	}

	values := map[string]interface{}{
		"name":               d.Id(),
		"primary_index_name": settings.Primary.Get(),
		"attributes_config": []interface{}{map[string]interface{}{
			"searchable_attributes":    settings.SearchableAttributes.Get(),
			"attributes_for_faceting":  settings.AttributesForFaceting.Get(),
			"unretrievable_attributes": settings.UnretrievableAttributes.Get(),
			"attributes_to_retrieve":   settings.AttributesToRetrieve.Get(),
		}},
		"ranking_config": []interface{}{map[string]interface{}{
			"ranking":              settings.Ranking.Get(),
			"custom_ranking":       settings.CustomRanking.Get(),
			"relevancy_strictness": settings.RelevancyStrictness.Get(),
		}},
		"faceting_config": []interface{}{map[string]interface{}{
			"max_values_per_facet": settings.MaxValuesPerFacet.Get(),
			"sort_facet_values_by": settings.SortFacetValuesBy.Get(),
		}},
		"highlight_and_snippet_config": []interface{}{map[string]interface{}{
			"attributes_to_highlight":               settings.AttributesToHighlight.Get(),
			"attributes_to_snippet":                 settings.AttributesToSnippet.Get(),
			"highlight_pre_tag":                     settings.HighlightPreTag.Get(),
			"highlight_post_tag":                    settings.HighlightPostTag.Get(),
			"snippet_ellipsis_text":                 settings.SnippetEllipsisText.Get(),
			"restrict_highlight_and_snippet_arrays": settings.RestrictHighlightAndSnippetArrays.Get(),
		}},
		"pagination_config": []interface{}{map[string]interface{}{
			"hits_per_page":         settings.HitsPerPage.Get(),
			"pagination_limited_to": settings.PaginationLimitedTo.Get(),
		}},
		"typos_config": []interface{}{map[string]interface{}{
			"min_word_size_for_1_typo":             settings.MinWordSizefor1Typo.Get(),
			"min_word_size_for_2_typos":            settings.MinWordSizefor2Typos.Get(),
			"typo_tolerance":                       typoTolerance,
			"allow_typos_on_numeric_tokens":        settings.AllowTyposOnNumericTokens.Get(),
			"disable_typo_tolerance_on_attributes": settings.DisableTypoToleranceOnAttributes.Get(),
			"disable_typo_tolerance_on_words":      settings.DisableTypoToleranceOnWords.Get(),
			"separators_to_index":                  settings.SeparatorsToIndex.Get(),
		}},
		"languages_config": []interface{}{map[string]interface{}{
			"ignore_plurals":                ignorePlurals,
			"ignore_plurals_for":            ignorePluralsFor,
			"attributes_to_transliterate":   settings.AttributesToTransliterate.Get(),
			"remove_stop_words":             removeStopWords,
			"remove_stop_words_for":         removeStopWordsFor,
			"camel_case_attributes":         settings.CamelCaseAttributes.Get(),
			"decompounded_attributes":       decompoundedAttributes,
			"keep_diacritics_on_characters": settings.KeepDiacriticsOnCharacters.Get(),
			"custom_normalization":          settings.CustomNormalization.Get()["default"],
			"query_languages":               settings.QueryLanguages.Get(),
			"index_languages":               settings.IndexLanguages.Get(),
			"decompound_query":              settings.DecompoundQuery.Get(),
		}},
		"enable_rules":           settings.EnableRules.Get(),
		"enable_personalization": settings.EnablePersonalization.Get(),
		"query_strategy_config": []interface{}{map[string]interface{}{
			"query_type":                   settings.QueryType.Get(),
			"remove_words_if_no_results":   settings.RemoveWordsIfNoResults.Get(),
			"advanced_syntax":              settings.AdvancedSyntax.Get(),
			"optional_words":               settings.OptionalWords.Get(),
			"disable_prefix_on_attributes": settings.DisablePrefixOnAttributes.Get(),
			"disable_exact_on_attributes":  settings.DisableExactOnAttributes.Get(),
			"exact_on_single_word_query":   settings.ExactOnSingleWordQuery.Get(),
			"alternatives_as_exact":        settings.AlternativesAsExact.Get(),
			"advanced_syntax_features":     settings.AdvancedSyntaxFeatures.Get(),
		}},
		"performance_config": []interface{}{map[string]interface{}{
			"numeric_attributes_for_filtering":   settings.NumericAttributesForFiltering.Get(),
			"allow_compression_of_integer_array": settings.AllowCompressionOfIntegerArray.Get(),
		}},
		"advanced_config": []interface{}{map[string]interface{}{
			"attribute_for_distinct":        settings.AttributeForDistinct.Get(),
			"distinct":                      func() int { _, i := settings.Distinct.Get(); return i }(),
			"replace_synonyms_in_highlight": settings.ReplaceSynonymsInHighlight.Get(),
			"min_proximity":                 settings.MinProximity.Get(),
			"response_fields":               settings.ResponseFields.Get(),
			"max_facet_hits":                settings.MaxFacetHits.Get(),
			"attribute_criteria_computed_by_min_proximity": settings.AttributeCriteriaComputedByMinProximity.Get(),
		}},
	}
	if err := setValues(d, values); err != nil {
		return err
	}

	return nil
}

func mapToVirtualIndexSettings(d *schema.ResourceData) search.Settings {
	settings := search.Settings{}
	if v, ok := d.GetOk("attributes_config"); ok {
		unmarshalAttributesConfig(v, &settings, true)
	}
	if v, ok := d.GetOk("ranking_config"); ok {
		unmarshalRankingConfig(v, &settings, true)
	}
	if v, ok := d.GetOk("faceting_config"); ok {
		unmarshalFacetingConfig(v, &settings)
	}
	if v, ok := d.GetOk("highlight_and_snippet_config"); ok {
		unmarshalHighlightAndSnippetConfig(v, &settings)
	}
	if v, ok := d.GetOk("pagination_config"); ok {
		unmarshalPaginationConfig(v, &settings)
	}
	if v, ok := d.GetOk("typos_config"); ok {
		unmarshalTyposConfig(v, &settings, true)
	}
	if v, ok := d.GetOk("languages_config"); ok {
		unmarshalLanguagesConfig(v, &settings, true)
	}
	if v, ok := d.GetOk("enable_rules"); ok {
		settings.EnableRules = opt.EnableRules(v.(bool))
	}
	if v, ok := d.GetOk("enable_personalization"); ok {
		settings.EnablePersonalization = opt.EnablePersonalization(v.(bool))
	}
	if v, ok := d.GetOk("query_strategy_config"); ok {
		unmarshalQueryStrategyConfig(v, &settings, true)
	}
	if v, ok := d.GetOk("performance_config"); ok {
		unmarshalPerformanceConfig(v, &settings, true)
	}
	if v, ok := d.GetOk("advanced_config"); ok {
		unmarshalAdvancedConfig(v, &settings, true)
	}

	return settings
}
