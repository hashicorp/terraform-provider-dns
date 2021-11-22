package provider

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/miekg/dns"
)

func resourceDnsPtrRecordSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceDnsPtrRecordSetCreate,
		Read:   resourceDnsPtrRecordSetRead,
		Update: resourceDnsPtrRecordSetUpdate,
		Delete: resourceDnsPtrRecordSetDelete,
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
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
			},
			"ptrs": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateZone,
				},
				Set: schema.HashString,
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

func resourceDnsPtrRecordSetCreate(d *schema.ResourceData, meta interface{}) error {

	d.SetId(resourceFQDN(d))

	return resourceDnsPtrRecordSetUpdate(d, meta)
}

func resourceDnsPtrRecordSetRead(d *schema.ResourceData, meta interface{}) error {

	answers, err := resourceDnsRead(d, meta, dns.TypePTR)
	if err != nil {
		return err
	}

	if len(answers) > 0 {

		var ttl sort.IntSlice

		ptrs := schema.NewSet(schema.HashString, nil)
		for _, record := range answers {
			ptr, t, err := getPtrVal(record)
			if err != nil {
				return fmt.Errorf("Error querying DNS record: %s", err)
			}
			ptrs.Add(ptr)
			ttl = append(ttl, t)
		}
		sort.Sort(ttl)

		d.Set("ptrs", ptrs)
		d.Set("ttl", ttl[0])
	} else {
		d.SetId("")
	}

	return nil

}

func resourceDnsPtrRecordSetUpdate(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		ttl := d.Get("ttl").(int)

		rec_fqdn := resourceFQDN(d)

		msg := new(dns.Msg)

		msg.SetUpdate(d.Get("zone").(string))

		if d.HasChange("ptrs") {
			o, n := d.GetChange("ptrs")
			os := o.(*schema.Set)
			ns := n.(*schema.Set)
			remove := os.Difference(ns).List()
			add := ns.Difference(os).List()

			// Loop through all the old ptrs and remove them
			for _, ptr := range remove {
				rr_remove, _ := dns.NewRR(fmt.Sprintf("%s %d PTR %s", rec_fqdn, ttl, ptr.(string)))
				msg.Remove([]dns.RR{rr_remove})
			}
			// Loop through all the new ptrs and insert them
			for _, ptr := range add {
				rr_insert, _ := dns.NewRR(fmt.Sprintf("%s %d PTR %s", rec_fqdn, ttl, ptr.(string)))
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

		return resourceDnsPtrRecordSetRead(d, meta)
	} else {
		return fmt.Errorf("update server is not set")
	}
}

func resourceDnsPtrRecordSetDelete(d *schema.ResourceData, meta interface{}) error {

	return resourceDnsDelete(d, meta, dns.TypePTR)
}
