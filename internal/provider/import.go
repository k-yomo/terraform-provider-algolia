package provider

import (
	"fmt"
	"strings"

	suggestions "github.com/algolia/algoliasearch-client-go/v4/algolia/query-suggestions"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
)

// parseImportRegionAndId will parse either {{id}} or {{region}}/{{id}} format import id.
func parseImportRegionAndId(id string) (suggestions.Region, string, error) {
	ids := strings.Split(id, "/")
	if len(ids) > 2 {
		return "", "", fmt.Errorf("'%s' is invalid format for import id. it must be either '{id}' or '{region}/{id}'", id)
	}
	if len(ids) == 1 {
		return "", id, nil
	}
	r := ids[0]
	if algoliautil.IsValidRegion(ids[0]) {
		return suggestions.Region(ids[0]), ids[1], nil
	} else {
		return "", "", fmt.Errorf("'%s' is invalid region, it must be either 'us', 'eu' or 'de'", r)
	}
}
