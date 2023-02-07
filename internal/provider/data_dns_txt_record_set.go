package provider

import (
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDnsTxtRecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsTxtRecordSetRead,

		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Host to look up.",
			},

			"record": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The first TXT record.",
			},

			"records": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of TXT records.",
			},
		},

		Description: "Use this data source to get DNS TXT record set of the host.",
	}
}

func dataSourceDnsTxtRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	//nolint:forcetypeassert
	host := d.Get("host").(string)

	records, err := net.LookupTXT(host)
	if err != nil {
		return fmt.Errorf("error looking up TXT records for %q: %s", host, err)
	}

	if len(records) > 0 {
		//nolint:errcheck
		d.Set("record", records[0])
	} else {
		//nolint:errcheck
		d.Set("record", "")
	}
	//nolint:errcheck
	d.Set("records", records)
	d.SetId(host)

	return nil
}
