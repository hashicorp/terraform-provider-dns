package dns

import (
	"fmt"
	"net"
	// "sort"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDnsNSRecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsARecordSetRead,
		Schema: map[string]*schema.Schema{
			"host": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"nameservers": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataSourceDnsNSRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	host := d.Get("host").(string)

	nameservers, err := net.LookupNS(host)
	if err != nil {
		return fmt.Errorf("error looking up NS records for %q: %s", host, err)
	}

	// nameservers := make([]string, 0)

	// sort.Strings(nameservers)

	d.Set("nameservers", nameservers)
	d.SetId(host)

	return nil
}
