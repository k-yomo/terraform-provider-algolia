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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
)

func resourceIndex() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIndexCreate,
		ReadContext:          resourceIndexRead,
		UpdateWithoutTimeout: resourceIndexUpdate,
		DeleteContext:        resourceIndexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceIndexStateContext,
		},
		Description: "A configuration for an index.",
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(1 * time.Hour),
		},
		// https://www.algolia.com/doc/api-reference/settings-api-parameters/
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the index / replica index. For creating virtual replica, use `algolia_virtual_index` resource instead.",
			},
			"primary_index_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of the existing primary index name. This field is used to create a replica index.",
			},
			"virtual": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "**Deprecated:** Use `algolia_virtual_index` resource instead. Whether the index is virtual index. If true, applying the params listed in the [doc](https://www.algolia.com/doc/guides/managing-results/refine-results/sorting/in-depth/replicas/#unsupported-parameters) will be ignored.",
				Deprecated:  "Use `algolia_virtual_index` resource instead",
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
							Optional:    true,
							Description: "The complete list of attributes used for searching.",
						},
						"attributes_for_faceting": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
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
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							DefaultFunc: func() (interface{}, error) {
								return []string{"typo", "geo", "words", "filters", "proximity", "attribute", "exact", "custom"}, nil
							},
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
							Optional:    true,
							Description: "List of attributes on which you want to disable typo tolerance.",
						},
						"disable_typo_tolerance_on_words": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
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
							Optional:    true,
							Description: "List of attributes on which to do a decomposition of camel case words.",
						},
						"decompounded_attributes": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of attributes to apply word segmentation, also known as decompounding.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"language": {
										Type:     schema.TypeString,
										Required: true,
									},
									"attributes": {
										Type:     schema.TypeSet,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
										Required: true,
									},
								},
							},
						},
						"keep_diacritics_on_characters": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "List of characters that the engine shouldn’t automatically normalize.",
						},
						"custom_normalization": {
							Type:        schema.TypeMap,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
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
							Optional:    true,
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
							Optional:    true,
							Description: "A list of words that should be considered as optional when found in the query.",
						},
						"disable_prefix_on_attributes": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							Description: "List of attributes on which you want to disable prefix matching.",
						},
						"disable_exact_on_attributes": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
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
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "The configuration for performance in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"numeric_attributes_for_filtering": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							Description: "List of numeric attributes that can be used as numerical filters.",
						},
						"allow_compression_of_integer_array": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
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
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"advanced_config.0.distinct"},
							Description:  "Name of the de-duplication attribute to be used with the `distinct` feature.",
						},
						"distinct": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
							// TODO: Uncomment once virtual index is migrated to `algolia_virtual_index` and `virtual` field is removed.
							// `distinct` requires `attribute_for_distinct` but disable the constraint here for virtual index.
							// since `attribute_for_distinct` can't be set in virtual index.
							// RequiredWith: []string{"advanced_config.0.attribute_for_distinct"},
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

func resourceIndexCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	indexName := d.Get("name").(string)

	if v, ok := d.GetOk("primary_index_name"); ok {
		primaryIndexName := v.(string)
		// Modifying the primary's replica setting on primary can cause problems if other replicas
		// are modifying it at the same time. Lock the primary until we're done in order to prevent that.
		mutexKV.Lock(algoliaIndexMutexKey(apiClient.appID, primaryIndexName))
		defer mutexKV.Unlock(algoliaIndexMutexKey(apiClient.appID, primaryIndexName))

		primaryIndex := apiClient.searchClient.InitIndex(primaryIndexName)
		primaryIndexSettings, err := primaryIndex.GetSettings(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		if !algoliautil.IndexExistsInReplicas(primaryIndexSettings.Replicas.Get(), indexName, false) {
			newReplicas := append(primaryIndexSettings.Replicas.Get(), indexName)
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
	}

	index := apiClient.searchClient.InitIndex(indexName)
	res, err := index.SetSettings(mapToIndexSettings(d))
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(indexName)

	return resourceIndexRead(ctx, d, m)
}

func resourceIndexRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := refreshIndexState(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceIndexUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	index := apiClient.searchClient.InitIndex(d.Id())
	res, err := index.SetSettings(mapToIndexSettings(d))
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	return resourceIndexRead(ctx, d, m)
}

func resourceIndexDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.Get("deletion_protection").(bool) {
		return diag.Errorf("cannot destroy index without setting deletion_protection=false and running `terraform apply`")
	}

	apiClient := m.(*apiClient)
	indexName := d.Id()

	if v, ok := d.GetOk("primary_index_name"); ok {
		primaryIndexName := v.(string)
		// Modifying the primary's replica setting on primary can cause problems if other replicas
		// are modifying it at the same time. Lock the primary until we're done in order to prevent that.
		mutexKV.Lock(algoliaIndexMutexKey(apiClient.appID, primaryIndexName))
		defer mutexKV.Unlock(algoliaIndexMutexKey(apiClient.appID, primaryIndexName))

		primaryIndex := apiClient.searchClient.InitIndex(primaryIndexName)
		primaryIndexSettings, err := primaryIndex.GetSettings(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		if algoliautil.IndexExistsInReplicas(primaryIndexSettings.Replicas.Get(), indexName, false) {
			newReplicas := algoliautil.RemoveIndexFromReplicas(primaryIndexSettings.Replicas.Get(), indexName, false)
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

func resourceIndexStateContext(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := refreshIndexState(ctx, d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func refreshIndexState(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*apiClient)

	index := apiClient.searchClient.InitIndex(d.Id())
	settings, err := index.GetSettings(ctx)
	if err != nil {
		if algoliautil.IsNotFoundError(err) {
			log.Printf("[WARN] index (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	if err := setValues(d, mapToIndexResourceValues(d, settings)); err != nil {
		return err
	}

	return nil
}

func mapToIndexResourceValues(d *schema.ResourceData, settings search.Settings) map[string]interface{} {
	isVirtualIndex := d.Get("virtual").(bool)

	return map[string]interface{}{
		"name":               d.Id(),
		"primary_index_name": settings.Primary.Get(),
		"virtual":            isVirtualIndex,
		"attributes_config":  marshalAttributesConfig(settings, isVirtualIndex),
		"ranking_config":     marshalRankingConfig(settings, isVirtualIndex),
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
		"typos_config":           marshalTyposConfig(settings, isVirtualIndex),
		"languages_config":       marshalLanguageConfig(settings, isVirtualIndex),
		"enable_rules":           settings.EnableRules.Get(),
		"enable_personalization": settings.EnablePersonalization.Get(),
		"query_strategy_config":  marshalQueryStrategyConfig(settings, isVirtualIndex),
		"performance_config":     marshalPerformanceConfig(settings, isVirtualIndex),
		"advanced_config":        marshalAdvancedConfig(settings, isVirtualIndex),
	}
}

func marshalAttributesConfig(settings search.Settings, isVirtualIndex bool) []interface{} {
	attributesConfig := map[string]interface{}{
		"unretrievable_attributes": settings.UnretrievableAttributes.Get(),
		"attributes_to_retrieve":   settings.AttributesToRetrieve.Get(),
	}
	if !isVirtualIndex {
		attributesConfig["searchable_attributes"] = settings.SearchableAttributes.Get()
		attributesConfig["attributes_for_faceting"] = settings.AttributesForFaceting.Get()
	}

	return []interface{}{attributesConfig}
}

func marshalRankingConfig(settings search.Settings, isVirtualIndex bool) []interface{} {
	rankingConfig := map[string]interface{}{
		"custom_ranking":       settings.CustomRanking.Get(),
		"relevancy_strictness": settings.RelevancyStrictness.Get(),
	}
	if !isVirtualIndex {
		rankingConfig["ranking"] = settings.Ranking.Get()
	}

	return []interface{}{rankingConfig}
}

func marshalTyposConfig(settings search.Settings, isVirtualIndex bool) []interface{} {
	var typoTolerance string
	if b, s := settings.TypoTolerance.Get(); s != "" {
		typoTolerance = s
	} else {
		typoTolerance = strconv.FormatBool(b)
	}

	typosConfig := map[string]interface{}{
		"min_word_size_for_1_typo":      settings.MinWordSizefor1Typo.Get(),
		"min_word_size_for_2_typos":     settings.MinWordSizefor2Typos.Get(),
		"typo_tolerance":                typoTolerance,
		"allow_typos_on_numeric_tokens": settings.AllowTyposOnNumericTokens.Get(),
		"separators_to_index":           settings.SeparatorsToIndex.Get(),
	}
	if !isVirtualIndex {
		typosConfig["disable_typo_tolerance_on_attributes"] = settings.DisableTypoToleranceOnAttributes.Get()
		typosConfig["disable_typo_tolerance_on_words"] = settings.DisableTypoToleranceOnWords.Get()
	}

	return []interface{}{typosConfig}
}

func marshalLanguageConfig(settings search.Settings, isVirtualIndex bool) []interface{} {
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

	languageConfig := map[string]interface{}{
		"ignore_plurals":              ignorePlurals,
		"ignore_plurals_for":          ignorePluralsFor,
		"attributes_to_transliterate": settings.AttributesToTransliterate.Get(),
		"remove_stop_words":           removeStopWords,
		"remove_stop_words_for":       removeStopWordsFor,
		"query_languages":             settings.QueryLanguages.Get(),
		"decompound_query":            settings.DecompoundQuery.Get(),
	}
	if !isVirtualIndex {
		languageConfig["camel_case_attributes"] = settings.CamelCaseAttributes.Get()
		languageConfig["custom_normalization"] = settings.CustomNormalization.Get()["default"]
		languageConfig["decompounded_attributes"] = decompoundedAttributes
		languageConfig["keep_diacritics_on_characters"] = settings.KeepDiacriticsOnCharacters.Get()
		languageConfig["index_languages"] = settings.IndexLanguages.Get()
	}

	return []interface{}{languageConfig}
}

func marshalQueryStrategyConfig(settings search.Settings, isVirtualIndex bool) []interface{} {
	queryStrategyConfig := map[string]interface{}{
		"query_type":                 settings.QueryType.Get(),
		"remove_words_if_no_results": settings.RemoveWordsIfNoResults.Get(),
		"advanced_syntax":            settings.AdvancedSyntax.Get(),
		"exact_on_single_word_query": settings.ExactOnSingleWordQuery.Get(),
		"alternatives_as_exact":      settings.AlternativesAsExact.Get(),
		"advanced_syntax_features":   settings.AdvancedSyntaxFeatures.Get(),
	}
	if !isVirtualIndex {
		queryStrategyConfig["optional_words"] = settings.OptionalWords.Get()
		queryStrategyConfig["disable_prefix_on_attributes"] = settings.DisablePrefixOnAttributes.Get()
		queryStrategyConfig["disable_exact_on_attributes"] = settings.DisableExactOnAttributes.Get()
	}

	return []interface{}{queryStrategyConfig}
}

func marshalPerformanceConfig(settings search.Settings, isVirtualIndex bool) []interface{} {
	if isVirtualIndex {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"numeric_attributes_for_filtering":   settings.NumericAttributesForFiltering.Get(),
		"allow_compression_of_integer_array": settings.AllowCompressionOfIntegerArray.Get(),
	}}
}

func marshalAdvancedConfig(settings search.Settings, isVirtualIndex bool) []interface{} {
	advancedConfig := map[string]interface{}{
		"distinct":                      func() int { _, i := settings.Distinct.Get(); return i }(),
		"replace_synonyms_in_highlight": settings.ReplaceSynonymsInHighlight.Get(),
		"min_proximity":                 settings.MinProximity.Get(),
		"response_fields":               settings.ResponseFields.Get(),
		"max_facet_hits":                settings.MaxFacetHits.Get(),
		"attribute_criteria_computed_by_min_proximity": settings.AttributeCriteriaComputedByMinProximity.Get(),
	}
	if !isVirtualIndex {
		advancedConfig["attribute_for_distinct"] = settings.AttributeForDistinct.Get()
	}

	return []interface{}{advancedConfig}
}

func mapToIndexSettings(d *schema.ResourceData) search.Settings {
	isVirtualIndex := d.Get("virtual").(bool)

	settings := search.Settings{}
	if v, ok := d.GetOk("attributes_config"); ok {
		unmarshalAttributesConfig(v, &settings, isVirtualIndex)
	}
	if v, ok := d.GetOk("ranking_config"); ok {
		unmarshalRankingConfig(v, &settings, isVirtualIndex)
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
		unmarshalTyposConfig(v, &settings, isVirtualIndex)
	}
	if v, ok := d.GetOk("languages_config"); ok {
		unmarshalLanguagesConfig(v, &settings, isVirtualIndex)
	}
	if v, ok := d.GetOk("enable_rules"); ok {
		settings.EnableRules = opt.EnableRules(v.(bool))
	}
	if v, ok := d.GetOk("enable_personalization"); ok {
		settings.EnablePersonalization = opt.EnablePersonalization(v.(bool))
	}
	if v, ok := d.GetOk("query_strategy_config"); ok {
		unmarshalQueryStrategyConfig(v, &settings, isVirtualIndex)
	}
	if v, ok := d.GetOk("performance_config"); ok {
		unmarshalPerformanceConfig(v, &settings, isVirtualIndex)
	}
	if v, ok := d.GetOk("advanced_config"); ok {
		unmarshalAdvancedConfig(v, &settings, isVirtualIndex)
	}

	return settings
}

func unmarshalAttributesConfig(configured interface{}, settings *search.Settings, isVirtualIndex bool) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}
	config := l[0].(map[string]interface{})
	settings.UnretrievableAttributes = opt.UnretrievableAttributes(castStringSet(config["unretrievable_attributes"])...)
	settings.AttributesToRetrieve = opt.AttributesToRetrieve(castStringSet(config["attributes_to_retrieve"])...)
	if !isVirtualIndex {
		settings.SearchableAttributes = opt.SearchableAttributes(castStringList(config["searchable_attributes"])...)
		settings.AttributesForFaceting = opt.AttributesForFaceting(castStringSet(config["attributes_for_faceting"])...)
	}
}

func unmarshalRankingConfig(configured interface{}, settings *search.Settings, isVirtualIndex bool) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}
	config := l[0].(map[string]interface{})
	settings.CustomRanking = opt.CustomRanking(castStringList(config["custom_ranking"])...)
	settings.RelevancyStrictness = opt.RelevancyStrictness(config["relevancy_strictness"].(int))
	if !isVirtualIndex {
		settings.Ranking = opt.Ranking(castStringList(config["ranking"])...)
	}
}

func unmarshalFacetingConfig(configured interface{}, settings *search.Settings) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}

	config := l[0].(map[string]interface{})

	if v, ok := config["max_values_per_facet"]; ok {
		settings.MaxValuesPerFacet = opt.MaxValuesPerFacet(v.(int))
	}
	if v, ok := config["sort_facet_values_by"]; ok {
		settings.SortFacetValuesBy = opt.SortFacetValuesBy(v.(string))
	}
}

func unmarshalHighlightAndSnippetConfig(configured interface{}, settings *search.Settings) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}

	config := l[0].(map[string]interface{})

	if v, ok := config["attributes_to_highlight"]; ok {
		settings.AttributesToHighlight = opt.AttributesToHighlight(castStringSet(v)...)
	}
	if v, ok := config["attributes_to_snippet"]; ok {
		settings.AttributesToSnippet = opt.AttributesToSnippet(castStringSet(v)...)
	}
	if v, ok := config["highlight_pre_tag"]; ok {
		settings.HighlightPreTag = opt.HighlightPreTag(v.(string))
	}
	if v, ok := config["highlight_post_tag"]; ok {
		settings.HighlightPostTag = opt.HighlightPostTag(v.(string))
	}
	if v, ok := config["snippet_ellipsis_text"]; ok {
		settings.SnippetEllipsisText = opt.SnippetEllipsisText(v.(string))
	}
	if v, ok := config["restrict_highlight_and_snippet_arrays"]; ok {
		settings.RestrictHighlightAndSnippetArrays = opt.RestrictHighlightAndSnippetArrays(v.(bool))
	}
}

func unmarshalPaginationConfig(configured interface{}, settings *search.Settings) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}
	config := l[0].(map[string]interface{})

	if v, ok := config["hits_per_page"]; ok {
		settings.HitsPerPage = opt.HitsPerPage(v.(int))
	}
	if v, ok := config["pagination_limited_to"]; ok {
		settings.PaginationLimitedTo = opt.PaginationLimitedTo(v.(int))
	}
}

func unmarshalTyposConfig(configured interface{}, settings *search.Settings, isVirtualIndex bool) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}

	config := l[0].(map[string]interface{})

	if v, ok := config["min_word_size_for_1_typo"]; ok {
		settings.MinWordSizefor1Typo = opt.MinWordSizefor1Typo(v.(int))
	}
	if v, ok := config["min_word_size_for_2_typos"]; ok {
		settings.MinWordSizefor2Typos = opt.MinWordSizefor2Typos(v.(int))
	}
	if v, ok := config["typo_tolerance"]; ok {
		typoTolerance := v.(string)
		if b, err := strconv.ParseBool(typoTolerance); err == nil {
			settings.TypoTolerance = opt.TypoTolerance(b)
		} else {
			if typoTolerance == "min" {
				settings.TypoTolerance = opt.TypoToleranceMin()
			} else {
				settings.TypoTolerance = opt.TypoToleranceStrict()
			}
		}
	}
	if v, ok := config["allow_typos_on_numeric_tokens"]; ok {
		settings.AllowTyposOnNumericTokens = opt.AllowTyposOnNumericTokens(v.(bool))
	}

	if !isVirtualIndex {
		if v, ok := config["disable_typo_tolerance_on_attributes"]; ok {
			settings.DisableTypoToleranceOnAttributes = opt.DisableTypoToleranceOnAttributes(castStringList(v)...)
		}
		if v, ok := config["disable_typo_tolerance_on_words"]; ok {
			settings.DisableTypoToleranceOnWords = opt.DisableTypoToleranceOnWords(castStringList(v)...)
		}
		if v, ok := config["separators_to_index"]; ok {
			settings.SeparatorsToIndex = opt.SeparatorsToIndex(v.(string))
		}
	}
}

func unmarshalLanguagesConfig(configured interface{}, settings *search.Settings, isVirtualIndex bool) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}

	config := l[0].(map[string]interface{})

	if v, ok := config["ignore_plurals"]; ok {
		settings.IgnorePlurals = opt.IgnorePlurals(v.(bool))
	}
	if v, ok := config["ignore_plurals_for"]; ok {
		set := castStringSet(v)
		if len(set) > 0 {
			settings.IgnorePlurals = opt.IgnorePluralsFor(set...)
		}
	}
	if v, ok := config["remove_stop_words"]; ok {
		settings.RemoveStopWords = opt.RemoveStopWords(v.(bool))
	}
	if v, ok := config["remove_stop_words_for"]; ok {
		set := castStringSet(v)
		if len(set) > 0 {
			settings.RemoveStopWords = opt.RemoveStopWordsFor(set...)
		}
	}
	if v, ok := config["decompounded_attributes"]; ok {
		unmarshalLanguagesConfigDecompoundedAttributes(v, settings)
	}
	if v, ok := config["query_languages"]; ok {
		settings.QueryLanguages = opt.QueryLanguages(castStringSet(v)...)
	}
	if v, ok := config["decompound_query"]; ok {
		settings.DecompoundQuery = opt.DecompoundQuery(v.(bool))
	}
	if !isVirtualIndex {
		if v, ok := config["attributes_to_transliterate"]; ok {
			settings.AttributesToTransliterate = opt.AttributesToTransliterate(castStringSet(v)...)
		}
		if v, ok := config["camel_case_attributes"]; ok {
			settings.CamelCaseAttributes = opt.CamelCaseAttributes(castStringSet(v)...)
		}
		if v, ok := config["keep_diacritics_on_characters"]; ok {
			settings.KeepDiacriticsOnCharacters = opt.KeepDiacriticsOnCharacters(v.(string))
		}
		if v, ok := config["decompounded_attributes"]; ok {
			unmarshalLanguagesConfigDecompoundedAttributes(v, settings)
		}
		if v, ok := config["custom_normalization"]; ok {
			settings.CustomNormalization = opt.CustomNormalization(map[string]map[string]string{"default": castStringMap(v)})
		}
		if v, ok := config["index_languages"]; ok {
			settings.IndexLanguages = opt.IndexLanguages(castStringSet(v)...)
		}
	}
}

func unmarshalLanguagesConfigDecompoundedAttributes(configured interface{}, settings *search.Settings) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}

	decompoundedAttributesMap := map[string][]string{}
	for _, v := range l {
		decompoundedAttributes := v.(map[string]interface{})
		decompoundedAttributesMap[decompoundedAttributes["language"].(string)] = castStringSet(decompoundedAttributes["attributes"])
	}

	settings.DecompoundedAttributes = opt.DecompoundedAttributes(decompoundedAttributesMap)
}

func unmarshalQueryStrategyConfig(configured interface{}, settings *search.Settings, isVirtualIndex bool) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}

	config := l[0].(map[string]interface{})

	if v, ok := config["query_type"]; ok {
		settings.QueryType = opt.QueryType(v.(string))
	}
	if v, ok := config["remove_words_if_no_results"]; ok {
		settings.RemoveWordsIfNoResults = opt.RemoveWordsIfNoResults(v.(string))
	}
	if v, ok := config["advanced_syntax"]; ok {
		settings.AdvancedSyntax = opt.AdvancedSyntax(v.(bool))
	}
	if v, ok := config["exact_on_single_word_query"]; ok {
		settings.ExactOnSingleWordQuery = opt.ExactOnSingleWordQuery(v.(string))
	}
	if v, ok := config["alternatives_as_exact"]; ok {
		settings.AlternativesAsExact = opt.AlternativesAsExact(castStringSet(v)...)
	}
	if v, ok := config["advanced_syntax_features"]; ok {
		settings.AdvancedSyntaxFeatures = opt.AdvancedSyntaxFeatures(castStringSet(v)...)
	}

	if !isVirtualIndex {
		if v, ok := config["optional_words"]; ok {
			settings.OptionalWords = opt.OptionalWords(castStringSet(v)...)
		}
		if v, ok := config["disable_prefix_on_attributes"]; ok {
			settings.DisablePrefixOnAttributes = opt.DisablePrefixOnAttributes(castStringSet(v)...)
		}
		if v, ok := config["disable_exact_on_attributes"]; ok {
			settings.DisableExactOnAttributes = opt.DisableExactOnAttributes(castStringSet(v)...)
		}
	}
}

func unmarshalPerformanceConfig(configured interface{}, settings *search.Settings, isVirtualIndex bool) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}

	config := l[0].(map[string]interface{})

	if !isVirtualIndex {
		if v, ok := config["numeric_attributes_for_filtering"]; ok {
			settings.NumericAttributesForFiltering = opt.NumericAttributesForFiltering(castStringSet(v)...)
		}
		if v, ok := config["allow_compression_of_integer_array"]; ok {
			settings.AllowCompressionOfIntegerArray = opt.AllowCompressionOfIntegerArray(v.(bool))
		}
	}
}

func unmarshalAdvancedConfig(configured interface{}, settings *search.Settings, isVirtualIndex bool) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}

	config := l[0].(map[string]interface{})

	if v, ok := config["distinct"]; ok {
		settings.Distinct = opt.DistinctOf(v.(int))
	}
	if v, ok := config["replace_synonyms_in_highlight"]; ok {
		settings.ReplaceSynonymsInHighlight = opt.ReplaceSynonymsInHighlight(v.(bool))
	}
	if v, ok := config["min_proximity"]; ok {
		settings.MinProximity = opt.MinProximity(v.(int))
	}
	if v, ok := config["response_fields"]; ok {
		settings.ResponseFields = opt.ResponseFields(castStringSet(v)...)
	}
	if v, ok := config["max_facet_hits"]; ok {
		settings.MaxFacetHits = opt.MaxFacetHits(v.(int))
	}
	if v, ok := config["attribute_criteria_computed_by_min_proximity"]; ok {
		settings.AttributeCriteriaComputedByMinProximity = opt.AttributeCriteriaComputedByMinProximity(v.(bool))
	}

	if !isVirtualIndex {
		if v, ok := config["attribute_for_distinct"]; ok {
			settings.AttributeForDistinct = opt.AttributeForDistinct(v.(string))
		}
	}
}

func algoliaIndexMutexKey(appID string, indexName string) string {
	return fmt.Sprintf("%s-algolia-index-%s", appID, indexName)
}
