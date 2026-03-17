package provider

import (
	"context"
	"fmt"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
)

func resourceSynonyms() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSynonymsCreate,
		ReadContext:   resourceSynonymsRead,
		UpdateContext: resourceSynonymsUpdate,
		DeleteContext: resourceSynonymsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSynonymsStateContext,
		},
		Description: `A configuration for synonyms. To get more information about synonyms, see the [Official Documentation](https://www.algolia.com/doc/guides/managing-results/optimize-search-results/adding-synonyms/).

※ **It replaces any existing synonyms set for the index.** So you can't have multiple ` + "`algolia_synonyms`" + ` resources for the same index.
`,
		// https://www.algolia.com/doc/api-reference/api-methods/batch-synonyms/
		Schema: map[string]*schema.Schema{
			"index_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the index to apply synonyms.",
			},
			"synonyms": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "A list of conditions that should apply to activate a Rule. You can use up to 25 conditions per Rule.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique identifier for the synonym.It can contain any character, and be of unlimited length.",
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"synonym", "oneWaySynonym", "altCorrection1", "altCorrection2", "placeholder"}, false),
							Description:  "The type of the synonym. Possible values are `synonym`, `oneWaySynonym`, `altCorrection1`, `altCorrection2` and `placeholder`.",
						},
						"synonyms": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of synonyms (up to `20 for type `synonym` and 100 for type `oneWaySynonym`). Required if type=`synonym` or type=`oneWaySynonym`.",
						},
						"input": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Defines the synonym. A word or expression, used as the basis for the array of synonyms. Required if type=`oneWaySynonym`.",
						},
						"word": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Single word, used as the basis for the below array of corrections. Required if type=`altCorrection1` or type=`altCorrection2`",
						},
						"corrections": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of corrections of the `word`. Required if type=`altCorrection1` or type=`altCorrection2`",
						},
						"placeholder": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Single word, used as the basis for the below array of replacements.  Required if type=`placeholder`",
						},
						"replacements": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of replacements of the placeholder. Required if type=`placeholder`",
						},
					},
				},
			},
		},
	}
}

func resourceSynonymsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	indexName := d.Get("index_name").(string)
	req := apiClient.searchClient.NewApiSaveSynonymsRequest(indexName, mapToSynonyms(d))
	req = req.WithReplaceExistingSynonyms(true)
	res, err := apiClient.searchClient.SaveSynonyms(req)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = apiClient.searchClient.WaitForTask(indexName, res.TaskID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(indexName)

	return resourceSynonymsRead(ctx, d, m)
}

func resourceSynonymsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := refreshSynonymsState(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSynonymsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	indexName := d.Get("index_name").(string)
	req := apiClient.searchClient.NewApiSaveSynonymsRequest(indexName, mapToSynonyms(d))
	req = req.WithReplaceExistingSynonyms(true)
	res, err := apiClient.searchClient.SaveSynonyms(req)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = apiClient.searchClient.WaitForTask(indexName, res.TaskID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(indexName)

	return resourceSynonymsRead(ctx, d, m)
}

func resourceSynonymsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	res, err := apiClient.searchClient.ClearSynonyms(apiClient.searchClient.NewApiClearSynonymsRequest(d.Id()))
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = apiClient.searchClient.WaitForTask(d.Id(), res.TaskID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSynonymsStateContext(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := d.Set("index_name", d.Id()); err != nil {
		return nil, err
	}
	if err := refreshSynonymsState(ctx, d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func refreshSynonymsState(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*apiClient)

	indexName := d.Id()

	var allSynonymHits []search.SynonymHit
	var page int32

	for {
		hitsPerPage := int32(1000)
		params := search.NewSearchSynonymsParams(
			search.WithSearchSynonymsParamsPage(page),
			search.WithSearchSynonymsParamsHitsPerPage(hitsPerPage),
		)
		req := apiClient.searchClient.NewApiSearchSynonymsRequest(indexName).WithSearchSynonymsParams(params)
		res, err := apiClient.searchClient.SearchSynonyms(req)
		if err != nil {
			if algoliautil.IsNotFoundError(err) {
				tflog.Warn(ctx, fmt.Sprintf("synonyms for (%s) not found, removing from state", d.Id()))
				d.SetId("")
				return nil
			}
			return err
		}

		allSynonymHits = append(allSynonymHits, res.Hits...)

		if len(res.Hits) < int(hitsPerPage) {
			break
		}
		page++
	}

	var synonyms []interface{}
	for _, synonym := range allSynonymHits {
		synonymData := map[string]interface{}{
			"object_id": synonym.ObjectID,
			"type":      string(synonym.Type),
		}
		switch synonym.Type {
		case search.SYNONYM_TYPE_SYNONYM:
			synonymData["synonyms"] = synonym.Synonyms
		case search.SYNONYM_TYPE_ONE_WAY_SYNONYM, search.SYNONYM_TYPE_ONEWAYSYNONYM:
			if synonym.Input != nil {
				synonymData["input"] = *synonym.Input
			}
			synonymData["synonyms"] = synonym.Synonyms
		case search.SYNONYM_TYPE_ALT_CORRECTION1, search.SYNONYM_TYPE_ALTCORRECTION1:
			if synonym.Word != nil {
				synonymData["word"] = *synonym.Word
			}
			synonymData["corrections"] = synonym.Corrections
		case search.SYNONYM_TYPE_ALT_CORRECTION2, search.SYNONYM_TYPE_ALTCORRECTION2:
			if synonym.Word != nil {
				synonymData["word"] = *synonym.Word
			}
			synonymData["corrections"] = synonym.Corrections
		case search.SYNONYM_TYPE_PLACEHOLDER:
			if synonym.Placeholder != nil {
				synonymData["placeholder"] = *synonym.Placeholder
			}
			synonymData["replacements"] = synonym.Replacements
		}
		synonyms = append(synonyms, synonymData)
	}

	values := map[string]interface{}{
		"synonyms": synonyms,
	}
	if err := setValues(d, values); err != nil {
		return err
	}

	return nil
}

func mapToSynonyms(d *schema.ResourceData) []search.SynonymHit {
	l := d.Get("synonyms").(*schema.Set)
	if l.Len() == 0 || l.List()[0] == nil {
		return nil
	}

	var synonyms []search.SynonymHit
	for _, v := range l.List() {
		synonymData := v.(map[string]interface{})
		objectID := synonymData["object_id"].(string)
		synonymType := search.SynonymType(synonymData["type"].(string))

		hit := search.SynonymHit{
			ObjectID: objectID,
			Type:     synonymType,
		}

		switch synonymType {
		case search.SYNONYM_TYPE_SYNONYM:
			hit.Synonyms = castStringSet(synonymData["synonyms"])
		case search.SYNONYM_TYPE_ONE_WAY_SYNONYM, search.SYNONYM_TYPE_ONEWAYSYNONYM:
			input := synonymData["input"].(string)
			hit.Input = &input
			hit.Synonyms = castStringSet(synonymData["synonyms"])
		case search.SYNONYM_TYPE_ALT_CORRECTION1, search.SYNONYM_TYPE_ALTCORRECTION1:
			word := synonymData["word"].(string)
			hit.Word = &word
			hit.Corrections = castStringSet(synonymData["corrections"])
		case search.SYNONYM_TYPE_ALT_CORRECTION2, search.SYNONYM_TYPE_ALTCORRECTION2:
			word := synonymData["word"].(string)
			hit.Word = &word
			hit.Corrections = castStringSet(synonymData["corrections"])
		case search.SYNONYM_TYPE_PLACEHOLDER:
			placeholder := synonymData["placeholder"].(string)
			hit.Placeholder = &placeholder
			hit.Replacements = castStringSet(synonymData["replacements"])
		}

		synonyms = append(synonyms, hit)
	}

	return synonyms
}
