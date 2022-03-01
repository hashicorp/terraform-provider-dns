package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsPtrRecordSet_Basic(t *testing.T) {

	var rec_name, rec_zone string
	resourceName := "dns_ptr_record_set.foo"
	resourceRoot := "dns_ptr_record_set.root"

	deletePtrRecordSet := func() {
		meta := testAccProvider.Meta()

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rec_fqdn := testResourceFQDN(rec_name, rec_zone)

		rr_remove, _ := dns.NewRR(fmt.Sprintf("%s 0 PTR", rec_fqdn))
		msg.RemoveRRset([]dns.RR{rr_remove})

		r, err := exchange(msg, true, meta)
		if err != nil {
			t.Fatalf("Error deleting DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			t.Fatalf("Error deleting DNS record: %v", r.Rcode)
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDnsPtrRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsPtrRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ptr.#", "1"),
					testAccCheckDnsPtrRecordSetExists(t, resourceName, []interface{}{"foo.example.com."}, &rec_name, &rec_zone),
				),
			},
			{
				Config: testAccDnsPtrRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ptr.#", "1"),
					testAccCheckDnsPtrRecordSetExists(t, resourceName, []interface{}{"bar.example.com."}, &rec_name, &rec_zone),
				),
			},
			{
				PreConfig: deletePtrRecordSet,
				Config:    testAccDnsPtrRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ptr.#", "1"),
					testAccCheckDnsPtrRecordSetExists(t, resourceName, []interface{}{"bar.example.com."}, &rec_name, &rec_zone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsPtrRecordSet_root,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "ptr.#", "1"),
					testAccCheckDnsPtrRecordSetExists(t, resourceRoot, []interface{}{"baz.example.com."}, &rec_name, &rec_zone),
				),
			},
			{
				ResourceName:      resourceRoot,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDnsPtrRecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_ptr_record_set", dns.TypePTR)
}

func testAccCheckDnsPtrRecordSetExists(t *testing.T, n string, addr []interface{}, rec_name, rec_zone *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		*rec_name = rs.Primary.Attributes["name"]
		*rec_zone = rs.Primary.Attributes["zone"]

		rec_fqdn := testResourceFQDN(*rec_name, *rec_zone)

		meta := testAccProvider.Meta()

		msg := new(dns.Msg)
		msg.SetQuestion(rec_fqdn, dns.TypePTR)
		r, err := exchange(msg, false, meta)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error querying DNS record")
		}

		ptr := schema.NewSet(schema.HashString, nil)
		expected := schema.NewSet(schema.HashString, addr)
		for _, record := range r.Answer {
			addr, _, err := getPtrVal(record)
			if err != nil {
				return fmt.Errorf("Error querying DNS record: %s", err)
			}
			ptr.Add(addr)
		}
		if !ptr.Equal(expected) {
			return fmt.Errorf("DNS record differs: expected %v, found %v", expected, ptr)
		}
		return nil
	}
}

var testAccDnsPtrRecordSet_basic = fmt.Sprintf(`
  resource "dns_ptr_record_set" "foo" {
    zone = "1.168.192.in-addr.arpa."
    name = "99"
    ptr = ["foo.example.com."]
    ttl = 300
  }`)

var testAccDnsPtrRecordSet_update = fmt.Sprintf(`
  resource "dns_ptr_record_set" "foo" {
    zone = "1.168.192.in-addr.arpa."
    name = "99"
    ptr = ["bar.example.com."]
    ttl = 300
  }`)

var testAccDnsPtrRecordSet_root = fmt.Sprintf(`
  resource "dns_ptr_record_set" "root" {
    zone = "1.168.192.in-addr.arpa."
    ptr = ["baz.example.com."]
    ttl = 300
  }`)
