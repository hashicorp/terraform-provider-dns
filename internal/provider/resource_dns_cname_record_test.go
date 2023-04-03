package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsCnameRecord_basic(t *testing.T) {

	var rec_name, rec_zone string
	resourceName := "dns_cname_record.foo"

	deleteCnameRecord := func() {
		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rec_fqdn := testResourceFQDN(rec_name, rec_zone)

		rrStr := fmt.Sprintf("%s 0 CNAME", rec_fqdn)

		rr_remove, err := dns.NewRR(rrStr)
		if err != nil {
			t.Fatalf("Error reading DNS record (%s): %s", rrStr, err)
		}

		msg.RemoveRRset([]dns.RR{rr_remove})

		r, err := exchange(msg, true, dnsClient)
		if err != nil {
			t.Fatalf("Error deleting DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			t.Fatalf("Error deleting DNS record: %v", r.Rcode)
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsCnameRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsCnameRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsCnameRecordExists(resourceName, "bar.example.com.", &rec_name, &rec_zone),
				),
			},
			{
				Config: testAccDnsCnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsCnameRecordExists(resourceName, "baz.example.com.", &rec_name, &rec_zone),
				),
			},
			{
				PreConfig: deleteCnameRecord,
				Config:    testAccDnsCnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsCnameRecordExists(resourceName, "baz.example.com.", &rec_name, &rec_zone),
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

func TestAccDnsCnameRecord_basic_upgrade(t *testing.T) {

	var rec_name, rec_zone string
	resourceName := "dns_cname_record.foo"

	deleteCnameRecord := func() {
		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rec_fqdn := testResourceFQDN(rec_name, rec_zone)

		rrStr := fmt.Sprintf("%s 0 CNAME", rec_fqdn)

		rr_remove, err := dns.NewRR(rrStr)
		if err != nil {
			t.Fatalf("Error reading DNS record (%s): %s", rrStr, err)
		}

		msg.RemoveRRset([]dns.RR{rr_remove})

		r, err := exchange(msg, true, dnsClient)
		if err != nil {
			t.Fatalf("Error deleting DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			t.Fatalf("Error deleting DNS record: %v", r.Rcode)
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDnsCnameRecordDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsCnameRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsCnameRecordExists(resourceName, "bar.example.com.", &rec_name, &rec_zone),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsCnameRecord_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsCnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsCnameRecordExists(resourceName, "baz.example.com.", &rec_name, &rec_zone),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsCnameRecord_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				PreConfig:         deleteCnameRecord,
				Config:            testAccDnsCnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsCnameRecordExists(resourceName, "baz.example.com.", &rec_name, &rec_zone),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsCnameRecord_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckDnsCnameRecordDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_cname_record", dns.TypeCNAME)
}

func testAccCheckDnsCnameRecordExists(n, expected string, rec_name, rec_zone *string) resource.TestCheckFunc {
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

		msg := new(dns.Msg)
		msg.SetQuestion(rec_fqdn, dns.TypeCNAME)
		r, err := exchange(msg, false, dnsClient)
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
		cname, _, err := getCnameVal(record)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if expected != cname {
			return fmt.Errorf("DNS record differs: expected %v, found %v", expected, cname)
		}
		return nil
	}
}

var testAccDnsCnameRecord_basic = `
  resource "dns_cname_record" "foo" {
    zone = "example.com."
    name = "foo"
    cname = "bar.example.com."
    ttl = 300
  }`

var testAccDnsCnameRecord_update = `
  resource "dns_cname_record" "foo" {
    zone = "example.com."
    name = "foo"
    cname = "baz.example.com."
    ttl = 300
  }`
