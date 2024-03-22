package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
)

func resourceRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRuleCreate,
		ReadContext:   resourceRuleRead,
		UpdateContext: resourceRuleUpdate,
		DeleteContext: resourceRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRuleStateContext,
		},
		Description: "A configuration for a Rule.  To get more information about rules, see the [Official Documentation](https://www.algolia.com/doc/guides/managing-results/rules/rules-overview/).",
		// https://www.algolia.com/doc/api-reference/api-methods/save-rule/#parameters
		Schema: map[string]*schema.Schema{
			"index_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the index to apply rule.",
			},
			"object_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique identifier for the Rule (format: `[A-Za-z0-9_-]+`).",
			},
			"conditions": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "A list of conditions that should apply to activate a Rule. You can use up to 25 conditions per Rule.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pattern": {
							Type:     schema.TypeString,
							Optional: true,
							Description: `Query pattern syntax.
Query patterns are expressed as a string with a specific syntax. A pattern is a sequence of tokens, which can be either:

- Facet value placeholder: ` + "`{facet:$facet_name}`" + `. Example: ` + "`{facet:brand}`" + `.
- Literal: the world itself. Example: Algolia.
Special characters (` + "`*`, `{`, `}`, `:` and `\\`" + `) must be escaped by preceding them with a backslash (` + "\\" + `) if they are to be treated as literals.

This parameter goes hand in hand with the ` + "`anchoring`" + ` parameter. If you’re creating a Rule that depends on a specific query, you must specify the pattern and anchoring. The empty ` + "`\"\"`" + ` pattern is only allowed when ` + "`anchoring`" + ` is set to ` + "`is`" + `.

Otherwise, you can omit both.
`,
						},
						"anchoring": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"is", "startsWith", "endsWith", "contains"}, false),
							Description: `Whether the pattern parameter must match the beginning or the end of the query string, or both, or none.
Possible values are ` + "`is`, `startsWith`, `endsWith` and `contains`." + `
This parameter goes hand in hand with the ` + "`pattern`" + ` parameter. If you’re creating a Rule that depends on a specific query, you must specify the ` + "`pattern` and `anchoring`." + `

Otherwise, you can omit both.
`,
						},
						"alternatives": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							Description: `Whether the ` + "`pattern`" + ` matches on plurals, synonyms, and typos.

This parameter goes hand in hand with the ` + "`pattern` " + ` parameter. If the ` + "`pattern` is “shoe” and `alternatives` is `true`, the `pattern`" + ` matches on “shoes”, as well as synonyms and typos of “shoe”.`,
						},
						"context": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Rule context (format: `[A-Za-z0-9_-]+`). When specified, the Rule is only applied when the same context is specified at query time (using the `ruleContexts` parameter). When absent, the Rule is generic and always applies (provided that its other conditions are met, of course).",
						},
					},
				},
			},
			"consequence": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Description: `Consequence of the Rule. 
At least one of the following object must be used:
- params
- promote
- hide
- user_data
`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"params": {
							Type:             schema.TypeString,
							Optional:         true,
							AtLeastOneOf:     []string{"consequence.0.params", "consequence.0.promote", "consequence.0.hide", "consequence.0.user_data"},
							Description:      "Additional search parameters in JSON format. Any valid search parameter is allowed. Specific treatment is applied to these fields: `query`, `automaticFacetFilters`, `automaticOptionalFacetFilters`.",
							DiffSuppressFunc: diffJsonSuppress,
							ValidateFunc:     validation.StringIsJSON,
						},
						"promote": {
							Type:         schema.TypeList,
							Optional:     true,
							AtLeastOneOf: []string{"consequence.0.params", "consequence.0.promote", "consequence.0.hide", "consequence.0.user_data"},
							Description:  "Objects to promote as hits.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"object_ids": {
										Type:     schema.TypeSet,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
										Required: true,
									},
									"position": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "The position to promote the object(s) to (zero-based). If you pass `object_ids`, we place the objects at this position as a group. For example, if you pass four `object_ids` to position `0`, the objects take the first four positions.",
									},
								},
							},
						},
						"hide": {
							Type:         schema.TypeSet,
							Elem:         &schema.Schema{Type: schema.TypeString},
							Set:          schema.HashString,
							Optional:     true,
							AtLeastOneOf: []string{"consequence.0.params", "consequence.0.promote", "consequence.0.hide", "consequence.0.user_data"},
							Description:  "List of object IDs to hide from hits.",
						},
						"user_data": {
							Type:         schema.TypeString,
							Optional:     true,
							AtLeastOneOf: []string{"consequence.0.params", "consequence.0.promote", "consequence.0.hide", "consequence.0.user_data"},
							Description:  "Custom JSON formatted string that will be appended to the userData array in the response. This object is not interpreted by the API. It is limited to 1kB of minified JSON.",
							ValidateFunc: validation.StringIsJSON,
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "This field is intended for Rule management purposes, in particular to ease searching for Rules and presenting them to human readers. It is not interpreted by the API.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the Rule is enabled. Disabled Rules remain in the index, but are not applied at query time.",
			},
			"validity": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Objects to promote as hits.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsRFC3339Time,
							Description:  "Lower bound of the time range. RFC3339 format.",
						},
						"until": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsRFC3339Time,
							Description:  "Upper bound of the time range. RFC3339 format.",
						},
					},
				},
			},
		},
	}
}

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	rule, err := mapToRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	index := apiClient.searchClient.InitIndex(d.Get("index_name").(string))
	res, err := index.SaveRule(rule, ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(rule.ObjectID)

	return resourceRuleRead(ctx, d, m)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := refreshRuleState(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	rule, err := mapToRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	index := apiClient.searchClient.InitIndex(d.Get("index_name").(string))
	res, err := index.SaveRule(rule, ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(rule.ObjectID)

	return resourceRuleRead(ctx, d, m)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	index := apiClient.searchClient.InitIndex(d.Get("index_name").(string))
	res, err := index.DeleteRule(d.Get("object_id").(string), ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRuleStateContext(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	tokens := strings.Split(d.Id(), "/")
	if len(tokens) != 2 {
		return nil, errors.New("import id must be {{index_name}}/{{object_id}}")
	}
	indexName := tokens[0]
	objectID := tokens[1]

	d.SetId(objectID)
	if err := d.Set("index_name", indexName); err != nil {
		return nil, err
	}

	if err := refreshRuleState(ctx, d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func refreshRuleState(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*apiClient)

	indexName := d.Get("index_name").(string)
	index := apiClient.searchClient.InitIndex(indexName)

	var rule search.Rule
	err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		var err error
		rule, err = index.GetRule(d.Id(), ctx)

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
			tflog.Warn(ctx, fmt.Sprintf("rule (%s) not found, removing from state", d.Id()))
			d.SetId("")
			return nil
		}
		return err
	}

	var conditions []interface{}
	for _, c := range rule.Conditions {
		// The code below is workaround since Alternatives.enable is a private field.
		alternativesJSONBytes, _ := c.Alternatives.MarshalJSON()
		alternatives, _ := strconv.ParseBool(string(alternativesJSONBytes))
		conditions = append(conditions, map[string]interface{}{
			"pattern":      c.Pattern,
			"anchoring":    c.Anchoring,
			"alternatives": alternatives,
			"context":      c.Context,
		})
	}

	consequence := map[string]interface{}{}
	{
		if rule.Consequence.Params != nil {
			paramsJSON, err := json.Marshal(rule.Consequence.Params)
			if err != nil {
				return fmt.Errorf("failed to marshal consequence params: %w", err)
			}
			consequence["params"] = string(paramsJSON)
		}
		var promotedObjects []interface{}
		for _, p := range rule.Consequence.Promote {
			promotedObject := map[string]interface{}{}
			if p.ObjectID != "" {
				promotedObject["object_ids"] = []string{p.ObjectID}
			}
			if len(p.ObjectIDs) > 0 {
				promotedObject["object_ids"] = p.ObjectIDs
			}
			promotedObject["position"] = p.Position
			promotedObjects = append(promotedObjects, promotedObject)
		}
		consequence["promote"] = promotedObjects

		var hiddenObjectIDs []string
		for _, hiddenObject := range rule.Consequence.Hide {
			hiddenObjectIDs = append(hiddenObjectIDs, hiddenObject.ObjectID)
		}
		consequence["hide"] = hiddenObjectIDs

		if rule.Consequence.UserData != nil {
			consequence["user_data"] = rule.Consequence.UserData
		}
	}

	var validty []interface{}
	for _, timeRange := range rule.Validity {
		validty = append(validty, map[string]string{
			"from":  timeRange.From.In(time.UTC).Format(time.RFC3339),
			"until": timeRange.Until.In(time.UTC).Format(time.RFC3339),
		})
	}

	values := map[string]interface{}{
		"index_name":  indexName,
		"object_id":   rule.ObjectID,
		"conditions":  conditions,
		"consequence": []interface{}{consequence},
		"description": rule.Description,
		"enabled":     rule.Enabled.Get(),
		"validity":    validty,
	}
	if err := setValues(d, values); err != nil {
		return err
	}

	d.SetId(rule.ObjectID)

	return nil
}

func mapToRule(d *schema.ResourceData) (search.Rule, error) {
	rule := search.Rule{
		ObjectID: d.Get("object_id").(string),
	}

	if v, ok := d.GetOk("conditions"); ok {
		unmarshalConditions(v, &rule)
	}
	var err error
	rule.Consequence, err = unmarshalConsequence(d.Get("consequence"))
	if err != nil {
		return rule, err
	}
	if v, ok := d.GetOk("description"); ok {
		rule.Description = v.(string)
	}
	if v, ok := d.GetOk("enabled"); ok {
		rule.Enabled = opt.Enabled(v.(bool))
	}
	if v, ok := d.GetOk("validity"); ok {
		rule.Validity = unmarshalValidity(v)
	}

	return rule, nil
}

func unmarshalConditions(configured interface{}, rule *search.Rule) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return
	}

	var conditions []search.RuleCondition
	for _, conditionInterface := range l {
		ruleCondition := search.RuleCondition{}
		c := conditionInterface.(map[string]interface{})
		if v, ok := c["pattern"]; ok {
			ruleCondition.Pattern = v.(string)
		}
		if v, ok := c["anchoring"]; ok {
			ruleCondition.Anchoring = search.RulePatternAnchoring(v.(string))
		}
		if v, ok := c["context"]; ok {
			ruleCondition.Context = v.(string)
		}
		if v, ok := c["alternatives"]; ok {
			if v.(bool) {
				ruleCondition.Alternatives = search.AlternativesEnabled()
			} else {
				ruleCondition.Alternatives = search.AlternativesDisabled()
			}
		}
		conditions = append(conditions, ruleCondition)
	}

	rule.Conditions = conditions
}

func unmarshalConsequence(configured interface{}) (search.RuleConsequence, error) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return search.RuleConsequence{}, nil
	}

	config := l[0].(map[string]interface{})
	consequence := search.RuleConsequence{}
	if v, ok := config["params"]; ok {
		var err error
		consequence.Params, err = unmarshalConsequenceParams(v)
		if err != nil {
			return search.RuleConsequence{}, err
		}
	}
	if v, ok := config["promote"]; ok {
		var promotedObjects []search.PromotedObject
		for _, v := range v.([]interface{}) {
			promotedObjectData := v.(map[string]interface{})
			promotedObject := search.PromotedObject{
				ObjectIDs: castStringSet(promotedObjectData["object_ids"]),
				Position:  promotedObjectData["position"].(int),
			}
			promotedObjects = append(promotedObjects, promotedObject)
		}
		consequence.Promote = promotedObjects
	}
	if v, ok := config["hide"]; ok {
		var hide []search.HiddenObject
		for _, objectID := range castStringSet(v) {
			hide = append(hide, search.HiddenObject{ObjectID: objectID})
		}
		consequence.Hide = hide
	}
	if v, ok := config["user_data"]; ok {
		consequence.UserData = v.(string)
	}
	consequence.UserData = nil
	return consequence, nil
}

func unmarshalConsequenceParams(configured interface{}) (*search.RuleParams, error) {
	paramsJSON := configured.(string)
	params := search.RuleParams{}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consequence params: %w", err)
	}

	return &params, nil
}

func unmarshalValidity(configured interface{}) []search.TimeRange {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var timeRanges []search.TimeRange
	for _, timeRangeData := range l {
		timeRange := timeRangeData.(map[string]interface{})
		from, _ := time.Parse(time.RFC3339, timeRange["from"].(string))
		until, _ := time.Parse(time.RFC3339, timeRange["until"].(string))
		timeRanges = append(timeRanges, search.TimeRange{
			From:  from,
			Until: until,
		})
	}

	return timeRanges
}
