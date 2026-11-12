package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
)

func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAPIKeyCreate,
		ReadContext:   resourceAPIKeyRead,
		UpdateContext: resourceAPIKeyUpdate,
		DeleteContext: resourceAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAPIKeyStateContext,
		},
		Description: "A configuration for an API key",
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
				Description: "List of targeted indices. You can target all indices starting with a prefix or ending with a suffix using the \u2018*\u2019 character. For example, \u201cdev_*\u201d matches all indices starting with \u201cdev_\u201d and \u201c*_dev\u201d matches all indices ending with \u201c_dev\u201d.",
			},
			"referers": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "List of referrers that can perform an operation. You can use the \"*\" (asterisk) character as a wildcard to match subdomains, or all pages of a website. For example, `\"https://algolia.com/\\*\"` matches all referrers starting with `\"https://algolia.com/\"`, and `\"\\*.algolia.com\"` matches all referrers ending with `\".algolia.com\"`. If you want to allow all possible referrers from the `algolia.com` domain, you can use `\"\\*algolia.com/\\*\"`.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the API key.",
			},
			"query_parameters": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) ([]string, []error) {
					if _, err := url.ParseQuery(v.(string)); err != nil {
						return nil, []error{fmt.Errorf("%q must be a URL-encoded query string: %w", k, err)}
					}
					return nil, nil
				},
				DiffSuppressFunc: func(_, oldValue, newValue string, _ *schema.ResourceData) bool {
					return queryParametersEqual(oldValue, newValue)
				},
				Description: "Search parameters enforced for every query made with this API key, as a URL-encoded query string. For example, `\"typoTolerance=strict&restrictSources=192.168.1.0/24\"`. Parameters set here can't be overridden by the query.",
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

	apiKey := mapToAPIKey(d)
	res, err := apiClient.searchClient.AddApiKey(
		apiClient.searchClient.NewApiAddApiKeyRequest(apiKey),
		search.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = apiClient.searchClient.WaitForApiKey(res.Key, search.API_KEY_OPERATION_ADD, search.WithContext(ctx)); err != nil {
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

	keyValue := d.Get("key").(string)
	apiKey := mapToAPIKey(d)
	_, err := apiClient.searchClient.UpdateApiKey(
		apiClient.searchClient.NewApiUpdateApiKeyRequest(keyValue, apiKey),
		search.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	// Custom wait loop for UPDATE: the built-in WaitForApiKey compares Validity
	// which is seconds-until-expiry and changes every second, causing infinite retries.
	// We poll until the non-Validity fields match.
	if err = waitForAPIKeyUpdate(ctx, apiClient.searchClient, keyValue, apiKey); err != nil {
		return diag.FromErr(err)
	}

	return resourceAPIKeyRead(ctx, d, m)
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	keyStr := d.Get("key").(string)
	_, err := apiClient.searchClient.DeleteApiKey(
		apiClient.searchClient.NewApiDeleteApiKeyRequest(keyStr),
		search.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = apiClient.searchClient.WaitForApiKey(keyStr, search.API_KEY_OPERATION_DELETE, search.WithContext(ctx)); err != nil {
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
	key, err := apiClient.searchClient.GetApiKey(
		apiClient.searchClient.NewApiGetApiKeyRequest(keyID),
		search.WithContext(ctx),
	)
	if err != nil {
		if algoliautil.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("api key (%s) not found, removing from state", d.Id()))
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId(strconv.FormatInt(key.CreatedAt, 10))

	acl := make([]string, len(key.Acl))
	for i, a := range key.Acl {
		acl[i] = string(a)
	}

	values := map[string]interface{}{
		"key":                         keyID,
		"acl":                         acl,
		"max_hits_per_query":          int(key.GetMaxHitsPerQuery()),
		"max_queries_per_ip_per_hour": int(key.GetMaxQueriesPerIPPerHour()),
		"referers":                    key.Referers,
		"description":                 key.GetDescription(),
		"query_parameters":            key.GetQueryParameters(),
		"indexes":                     key.Indexes,
		"created_at":                  key.CreatedAt,
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

// waitForAPIKeyUpdate polls until the API key reflects the update, with a bounded number of retries.
//
// The client's built-in waiters compare the whole key exactly. That never succeeds when Algolia
// normalizes a field (e.g. reorders `indexes` or returns [] for an unset list), and older versions
// of this provider hung forever because of it (#158, #278). Here only the fields this provider
// manages are compared, order-insensitively, and Validity (seconds until expiry, which changes
// every second) is ignored.
func waitForAPIKeyUpdate(ctx context.Context, client *search.APIClient, key string, expected *search.ApiKey) error {
	const maxRetries = 50
	for i := 0; i < maxRetries; i++ {
		resp, err := client.GetApiKey(client.NewApiGetApiKeyRequest(key), search.WithContext(ctx))
		if err != nil {
			return err
		}
		if apiKeyUpdateApplied(expected, resp) {
			return nil
		}

		delay := time.Duration(min(200*(i+1), 5000)) * time.Millisecond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return fmt.Errorf("timed out waiting for API key update to be applied after %d retries; run `terraform refresh` to sync the state with the actual key", maxRetries)
}

// apiKeyUpdateApplied reports whether actual reflects every field of expected that this provider manages.
func apiKeyUpdateApplied(expected *search.ApiKey, actual *search.GetApiKeyResponse) bool {
	expectedACL := make([]string, len(expected.Acl))
	for i, a := range expected.Acl {
		expectedACL[i] = string(a)
	}
	actualACL := make([]string, len(actual.Acl))
	for i, a := range actual.Acl {
		actualACL[i] = string(a)
	}

	return stringSetsEqual(expectedACL, actualACL) &&
		stringSetsEqual(expected.Indexes, actual.Indexes) &&
		stringSetsEqual(expected.Referers, actual.Referers) &&
		expected.GetDescription() == actual.GetDescription() &&
		queryParametersEqual(expected.GetQueryParameters(), actual.GetQueryParameters()) &&
		expected.GetMaxHitsPerQuery() == actual.GetMaxHitsPerQuery() &&
		expected.GetMaxQueriesPerIPPerHour() == actual.GetMaxQueriesPerIPPerHour()
}

// stringSetsEqual compares two string slices as sets; nil and empty are equal.
func stringSetsEqual(a, b []string) bool {
	setA := make(map[string]struct{}, len(a))
	for _, v := range a {
		setA[v] = struct{}{}
	}
	setB := make(map[string]struct{}, len(b))
	for _, v := range b {
		setB[v] = struct{}{}
	}
	if len(setA) != len(setB) {
		return false
	}
	for v := range setA {
		if _, ok := setB[v]; !ok {
			return false
		}
	}
	return true
}

// queryParametersEqual compares two URL-encoded query strings by their decoded parameters,
// so that parameter order and encoding differences don't cause a diff.
func queryParametersEqual(a, b string) bool {
	if a == b {
		return true
	}
	valuesA, errA := url.ParseQuery(a)
	valuesB, errB := url.ParseQuery(b)
	if errA != nil || errB != nil {
		return false
	}
	if len(valuesA) == 0 && len(valuesB) == 0 {
		return true
	}
	return reflect.DeepEqual(valuesA, valuesB)
}

func mapToAPIKey(d *schema.ResourceData) *search.ApiKey {
	var validity *int32
	if expiresAtRFC3339, ok := d.GetOk("expires_at"); ok && expiresAtRFC3339 != "" {
		t, _ := time.Parse(time.RFC3339, expiresAtRFC3339.(string))
		v := int32(t.Unix() - time.Now().Unix())
		validity = &v
	}

	aclStrings := castStringSet(d.Get("acl"))
	acl := make([]search.Acl, len(aclStrings))
	for i, s := range aclStrings {
		acl[i] = search.Acl(s)
	}

	maxHitsPerQuery := int32(d.Get("max_hits_per_query").(int))
	maxQueriesPerIPPerHour := int32(d.Get("max_queries_per_ip_per_hour").(int))
	description := d.Get("description").(string)
	// Always send the value, even when empty, so that removing it from the config clears it.
	queryParameters := d.Get("query_parameters").(string)

	apiKey := &search.ApiKey{
		Acl:                    acl,
		Description:            &description,
		Indexes:                castStringSet(d.Get("indexes")),
		MaxHitsPerQuery:        &maxHitsPerQuery,
		MaxQueriesPerIPPerHour: &maxQueriesPerIPPerHour,
		QueryParameters:        &queryParameters,
		Referers:               castStringSet(d.Get("referers")),
		Validity:               validity,
	}

	return apiKey
}
