package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
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

This parameter goes hand in hand with the ` + "`anchoring`" + ` parameter. If you're creating a Rule that depends on a specific query, you must specify the pattern and anchoring. The empty ` + "`\"\"`" + ` pattern is only allowed when ` + "`anchoring`" + ` is set to ` + "`is`" + `.

Otherwise, you can omit both.
`,
						},
						"anchoring": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"is", "startsWith", "endsWith", "contains"}, false),
							Description: `Whether the pattern parameter must match the beginning or the end of the query string, or both, or none.
Possible values are ` + "`is`, `startsWith`, `endsWith` and `contains`." + `
This parameter goes hand in hand with the ` + "`pattern`" + ` parameter. If you're creating a Rule that depends on a specific query, you must specify the ` + "`pattern` and `anchoring`." + `

Otherwise, you can omit both.
`,
						},
						"alternatives": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							Description: `Whether the ` + "`pattern`" + ` matches on plurals, synonyms, and typos.

This parameter goes hand in hand with the ` + "`pattern` " + ` parameter. If the ` + "`pattern` is \u201cshoe\u201d and `alternatives` is `true`, the `pattern`" + ` matches on \u201cshoes\u201d, as well as synonyms and typos of \u201cshoe\u201d.`,
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
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							AtLeastOneOf: []string{"consequence.0.params", "consequence.0.params_json", "consequence.0.promote", "consequence.0.hide", "consequence.0.user_data"},
							Description:  "**Deprecated:** Use `params_json` instead. Additional search parameters. Any valid search parameter is allowed. Specific treatment is applied to these fields: `query`, `automaticFacetFilters`, `automaticOptionalFacetFilters`.",
							Deprecated:   "Use `params_json` instead",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"query": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"consequence.0.params.0.object_query"},
										Description:   "It replaces the entire query string. Either one of `query` or `object_query` can be set.",
									},
									"object_query": {
										Type:          schema.TypeList,
										Optional:      true,
										ConflictsWith: []string{"consequence.0.params.0.query"},
										Description:   "It describes incremental edits to be made to the query string. Either one of `query` or `object_query` can be set.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"remove", "replace"}, false),
													Description: `Type of edit. Must be one of:
	- ` + "`remove`" + `: when you want to delete some text and not replace it with anything
	- ` + "`replace`" + `: when you want to delete some text and replace it with something else
`,
												},
												"delete": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Text or patterns to remove from the query string.",
												},
												"insert": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Text that should be inserted in place of the removed text inside the query string.",
												},
											},
										},
									},
									"automatic_facet_filters": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "Names of facets to which automatic filtering must be applied; they must match the facet name of a facet value placeholder in the query pattern.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"facet": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Attribute to filter on. This must match a facet placeholder in the Rule's pattern.",
												},
												"score": {
													Type:        schema.TypeInt,
													Optional:    true,
													Default:     1,
													Description: "Score for the filter. Typically used for optional or disjunctive filters.",
												},
												"disjunctive": {
													Type:        schema.TypeBool,
													Optional:    true,
													Default:     false,
													Description: "Whether the filter is disjunctive (true) or conjunctive (false). If the filter applies multiple times, e.g. because the query string contains multiple values of the same facet, the multiple occurrences are combined with an `AND` operator by default (conjunctive mode). If the filter is specified as disjunctive, however, multiple occurrences are combined with an `OR` operator instead.",
												},
											},
										},
									},
									"automatic_optional_facet_filters": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "Same syntax as `automatic_facet_filters`, but the engine treats the filters as optional. Behaves like [optionalFilters](https://www.algolia.com/doc/api-reference/api-parameters/optionalFilters/).",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"facet": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Attribute to filter on. This must match a facet placeholder in the Rule's pattern.",
												},
												"score": {
													Type:        schema.TypeInt,
													Optional:    true,
													Default:     1,
													Description: "Score for the filter. Typically used for optional or disjunctive filters.",
												},
												"disjunctive": {
													Type:        schema.TypeBool,
													Optional:    true,
													Default:     false,
													Description: "Whether the filter is disjunctive (true) or conjunctive (false). If the filter applies multiple times, e.g. because the query string contains multiple values of the same facet, the multiple occurrences are combined with an `AND` operator by default (conjunctive mode). If the filter is specified as disjunctive, however, multiple occurrences are combined with an `OR` operator instead.",
												},
											},
										},
									},
								},
							},
						},
						"params_json": {
							Type:             schema.TypeString,
							Optional:         true,
							AtLeastOneOf:     []string{"consequence.0.params", "consequence.0.params_json", "consequence.0.promote", "consequence.0.hide", "consequence.0.user_data"},
							Description:      "Additional search parameters in JSON format. Any valid search parameter is allowed. Specific treatment is applied to these fields: `query`, `automaticFacetFilters`, `automaticOptionalFacetFilters`.",
							DiffSuppressFunc: diffJsonSuppress,
							ValidateFunc:     validation.StringIsJSON,
						},
						"promote": {
							Type:         schema.TypeList,
							Optional:     true,
							AtLeastOneOf: []string{"consequence.0.params", "consequence.0.params_json", "consequence.0.promote", "consequence.0.hide", "consequence.0.user_data"},
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
							AtLeastOneOf: []string{"consequence.0.params", "consequence.0.params_json", "consequence.0.promote", "consequence.0.hide", "consequence.0.user_data"},
							Description:  "List of object IDs to hide from hits.",
						},
						"user_data": {
							Type:         schema.TypeString,
							Optional:     true,
							AtLeastOneOf: []string{"consequence.0.params", "consequence.0.params_json", "consequence.0.promote", "consequence.0.hide", "consequence.0.user_data"},
							Description:  "Custom JSON formatted string that will be appended to the userData array in the response. This object is not interpreted by the API. It is limited to 1kB of minified JSON.",
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

	indexName := d.Get("index_name").(string)
	rule, err := mapToRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := apiClient.searchClient.SaveRule(
		apiClient.searchClient.NewApiSaveRuleRequest(indexName, rule.ObjectID, rule),
		search.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = apiClient.searchClient.WaitForTask(indexName, res.TaskID); err != nil {
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

	indexName := d.Get("index_name").(string)
	rule, err := mapToRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := apiClient.searchClient.SaveRule(
		apiClient.searchClient.NewApiSaveRuleRequest(indexName, rule.ObjectID, rule),
		search.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = apiClient.searchClient.WaitForTask(indexName, res.TaskID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(rule.ObjectID)

	return resourceRuleRead(ctx, d, m)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	indexName := d.Get("index_name").(string)
	objectID := d.Get("object_id").(string)
	res, err := apiClient.searchClient.DeleteRule(
		apiClient.searchClient.NewApiDeleteRuleRequest(indexName, objectID),
		search.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = apiClient.searchClient.WaitForTask(indexName, res.TaskID); err != nil {
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

	var rule *search.Rule
	err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		var err error
		rule, err = apiClient.searchClient.GetRule(
			apiClient.searchClient.NewApiGetRuleRequest(indexName, d.Id()),
			search.WithContext(ctx),
		)

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
		alternatives := c.GetAlternatives()
		conditions = append(conditions, map[string]interface{}{
			"pattern":      c.GetPattern(),
			"anchoring":    string(c.GetAnchoring()),
			"alternatives": alternatives,
			"context":      c.GetContext(),
		})
	}

	consequence := map[string]interface{}{}
	{
		if rule.Consequence.Params != nil {
			if isParamsJSONSet(d) {
				paramsJSON, err := json.Marshal(rule.Consequence.Params)
				if err != nil {
					return fmt.Errorf("failed to marshal consequence params: %w", err)
				}
				consequence["params_json"] = string(paramsJSON)
			} else {
				params := rule.Consequence.Params
				paramsData := map[string]interface{}{}
				if params.Query != nil {
					if params.Query.String != nil {
						paramsData["query"] = *params.Query.String
					} else if params.Query.ConsequenceQueryObject != nil {
						var edits []interface{}
						for _, edit := range params.Query.ConsequenceQueryObject.Edits {
							editMap := map[string]interface{}{
								"type":   string(edit.GetType()),
								"delete": edit.GetDelete(),
								"insert": edit.GetInsert(),
							}
							edits = append(edits, editMap)
						}
						paramsData["object_query"] = edits
					}
				}

				var automaticFacetFilters []interface{}
				if params.AutomaticFacetFilters != nil {
					if filters := params.AutomaticFacetFilters.ArrayOfAutomaticFacetFilter; filters != nil {
						for _, aff := range *filters {
							automaticFacetFilters = append(automaticFacetFilters, map[string]interface{}{
								"facet":       aff.Facet,
								"score":       int(aff.GetScore()),
								"disjunctive": aff.GetDisjunctive(),
							})
						}
					}
				}
				paramsData["automatic_facet_filters"] = automaticFacetFilters

				var automaticOptionalFacetFilters []interface{}
				if params.AutomaticOptionalFacetFilters != nil {
					if filters := params.AutomaticOptionalFacetFilters.ArrayOfAutomaticFacetFilter; filters != nil {
						for _, aff := range *filters {
							automaticOptionalFacetFilters = append(automaticOptionalFacetFilters, map[string]interface{}{
								"facet":       aff.Facet,
								"score":       int(aff.GetScore()),
								"disjunctive": aff.GetDisjunctive(),
							})
						}
					}
				}
				paramsData["automatic_optional_facet_filters"] = automaticOptionalFacetFilters

				consequence["params"] = []interface{}{paramsData}
			}
		}
		var promotedObjects []interface{}
		for _, p := range rule.Consequence.Promote {
			promotedObject := map[string]interface{}{}
			if p.PromoteObjectID != nil {
				promotedObject["object_ids"] = []string{p.PromoteObjectID.ObjectID}
				promotedObject["position"] = int(p.PromoteObjectID.Position)
			}
			if p.PromoteObjectIDs != nil {
				promotedObject["object_ids"] = p.PromoteObjectIDs.ObjectIDs
				promotedObject["position"] = int(p.PromoteObjectIDs.Position)
			}
			promotedObjects = append(promotedObjects, promotedObject)
		}
		consequence["promote"] = promotedObjects

		var hiddenObjectIDs []string
		for _, hiddenObject := range rule.Consequence.Hide {
			hiddenObjectIDs = append(hiddenObjectIDs, hiddenObject.ObjectID)
		}
		consequence["hide"] = hiddenObjectIDs

		if rule.Consequence.UserData != nil {
			userDataJSON, err := json.Marshal(rule.Consequence.UserData)
			if err != nil {
				return fmt.Errorf("failed to marshal consequence userData: %w", err)
			}
			consequence["user_data"] = string(userDataJSON)
		}
	}

	var validty []interface{}
	for _, timeRange := range rule.Validity {
		from := time.Unix(timeRange.GetFrom(), 0).UTC().Format(time.RFC3339)
		until := time.Unix(timeRange.GetUntil(), 0).UTC().Format(time.RFC3339)
		validty = append(validty, map[string]string{
			"from":  from,
			"until": until,
		})
	}

	enabled := true
	if rule.Enabled != nil {
		enabled = *rule.Enabled
	}

	values := map[string]interface{}{
		"index_name":  indexName,
		"object_id":   rule.ObjectID,
		"conditions":  conditions,
		"consequence": []interface{}{consequence},
		"description": rule.GetDescription(),
		"enabled":     enabled,
		"validity":    validty,
	}
	if err := setValues(d, values); err != nil {
		return err
	}

	d.SetId(rule.ObjectID)

	return nil
}

func isParamsJSONSet(d *schema.ResourceData) bool {
	l, ok := d.Get("consequence").([]interface{})
	if !ok || len(l) == 0 {
		return false
	}

	consequence := l[0].(map[string]interface{})
	_, ok = consequence["params_json"]
	return ok
}

func mapToRule(d *schema.ResourceData) (*search.Rule, error) {
	rule := &search.Rule{
		ObjectID: d.Get("object_id").(string),
	}

	if v, ok := d.GetOk("conditions"); ok {
		unmarshalConditions(v, rule)
	}
	var err error
	rule.Consequence, err = unmarshalConsequence(d.Get("consequence"))
	if err != nil {
		return rule, err
	}
	if v, ok := d.GetOk("description"); ok {
		s := v.(string)
		rule.Description = &s
	}
	if v, ok := d.GetOk("enabled"); ok {
		b := v.(bool)
		rule.Enabled = &b
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

	var conditions []search.Condition
	for _, conditionInterface := range l {
		ruleCondition := search.Condition{}
		c := conditionInterface.(map[string]interface{})
		if v, ok := c["pattern"]; ok {
			s := v.(string)
			ruleCondition.Pattern = &s
		}
		if v, ok := c["anchoring"]; ok {
			a := search.Anchoring(v.(string))
			ruleCondition.Anchoring = &a
		}
		if v, ok := c["context"]; ok {
			s := v.(string)
			ruleCondition.Context = &s
		}
		if v, ok := c["alternatives"]; ok {
			b := v.(bool)
			ruleCondition.Alternatives = &b
		}
		conditions = append(conditions, ruleCondition)
	}

	rule.Conditions = conditions
}

func unmarshalConsequence(configured interface{}) (search.Consequence, error) {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return search.Consequence{}, nil
	}

	config := l[0].(map[string]interface{})
	consequence := search.Consequence{}
	if v, ok := config["params"]; ok {
		consequence.Params = unmarshalConsequenceParams(v)
	}
	if v, ok := config["params_json"]; ok {
		var err error
		consequence.Params, err = unmarshalConsequenceParamsJSON(v)
		if err != nil {
			return search.Consequence{}, err
		}
	}
	if v, ok := config["promote"]; ok {
		var promotes []search.Promote
		for _, v := range v.([]interface{}) {
			promotedObjectData := v.(map[string]interface{})
			objectIDs := castStringSet(promotedObjectData["object_ids"])
			position := int32(promotedObjectData["position"].(int))
			promote := search.PromoteObjectIDsAsPromote(
				search.NewPromoteObjectIDs(objectIDs, position),
			)
			promotes = append(promotes, *promote)
		}
		consequence.Promote = promotes
	}
	if v, ok := config["hide"]; ok {
		var hide []search.ConsequenceHide
		for _, objectID := range castStringSet(v) {
			hide = append(hide, search.ConsequenceHide{ObjectID: objectID})
		}
		consequence.Hide = hide
	}
	if v, ok := config["user_data"]; ok {
		s := v.(string)
		if s != "" {
			var userData map[string]any
			if err := json.Unmarshal([]byte(s), &userData); err != nil {
				return search.Consequence{}, fmt.Errorf("failed to unmarshal user_data: %w", err)
			}
			consequence.UserData = userData
		}
	}
	return consequence, nil
}

func unmarshalConsequenceParams(configured interface{}) *search.ConsequenceParams {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	paramsData := l[0].(map[string]interface{})

	params := search.ConsequenceParams{}
	if v, ok := paramsData["query"]; ok {
		s := v.(string)
		if s != "" {
			params.Query = search.StringAsConsequenceQuery(s)
		}
	}
	if v, ok := paramsData["object_query"]; ok {
		var edits []search.Edit
		for _, e := range v.([]interface{}) {
			editData := e.(map[string]interface{})
			editType := search.EditType(editData["type"].(string))
			deleteStr := editData["delete"].(string)
			edit := search.Edit{
				Type:   &editType,
				Delete: &deleteStr,
			}
			if insert, ok := editData["insert"]; ok {
				s := insert.(string)
				if s != "" {
					edit.Insert = &s
				}
			}
			edits = append(edits, edit)
		}
		params.Query = search.ConsequenceQueryObjectAsConsequenceQuery(
			&search.ConsequenceQueryObject{Edits: edits},
		)
	}
	if v, ok := paramsData["automatic_facet_filters"]; ok {
		filters := unmarshalAutomaticFacetFilters(v)
		if len(filters) > 0 {
			params.AutomaticFacetFilters = search.ArrayOfAutomaticFacetFilterAsAutomaticFacetFilters(filters)
		}
	}
	if v, ok := paramsData["automatic_optional_facet_filters"]; ok {
		filters := unmarshalAutomaticFacetFilters(v)
		if len(filters) > 0 {
			params.AutomaticOptionalFacetFilters = search.ArrayOfAutomaticFacetFilterAsAutomaticFacetFilters(filters)
		}
	}

	return &params
}

func unmarshalConsequenceParamsJSON(configured interface{}) (*search.ConsequenceParams, error) {
	paramsJSON := configured.(string)
	if paramsJSON == "" {
		return nil, nil
	}
	params := search.ConsequenceParams{}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consequence params: %w", err)
	}
	return &params, nil
}

func unmarshalAutomaticFacetFilters(configured interface{}) []search.AutomaticFacetFilter {
	var automaticFacetFilters []search.AutomaticFacetFilter
	for _, v := range configured.([]interface{}) {
		data := v.(map[string]interface{})
		score := int32(data["score"].(int))
		disjunctive := data["disjunctive"].(bool)
		aff := search.AutomaticFacetFilter{
			Facet:       data["facet"].(string),
			Score:       &score,
			Disjunctive: &disjunctive,
		}
		automaticFacetFilters = append(automaticFacetFilters, aff)
	}
	return automaticFacetFilters
}

func unmarshalValidity(configured interface{}) []search.TimeRange {
	l := configured.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var timeRanges []search.TimeRange
	for _, timeRangeData := range l {
		timeRange := timeRangeData.(map[string]interface{})
		fromTime, _ := time.Parse(time.RFC3339, timeRange["from"].(string))
		untilTime, _ := time.Parse(time.RFC3339, timeRange["until"].(string))
		fromUnix := fromTime.Unix()
		untilUnix := untilTime.Unix()
		timeRanges = append(timeRanges, search.TimeRange{
			From:  &fromUnix,
			Until: &untilUnix,
		})
	}

	return timeRanges
}
