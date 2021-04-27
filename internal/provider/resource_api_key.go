package provider

import (
	"context"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strconv"
	"time"
)

func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		Description:   "A configuration for an API key",
		CreateContext: resourceAPIKeyCreate,
		ReadContext:   resourceAPIKeyRead,
		UpdateContext: resourceAPIKeyUpdate,
		DeleteContext: resourceAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAPIKeyStateContext,
		},
		// https://www.algolia.com/doc/api-reference/api-methods/add-api-key/
		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The created key.",
			},
			"acl": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
				Description: `Set of permissions associated with the key.
The possible ACLs are:
  - ` + "`search`" + `: allowed to perform search operations.
  - ` + "`browse`" + `: allowed to retrieve all index data with the browse endpoint.
  - ` + "`addObject`" + `: allowed to add or update a records in the index.
  - ` + "`deleteObject`" + `: allowed to delete an existing record.
  - ` + "`listIndexes`" + `: allowed to get a list of all existing indices.
  - ` + "`deleteIndex`" + `: allowed to delete an index.
  - ` + "`settings`" + `: allowed to read all index settings.
  - ` + "`editSettings`" + `: allowed to update all index settings.
  - ` + "`analytics`" + `: allowed to retrieve data with the Analytics API.
  - ` + "`recommendation`" + `: allowed to interact with the Recommendation API.
  - ` + "`usage`" + ` allowed to retrieve data with the Usage API.
  - ` + "`nluReadAnswers`" + `: allowed to perform semantic search with the Answers API.
  - ` + "`logs`" + `: allowed to query the logs.
  - ` + "`seeUnretrievableAttributes`" + `: allowed to retrieve unretrievableAttributes for all operations that return records.
`,
			},
			"expires_at": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "Unix timestamp of the date at which the key expires. RFC3339 format. Will not expire per default.",
			},
			"max_hits_per_query": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Maximum number of hits this API key can retrieve in one call. This parameter can be used to protect you from attempts at retrieving your entire index contents by massively querying the index.",
			},
			"max_queries_per_ip_per_hour": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				Description: `Maximum number of API calls allowed from an IP address per hour.Each time an API call is performed with this key, a check is performed. If the IP at the source of the call did more than this number of calls in the last hour, a 429 code is returned.

This parameter can be used to protect you from attempts at retrieving your entire index contents by massively querying the index.`,
			},
			"indexes": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "List of targeted indices. You can target all indices starting with a prefix or ending with a suffix using the ‘*’ character. For example, “dev_*” matches all indices starting with “dev_” and “*_dev” matches all indices ending with “_dev”.",
			},
			"referers": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "List of referrers that can perform an operation. You can use the “*” (asterisk) character as a wildcard to match subdomains, or all pages of a website. For example, `\"https://algolia.com/\\*\"` matches all referrers starting with `\"https://algolia.com/\"`, and `\"\\*.algolia.com\"` matches all referrers ending with `\".algolia.com\"`. If you want to allow all possible referrers from the `algolia.com` domain, you can use `\"\\*algolia.com/\\*\"`.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the API key.",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The unix time at which the key has been created.",
			},
		},
	}
}

func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	res, err := apiClient.searchClient.AddAPIKey(mapToAPIKey(d), ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("key", res.Key); err != nil {
		return diag.FromErr(err)
	}

	return resourceAPIKeyRead(ctx, d, m)
}

func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := refreshAPIKeyState(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	res, err := apiClient.searchClient.UpdateAPIKey(mapToAPIKey(d))
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	return resourceAPIKeyRead(ctx, d, m)
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	res, err := apiClient.searchClient.DeleteAPIKey(d.Get("key").(string), ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = res.Wait(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAPIKeyStateContext(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := d.Set("key", d.Id()); err != nil {
		return nil, err
	}

	if err := refreshAPIKeyState(ctx, d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func refreshAPIKeyState(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*apiClient)

	keyID := d.Get("key").(string)
	key, err := apiClient.searchClient.GetAPIKey(keyID, ctx)
	if err != nil {
		d.SetId("")
		return err
	}

	d.SetId(strconv.FormatInt(key.CreatedAt.Unix(), 10))

	values := map[string]interface{}{
		"key":                         keyID,
		"acl":                         key.ACL,
		"max_hits_per_query":          key.MaxHitsPerQuery,
		"max_queries_per_ip_per_hour": key.MaxQueriesPerIPPerHour,
		"referers":                    key.Referers,
		"description":                 key.Description,
		"indexes":                     key.Indexes,
		"created_at":                  key.CreatedAt.Unix(),
	}
	// we can't set from key.Validity since it is remaining valid time and the value changes every second.
	// TODO: fix to work with import
	if expiresAtRFC3339, ok := d.GetOk("expires_at"); ok {
		values["expires_at"] = expiresAtRFC3339
	}
	if err := setValues(d, values); err != nil {
		return err
	}

	return nil
}

func mapToAPIKey(d *schema.ResourceData) search.Key {
	var validity time.Duration
	if expiresAtRFC3339, ok := d.GetOk("expires_at"); ok && expiresAtRFC3339 != "" {
		t, _ := time.Parse(time.RFC3339, expiresAtRFC3339.(string))
		validity = time.Duration(int(t.Unix())-int(time.Now().Unix())) * time.Second
	}

	return search.Key{
		Value:                  d.Get("key").(string),
		ACL:                    castStringSet(d.Get("acl")),
		Validity:               validity,
		MaxHitsPerQuery:        d.Get("max_hits_per_query").(int),
		MaxQueriesPerIPPerHour: d.Get("max_queries_per_ip_per_hour").(int),
		Indexes:                castStringSet(d.Get("indexes")),
		Referers:               castStringSet(d.Get("referers")),
		Description:            d.Get("description").(string),
	}
}
