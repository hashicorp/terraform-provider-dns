package dns

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/miekg/dns"
	"time"
)

func resourceDnsNSRecordSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceDnsNSRecordSetCreate,
		Read:   resourceDnsNSRecordSetRead,
		Update: resourceDnsNSRecordSetUpdate,
		Delete: resourceDnsNSRecordSetDelete,

		Schema: map[string]*schema.Schema{
			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"nameservers": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
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

	if rec_zone != dns.Fqdn(rec_zone) {
		return fmt.Errorf("Error creating DNS record: \"zone\" should be an FQDN")
	}

	rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

	d.SetId(rec_fqdn)

	return resourceDnsNSRecordSetUpdate(d, meta)
}

func resourceDnsNSRecordSetRead(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		rec_name := d.Get("name").(string)
		rec_zone := d.Get("zone").(string)

		if rec_zone != dns.Fqdn(rec_zone) {
			return fmt.Errorf("Error reading DNS record: \"zone\" should be an FQDN")
		}

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		c := meta.(*DNSClient).c

		srv_addr := meta.(*DNSClient).srv_addr

		msg := new(dns.Msg)
		msg.SetQuestion(rec_fqdn, dns.TypeNS)
		msg.RecursionDesired = false

		r, _, err := c.Exchange(msg, srv_addr)

		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error querying DNS record: %s", dns.RcodeToString[r.Rcode])
		}

		nameservers := schema.NewSet(schema.HashString, nil)

		for _, record := range r.Ns {
			nameserver, err := getNSVal(record)

			if err != nil {
				return fmt.Errorf("Error querying DNS record: %s", err)
			}

			nameservers.Add(nameserver)
		}

		if !nameservers.Equal(d.Get("nameservers")) {
			d.SetId("")
			return fmt.Errorf("DNS record differs")
		}
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

		if rec_zone != dns.Fqdn(rec_zone) {
			return fmt.Errorf("Error updating DNS record: \"zone\" should be an FQDN")
		}

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		c := meta.(*DNSClient).c
		srv_addr := meta.(*DNSClient).srv_addr
		keyname := meta.(*DNSClient).keyname
		keyalgo := meta.(*DNSClient).keyalgo

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

			if keyname != "" {
				msg.SetTsig(keyname, keyalgo, 300, time.Now().Unix())
			}

			r, _, err := c.Exchange(msg, srv_addr)
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

		if rec_zone != dns.Fqdn(rec_zone) {
			return fmt.Errorf("Error updating DNS record: \"zone\" should be an FQDN")
		}

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		c := meta.(*DNSClient).c
		srv_addr := meta.(*DNSClient).srv_addr
		keyname := meta.(*DNSClient).keyname
		keyalgo := meta.(*DNSClient).keyalgo

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rr_remove, _ := dns.NewRR(fmt.Sprintf("%s 0 NS", rec_fqdn))
		msg.RemoveRRset([]dns.RR{rr_remove})

		if keyname != "" {
			msg.SetTsig(keyname, keyalgo, 300, time.Now().Unix())
		}

		r, _, err := c.Exchange(msg, srv_addr)
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
