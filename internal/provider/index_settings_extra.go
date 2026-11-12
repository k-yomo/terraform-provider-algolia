package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/algolia/algoliasearch-client-go/v4/algolia/utils"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// The settings in this file were added after the original index schema was designed.
// Unlike the older settings, they are Optional+Computed without defaults, and they are only
// sent to Algolia when present in the configuration. This way, upgrading the provider never
// resets values that were configured outside Terraform (e.g. NeuralSearch or Dynamic
// Re-Ranking enabled from the dashboard), and removing an attribute from the configuration
// leaves the current value in Algolia untouched.

const (
	indexSettingModeKey                 = "mode"
	indexSettingSemanticSearchKey       = "semantic_search"
	indexSettingEnableReRankingKey      = "enable_re_ranking"
	indexSettingReRankingApplyFilterKey = "re_ranking_apply_filter"
	indexSettingRenderingContentKey     = "rendering_content_json"
	indexSettingUserDataKey             = "user_data_json"
)

const (
	indexSettingModeDescription                 = "Search mode the index will use to query for results. Possible values are `keywordSearch` and `neuralSearch`. `neuralSearch` requires the NeuralSearch feature to be available on your plan."
	indexSettingSemanticSearchDescription       = "Settings for the semantic search part of NeuralSearch. Only used when `mode` is `neuralSearch`."
	indexSettingEventSourcesDescription         = "Indices from which to collect click and conversion events. If empty, the current index and all its replicas are used."
	indexSettingEnableReRankingDescription      = "Whether this search will use [Dynamic Re-Ranking](https://www.algolia.com/doc/guides/algolia-ai/re-ranking/). This setting only has an effect if Dynamic Re-Ranking is enabled for the index in the Algolia dashboard."
	indexSettingReRankingApplyFilterDescription = "Restrict [Dynamic Re-Ranking](https://www.algolia.com/doc/guides/algolia-ai/re-ranking/) to records that match these filters, using the [filters](https://www.algolia.com/doc/api-reference/api-parameters/filters/) syntax. Set to an empty string to remove the restriction."
	indexSettingRenderingContentDescription     = "JSON-encoded [renderingContent](https://www.algolia.com/doc/api-reference/api-parameters/renderingContent/): extra data that can be used to render the search UI, such as the order of facets and facet values (`facetOrdering`), a `redirect`, or `widgets`. Use `jsonencode` to build the value. Set to `jsonencode({})` to clear it."
	indexSettingUserDataDescription             = "JSON-encoded custom data that you can retrieve in the search response. Use `jsonencode` to build the value. Set to `jsonencode({})` to clear it."
	indexSettingsExtraClearingNote              = " Removing this attribute from the configuration leaves the current value in Algolia unchanged."
)

// extraIndexSettingsResourceSchema returns the schema of the newer index settings for the
// `algolia_index` and `algolia_virtual_index` resources.
func extraIndexSettingsResourceSchema(isVirtualIndex bool) map[string]*schema.Schema {
	return withoutVirtualIndexForbiddenSettings(map[string]*schema.Schema{
		indexSettingModeKey: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice([]string{string(search.MODE_KEYWORD_SEARCH), string(search.MODE_NEURAL_SEARCH)}, false),
			Description:  indexSettingModeDescription + indexSettingsExtraClearingNote,
		},
		indexSettingSemanticSearchKey: {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Description: indexSettingSemanticSearchDescription + indexSettingsExtraClearingNote,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"event_sources": {
						Type:        schema.TypeSet,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Set:         schema.HashString,
						Optional:    true,
						Description: indexSettingEventSourcesDescription,
					},
				},
			},
		},
		indexSettingEnableReRankingKey: {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: indexSettingEnableReRankingDescription + indexSettingsExtraClearingNote,
		},
		indexSettingReRankingApplyFilterKey: {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: indexSettingReRankingApplyFilterDescription + indexSettingsExtraClearingNote,
		},
		indexSettingRenderingContentKey: {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: diffEmptyJSONObjectSuppress,
			Description:      indexSettingRenderingContentDescription + indexSettingsExtraClearingNote,
		},
		indexSettingUserDataKey: {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: diffEmptyJSONObjectSuppress,
			Description:      indexSettingUserDataDescription + indexSettingsExtraClearingNote,
		},
	}, isVirtualIndex)
}

// extraIndexSettingsDataSourceSchema returns the schema of the newer index settings for the
// `algolia_index` and `algolia_virtual_index` data sources.
func extraIndexSettingsDataSourceSchema(isVirtualIndex bool) map[string]*schema.Schema {
	return withoutVirtualIndexForbiddenSettings(map[string]*schema.Schema{
		indexSettingModeKey: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: indexSettingModeDescription,
		},
		indexSettingSemanticSearchKey: {
			Type:        schema.TypeList,
			Computed:    true,
			Description: indexSettingSemanticSearchDescription,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"event_sources": {
						Type:        schema.TypeSet,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Set:         schema.HashString,
						Computed:    true,
						Description: indexSettingEventSourcesDescription,
					},
				},
			},
		},
		indexSettingEnableReRankingKey: {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: indexSettingEnableReRankingDescription,
		},
		indexSettingReRankingApplyFilterKey: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: indexSettingReRankingApplyFilterDescription,
		},
		indexSettingRenderingContentKey: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: indexSettingRenderingContentDescription,
		},
		indexSettingUserDataKey: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: indexSettingUserDataDescription,
		},
	}, isVirtualIndex)
}

// withoutVirtualIndexForbiddenSettings removes the settings that Algolia rejects on virtual
// replicas ("The following settings are forbidden in virtual replicas: userData").
func withoutVirtualIndexForbiddenSettings(s map[string]*schema.Schema, isVirtualIndex bool) map[string]*schema.Schema {
	if isVirtualIndex {
		delete(s, indexSettingUserDataKey)
	}
	return s
}

// withExtraIndexSettings adds the given extra settings schema to an index schema.
func withExtraIndexSettings(s map[string]*schema.Schema, extra map[string]*schema.Schema) map[string]*schema.Schema {
	maps.Copy(s, extra)
	return s
}

// marshalExtraIndexSettings converts the newer settings returned by Algolia into state values.
func marshalExtraIndexSettings(settings *search.SettingsResponse, isVirtualIndex bool) (map[string]interface{}, error) {
	values := map[string]interface{}{
		indexSettingModeKey:                 "",
		indexSettingSemanticSearchKey:       []interface{}{},
		indexSettingEnableReRankingKey:      derefBool(settings.EnableReRanking, false),
		indexSettingReRankingApplyFilterKey: "",
		indexSettingRenderingContentKey:     "",
		indexSettingUserDataKey:             "",
	}

	if settings.Mode != nil {
		values[indexSettingModeKey] = string(*settings.Mode)
	}
	if settings.SemanticSearch != nil {
		values[indexSettingSemanticSearchKey] = []interface{}{map[string]interface{}{
			"event_sources": settings.SemanticSearch.EventSources,
		}}
	}
	if filter := settings.ReRankingApplyFilter.Get(); filter != nil {
		s, err := reRankingApplyFilterToString(filter)
		if err != nil {
			return nil, err
		}
		values[indexSettingReRankingApplyFilterKey] = s
	}
	if settings.RenderingContent != nil {
		b, err := json.Marshal(settings.RenderingContent)
		if err != nil {
			return nil, fmt.Errorf("failed to encode renderingContent: %w", err)
		}
		values[indexSettingRenderingContentKey] = string(b)
	}
	if settings.UserData != nil {
		b, err := json.Marshal(settings.UserData)
		if err != nil {
			return nil, fmt.Errorf("failed to encode userData: %w", err)
		}
		values[indexSettingUserDataKey] = string(b)
	}
	if isVirtualIndex {
		delete(values, indexSettingUserDataKey)
	}

	return values, nil
}

// reRankingApplyFilterToString returns the filter as a string. Algolia stores a filter string
// verbatim as a one-element array (`"a AND b"` is returned as `["a AND b"]`), so that form is
// unwrapped. Other array forms can only be set outside Terraform; they are returned
// JSON-encoded so the value is still visible in state.
func reRankingApplyFilterToString(filter *search.ReRankingApplyFilter) (string, error) {
	if filter.String != nil {
		return *filter.String, nil
	}
	if filter.ArrayOfReRankingApplyFilter == nil {
		return "", nil
	}
	if arr := *filter.ArrayOfReRankingApplyFilter; len(arr) == 1 && arr[0].String != nil {
		return *arr[0].String, nil
	}
	b, err := json.Marshal(filter)
	if err != nil {
		return "", fmt.Errorf("failed to encode reRankingApplyFilter: %w", err)
	}
	return string(b), nil
}

// unmarshalExtraIndexSettings copies the newer settings that are present in the configuration
// into the settings sent to Algolia. Values are read from the raw configuration rather than
// with d.Get, because d.Get also returns values that only exist in state (these attributes are
// Computed) and returns the previous value for an explicitly configured empty string.
func unmarshalExtraIndexSettings(d *schema.ResourceData, settings *search.IndexSettings) error {
	return unmarshalExtraIndexSettingsFromConfig(d.GetRawConfig(), settings)
}

func unmarshalExtraIndexSettingsFromConfig(config cty.Value, settings *search.IndexSettings) error {
	if config.IsNull() || !config.IsKnown() {
		return nil
	}
	configured := func(key string) (cty.Value, bool) {
		if !config.Type().HasAttribute(key) {
			return cty.NilVal, false
		}
		v := config.GetAttr(key)
		return v, v.IsKnown() && !v.IsNull()
	}

	if v, ok := configured(indexSettingModeKey); ok {
		settings.Mode = utils.ToPtr(search.Mode(v.AsString()))
	}
	if v, ok := configured(indexSettingSemanticSearchKey); ok && v.LengthInt() > 0 {
		semanticSearch := search.SemanticSearch{EventSources: []string{}}
		if block := v.Index(cty.NumberIntVal(0)); !block.IsNull() {
			if sources := block.GetAttr("event_sources"); sources.IsKnown() && !sources.IsNull() {
				for it := sources.ElementIterator(); it.Next(); {
					_, source := it.Element()
					semanticSearch.EventSources = append(semanticSearch.EventSources, source.AsString())
				}
			}
		}
		settings.SemanticSearch = &semanticSearch
	}
	if v, ok := configured(indexSettingEnableReRankingKey); ok {
		settings.EnableReRanking = utils.ToPtr(v.True())
	}
	if v, ok := configured(indexSettingReRankingApplyFilterKey); ok {
		settings.ReRankingApplyFilter = *utils.NewNullable(search.StringAsReRankingApplyFilter(v.AsString()))
	}
	if v, ok := configured(indexSettingRenderingContentKey); ok {
		var renderingContent search.RenderingContent
		if err := json.Unmarshal([]byte(v.AsString()), &renderingContent); err != nil {
			return fmt.Errorf("failed to decode %s: %w", indexSettingRenderingContentKey, err)
		}
		settings.RenderingContent = &renderingContent
	}
	if v, ok := configured(indexSettingUserDataKey); ok {
		var userData interface{}
		if err := json.Unmarshal([]byte(v.AsString()), &userData); err != nil {
			return fmt.Errorf("failed to decode %s: %w", indexSettingUserDataKey, err)
		}
		settings.UserData = userData
	}

	return nil
}

// customizeDiffExtraIndexSettings plans an explicitly configured empty
// `re_ranking_apply_filter`. Terraform treats an empty string on an Optional+Computed attribute
// as unset and would otherwise keep the previous value, which makes the filter impossible to clear.
func customizeDiffExtraIndexSettings(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	config := d.GetRawConfig()
	if config.IsNull() || !config.Type().HasAttribute(indexSettingReRankingApplyFilterKey) {
		return nil
	}
	v := config.GetAttr(indexSettingReRankingApplyFilterKey)
	if !v.IsKnown() || v.IsNull() || v.AsString() != "" {
		return nil
	}
	if old, _ := d.GetChange(indexSettingReRankingApplyFilterKey); old.(string) != "" {
		return d.SetNew(indexSettingReRankingApplyFilterKey, "")
	}
	return nil
}

// diffEmptyJSONObjectSuppress is like diffJsonSuppress, but also treats an empty value as
// `{}`, since Algolia omits empty objects from the settings response.
func diffEmptyJSONObjectSuppress(k, old, new string, d *schema.ResourceData) bool {
	normalize := func(s string) string {
		if s == "" {
			return "{}"
		}
		return s
	}
	return diffJsonSuppress(k, normalize(old), normalize(new), d)
}

// computedListSettings are list and set attributes whose Algolia default is not empty
// (e.g. `attributes_to_retrieve` defaults to `["*"]`). They are Computed, and must only be sent
// when configured: sending the empty value of an unconfigured attribute overrides Algolia's
// default, e.g. `attributesToRetrieve: []` makes search return only `objectID`.
var computedListSettings = []struct {
	block, attribute string
	clear            func(*search.IndexSettings)
}{
	{"attributes_config", "attributes_to_retrieve", func(s *search.IndexSettings) { s.AttributesToRetrieve = nil }},
	{"ranking_config", "ranking", func(s *search.IndexSettings) { s.Ranking = nil }},
	{"query_strategy_config", "alternatives_as_exact", func(s *search.IndexSettings) { s.AlternativesAsExact = nil }},
	{"query_strategy_config", "advanced_syntax_features", func(s *search.IndexSettings) { s.AdvancedSyntaxFeatures = nil }},
	{"advanced_config", "response_fields", func(s *search.IndexSettings) { s.ResponseFields = nil }},
}

// dropUnconfiguredComputedListSettings removes the computedListSettings that are not in the
// configuration from the settings sent to Algolia, so the value stored in Algolia is kept.
func dropUnconfiguredComputedListSettings(d *schema.ResourceData, settings *search.IndexSettings) {
	config := d.GetRawConfig()
	for _, s := range computedListSettings {
		if !isNestedAttributeConfigured(config, s.block, s.attribute) {
			s.clear(settings)
		}
	}
}

// isNestedAttributeConfigured reports whether `<block>.0.<attribute>` is set in the configuration
// of a block with MaxItems 1.
func isNestedAttributeConfigured(config cty.Value, block, attribute string) bool {
	if config.IsNull() || !config.IsKnown() || !config.Type().HasAttribute(block) {
		return false
	}
	blocks := config.GetAttr(block)
	if blocks.IsNull() || !blocks.IsKnown() || blocks.LengthInt() == 0 {
		return false
	}
	b := blocks.Index(cty.NumberIntVal(0))
	if b.IsNull() || !b.Type().HasAttribute(attribute) {
		return false
	}
	v := b.GetAttr(attribute)
	return !v.IsNull()
}
