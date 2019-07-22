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
			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"ignore_errors": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"mx": &schema.Schema{
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
			},
		},
	}
}

func dataSourceDnsMXRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	domain := d.Get("domain").(string)
	ignore := d.Get("ignore_errors").(bool)

	records, err := net.LookupMX(domain)
	if err != nil && !ignore {
		return fmt.Errorf("error looking up MX records for %q: %s", domain, err)
	}
	if records == nil {
		records = []*net.MX{}
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