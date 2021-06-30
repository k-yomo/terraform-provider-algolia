package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	var strs []string
	for _, v := range list.([]interface{}) {
		strs = append(strs, v.(string))
	}
	return strs
}

func castStringSet(set interface{}) []string {
	var strs []string
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
	strMap := map[string]interface{}{}
	for k, v := range m.(map[string]interface{}) {
		strMap[k] = v
	}
	return strMap
}
