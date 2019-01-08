package dns

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/miekg/dns"
)

func resourceDnsAAAARecordSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceDnsAAAARecordSetCreate,
		Read:   resourceDnsAAAARecordSetRead,
		Update: resourceDnsAAAARecordSetUpdate,
		Delete: resourceDnsAAAARecordSetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDnsImport,
		},

		Schema: map[string]*schema.Schema{
			"zone": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateZone,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
			},
			"addresses": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      hashIPString,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  3600,
			},
		},
	}
}

func resourceDnsAAAARecordSetCreate(d *schema.ResourceData, meta interface{}) error {

	rec_name := d.Get("name").(string)
	rec_zone := d.Get("zone").(string)

	rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

	d.SetId(rec_fqdn)

	return resourceDnsAAAARecordSetUpdate(d, meta)
}

func resourceDnsAAAARecordSetRead(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		rec_name := d.Get("name").(string)
		rec_zone := d.Get("zone").(string)

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		msg := new(dns.Msg)
		msg.SetQuestion(rec_fqdn, dns.TypeAAAA)

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

		addresses := schema.NewSet(hashIPString, nil)
		for _, record := range r.Answer {
			addr, t, err := getAAAAVal(record)
			if err != nil {
				return fmt.Errorf("Error querying DNS record: %s", err)
			}
			addresses.Add(addr)
			ttl = append(ttl, t)
		}
		sort.Sort(ttl)

		d.Set("name", rec_name)
		d.Set("zone", rec_zone)
		d.Set("addresses", addresses)
		d.Set("ttl", ttl[0])

		return nil
	} else {
		return fmt.Errorf("update server is not set")
	}
}

func resourceDnsAAAARecordSetUpdate(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		rec_name := d.Get("name").(string)
		rec_zone := d.Get("zone").(string)
		ttl := d.Get("ttl").(int)

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		if d.HasChange("addresses") {
			o, n := d.GetChange("addresses")
			os := o.(*schema.Set)
			ns := n.(*schema.Set)
			remove := os.Difference(ns).List()
			add := ns.Difference(os).List()

			// Loop through all the old addresses and remove them
			for _, addr := range remove {
				rr_remove, _ := dns.NewRR(fmt.Sprintf("%s %d AAAA %s", rec_fqdn, ttl, addr.(string)))
				msg.Remove([]dns.RR{rr_remove})
			}
			// Loop through all the new addresses and insert them
			for _, addr := range add {
				rr_insert, _ := dns.NewRR(fmt.Sprintf("%s %d AAAA %s", rec_fqdn, ttl, addr.(string)))
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

			addresses := ns
			d.Set("addresses", addresses)
		}

		return resourceDnsAAAARecordSetRead(d, meta)
	} else {
		return fmt.Errorf("update server is not set")
	}
}

func resourceDnsAAAARecordSetDelete(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		rec_name := d.Get("name").(string)
		rec_zone := d.Get("zone").(string)

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rr_remove, _ := dns.NewRR(fmt.Sprintf("%s 0 AAAA", rec_fqdn))
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
