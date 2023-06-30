resource "algolia_api_key" "example" {
  acl                         = ["search", "browse"]
  expires_at                  = "2030-01-01T00:00:00Z"
  max_hits_per_query          = 100
  max_queries_per_ip_per_hour = 10000
  description                 = "This is a example api key"
  indexes                     = ["*"]
  referers                    = ["https://algolia.com/\\*"]
  query_parameters            = ""
}
