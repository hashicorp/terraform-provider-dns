package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsCnameRecord_Basic(t *testing.T) {
	resourceName := "dns_cname_record.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsCnameRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsCnameRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cname", "bar.example.com."),
				),
			},
			{
				Config: testAccDnsCnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cname", "baz.example.com."),
				),
			},
			{
				PreConfig: func() { deleteCnameRecord(t) },
				Config:    testAccDnsCnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cname", "baz.example.com."),
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

func TestAccDnsCnameRecord_Basic_Upgrade(t *testing.T) {
	resourceName := "dns_cname_record.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDnsCnameRecordDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsCnameRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cname", "bar.example.com."),
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
					resource.TestCheckResourceAttr(resourceName, "cname", "baz.example.com."),
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
				PreConfig:         func() { deleteCnameRecord(t) },
				Config:            testAccDnsCnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cname", "baz.example.com."),
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

func deleteCnameRecord(t *testing.T) {
	rec_name := "baz"
	rec_zone := "example.com."

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

func testAccCheckDnsCnameRecordDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_cname_record", dns.TypeCNAME)
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
