package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsNSRecordSet_Basic(t *testing.T) {

	var rec_name, rec_zone string
	resourceName := "dns_ns_record_set.foo"

	deleteNSRecordSet := func() {
		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rec_fqdn := testResourceFQDN(rec_name, rec_zone)

		rrStr := fmt.Sprintf("%s 0 NS", rec_fqdn)

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
		CheckDestroy:             testAccCheckDnsNSRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsNSRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "2"),
					testAccCheckDnsNSRecordSetExists(resourceName, []interface{}{"ns1.testdns.co.uk.", "ns2.testdns.co.uk."}, &rec_name, &rec_zone),
				),
			},
			{
				Config: testAccDnsNSRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "3"),
					testAccCheckDnsNSRecordSetExists(resourceName, []interface{}{"ns1.test2dns.co.uk.", "ns2.test2dns.co.uk.", "ns3.test2dns.co.uk."}, &rec_name, &rec_zone),
				),
			},
			{
				PreConfig: deleteNSRecordSet,
				Config:    testAccDnsNSRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "3"),
					testAccCheckDnsNSRecordSetExists(resourceName, []interface{}{"ns1.test2dns.co.uk.", "ns2.test2dns.co.uk.", "ns3.test2dns.co.uk."}, &rec_name, &rec_zone),
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

func TestAccDnsNSRecordSet_Basic_Upgrade(t *testing.T) {

	var rec_name, rec_zone string
	resourceName := "dns_ns_record_set.foo"

	deleteNSRecordSet := func() {
		msg := new(dns.Msg)

		msg.SetUpdate(rec_zone)

		rec_fqdn := testResourceFQDN(rec_name, rec_zone)

		rrStr := fmt.Sprintf("%s 0 NS", rec_fqdn)

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
		CheckDestroy: testAccCheckDnsNSRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsNSRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "2"),
					testAccCheckDnsNSRecordSetExists(resourceName, []interface{}{"ns1.testdns.co.uk.", "ns2.testdns.co.uk."}, &rec_name, &rec_zone),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsNSRecordSet_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsNSRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "3"),
					testAccCheckDnsNSRecordSetExists(resourceName, []interface{}{"ns1.test2dns.co.uk.", "ns2.test2dns.co.uk.", "ns3.test2dns.co.uk."}, &rec_name, &rec_zone),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsNSRecordSet_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				PreConfig:                deleteNSRecordSet,
				Config:                   testAccDnsNSRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "3"),
					testAccCheckDnsNSRecordSetExists(resourceName, []interface{}{"ns1.test2dns.co.uk.", "ns2.test2dns.co.uk.", "ns3.test2dns.co.uk."}, &rec_name, &rec_zone),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsNSRecordSet_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckDnsNSRecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_ns_record_set", dns.TypeNS)
}

func testAccCheckDnsNSRecordSetExists(n string, nameservers []interface{}, rec_name, rec_zone *string) resource.TestCheckFunc {
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
		msg.SetQuestion(rec_fqdn, dns.TypeNS)

		r, err := exchange(msg, false, dnsClient)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error querying DNS record")
		}

		var answers []string
		for _, record := range r.Ns {
			nameserver, _, err := getNSVal(record)
			if err != nil {
				return fmt.Errorf("Error querying DNS record: %s", err)
			}
			answers = append(answers, nameserver)
		}

		existing, diags := types.SetValueFrom(context.Background(), types.StringType, answers)
		if diags.HasError() {
			return fmt.Errorf("couldn't create set from answers")
		}
		expected, diags := types.SetValueFrom(context.Background(), types.StringType, nameservers)
		if diags.HasError() {
			return fmt.Errorf("couldn't create set from given nameservers param")
		}

		if !existing.Equal(expected) {
			return fmt.Errorf("DNS record differs: expected %v, found %v", expected, existing)
		}
		return nil
	}
}

var testAccDnsNSRecordSet_basic = `
  resource "dns_ns_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    nameservers = [
		"ns1.testdns.co.uk.",
		"ns2.testdns.co.uk.",
		]
    ttl = 60
  }`

var testAccDnsNSRecordSet_update = `
  resource "dns_ns_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    nameservers = ["ns1.test2dns.co.uk.", "ns2.test2dns.co.uk.", "ns3.test2dns.co.uk.",]
    ttl = 60
  }`
