package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/transport"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

func setValues(d *schema.ResourceData, values map[string]interface{}) error {
	for k, v := range values {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func castStringList(list interface{}) []string {
	// we are initializing non nil array to be marshaled to [] in JSON
	strs := []string{}
	for _, v := range list.([]interface{}) {
		strs = append(strs, v.(string))
	}
	return strs
}

func castStringSet(set interface{}) []string {
	// we are initializing non nil array to be marshaled to [] in JSON
	strs := []string{}
	for _, v := range set.(*schema.Set).List() {
		strs = append(strs, v.(string))
	}
	return strs
}

func castStringMap(m interface{}) map[string]string {
	strMap := map[string]string{}
	for k, v := range m.(map[string]interface{}) {
		strMap[k] = v.(string)
	}
	return strMap
}

func castInterfaceMap(m interface{}) map[string]interface{} {
	interfaceMap := map[string]interface{}{}
	for k, v := range m.(map[string]interface{}) {
		interfaceMap[k] = v
	}
	return interfaceMap
}

func castKeyQueryParams(qp string) KeyQueryParams {

	type QueryParameters KeyQueryParams

	err = transport.URLDecode(
		[]byte(qp),
		&QueryParameters,
	)
	if err != nil {
		return fmt.Errorf("cannot decode QueryParameters %q: %v", tmp.QueryParameters, err)
	}

	return QueryParameters
}
