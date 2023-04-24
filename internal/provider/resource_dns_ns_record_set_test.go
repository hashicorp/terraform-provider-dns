package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsNSRecordSet_Basic(t *testing.T) {
	resourceName := "dns_ns_record_set.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsNSRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsNSRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns1.testdns.co.uk."),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns2.testdns.co.uk."),
				),
			},
			{
				Config: testAccDnsNSRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns1.test2dns.co.uk."),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns2.test2dns.co.uk."),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns3.test2dns.co.uk."),
				),
			},
			{
				PreConfig: func() { testRemoveRecord(t, "NS", "foo") },
				Config:    testAccDnsNSRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns1.test2dns.co.uk."),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns2.test2dns.co.uk."),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns3.test2dns.co.uk."),
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
	resourceName := "dns_ns_record_set.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDnsNSRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsNSRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns1.testdns.co.uk."),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns2.testdns.co.uk."),
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
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns1.test2dns.co.uk."),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns2.test2dns.co.uk."),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns3.test2dns.co.uk."),
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
				PreConfig:                func() { testRemoveRecord(t, "NS", "foo") },
				Config:                   testAccDnsNSRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "nameservers.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns1.test2dns.co.uk."),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns2.test2dns.co.uk."),
					resource.TestCheckTypeSetElemAttr(resourceName, "nameservers.*", "ns3.test2dns.co.uk."),
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
