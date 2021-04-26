package provider

import (
	"context"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strconv"
)

func resourceIndex() *schema.Resource {
	return &schema.Resource{
		Description:   "A configuration for an index",
		CreateContext: resourceIndexCreate,
		ReadContext:   resourceIndexRead,
		UpdateContext: resourceIndexUpdate,
		DeleteContext: resourceIndexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceIndexStateContext,
		},
		// https://www.algolia.com/doc/api-reference/settings-api-parameters/
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the index.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"attributes_config": {
				Description: "The configuration for attributes.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"searchable_attributes": {
							Description: "The complete list of attributes used for searching.",
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"attributes_for_faceting": {
							Description: "The complete list of attributes that will be used for faceting.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
						"unretrievable_attributes": {
							Description: "List of attributes that cannot be retrieved at query time.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
						"attributes_to_retrieve": {
							Description: "List of attributes to be retrieved at query time.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							DefaultFunc: func() (interface{}, error) {
								return []string{"*"}, nil
							},
						},
					},
				},
			},
			"ranking_config": {
				Description: "The configuration for ranking.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ranking": {
							Description: "List of ranking criteria.",
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
							DefaultFunc: func() (interface{}, error) {
								return []string{"typo", "geo", "words", "filters", "proximity", "attribute", "exact", "custom"}, nil
							},
						},
						"custom_ranking": {
							Description: "List of attributes for custom ranking criterion.",
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						// TODO: Add after the PR below merged.
						//  https://github.com/algolia/algoliasearch-client-go/pull/661
						// "relevancy_strictness": {
						// 	Description:  "Relevancy threshold below which less relevant results aren’t included in the results",
						// 	Type:         schema.TypeInt,
						// 	Optional:     true,
						// 	Default:      100,
						// 	ValidateFunc: validation.IntBetween(0, 100),
						// },
						"replicas": {
							Description: "List of replica names.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
					},
				},
			},
			"faceting_config": {
				Description: "The configuration for faceting.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_values_per_facet": {
							Description:  "Maximum number of facet values to return for each facet during a regular search.",
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      100,
							ValidateFunc: validation.IntAtMost(1000),
						},
						"sort_facet_values_by": {
							Description:  "Parameter to controls how the facet values are sorted within each faceted attribute.",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "count",
							ValidateFunc: validation.StringInSlice([]string{"alpha", "count"}, false),
						},
					},
				},
			},
			"highlight_and_snippet_config": {
				Description: "The configuration for highlight / snippet in index setting.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attributes_to_highlight": {
							Description: "List of attributes to highlight.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							Computed:    true,
						},
						"attributes_to_snippet": {
							Description: "List of attributes to snippet, with an optional maximum number of words to snippet.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
						"highlight_pre_tag": {
							Description: "The HTML string to insert before the highlighted parts in all highlight and snippet results.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "<em>",
						},
						"highlight_post_tag": {
							Description: "The HTML string to insert after the highlighted parts in all highlight and snippet results.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "</em>",
						},
						"snippet_ellipsis_text": {
							Description: "String used as an ellipsis indicator when a snippet is truncated.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "…",
						},
						"restrict_highlight_and_snippet_arrays": {
							Description: "Restrict highlighting and snippeting to items that matched the query.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"pagination_config": {
				Description: "The configuration for pagination in index setting.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hits_per_page": {
							Description:  "The number of hits per page.",
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      200,
							ValidateFunc: validation.IntAtMost(1000),
						},
						"pagination_limited_to": {
							Description: "The maximum number of hits accessible via pagination",
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1000,
						},
					},
				},
			},
			"typos_config": {
				Description: "The configuration for typos in index setting.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_word_size_for_1_typo": {
							Description:  "Minimum number of characters a word in the query string must contain to accept matches with 1 typo.",
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      4,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"min_word_size_for_2_typos": {
							Description:  "Minimum number of characters a word in the query string must contain to accept matches with 2 typos.",
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      8,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"typo_tolerance": {
							Description:  "Whether typo tolerance is enabled and how it is applied",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "true",
							ValidateFunc: validation.StringInSlice([]string{"true", "false", "min", "strict"}, false),
						},
						"allow_typos_on_numeric_tokens": {
							Description: "Whether to allow typos on numbers (“numeric tokens”) in the query str",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"disable_typo_tolerance_on_attributes": {
							Description: "List of attributes on which you want to disable typo tolerance.",
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"disable_typo_tolerance_on_words": {
							Description: "List of words on which typo tolerance will be disabled.",
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"separators_to_index": {
							Description: "Separators (punctuation characters) to index. By default, separators are not indexed.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
						},
					},
				},
			},
			"languages_config": {
				Description: "The configuration for languages in index setting.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ignore_plurals": {
							Description:   "Weather to treat singular, plurals, and other forms of declensions as matching terms.",
							Type:          schema.TypeBool,
							Optional:      true,
							Default:       false,
							ConflictsWith: []string{"languages_config.0.ignore_plurals_for"},
						},
						"ignore_plurals_for": {
							Description: `Weather to treat singular, plurals, and other forms of declensions as matching terms in target languages.
List of supported languages are listed on http://nhttps//www.algolia.com/doc/api-reference/api-parameters/ignorePlurals/#usage-notes`,
							Type:          schema.TypeSet,
							Elem:          &schema.Schema{Type: schema.TypeString},
							Set:           schema.HashString,
							Optional:      true,
							ConflictsWith: []string{"languages_config.0.ignore_plurals"},
						},
						"attributes_to_transliterate": {
							Description: "List of attributes to apply transliteration",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							Computed:    true,
						},
						"remove_stop_words": {
							Description:   "Weather to removes stop (common) words from the query before executing it.",
							Type:          schema.TypeBool,
							Optional:      true,
							Default:       false,
							ConflictsWith: []string{"languages_config.0.remove_stop_words_for"},
						},
						"remove_stop_words_for": {
							Description:   "List of languages to removes stop (common) words from the query before executing it.",
							Type:          schema.TypeSet,
							Elem:          &schema.Schema{Type: schema.TypeString},
							Set:           schema.HashString,
							Optional:      true,
							ConflictsWith: []string{"languages_config.0.remove_stop_words"},
						},
						"camel_case_attributes": {
							Description: "List of attributes on which to do a decomposition of camel case words.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
						"decompounded_attributes": {
							Description: "List of attributes to apply word segmentation, also known as decompounding.",
							Type:        schema.TypeList,
							Optional:    true,
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
							Description: "List of characters that the engine shouldn’t automatically normalize.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
						},
						"custom_normalization": {
							Description: "Custom normalization which overrides the engine’s default normalization",
							Type:        schema.TypeMap,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"query_languages": {
							Description: "List of languages to be used by language-specific settings and functionalities such as ignorePlurals, removeStopWords, and CJK word-detection.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
						"index_languages": {
							Description: "List of languages at the index level for language-specific processing such as tokenization and normalization.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
						"decompound_query": {
							Description: "Weather to split compound words into their composing atoms in the query.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
					},
				},
			},
			"enable_rules": {
				Description: "Whether Rules should be globally enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"enable_personalization": {
				Description: "Weather to enable the Personalization feature.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"query_strategy_config": {
				Description: "The configuration for query strategy in index setting.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_type": {
							Description:  "Query type to control if and how query words are interpreted as prefixes.",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "prefixLast",
							ValidateFunc: validation.StringInSlice([]string{"prefixLast", "prefixAll", "prefixNone"}, false),
						},
						"remove_words_if_no_results": {
							Description:  "Strategy to remove words from the query when it doesn’t match any hits.",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "none",
							ValidateFunc: validation.StringInSlice([]string{"none", "lastWords", "firstWords", "allOptional"}, false),
						},
						"advanced_syntax": {
							Description: "Weather to enable the advanced query syntax.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"optional_words": {
							Description: "A list of words that should be considered as optional when found in the query.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
						"disable_prefix_on_attributes": {
							Description: "List of attributes on which you want to disable prefix matching.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
						"disable_exact_on_attributes": {
							Description: "List of attributes on which you want to disable the exact ranking criterion.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
						"exact_on_single_word_query": {
							Description:  "Controls how the exact ranking criterion is computed when the query contains only one word.",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "attribute",
							ValidateFunc: validation.StringInSlice([]string{"none", "lastWords", "firstWords", "allOptional"}, false),
						},
						"alternatives_as_exact": {
							Description: "List of alternatives that should be considered an exact match by the exact ranking criterion.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							DefaultFunc: func() (interface{}, error) {
								return []string{"ignorePlurals", "singleWordSynonym"}, nil
							},
						},
						"advanced_syntax_features": {
							Description: "Advanced syntax features to be activated when ‘advancedSyntax’ is enabled",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
							DefaultFunc: func() (interface{}, error) {
								return []string{"exactPhrase", "excludeWords"}, nil
							},
						},
						// TODO: Add params for advanced setting
						//  https://www.algolia.com/doc/api-reference/settings-api-parameters/
					},
				},
			},
			"performance_config": {
				Description: "The configuration for performance in index setting.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"numeric_attributes_for_filtering": {
							Description: "List of numeric attributes that can be used as numerical filters.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Optional:    true,
						},
						"allow_compression_of_integer_array": {
							Description: "Weather to enable compression of large integer arrays.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
		},
	}
}

func resourceIndexCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	indexName := d.Get("name").(string)
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
	apiClient := m.(*apiClient)

	index := apiClient.searchClient.InitIndex(d.Id())
	res, err := index.Delete(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := res.Wait(ctx); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceIndexStateContext(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := refreshAPIKeyState(ctx, d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func refreshIndexState(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*apiClient)

	index := apiClient.searchClient.InitIndex(d.Id())
	settings, err := index.GetSettings(ctx)
	if err != nil {
		d.SetId("")
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
		"name": d.Id(),
		"attributes_config": []interface{}{map[string]interface{}{
			"searchable_attributes":    settings.SearchableAttributes.Get(),
			"attributes_for_faceting":  settings.AttributesForFaceting.Get(),
			"unretrievable_attributes": settings.UnretrievableAttributes.Get(),
			"attributes_to_retrieve":   settings.AttributesToRetrieve.Get(),
		}},
		"ranking_config": []interface{}{map[string]interface{}{
			"ranking":        settings.Ranking.Get(),
			"custom_ranking": settings.CustomRanking.Get(),
			"replicas":       settings.Replicas.Get(),
		}},
		"faceting_config": []interface{}{map[string]interface{}{
			"max_values_per_facet": settings.MaxValuesPerFacet.Get(),
			"sort_facet_values_by": settings.SortFacetValuesBy.Get(),
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
	}
	if err := setValues(d, values); err != nil {
		return err
	}

	return nil
}

func mapToIndexSettings(d *schema.ResourceData) search.Settings {
	settings := search.Settings{}
	if v, ok := d.GetOk("attributes_config"); ok {
		unmarshalAttributesConfig(v, &settings)
	}
	if v, ok := d.GetOk("ranking_config"); ok {
		unmarshalRankingConfig(v, &settings)
	}
	if v, ok := d.GetOk("faceting_config"); ok {
		unmarshalFacetingConfig(v, &settings)
	}
	if v, ok := d.GetOk("pagination_config"); ok {
		unmarshalPaginationConfig(v, &settings)
	}
	if v, ok := d.GetOk("typos_config"); ok {
		unmarshalTyposConfig(v, &settings)
	}
	if v, ok := d.GetOk("languages_config"); ok {
		unmarshalLanguagesConfig(v, &settings)
	}
	if v, ok := d.GetOk("enable_rules"); ok {
		settings.EnableRules = opt.EnableRules(v.(bool))
	}
	if v, ok := d.GetOk("enable_personalization"); ok {
		settings.EnablePersonalization = opt.EnablePersonalization(v.(bool))
	}
	if v, ok := d.GetOk("query_strategy_config"); ok {
		unmarshalQueryStrategyConfig(v, &settings)
	}
	if v, ok := d.GetOk("performance_config"); ok {
		unmarshalPerformanceConfig(v, &settings)
	}

	return settings
}

func unmarshalAttributesConfig(configured interface{}, settings *search.Settings) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}
	config := l[0].(map[string]interface{})
	settings.SearchableAttributes = opt.SearchableAttributes(castStringList(config["searchable_attributes"])...)
	settings.AttributesForFaceting = opt.AttributesForFaceting(castStringSet(config["attributes_for_faceting"])...)
	settings.UnretrievableAttributes = opt.UnretrievableAttributes(castStringSet(config["unretrievable_attributes"])...)
	settings.AttributesToRetrieve = opt.AttributesToRetrieve(castStringSet(config["attributes_to_retrieve"])...)
}

func unmarshalRankingConfig(configured interface{}, settings *search.Settings) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}
	config := l[0].(map[string]interface{})
	settings.Ranking = opt.Ranking(castStringList(config["ranking"])...)
	settings.CustomRanking = opt.CustomRanking(castStringList(config["custom_ranking"])...)
	settings.Replicas = opt.Replicas(castStringSet(config["replicas"])...)
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

func unmarshalTyposConfig(configured interface{}, settings *search.Settings) {
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

func unmarshalLanguagesConfig(configured interface{}, settings *search.Settings) {
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
	if v, ok := config["attributes_to_transliterate"]; ok {
		settings.AttributesToTransliterate = opt.AttributesToTransliterate(castStringSet(v)...)
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
	if v, ok := config["camel_case_attributes"]; ok {
		settings.CamelCaseAttributes = opt.CamelCaseAttributes(castStringSet(v)...)
	}
	if v, ok := config["decompounded_attributes"]; ok {
		unmarshalLanguagesConfigDecompoundedAttributes(v, settings)
	}
	if v, ok := config["keep_diacritics_on_characters"]; ok {
		settings.KeepDiacriticsOnCharacters = opt.KeepDiacriticsOnCharacters(v.(string))
	}
	if v, ok := config["custom_normalization"]; ok {
		settings.CustomNormalization = opt.CustomNormalization(map[string]map[string]string{"default": castStringMap(v)})
	}
	if v, ok := config["query_languages"]; ok {
		settings.QueryLanguages = opt.QueryLanguages(castStringSet(v)...)
	}
	if v, ok := config["index_languages"]; ok {
		settings.IndexLanguages = opt.IndexLanguages(castStringSet(v)...)
	}
	if v, ok := config["decompound_query"]; ok {
		settings.DecompoundQuery = opt.DecompoundQuery(v.(bool))
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

func unmarshalQueryStrategyConfig(configured interface{}, settings *search.Settings) {
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
	if v, ok := config["optional_words"]; ok {
		settings.OptionalWords = opt.OptionalWords(castStringSet(v)...)
	}
	if v, ok := config["disable_prefix_on_attributes"]; ok {
		settings.DisablePrefixOnAttributes = opt.DisablePrefixOnAttributes(castStringSet(v)...)
	}
	if v, ok := config["disable_exact_on_attributes"]; ok {
		settings.DisableExactOnAttributes = opt.DisableExactOnAttributes(castStringSet(v)...)
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
}

func unmarshalPerformanceConfig(configured interface{}, settings *search.Settings) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}

	config := l[0].(map[string]interface{})

	if v, ok := config["numeric_attributes_for_filtering"]; ok {
		settings.NumericAttributesForFiltering = opt.NumericAttributesForFiltering(castStringSet(v)...)
	}
	if v, ok := config["allow_compression_of_integer_array"]; ok {
		settings.AllowCompressionOfIntegerArray = opt.AllowCompressionOfIntegerArray(v.(bool))
	}
}
