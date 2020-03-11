package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsAAAARecordSet_basic(t *testing.T) {

	var rec_name, rec_zone string
	resourceName := "dns_aaaa_record_set.bar"
	resourceRoot := "dns_aaaa_record_set.root"

	deleteAAAARecordSet := func() {
		meta := testAccProvider.Meta()

		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rec_fqdn := testResourceFQDN(rec_name, rec_zone)

		rr_remove, _ := dns.NewRR(fmt.Sprintf("%s 0 AAAA", rec_fqdn))
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
		CheckDestroy: testAccCheckDnsAAAARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsAAAARecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					testAccCheckDnsAAAARecordSetExists(t, resourceName, []interface{}{"fdd5:e282::dead:beef:cafe:babe", "fdd5:e282::cafe:babe:dead:beef"}, &rec_name, &rec_zone),
				),
			},
			{
				Config: testAccDnsAAAARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					testAccCheckDnsAAAARecordSetExists(t, resourceName, []interface{}{"fdd5:e282::beef:dead:babe:cafe", "fdd5:e282::babe:cafe:beef:dead"}, &rec_name, &rec_zone),
				),
			},
			{
				PreConfig: deleteAAAARecordSet,
				Config:    testAccDnsAAAARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					testAccCheckDnsAAAARecordSetExists(t, resourceName, []interface{}{"fdd5:e282::beef:dead:babe:cafe", "fdd5:e282::babe:cafe:beef:dead"}, &rec_name, &rec_zone),
				),
			},
			{
				Config: testAccDnsAAAARecordSet_retry,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "14"),
					testAccCheckDnsAAAARecordSetExists(t, resourceName, []interface{}{"fdd5:e282::beef:dead:babe:cafe", "fdd5:e282::babe:cafe:beef:dead", "fdd5:e282::beef:babe:dead:cafe", "fdd5:e282::babe:beef:cafe:dead", "fdd5:e282::cafe:beef:babe:dead", "fdd5:e282::cafe:beef:dead:babe", "fdd5:e282::cafe:babe:dead:beef", "fdd5:e282::cafe:babe:beef:dead", "fdd5:e282::dead:babe:cafe:beef", "fdd5:e282::dead:babe:beef:cafe", "fdd5:e282::dead:cafe:babe:beef", "fdd5:e282::dead:cafe:beef:babe", "fdd5:e282::dead:beef:cafe:babe", "fdd5:e282::dead:beef:babe:cafe"}, &rec_name, &rec_zone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsAAAARecordSet_root,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "addresses.#", "1"),
					testAccCheckDnsAAAARecordSetExists(t, resourceRoot, []interface{}{"fdd5:e282::beef:dead:babe:cafe"}, &rec_name, &rec_zone),
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

func testAccCheckDnsAAAARecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_aaaa_record_set", dns.TypeAAAA)
}

func testAccCheckDnsAAAARecordSetExists(t *testing.T, n string, addr []interface{}, rec_name, rec_zone *string) resource.TestCheckFunc {
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
		msg.SetQuestion(rec_fqdn, dns.TypeAAAA)
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
			addr, _, err := getAAAAVal(record)
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

var testAccDnsAAAARecordSet_basic = fmt.Sprintf(`
  resource "dns_aaaa_record_set" "bar" {
    zone = "example.com."
    name = "bar"
    addresses = ["fdd5:e282:0000:0000:dead:beef:cafe:babe", "fdd5:e282:0000:0000:cafe:babe:dead:beef"]
    ttl = 300
  }`)

var testAccDnsAAAARecordSet_update = fmt.Sprintf(`
  resource "dns_aaaa_record_set" "bar" {
    zone = "example.com."
    name = "bar"
    addresses = ["fdd5:e282:0000:0000:beef:dead:babe:cafe", "fdd5:e282:0000:0000:babe:cafe:beef:dead"]
    ttl = 300
  }`)

var testAccDnsAAAARecordSet_retry = fmt.Sprintf(`
  resource "dns_aaaa_record_set" "bar" {
    zone = "example.com."
    name = "bar"
    addresses = ["fdd5:e282::beef:dead:babe:cafe", "fdd5:e282::babe:cafe:beef:dead", "fdd5:e282::beef:babe:dead:cafe", "fdd5:e282::babe:beef:cafe:dead", "fdd5:e282::cafe:beef:babe:dead", "fdd5:e282::cafe:beef:dead:babe", "fdd5:e282::cafe:babe:dead:beef", "fdd5:e282::cafe:babe:beef:dead", "fdd5:e282::dead:babe:cafe:beef", "fdd5:e282::dead:babe:beef:cafe", "fdd5:e282::dead:cafe:babe:beef", "fdd5:e282::dead:cafe:beef:babe", "fdd5:e282::dead:beef:cafe:babe", "fdd5:e282::dead:beef:babe:cafe"]
    ttl = 300
  }`)

var testAccDnsAAAARecordSet_root = fmt.Sprintf(`
  resource "dns_aaaa_record_set" "root" {
    zone = "example.com."
    addresses = ["fdd5:e282:0000:0000:beef:dead:babe:cafe"]
    ttl = 300
  }`)
