package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsARecordSet_Basic(t *testing.T) {

	var rec_name, rec_zone string
	resourceName := "dns_a_record_set.foo"
	resourceRoot := "dns_a_record_set.root"

	deleteARecordSet := func() {
		meta := testAccProvider.Meta()

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rec_fqdn := testResourceFQDN(rec_name, rec_zone)

		rrStr := fmt.Sprintf("%s 0 A", rec_fqdn)

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
		CheckDestroy: testAccCheckDnsARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsARecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					testAccCheckDnsARecordSetExists(t, resourceName, []interface{}{"192.168.0.2", "192.168.0.1"}, &rec_name, &rec_zone),
				),
			},
			{
				Config: testAccDnsARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "3"),
					testAccCheckDnsARecordSetExists(t, resourceName, []interface{}{"10.0.0.3", "10.0.0.2", "10.0.0.1"}, &rec_name, &rec_zone),
				),
			},
			{
				PreConfig: deleteARecordSet,
				Config:    testAccDnsARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "3"),
					testAccCheckDnsARecordSetExists(t, resourceName, []interface{}{"10.0.0.3", "10.0.0.2", "10.0.0.1"}, &rec_name, &rec_zone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsARecordSet_root,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "addresses.#", "1"),
					testAccCheckDnsARecordSetExists(t, resourceRoot, []interface{}{"192.168.0.1"}, &rec_name, &rec_zone),
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

func testAccCheckDnsARecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_a_record_set", dns.TypeA)
}

func testAccCheckDnsARecordSetExists(t *testing.T, n string, addr []interface{}, rec_name, rec_zone *string) resource.TestCheckFunc {
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
		msg.SetQuestion(rec_fqdn, dns.TypeA)
		r, err := exchange(msg, false, meta)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error querying DNS record")
		}

		addresses := schema.NewSet(schema.HashString, nil)
		expected := schema.NewSet(schema.HashString, addr)
		for _, record := range r.Answer {
			addr, _, err := getAVal(record)
			if err != nil {
				return fmt.Errorf("Error querying DNS record: %s", err)
			}
			addresses.Add(addr)
		}
		if !addresses.Equal(expected) {
			return fmt.Errorf("DNS record differs: expected %v, found %v", expected, addresses)
		}
		return nil
	}
}

var testAccDnsARecordSet_basic = fmt.Sprintf(`
  resource "dns_a_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    addresses = ["192.168.000.001", "192.168.000.002"]
    ttl = 300
  }`)

var testAccDnsARecordSet_update = fmt.Sprintf(`
  resource "dns_a_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    addresses = ["10.0.0.1", "10.0.0.2", "10.0.0.3"]
    ttl = 300
  }`)

var testAccDnsARecordSet_root = fmt.Sprintf(`
  resource "dns_a_record_set" "root" {
    zone = "example.com."
    addresses = ["192.168.0.1"]
    ttl = 300
  }`)
