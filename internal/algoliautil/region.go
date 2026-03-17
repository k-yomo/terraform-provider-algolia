package algoliautil

import "slices"

var ValidRegionStrings = []string{"us", "eu", "de"}

func IsValidRegion(r string) bool {
	return slices.Contains(ValidRegionStrings, r)
}
