package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
	"github.com/rs/xid"
)

func testCheckResourceListAttr(name, key string, values []string) resource.TestCheckFunc {
	var testCheckFuncs []resource.TestCheckFunc
	resource.ComposeTestCheckFunc()
	for i, v := range values {
		testCheckFuncs = append(testCheckFuncs, resource.TestCheckResourceAttr(name, fmt.Sprintf("%s.%d", key, i), v))
	}
	return resource.ComposeTestCheckFunc(testCheckFuncs...)
}

// randResourceID generates unique id string
// id length must be longer than (prefix + uuid length)
func randResourceID(length int) string {
	// The first character must be alphabet for algolia resources
	uuid := algoliautil.TestIndexNamePrefix + xid.New().String()

	if length < len(uuid) {
		panic(fmt.Sprintf("length must be longer than %d", len(uuid)))
	}

	return uuid + acctest.RandStringFromCharSet(length-len(uuid), acctest.CharSetAlphaNum)
}
