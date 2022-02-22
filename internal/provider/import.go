package provider

import (
	"fmt"
	"strings"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/region"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
)

// parseImportRegionAndId will parse either {{id}} or {{region}}/{{id}} format import id.
func parseImportRegionAndId(id string) (region.Region, string, error) {
	ids := strings.Split(id, "/")
	if len(ids) > 2 {
		return "", "", fmt.Errorf("'%s' is invalid format for import id. it must be either '{id}' or '{region}/{id}'", id)
	}
	if len(ids) == 1 {
		return "", id, nil
	}
	r := ids[0]
	if algoliautil.IsValidRegion(ids[0]) {
		return region.Region(ids[0]), ids[1], nil
	} else {
		return "", "", fmt.Errorf("'%s' is invalid region, it must be either 'us', 'eu' or 'de'", r)
	}
}
