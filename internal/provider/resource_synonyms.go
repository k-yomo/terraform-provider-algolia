package provider

import (
	"context"
	"io"
	"log"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
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
	res, err := apiClient.searchClient.InitIndex(indexName).ReplaceAllSynonyms(mapToSynonyms(d), ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
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
	res, err := apiClient.searchClient.InitIndex(indexName).ReplaceAllSynonyms(mapToSynonyms(d), ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(indexName)

	return resourceSynonymsRead(ctx, d, m)
}

func resourceSynonymsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	res, err := apiClient.searchClient.InitIndex(d.Id()).ClearSynonyms(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
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
	iter, err := apiClient.searchClient.InitIndex(indexName).BrowseSynonyms(ctx)
	if err != nil {
		if algoliautil.IsNotFoundError(err) {
			log.Printf("[WARN] synonyms for (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	var synonyms []interface{}
	for {
		synonym, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		synonymData := map[string]interface{}{
			"object_id": synonym.ObjectID(),
			"type":      string(synonym.Type()),
		}
		switch synonym.Type() {
		case search.RegularSynonymType:
			rs := synonym.(search.RegularSynonym)
			synonymData["synonyms"] = rs.Synonyms
		case search.OneWaySynonymType:
			ows := synonym.(search.OneWaySynonym)
			synonymData["input"] = ows.Input
			synonymData["synonyms"] = ows.Synonyms
		case search.AltCorrection1Type:
			ac1 := synonym.(search.AltCorrection1)
			synonymData["word"] = ac1.Word
			synonymData["corrections"] = ac1.Corrections
		case search.AltCorrection2Type:
			ac2 := synonym.(search.AltCorrection2)
			synonymData["word"] = ac2.Word
			synonymData["corrections"] = ac2.Corrections
		case search.PlaceholderType:
			p := synonym.(search.Placeholder)
			synonymData["placeholder"] = p.Placeholder
			synonymData["replacements"] = p.Replacements
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

func mapToSynonyms(d *schema.ResourceData) []search.Synonym {
	l := d.Get("synonyms").(*schema.Set)
	if l.Len() == 0 || l.List()[0] == nil {
		return nil
	}

	var synonyms []search.Synonym
	for _, v := range l.List() {
		synonymData := v.(map[string]interface{})
		objectID := synonymData["object_id"].(string)

		var synonym search.Synonym
		switch search.SynonymType(synonymData["type"].(string)) {
		case search.RegularSynonymType:
			synonym = search.NewRegularSynonym(objectID, castStringSet(synonymData["synonyms"])...)
		case search.OneWaySynonymType:
			synonym = search.NewOneWaySynonym(objectID, synonymData["input"].(string), castStringSet(synonymData["synonyms"])...)
		case search.AltCorrection1Type:
			synonym = search.NewAltCorrection1(objectID, synonymData["word"].(string), castStringSet(synonymData["corrections"])...)
		case search.AltCorrection2Type:
			synonym = search.NewAltCorrection2(objectID, synonymData["word"].(string), castStringSet(synonymData["corrections"])...)
		case search.PlaceholderType:
			synonym = search.NewPlaceholder(objectID, synonymData["placeholder"].(string), castStringSet(synonymData["replacements"])...)
		}
		synonyms = append(synonyms, synonym)
	}

	return synonyms
}
