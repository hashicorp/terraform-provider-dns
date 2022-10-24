package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsPtrRecord_basic(t *testing.T) {

	var rec_name, rec_zone string
	resourceName := "dns_ptr_record.foo"
	resourceRoot := "dns_ptr_record.root"

	deletePtrRecord := func() {
		meta := testAccProvider.Meta()

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rec_fqdn := testResourceFQDN(rec_name, rec_zone)

		rrStr := fmt.Sprintf("%s 0 PTR", rec_fqdn)

		rr_remove, err := dns.NewRR(rrStr)
		if err != nil {
			t.Fatalf("Error reading DNS record (%s): %s", rrStr, err)
		}

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
		CheckDestroy: testAccCheckDnsPtrRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsPtrRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsPtrRecordExists(t, resourceName, "bar.example.com.", &rec_name, &rec_zone),
				),
			},
			{
				Config: testAccDnsPtrRecord_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsPtrRecordExists(t, resourceName, "baz.example.com.", &rec_name, &rec_zone),
				),
			},
			{
				PreConfig: deletePtrRecord,
				Config:    testAccDnsPtrRecord_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsPtrRecordExists(t, resourceName, "baz.example.com.", &rec_name, &rec_zone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsPtrRecord_root,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsPtrRecordExists(t, resourceRoot, "baz.example.com.", &rec_name, &rec_zone),
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

func testAccCheckDnsPtrRecordDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_ptr_record", dns.TypePTR)
}

func testAccCheckDnsPtrRecordExists(t *testing.T, n string, expected string, rec_name, rec_zone *string) resource.TestCheckFunc {
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

		if len(r.Answer) > 1 {
			return fmt.Errorf("Error querying DNS record: multiple responses received")
		}
		record := r.Answer[0]
		ptr, _, err := getPtrVal(record)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if expected != ptr {
			return fmt.Errorf("DNS record differs: expected %v, found %v", expected, ptr)
		}
		return nil
	}
}

var testAccDnsPtrRecord_basic = fmt.Sprintf(`
  resource "dns_ptr_record" "foo" {
    zone = "example.com."
    name = "r._dns-sd._udp"
    ptr = "bar.example.com."
    ttl = 300
  }`)

var testAccDnsPtrRecord_update = fmt.Sprintf(`
  resource "dns_ptr_record" "foo" {
    zone = "example.com."
    name = "r._dns-sd._udp"
    ptr = "baz.example.com."
    ttl = 300
  }`)

var testAccDnsPtrRecord_root = fmt.Sprintf(`
  resource "dns_ptr_record" "root" {
    zone = "example.com."
    ptr = "baz.example.com."
    ttl = 300
  }`)
