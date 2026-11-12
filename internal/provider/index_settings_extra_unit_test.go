package provider

import (
	"encoding/json"
	"testing"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/algolia/algoliasearch-client-go/v4/algolia/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
)

func TestUnmarshalExtraIndexSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config cty.Value
		want   search.IndexSettings
	}{
		{
			name:   "nothing configured sends nothing",
			config: cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("test")}),
			want:   search.IndexSettings{},
		},
		{
			name: "all settings configured",
			config: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("test"),
				"mode": cty.StringVal("neuralSearch"),
				"semantic_search": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"event_sources": cty.SetVal([]cty.Value{cty.StringVal("products")}),
				})}),
				"enable_re_ranking":       cty.True,
				"re_ranking_apply_filter": cty.StringVal("brand:Apple"),
				"rendering_content_json":  cty.StringVal(`{"facetOrdering":{"facets":{"order":["brand"]}}}`),
				"user_data_json":          cty.StringVal(`{"team":"search"}`),
			}),
			want: search.IndexSettings{
				Mode:                 utils.ToPtr(search.MODE_NEURAL_SEARCH),
				SemanticSearch:       &search.SemanticSearch{EventSources: []string{"products"}},
				EnableReRanking:      utils.ToPtr(true),
				ReRankingApplyFilter: *utils.NewNullable(search.StringAsReRankingApplyFilter("brand:Apple")),
				RenderingContent: &search.RenderingContent{
					FacetOrdering: &search.FacetOrdering{Facets: &search.Facets{Order: []string{"brand"}}},
				},
				UserData: map[string]interface{}{"team": "search"},
			},
		},
		{
			name: "false and empty values are sent",
			config: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("test"),
				"semantic_search": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"event_sources": cty.SetValEmpty(cty.String),
				})}),
				"enable_re_ranking":       cty.False,
				"re_ranking_apply_filter": cty.StringVal(""),
			}),
			want: search.IndexSettings{
				SemanticSearch:       &search.SemanticSearch{EventSources: []string{}},
				EnableReRanking:      utils.ToPtr(false),
				ReRankingApplyFilter: *utils.NewNullable(search.StringAsReRankingApplyFilter("")),
			},
		},
		{
			name: "null attributes (not in the configuration) are not sent",
			config: cty.ObjectVal(map[string]cty.Value{
				"mode":                    cty.NullVal(cty.String),
				"semantic_search":         cty.ListValEmpty(cty.Object(map[string]cty.Type{"event_sources": cty.Set(cty.String)})),
				"enable_re_ranking":       cty.NullVal(cty.Bool),
				"re_ranking_apply_filter": cty.NullVal(cty.String),
				"rendering_content_json":  cty.NullVal(cty.String),
				"user_data_json":          cty.NullVal(cty.String),
			}),
			want: search.IndexSettings{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := search.IndexSettings{}
			if err := unmarshalExtraIndexSettingsFromConfig(tt.config, &got); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(mustMarshalJSON(t, tt.want), mustMarshalJSON(t, got)); diff != "" {
				t.Errorf("unmarshalExtraIndexSettings() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMarshalExtraIndexSettings(t *testing.T) {
	t.Parallel()

	nested := []search.ReRankingApplyFilter{
		*search.ArrayOfReRankingApplyFilterAsReRankingApplyFilter([]search.ReRankingApplyFilter{
			*search.StringAsReRankingApplyFilter("brand:Apple"),
			*search.StringAsReRankingApplyFilter("brand:Samsung"),
		}),
		*search.StringAsReRankingApplyFilter("color:red"),
	}
	tests := []struct {
		name           string
		settings       search.SettingsResponse
		isVirtualIndex bool
		want           map[string]interface{}
	}{
		{
			name:     "unset settings",
			settings: search.SettingsResponse{},
			want: map[string]interface{}{
				"mode":                    "",
				"semantic_search":         []interface{}{},
				"enable_re_ranking":       false,
				"re_ranking_apply_filter": "",
				"rendering_content_json":  "",
				"user_data_json":          "",
			},
		},
		{
			name: "set settings",
			settings: search.SettingsResponse{
				Mode:            utils.ToPtr(search.MODE_KEYWORD_SEARCH),
				SemanticSearch:  &search.SemanticSearch{EventSources: []string{"products"}},
				EnableReRanking: utils.ToPtr(true),
				// Algolia stores a filter string as a one-element array.
				ReRankingApplyFilter: *utils.NewNullable(search.ArrayOfReRankingApplyFilterAsReRankingApplyFilter(
					[]search.ReRankingApplyFilter{*search.StringAsReRankingApplyFilter("brand:Apple AND color:red")},
				)),
				RenderingContent: &search.RenderingContent{
					FacetOrdering: &search.FacetOrdering{Facets: &search.Facets{Order: []string{"brand"}}},
				},
				UserData: map[string]interface{}{"team": "search"},
			},
			want: map[string]interface{}{
				"mode":                    "keywordSearch",
				"semantic_search":         []interface{}{map[string]interface{}{"event_sources": []string{"products"}}},
				"enable_re_ranking":       true,
				"re_ranking_apply_filter": "brand:Apple AND color:red",
				"rendering_content_json":  `{"facetOrdering":{"facets":{"order":["brand"]}}}`,
				"user_data_json":          `{"team":"search"}`,
			},
		},
		{
			name: "nested filter set outside Terraform is kept visible",
			settings: search.SettingsResponse{
				ReRankingApplyFilter: *utils.NewNullable(search.ArrayOfReRankingApplyFilterAsReRankingApplyFilter(nested)),
			},
			want: map[string]interface{}{
				"mode":                    "",
				"semantic_search":         []interface{}{},
				"enable_re_ranking":       false,
				"re_ranking_apply_filter": `[["brand:Apple","brand:Samsung"],"color:red"]`,
				"rendering_content_json":  "",
				"user_data_json":          "",
			},
		},
		{
			name:           "virtual index has no user data",
			settings:       search.SettingsResponse{UserData: map[string]interface{}{"team": "search"}},
			isVirtualIndex: true,
			want: map[string]interface{}{
				"mode":                    "",
				"semantic_search":         []interface{}{},
				"enable_re_ranking":       false,
				"re_ranking_apply_filter": "",
				"rendering_content_json":  "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := marshalExtraIndexSettings(&tt.settings, tt.isVirtualIndex)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("marshalExtraIndexSettings() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExtraIndexSettingsSchema_virtualIndexHasNoUserData(t *testing.T) {
	t.Parallel()

	// Algolia rejects userData on virtual replicas.
	if _, ok := resourceVirtualIndex().Schema["user_data_json"]; ok {
		t.Error("algolia_virtual_index resource must not have user_data_json")
	}
	if _, ok := dataSourceVirtualIndex().Schema["user_data_json"]; ok {
		t.Error("algolia_virtual_index data source must not have user_data_json")
	}
	if _, ok := resourceIndex().Schema["user_data_json"]; !ok {
		t.Error("algolia_index resource must have user_data_json")
	}
}

func mustMarshalJSON(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestIsNestedAttributeConfigured(t *testing.T) {
	t.Parallel()

	blockType := cty.Object(map[string]cty.Type{"attributes_to_retrieve": cty.Set(cty.String), "searchable_attributes": cty.List(cty.String)})
	withBlock := func(attrs map[string]cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{"attributes_config": cty.ListVal([]cty.Value{cty.ObjectVal(attrs)})})
	}
	tests := []struct {
		name   string
		config cty.Value
		want   bool
	}{
		{"null config", cty.NullVal(cty.Object(map[string]cty.Type{"attributes_config": cty.List(blockType)})), false},
		{"no block", cty.ObjectVal(map[string]cty.Value{"attributes_config": cty.ListValEmpty(blockType)}), false},
		{"block without attribute", withBlock(map[string]cty.Value{
			"attributes_to_retrieve": cty.NullVal(cty.Set(cty.String)),
			"searchable_attributes":  cty.ListVal([]cty.Value{cty.StringVal("title")}),
		}), false},
		{"attribute set", withBlock(map[string]cty.Value{
			"attributes_to_retrieve": cty.SetVal([]cty.Value{cty.StringVal("title")}),
			"searchable_attributes":  cty.NullVal(cty.List(cty.String)),
		}), true},
		{"attribute explicitly empty", withBlock(map[string]cty.Value{
			"attributes_to_retrieve": cty.SetValEmpty(cty.String),
			"searchable_attributes":  cty.NullVal(cty.List(cty.String)),
		}), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isNestedAttributeConfigured(tt.config, "attributes_config", "attributes_to_retrieve"); got != tt.want {
				t.Errorf("isNestedAttributeConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}
