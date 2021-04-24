package provider

import (
	"context"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Description: "The created key.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"acl": {
				Description: "Set of permissions associated with the key.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Required:    true,
			},
			"expires_at": {
				Description: "Unix timestamp of the date at which the key expires. A value of 0 means the API key doesn’t expire.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"max_hits_per_query": {
				Description: "Maximum number of hits this API key can retrieve in one call.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"max_queries_per_ip_per_hour": {
				Description: "Maximum number of API calls allowed from an IP address per hour.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"indexes": {
				Description: "List of targeted indices. You can target all indices starting with a prefix or ending with a suffix using the ‘*’ character.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
			},
			"referers": {
				Description: "List of referrers that can perform an operation. You can use the “*” (asterisk) character as a wildcard to match subdomains, or all pages of a website.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "Description of the API key.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"created_at": {
				Description: "The unix time at which the key has been created.",
				Type:        schema.TypeInt,
				Computed:    true,
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

	d.SetId(strconv.FormatInt(res.CreatedAt.Unix(), 10))

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
	key, err := apiClient.searchClient.GetAPIKey(d.Get("key").(string), ctx)
	if err != nil {
		d.SetId("")
		return err
	}

	values := map[string]interface{}{
		"key":                         d.Get("key"),
		"acl":                         key.ACL,
		"expires_at":                  d.Get("expires_at"), // we don't set from key.Validity since it is remaining valid time and it's unstable
		"max_hits_per_query":          key.MaxHitsPerQuery,
		"max_queries_per_ip_per_hour": key.MaxQueriesPerIPPerHour,
		"referers":                    key.Referers,
		"description":                 key.Description,
		"indexes":                     key.Indexes,
		"created_at":                  key.CreatedAt.Unix(),
	}
	if err := setValues(d, values); err != nil {
		return err
	}

	return nil
}

func mapToAPIKey(d *schema.ResourceData) search.Key {
	var acl []string
	if value, ok := d.GetOk("acl"); ok {
		for _, v := range value.(*schema.Set).List() {
			acl = append(acl, v.(string))
		}
	}

	var indexes []string
	if value, ok := d.GetOk("indexes"); ok {
		for _, v := range value.(*schema.Set).List() {
			indexes = append(indexes, v.(string))
		}
	}

	var referers []string
	if value, ok := d.GetOk("referers"); ok {
		for _, v := range value.(*schema.Set).List() {
			referers = append(referers, v.(string))
		}
	}

	var validity time.Duration
	if expiresAt, ok := d.GetOk("expires_at"); ok {
		validity = time.Duration(expiresAt.(int)-int(time.Now().Unix())) * time.Second
	}

	return search.Key{
		Value:                  d.Get("key").(string),
		ACL:                    acl,
		Validity:               validity,
		MaxHitsPerQuery:        d.Get("max_hits_per_query").(int),
		MaxQueriesPerIPPerHour: d.Get("max_queries_per_ip_per_hour").(int),
		Indexes:                indexes,
		Referers:               referers,
		Description:            d.Get("description").(string),
	}
}
