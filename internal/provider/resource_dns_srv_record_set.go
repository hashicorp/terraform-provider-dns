package provider

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/miekg/dns"

	"github.com/hashicorp/terraform-provider-dns/internal/hashcode"
)

func resourceDnsSRVRecordSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceDnsSRVRecordSetCreate,
		Read:   resourceDnsSRVRecordSetRead,
		Update: resourceDnsSRVRecordSetUpdate,
		Delete: resourceDnsSRVRecordSetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDnsImport,
		},

		Schema: map[string]*schema.Schema{
			"zone": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateZone,
				Description: "DNS zone the record set belongs to. It must be an FQDN, that is, include the trailing " +
					"dot.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
				Description: "The name of the record set. The `zone` argument will be appended to this value to " +
					"create the full record path.",
			},
			"srv": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"priority": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The priority for the record.",
						},
						"weight": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The weight for the record.",
						},
						"port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The port for the service on the target.",
						},
						"target": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateZone,
							Description:  "The FQDN of the target, include the trailing dot.",
						},
					},
				},
				Set:         resourceDnsSRVRecordSetHash,
				Description: "Can be specified multiple times for each SRV record.",
			},
			"ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Default:     3600,
				Description: "The TTL of the record set. Defaults to `3600`.",
			},
		},

		Description: "Creates an SRV type DNS record set.",
	}
}

func resourceDnsSRVRecordSetCreate(d *schema.ResourceData, meta interface{}) error {

	d.SetId(resourceFQDN(d))

	return resourceDnsSRVRecordSetUpdate(d, meta)
}

func resourceDnsSRVRecordSetRead(d *schema.ResourceData, meta interface{}) error {

	answers, err := resourceDnsRead(d, meta, dns.TypeSRV)
	if err != nil {
		return err
	}

	if len(answers) > 0 {

		var ttl sort.IntSlice

		srv := schema.NewSet(resourceDnsSRVRecordSetHash, nil)
		for _, record := range answers {
			switch r := record.(type) {
			case *dns.SRV:
				s := map[string]interface{}{
					"priority": int(r.Priority),
					"weight":   int(r.Weight),
					"port":     int(r.Port),
					"target":   r.Target,
				}
				srv.Add(s)
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				return fmt.Errorf("didn't get an SRV record")
			}
		}
		sort.Sort(ttl)

		//nolint:errcheck
		d.Set("srv", srv)
		//nolint:errcheck
		d.Set("ttl", ttl[0])
	} else {
		d.SetId("")
	}

	return nil
}

func resourceDnsSRVRecordSetUpdate(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		//nolint:forcetypeassert
		ttl := d.Get("ttl").(int)
		fqdn := resourceFQDN(d)

		msg := new(dns.Msg)

		//nolint:forcetypeassert
		msg.SetUpdate(d.Get("zone").(string))

		if d.HasChange("srv") {
			o, n := d.GetChange("srv")
			//nolint:forcetypeassert
			os := o.(*schema.Set)
			//nolint:forcetypeassert
			ns := n.(*schema.Set)
			remove := os.Difference(ns).List()
			add := ns.Difference(os).List()

			// Loop through all the old addresses and remove them
			for _, srv := range remove {
				//nolint:forcetypeassert
				s := srv.(map[string]interface{})
				rrStr := fmt.Sprintf("%s %d SRV %d %d %d %s", fqdn, ttl, s["priority"], s["weight"], s["port"], s["target"])

				rr_remove, err := dns.NewRR(rrStr)
				if err != nil {
					return fmt.Errorf("error reading DNS record (%s): %s", rrStr, err)
				}

				msg.Remove([]dns.RR{rr_remove})
			}
			// Loop through all the new addresses and insert them
			for _, srv := range add {
				//nolint:forcetypeassert
				s := srv.(map[string]interface{})
				rrStr := fmt.Sprintf("%s %d SRV %d %d %d %s", fqdn, ttl, s["priority"], s["weight"], s["port"], s["target"])

				rr_insert, err := dns.NewRR(rrStr)
				if err != nil {
					return fmt.Errorf("error reading DNS record (%s): %s", rrStr, err)
				}

				msg.Insert([]dns.RR{rr_insert})
			}

			r, err := exchange(msg, true, meta)
			if err != nil {
				d.SetId("")
				return fmt.Errorf("Error updating DNS record: %s", err)
			}
			if r.Rcode != dns.RcodeSuccess {
				d.SetId("")
				return fmt.Errorf("Error updating DNS record: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode])
			}
		}

		return resourceDnsSRVRecordSetRead(d, meta)
	} else {
		return fmt.Errorf("update server is not set")
	}
}

func resourceDnsSRVRecordSetDelete(d *schema.ResourceData, meta interface{}) error {

	return resourceDnsDelete(d, meta, dns.TypeSRV)
}

func resourceDnsSRVRecordSetHash(v interface{}) int {
	var buf bytes.Buffer
	//nolint:forcetypeassert
	m := v.(map[string]interface{})
	//nolint:forcetypeassert
	buf.WriteString(fmt.Sprintf("%d-", m["priority"].(int)))
	//nolint:forcetypeassert
	buf.WriteString(fmt.Sprintf("%d-", m["weight"].(int)))
	//nolint:forcetypeassert
	buf.WriteString(fmt.Sprintf("%d-", m["port"].(int)))
	//nolint:forcetypeassert
	buf.WriteString(fmt.Sprintf("%s-", m["target"].(string)))

	return hashcode.String(buf.String())
}
