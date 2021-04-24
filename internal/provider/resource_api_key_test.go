package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccResourceAPIKey(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAPIKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("algolia_api_key.test", "key", regexp.MustCompile("^.{1,}$")),
					testCheckResourceListAttr("algolia_api_key.test", "acl", []string{"search"}),
					resource.TestCheckResourceAttr("algolia_api_key.test", "expires_at", "0"),
					resource.TestCheckResourceAttr("algolia_api_key.test", "max_hits_per_query", "0"),
					resource.TestCheckResourceAttr("algolia_api_key.test", "max_queries_per_ip_per_hour", "0"),
					resource.TestCheckNoResourceAttr("algolia_api_key.test", "indexes.0"),
					resource.TestCheckResourceAttr("algolia_api_key.test", "description", ""),
				),
			},
			{
				Config: testAccResourceAPIKeyUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("algolia_api_key.test", "key", regexp.MustCompile("^.{1,}$")),
					testCheckResourceListAttr("algolia_api_key.test", "acl", []string{"browse", "search"}),
					resource.TestCheckResourceAttr("algolia_api_key.test", "expires_at", "2524608000"),
					resource.TestCheckResourceAttr("algolia_api_key.test", "max_hits_per_query", "100"),
					resource.TestCheckResourceAttr("algolia_api_key.test", "max_queries_per_ip_per_hour", "10000"),
					testCheckResourceListAttr("algolia_api_key.test", "indexes", []string{"dev_*"}),
					testCheckResourceListAttr("algolia_api_key.test", "referers", []string{"https://algolia.com/\\*"}),
					resource.TestCheckResourceAttr("algolia_api_key.test", "description", "This is a test api key"),
				),
			},
		},
	})
}

const testAccResourceAPIKey = `
resource "algolia_api_key" "test" {
  acl = ["search"]
}
`

const testAccResourceAPIKeyUpdate = `
resource "algolia_api_key" "test" {
  acl                         = ["browse", "search"]
  expires_at                  = 2524608000 # 01 Jan 2050 00:00:00 GMT
  max_hits_per_query          = 100
  max_queries_per_ip_per_hour = 10000
  indexes                     = ["dev_*"]
  referers                    = ["https://algolia.com/\\*"]
  description                 = "This is a test api key"
}
`
