package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

// The first character must be alphabet for algolia resources
func randStringStartWithAlpha(length int) string {
	const uuidLen = 21 // firstChar + xid(20 chars)
	uuid := acctest.RandStringFromCharSet(1, acctest.CharSetAlpha) + xid.New().String()

	if length < uuidLen {
		return uuid[:length]
	}

	return uuid + acctest.RandStringFromCharSet(length-uuidLen, acctest.CharSetAlphaNum)
}
