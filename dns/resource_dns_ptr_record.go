package dns

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/miekg/dns"
)

func resourceDnsPtrRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceDnsPtrRecordCreate,
		Read:   resourceDnsPtrRecordRead,
		Update: resourceDnsPtrRecordUpdate,
		Delete: resourceDnsPtrRecordDelete,
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
			"ptr": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateZone,
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

func resourceDnsPtrRecordCreate(d *schema.ResourceData, meta interface{}) error {

	rec_name := d.Get("name").(string)
	rec_zone := d.Get("zone").(string)

	rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

	d.SetId(rec_fqdn)

	return resourceDnsPtrRecordUpdate(d, meta)
}

func resourceDnsPtrRecordRead(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		rec_name := d.Get("name").(string)
		rec_zone := d.Get("zone").(string)

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

	Retry:
		msg := new(dns.Msg)
		msg.SetQuestion(rec_fqdn, dns.TypePTR)

		r, err := exchange(msg, true, meta)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		switch r.Rcode {
		case dns.RcodeServerFailure:
			goto Retry
		case dns.RcodeSuccess:
			break
		case dns.RcodeNameError:
			d.SetId("")
			return nil
		default:
			return fmt.Errorf("Error querying DNS record: %s", dns.RcodeToString[r.Rcode])
		}

		if len(r.Answer) > 1 {
			return fmt.Errorf("Error querying DNS record: multiple responses received")
		}
		record := r.Answer[0]
		ptr, ttl, err := getPtrVal(record)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		d.Set("name", rec_name)
		d.Set("zone", rec_zone)
		d.Set("ptr", ptr)
		d.Set("ttl", ttl)

		return nil
	} else {
		return fmt.Errorf("update server is not set")
	}
}

func resourceDnsPtrRecordUpdate(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		rec_name := d.Get("name").(string)
		rec_zone := d.Get("zone").(string)
		ttl := d.Get("ttl").(int)

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		if d.HasChange("ptr") {
			o, n := d.GetChange("ptr")

			if o != "" {
				rr_remove, _ := dns.NewRR(fmt.Sprintf("%s %d PTR %s", rec_fqdn, ttl, o))
				msg.Remove([]dns.RR{rr_remove})
			}
			if n != "" {
				rr_insert, _ := dns.NewRR(fmt.Sprintf("%s %d PTR %s", rec_fqdn, ttl, n))
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

			ptr := n
			d.Set("ptr", ptr)
		}

		return resourceDnsPtrRecordRead(d, meta)
	} else {
		return fmt.Errorf("update server is not set")
	}
}

func resourceDnsPtrRecordDelete(d *schema.ResourceData, meta interface{}) error {

	if meta != nil {

		rec_name := d.Get("name").(string)
		rec_zone := d.Get("zone").(string)

		rec_fqdn := fmt.Sprintf("%s.%s", rec_name, rec_zone)

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rr_remove, _ := dns.NewRR(fmt.Sprintf("%s 0 PTR", rec_fqdn))
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
