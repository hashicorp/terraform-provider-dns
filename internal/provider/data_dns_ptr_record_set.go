package provider

import (
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDnsPtrRecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsPtrRecordSetRead,
		Schema: map[string]*schema.Schema{
			"ip_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP address to look up.",
			},
			"ptr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A PTR record associated with `ip_address`.",
			},
		},
		Description: "Use this data source to get DNS PTR record set of the ip address.",
	}
}

func dataSourceDnsPtrRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	//nolint:forcetypeassert
	ipAddress := d.Get("ip_address").(string)
	names, err := net.LookupAddr(ipAddress)
	if err != nil {
		return fmt.Errorf("error looking up PTR records for %q: %s", ipAddress, err)
	}
	if len(names) == 0 {
		return fmt.Errorf("error looking up PTR records for %q: no records found", ipAddress)
	}

	//nolint:errcheck
	d.Set("ptr", names[0])
	d.SetId(ipAddress)

	return nil
}
