package provider

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDnsAAAARecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsAAAARecordSetRead,
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Host to look up.",
			},
			"addrs": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of IP addresses. IP addresses are always sorted to avoid constant changing plans.",
			},
		},
		Description: "Use this data source to get DNS AAAA records of the host.",
	}
}

func dataSourceDnsAAAARecordSetRead(d *schema.ResourceData, meta interface{}) error {
	//nolint:forcetypeassert
	host := d.Get("host").(string)

	_, aaaa, err := lookupIP(host)
	if err != nil {
		return fmt.Errorf("error looking up AAAA records for %q: %s", host, err)
	}
	sort.Strings(aaaa)

	//nolint:errcheck
	d.Set("addrs", aaaa)
	d.SetId(host)

	return nil
}
