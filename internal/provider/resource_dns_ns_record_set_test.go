package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsNSRecordSet_Basic(t *testing.T) {

	var rec_name, rec_zone string
	resourceName := "dns_ns_record_set.foo"

	deleteNSRecordSet := func() {
		meta := testAccProvider.Meta()

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rec_fqdn := testResourceFQDN(rec_name, rec_zone)

		rr_remove, err := dns.NewRR(fmt.Sprintf("%s 0 NS", rec_fqdn))
		if err != nil {
			t.Fatalf("Error reading DNS record: %s", err)
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
		CheckDestroy: testAccCheckDnsNSRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsNSRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "2"),
					testAccCheckDnsNSRecordSetExists(t, resourceName, []interface{}{"ns1.testdns.co.uk.", "ns2.testdns.co.uk."}, &rec_name, &rec_zone),
				),
			},
			{
				Config: testAccDnsNSRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "3"),
					testAccCheckDnsNSRecordSetExists(t, resourceName, []interface{}{"ns1.test2dns.co.uk.", "ns2.test2dns.co.uk.", "ns3.test2dns.co.uk."}, &rec_name, &rec_zone),
				),
			},
			{
				PreConfig: deleteNSRecordSet,
				Config:    testAccDnsNSRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "3"),
					testAccCheckDnsNSRecordSetExists(t, resourceName, []interface{}{"ns1.test2dns.co.uk.", "ns2.test2dns.co.uk.", "ns3.test2dns.co.uk."}, &rec_name, &rec_zone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDnsNSRecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_ns_record_set", dns.TypeNS)
}

func testAccCheckDnsNSRecordSetExists(t *testing.T, n string, nameserver []interface{}, rec_name, rec_zone *string) resource.TestCheckFunc {
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
		msg.SetQuestion(rec_fqdn, dns.TypeNS)

		r, err := exchange(msg, false, meta)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error querying DNS record")
		}

		nameservers := schema.NewSet(schema.HashString, nil)
		for _, record := range r.Ns {
			nameserver, _, err := getNSVal(record)
			if err != nil {
				return fmt.Errorf("Error querying DNS record: %s", err)
			}
			nameservers.Add(nameserver)
		}
		expected := schema.NewSet(schema.HashString, nameserver)

		if !nameservers.Equal(expected) {
			return fmt.Errorf("DNS record differs: expected %v, found %v", expected, nameservers)
		}
		return nil
	}
}

var testAccDnsNSRecordSet_basic = fmt.Sprintf(`
  resource "dns_ns_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    nameservers = [
		"ns1.testdns.co.uk.",
		"ns2.testdns.co.uk.",
		]
    ttl = 60
  }`)

var testAccDnsNSRecordSet_update = fmt.Sprintf(`
  resource "dns_ns_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    nameservers = ["ns1.test2dns.co.uk.", "ns2.test2dns.co.uk.", "ns3.test2dns.co.uk.",]
    ttl = 60
  }`)
