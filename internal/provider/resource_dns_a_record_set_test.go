package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsARecordSet_Basic(t *testing.T) {

	resourceName := "dns_a_record_set.foo"
	resourceRoot := "dns_a_record_set.root"

	deleteARecordSet := func() {
		rec_name := "foo"
		rec_zone := "example.com."

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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testSDKProviderFactories,
		CheckDestroy:      testAccCheckDnsARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsARecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "192.168.000.002"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "192.168.000.001"),
				),
			},
			{
				Config: testAccDnsARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.1"),
				),
			},
			{
				PreConfig: deleteARecordSet,
				Config:    testAccDnsARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.1"),
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
					resource.TestCheckTypeSetElemAttr(resourceRoot, "addresses.*", "192.168.0.1"),
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

var testAccDnsARecordSet_basic = `
  resource "dns_a_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    addresses = ["192.168.000.001", "192.168.000.002"]
    ttl = 300
  }`

var testAccDnsARecordSet_update = `
  resource "dns_a_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    addresses = ["10.0.0.1", "10.0.0.2", "10.0.0.3"]
    ttl = 300
  }`

var testAccDnsARecordSet_root = `
  resource "dns_a_record_set" "root" {
    zone = "example.com."
    addresses = ["192.168.0.1"]
    ttl = 300
  }`
