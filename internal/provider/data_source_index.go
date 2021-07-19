package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceIndex() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for an index.",
		ReadContext: dataSourceIndexRead,
		// https://www.algolia.com/doc/api-reference/settings-api-parameters/
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the index.",
			},
			"attributes_config": {
				Type:        schema.TypeList,
				Computed:    true,
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
							Computed:    true,
							Description: "List of attributes that cannot be retrieved at query time.",
						},
						"attributes_to_retrieve": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of attributes to be retrieved at query time.",
						},
					},
				},
			},
			"ranking_config": {
				Type:        schema.TypeList,
				Computed:    true,
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
							Computed:    true,
							Description: "List of attributes for custom ranking criterion.",
						},
						// TODO: Add after the PR below merged.
						//  https://github.com/algolia/algoliasearch-client-go/pull/661
						// "relevancy_strictness": {
						// 	Type:         schema.TypeInt,
						//  Computed:    true,
						// 	Description:  "Relevancy threshold below which less relevant results aren’t included in the results",
						// },
						"replicas": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of replica names.",
						},
					},
				},
			},
			"faceting_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The configuration for faceting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_values_per_facet": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of facet values to return for each facet during a regular search.",
						},
						"sort_facet_values_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Parameter to controls how the facet values are sorted within each faceted attribute.",
						},
					},
				},
			},
			"highlight_and_snippet_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The configuration for highlight / snippet in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attributes_to_highlight": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of attributes to highlight.",
						},
						"attributes_to_snippet": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of attributes to snippet, with an optional maximum number of words to snippet.",
						},
						"highlight_pre_tag": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The HTML string to insert before the highlighted parts in all highlight and snippet results.",
						},
						"highlight_post_tag": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The HTML string to insert after the highlighted parts in all highlight and snippet results.",
						},
						"snippet_ellipsis_text": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "String used as an ellipsis indicator when a snippet is truncated.",
						},
						"restrict_highlight_and_snippet_arrays": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Restrict highlighting and snippeting to items that matched the query.",
						},
					},
				},
			},
			"pagination_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "The configuration for pagination in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hits_per_page": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of hits per page.",
						},
						"pagination_limited_to": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum number of hits accessible via pagination",
						},
					},
				},
			},
			"typos_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The configuration for typos in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_word_size_for_1_typo": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Minimum number of characters a word in the query string must contain to accept matches with 1 typo.",
						},
						"min_word_size_for_2_typos": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Minimum number of characters a word in the query string must contain to accept matches with 2 typos.",
						},
						"typo_tolerance": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Whether typo tolerance is enabled and how it is applied",
						},
						"allow_typos_on_numeric_tokens": {
							Type:        schema.TypeBool,
							Computed:    true,
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
							Computed:    true,
							Description: "Separators (punctuation characters) to index. By default, separators are not indexed.",
						},
					},
				},
			},
			"languages_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The configuration for languages in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ignore_plurals": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to treat singular, plurals, and other forms of declensions as matching terms.",
						},
						"ignore_plurals_for": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Computed: true,
							Description: `Whether to treat singular, plurals, and other forms of declensions as matching terms in target languages.
List of supported languages are listed on http://nhttps//www.algolia.com/doc/api-reference/api-parameters/ignorePlurals/#usage-notes`,
						},
						"attributes_to_transliterate": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of attributes to apply transliteration",
						},
						"remove_stop_words": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to removes stop (common) words from the query before executing it.",
						},
						"remove_stop_words_for": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of languages to removes stop (common) words from the query before executing it.",
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
							Computed:    true,
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
							Computed:    true,
							Description: "Whether to split compound words into their composing atoms in the query.",
						},
					},
				},
			},
			"enable_rules": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether Rules should be globally enabled.",
			},
			"enable_personalization": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether to enable the Personalization feature.",
			},
			"query_strategy_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The configuration for query strategy in index setting.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Query type to control if and how query words are interpreted as prefixes.",
						},
						"remove_words_if_no_results": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Strategy to remove words from the query when it doesn’t match any hits.",
						},
						"advanced_syntax": {
							Type:        schema.TypeBool,
							Computed:    true,
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
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Controls how the exact ranking criterion is computed when the query contains only one word.",
						},
						"alternatives_as_exact": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "List of alternatives that should be considered an exact match by the exact ranking criterion.",
						},
						"advanced_syntax_features": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
							Description: "Advanced syntax features to be activated when ‘advancedSyntax’ is enabled",
						},
						// TODO: Add params for advanced setting
						//  https://www.algolia.com/doc/api-reference/settings-api-parameters/
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
				Computed:    true,
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
							Computed: true,
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
							Computed:    true,
							Description: "Whether to highlight and snippet the original word that matches the synonym or the synonym itself.",
						},
						"min_proximity": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Precision of the `proximity` ranking criterion.",
						},
						"response_fields": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Computed: true,
							Description: `The fields the response will contain. Applies to search and browse queries.
This parameter is mainly intended to **limit the response size.** For example, in complex queries, echoing of request parameters in the response’s params field can be undesirable.`,
						},
						"max_facet_hits": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of facet hits to return during a search for facet values.",
						},
						"attribute_criteria_computed_by_min_proximity": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "When attribute is ranked above proximity in your ranking formula, proximity is used to select which searchable attribute is matched in the **attribute ranking stage**.",
						},
					},
				},
			},
		},
	}
}

func dataSourceIndexRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(d.Get("name").(string))
	if err := refreshIndexState(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
