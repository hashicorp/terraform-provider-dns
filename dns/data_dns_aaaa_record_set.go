package dns

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDnsAAAARecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsAAAARecordSetRead,
		Schema: map[string]*schema.Schema{
			"host": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ignore_errors": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"addrs": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataSourceDnsAAAARecordSetRead(d *schema.ResourceData, meta interface{}) error {
	host := d.Get("host").(string)
	ignore := d.Get("ignore_errors").(bool)

	_, aaaa, err := lookupIP(host)
	if err != nil && !ignore {
		return fmt.Errorf("error looking up AAAA records for %q: %s", host, err)
	}
	if aaaa == nil {
		aaaa = []string{}
	}
	sort.Strings(aaaa)

	d.Set("addrs", aaaa)
	d.SetId(host)

	return nil
}
