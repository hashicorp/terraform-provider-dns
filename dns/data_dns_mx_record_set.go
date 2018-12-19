package dns

import (
	"fmt"
	"net"

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
		},
	}
}

func dataSourceDnsMXRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	zone := d.Get("zone").(string)

	mxRecords, err := net.LookupMX(zone)
	if err != nil {
		return fmt.Errorf("error looking up MX records for %q: %s", zone, err)
	}

	mxservers := make([]string, len(mxRecords))
	for i, record := range mxRecords {
		mxservers[i] = record.Host
	}

	err = d.Set("mxservers", mxservers)
	if err != nil {
		return err
	}
	d.SetId(zone)

	return nil
}
