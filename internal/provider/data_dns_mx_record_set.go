package provider

import (
	"fmt"
	"net"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDnsMXRecordSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDnsMXRecordSetRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Domain to look up.",
			},
			"mx": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"preference": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"exchange": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
				Description: "A list of records. They are sorted by ascending preference then alphabetically by " +
					"exchange to stay consistent across runs.",
			},
		},
		Description: "Use this data source to get DNS MX records for a domain.",
	}
}

func dataSourceDnsMXRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	//nolint:forcetypeassert
	domain := d.Get("domain").(string)

	records, err := net.LookupMX(domain)
	if err != nil {
		return fmt.Errorf("error looking up MX records for %q: %s", domain, err)
	}

	// Sort by preference ascending, and host alphabetically
	sort.Slice(records, func(i, j int) bool {
		if records[i].Pref < records[j].Pref {
			return true
		}
		if records[i].Pref > records[j].Pref {
			return false
		}
		return records[i].Host < records[j].Host
	})

	mx := make([]map[string]interface{}, len(records))
	for i, record := range records {
		mx[i] = map[string]interface{}{
			"preference": int(record.Pref),
			"exchange":   record.Host,
		}
	}

	if err = d.Set("mx", mx); err != nil {
		return err
	}
	d.SetId(domain)

	return nil
}
