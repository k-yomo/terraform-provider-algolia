package provider

import (
	"fmt"
	"strings"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/region"
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
	r := region.Region(ids[0])
	switch r {
	case region.US, region.EU, region.DE:
		return r, ids[1], nil
	default:
		return "", "", fmt.Errorf("'%s' is invalid region, it must be either 'us', 'eu' or 'de'", r)
	}
}
