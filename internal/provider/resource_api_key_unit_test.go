package provider

import (
	"testing"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/algolia/algoliasearch-client-go/v4/algolia/utils"
)

func TestAPIKeyUpdateApplied(t *testing.T) {
	t.Parallel()

	desc := "test key"
	maxHits, maxQueries := int32(100), int32(1000)
	expected := &search.ApiKey{
		Acl:                    []search.Acl{"addObject", "browse"},
		Indexes:                []string{"events", "events_tmp_*"},
		Description:            &desc,
		MaxHitsPerQuery:        &maxHits,
		MaxQueriesPerIPPerHour: &maxQueries,
	}
	matching := func() *search.GetApiKeyResponse {
		return &search.GetApiKeyResponse{
			Acl:                    []search.Acl{"addObject", "browse"},
			Indexes:                []string{"events", "events_tmp_*"},
			Referers:               []string{},
			Description:            &desc,
			MaxHitsPerQuery:        &maxHits,
			MaxQueriesPerIPPerHour: &maxQueries,
			Validity:               utils.ToPtr[int32](12345),
		}
	}

	tests := []struct {
		name   string
		mutate func(r *search.GetApiKeyResponse)
		want   bool
	}{
		{name: "identical", mutate: func(r *search.GetApiKeyResponse) {}, want: true},
		{
			name: "different order of acl and indexes (e.g. wildcards reordered, #278)",
			mutate: func(r *search.GetApiKeyResponse) {
				r.Acl = []search.Acl{"browse", "addObject"}
				r.Indexes = []string{"events_tmp_*", "events"}
			},
			want: true,
		},
		{name: "nil vs empty referers", mutate: func(r *search.GetApiKeyResponse) { r.Referers = nil }, want: true},
		{name: "validity is ignored", mutate: func(r *search.GetApiKeyResponse) { r.Validity = utils.ToPtr[int32](1) }, want: true},
		{name: "acl not yet updated (#158)", mutate: func(r *search.GetApiKeyResponse) { r.Acl = []search.Acl{"addObject"} }, want: false},
		{name: "indexes not yet updated", mutate: func(r *search.GetApiKeyResponse) { r.Indexes = []string{"events"} }, want: false},
		{name: "extra referer", mutate: func(r *search.GetApiKeyResponse) { r.Referers = []string{"https://example.com/*"} }, want: false},
		{name: "description differs", mutate: func(r *search.GetApiKeyResponse) { r.Description = utils.ToPtr("old") }, want: false},
		{name: "max hits differs", mutate: func(r *search.GetApiKeyResponse) { r.MaxHitsPerQuery = utils.ToPtr[int32](1) }, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := matching()
			tt.mutate(actual)
			if got := apiKeyUpdateApplied(expected, actual); got != tt.want {
				t.Errorf("apiKeyUpdateApplied() = %v, want %v", got, tt.want)
			}
		})
	}
}
