package algoliautil

import "github.com/algolia/algoliasearch-client-go/v3/algolia/region"

var ValidRegionStrings = []string{string(region.US), string(region.EU), string(region.DE)}

func IsValidRegion(r string) bool {
	for _, validRegionStr := range ValidRegionStrings {
		if r == validRegionStr {
			return true
		}
	}
	return false
}
