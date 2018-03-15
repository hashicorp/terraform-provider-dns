package dns

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/miekg/dns"
)

func resourceDnsNSRecordSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceDnsNSRecordSetCreate,
		Read:   resourceDnsNSRecordSetRead,
		Update: resourceDnsNSRecordSetUpdate,
		Delete: resourceDnsNSRecordSetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDnsImport,
		},

		Schema: map[string]*schema.Schema{
			"zone": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateZone,
			},
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
			},
			"nameservers": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateZone,
				},
				Set: schema.HashString,
			},
			"ttl": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  3600,
			},
		},
	}
}

func resourceDnsNSRecordSetCreate(d *schema.ResourceData, meta interface{}) error {

	rec_name := d.Get("name").(string)
	rec_zone := d.Get("zone").(string)

	rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

	d.SetId(rec_fqdn)

	return resourceDnsNSRecordSetUpdate(d, meta)
}

func resourceDnsNSRecordSetRead(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		rec_name := d.Get("name").(string)
		rec_zone := d.Get("zone").(string)

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		msg := new(dns.Msg)
		msg.SetQuestion(rec_fqdn, dns.TypeNS)

		r, err := exchange(msg, true, meta)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		switch r.Rcode {
		case dns.RcodeSuccess:
			break
		case dns.RcodeNameError:
			d.SetId("")
			return nil
		default:
			return fmt.Errorf("Error querying DNS record: %s", dns.RcodeToString[r.Rcode])
		}

		var ttl sort.IntSlice

		nameservers := schema.NewSet(schema.HashString, nil)

		for _, record := range r.Ns {
			nameserver, t, err := getNSVal(record)

			if err != nil {
				return fmt.Errorf("Error querying DNS record: %s", err)
			}

			nameservers.Add(nameserver)
			ttl = append(ttl, t)
		}
		sort.Sort(ttl)

		d.Set("name", rec_name)
		d.Set("zone", rec_zone)
		d.Set("nameservers", nameservers)
		d.Set("ttl", ttl[0])

		return nil
	} else {
		return fmt.Errorf("update server is not set")
	}
}

func resourceDnsNSRecordSetUpdate(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		rec_name := d.Get("name").(string)
		rec_zone := d.Get("zone").(string)
		ttl := d.Get("ttl").(int)

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		if d.HasChange("nameservers") {
			o, n := d.GetChange("nameservers")
			os := o.(*schema.Set)
			ns := n.(*schema.Set)
			remove := os.Difference(ns).List()
			add := ns.Difference(os).List()

			// Loop through all the old nameservers and remove them
			for _, nameserver := range remove {
				rr_remove, _ := dns.NewRR(fmt.Sprintf("%s %d NS %s", rec_fqdn, ttl, nameserver.(string)))
				msg.Remove([]dns.RR{rr_remove})
			}
			// Loop through all the new nameservers and insert them
			for _, nameserver := range add {
				rr_insert, _ := dns.NewRR(fmt.Sprintf("%s %d NS %s", rec_fqdn, ttl, nameserver.(string)))
				msg.Insert([]dns.RR{rr_insert})
			}

			r, err := exchange(msg, true, meta)
			if err != nil {
				d.SetId("")
				return fmt.Errorf("Error updating DNS record: %s", err)
			}
			if r.Rcode != dns.RcodeSuccess {
				d.SetId("")
				return fmt.Errorf("Error updating DNS record: %v", r.Rcode)
			}

			nameservers := ns
			d.Set("nameservers", nameservers)
		}

		return resourceDnsNSRecordSetRead(d, meta)
	} else {
		return fmt.Errorf("update server is not set")
	}
}

func resourceDnsNSRecordSetDelete(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		rec_name := d.Get("name").(string)
		rec_zone := d.Get("zone").(string)

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rr_remove, _ := dns.NewRR(fmt.Sprintf("%s 0 NS", rec_fqdn))
		msg.RemoveRRset([]dns.RR{rr_remove})

		r, err := exchange(msg, true, meta)
		if err != nil {
			return fmt.Errorf("Error deleting DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error deleting DNS record: %v", r.Rcode)
		}

		return nil
	} else {
		return fmt.Errorf("update server is not set")
	}
}
