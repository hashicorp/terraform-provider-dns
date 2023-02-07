package provider

import (
	"fmt"
	"net"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDnsNSRecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsNSRecordSetRead,
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Host to look up.",
			},
			"nameservers": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of nameservers. Nameservers are always sorted to avoid constant changing plans.",
			},
		},
		Description: "Use this data source to get DNS ns records of the host.",
	}
}

func dataSourceDnsNSRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	//nolint:forcetypeassert
	host := d.Get("host").(string)

	nsRecords, err := net.LookupNS(host)
	if err != nil {
		return fmt.Errorf("error looking up NS records for %q: %s", host, err)
	}

	nameservers := make([]string, len(nsRecords))
	for i, record := range nsRecords {
		nameservers[i] = record.Host
	}
	sort.Strings(nameservers)

	err = d.Set("nameservers", nameservers)
	if err != nil {
		return err
	}
	d.SetId(host)

	return nil
}
