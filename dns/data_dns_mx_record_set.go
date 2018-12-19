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

	type MxServer struct {
		Host     string
		Priority int
	}

	//var m []MxServer
	m := make([]MxServer, 0)
	for _, record := range mxRecords {
		m = append(m, MxServer{record.Host, int(record.Pref)})
	}

	sort.Slice(m, func(i, j int) bool {
		if m[i].Priority < m[j].Priority {
			return true
		}
		if m[i].Priority > m[j].Priority {
			return false
		}
		return m[i].Host < m[j].Host
	})

	mxservers := make([]string, 0, len(m))
	priorities := make([]int, 0, len(m))

	for _, MxServer := range m {
		mxservers = append(mxservers, MxServer.Host)
		priorities = append(priorities, MxServer.Priority)
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
