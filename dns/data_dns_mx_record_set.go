package dns

import (
	"fmt"
	"net"
	"sort"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDnsMXRecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsMXRecordSetRead,
		Schema: map[string]*schema.Schema{
			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"mxservers": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"priorities": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Computed: true,
			},
		},
	}
}

func dataSourceDnsMXRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	zone := d.Get("zone").(string)

	mxRecords, err := net.LookupMX(zone)
	if err != nil {
		return fmt.Errorf("error looking up MX records for %q: %s", zone, err)
	}

	m := make(map[string]int)
	for _, record := range mxRecords {
		m[record.Host] = int(record.Pref)
	}
	//List of mxservers sorted by name, not priority
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	mxservers := make([]string, 0, len(m))
	priorities := make([]int, 0, len(m))
	for _, k := range keys {
		mxservers = append(mxservers, k)
		priorities = append(priorities, m[k])
	}

	err = d.Set("mxservers", mxservers)
	if err != nil {
		return err
	}

	err = d.Set("priorities", priorities)
	if err != nil {
		return err
	}

	d.SetId(zone)

	return nil
}
