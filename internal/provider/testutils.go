package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"math/rand"
	"time"
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
	rand.Seed(time.Now().UnixNano())
	return acctest.RandStringFromCharSet(1, acctest.CharSetAlpha) + acctest.RandStringFromCharSet(length-1, acctest.CharSetAlphaNum)
}
