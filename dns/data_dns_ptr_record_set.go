package dns

import (
	"fmt"
	"net"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDnsPtrRecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsPtrRecordSetRead,
		Schema: map[string]*schema.Schema{
			"ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ignore_errors": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ptr": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDnsPtrRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	ipAddress := d.Get("ip_address").(string)
	ignore := d.Get("ignore_errors").(bool)

	names, err := net.LookupAddr(ipAddress)
	if err != nil && !ignore {
		return fmt.Errorf("error looking up PTR records for %q: %s", ipAddress, err)
	}
	if len(names) > 0 {
		d.Set("ptr", names[0])
	} else if ignore {
		d.Set("ptr", "")
	} else {
		return fmt.Errorf("error looking up PTR records for %q: no records found", ipAddress)
	}

	d.SetId(ipAddress)

	return nil
}
