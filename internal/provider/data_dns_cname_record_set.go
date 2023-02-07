package provider

import (
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDnsCnameRecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsCnameRecordSetRead,

		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Host to look up.",
			},

			"cname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A CNAME record associated with host.",
			},
		},

		Description: "Use this data source to get DNS CNAME record set of the host.",
	}
}

func dataSourceDnsCnameRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	//nolint:forcetypeassert
	host := d.Get("host").(string)

	cname, err := net.LookupCNAME(host)
	if err != nil {
		return fmt.Errorf("error looking up CNAME records for %q: %s", host, err)
	}

	//nolint:errcheck
	d.Set("cname", cname)
	d.SetId(host)

	return nil
}
