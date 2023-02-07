package provider

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDnsARecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsARecordSetRead,
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
		Description: "Use this data source to get DNS A records of the host.",
	}
}

func dataSourceDnsARecordSetRead(d *schema.ResourceData, meta interface{}) error {
	//nolint:forcetypeassert
	host := d.Get("host").(string)

	a, _, err := lookupIP(host)
	if err != nil {
		return fmt.Errorf("error looking up A records for %q: %s", host, err)
	}
	sort.Strings(a)

	//nolint:errcheck
	d.Set("addrs", a)
	d.SetId(host)

	return nil
}
