package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Unit tests (no Algolia credentials needed) for mapping rule config to the API payload.
func TestMapToRule_ConsequenceParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		raw        map[string]interface{}
		wantParams string // JSON of consequence.params; "" means no params
	}{
		{
			name: "deprecated params block with string query",
			raw: map[string]interface{}{
				"index_name": "idx", "object_id": "r1",
				"consequence": []interface{}{map[string]interface{}{
					"params": []interface{}{map[string]interface{}{"query": "budget shoes"}},
				}},
			},
			wantParams: `{"query":"budget shoes"}`,
		},
		{
			name: "deprecated params block with object query and facet filters",
			raw: map[string]interface{}{
				"index_name": "idx", "object_id": "r1",
				"consequence": []interface{}{map[string]interface{}{
					"params": []interface{}{map[string]interface{}{
						"object_query": []interface{}{map[string]interface{}{"type": "remove", "delete": "cheap"}},
						"automatic_facet_filters": []interface{}{map[string]interface{}{
							"facet": "brand", "score": 2, "disjunctive": true,
						}},
					}},
				}},
			},
			wantParams: `{"query":{"edits":[{"type":"remove","delete":"cheap"}]},"automaticFacetFilters":[{"facet":"brand","score":2,"disjunctive":true}]}`,
		},
		{
			name: "params_json",
			raw: map[string]interface{}{
				"index_name": "idx", "object_id": "r1",
				"consequence": []interface{}{map[string]interface{}{
					"params_json": `{"filters":"brand:apple"}`,
				}},
			},
			wantParams: `{"filters":"brand:apple"}`,
		},
		{
			name: "no params",
			raw: map[string]interface{}{
				"index_name": "idx", "object_id": "r1",
				"consequence": []interface{}{map[string]interface{}{
					"hide": []interface{}{"obj-1"},
				}},
			},
			wantParams: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := schema.TestResourceDataRaw(t, resourceRule().Schema, tt.raw)
			rule, err := mapToRule(d)
			if err != nil {
				t.Fatalf("mapToRule() error = %v", err)
			}
			params := rule.Consequence.Params
			if tt.wantParams == "" {
				if params != nil {
					got, _ := json.Marshal(params)
					t.Fatalf("consequence.params = %s, want none", got)
				}
				return
			}
			if params == nil {
				t.Fatalf("consequence.params = nil, want %s", tt.wantParams)
			}
			got, err := json.Marshal(params)
			if err != nil {
				t.Fatal(err)
			}
			equal, err := jsonBytesEqual([]byte(tt.wantParams), got)
			if err != nil {
				t.Fatal(err)
			}
			if !equal {
				t.Errorf("consequence.params = %s, want %s", got, tt.wantParams)
			}
		})
	}
}
