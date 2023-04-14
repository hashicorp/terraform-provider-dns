package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/miekg/dns"
)

func TestAccDnsTXTRecordSet_Basic(t *testing.T) {
	resourceName := "dns_txt_record_set.foo"
	resourceRoot := "dns_txt_record_set.root"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsTXTRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "txt.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "foo"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "bar"),
				),
			},
			{
				Config: testAccDnsTXTRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "txt.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "foo"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "bar"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "baz"),
				),
			},
			{
				PreConfig: func() { deleteTXTRecordSet(t) },
				Config:    testAccDnsTXTRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "txt.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "foo"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "bar"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "baz"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsTXTRecordSet_root,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "txt.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceRoot, "txt.*", "foo"),
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

func TestAccDnsTXTRecordSet_Basic_Upgrade(t *testing.T) {
	resourceName := "dns_txt_record_set.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDnsARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsTXTRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "txt.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "foo"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "bar"),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsTXTRecordSet_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsTXTRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "txt.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "foo"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "bar"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "baz"),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsTXTRecordSet_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				PreConfig:         func() { deleteTXTRecordSet(t) },
				Config:            testAccDnsTXTRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "txt.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "foo"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "bar"),
					resource.TestCheckTypeSetElemAttr(resourceName, "txt.*", "baz"),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsTXTRecordSet_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func deleteTXTRecordSet(t *testing.T) {
	name := "foo"
	zone := "example.com."

	msg := new(dns.Msg)

	msg.SetUpdate(zone)

	fqdn := testResourceFQDN(name, zone)

	rrStr := fmt.Sprintf("%s 0 TXT", fqdn)

	rr_remove, err := dns.NewRR(rrStr)
	if err != nil {
		t.Fatalf("Error reading DNS record: %s", err)
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

var testAccDnsTXTRecordSet_basic = `
  resource "dns_txt_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    txt = ["foo", "bar"]
    ttl = 300
  }`

var testAccDnsTXTRecordSet_update = `
  resource "dns_txt_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    txt = ["foo", "bar", "baz"]
    ttl = 300
  }`

var testAccDnsTXTRecordSet_root = `
  resource "dns_txt_record_set" "root" {
    zone = "example.com."
    txt = ["foo"]
    ttl = 300
  }`
