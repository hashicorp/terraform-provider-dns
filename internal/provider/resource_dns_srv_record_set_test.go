package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsSRVRecordSet_Basic(t *testing.T) {
	resourceName := "dns_srv_record_set.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsSRVRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsSRVRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srv.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "srv.*", map[string]string{"priority": "10", "weight": "60", "port": "5060", "target": "bigbox.example.com."}),
				),
			},
			{
				Config: testAccDnsSRVRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srv.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "srv.*", map[string]string{"priority": "10", "weight": "60", "port": "5060", "target": "bigbox.example.com."}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "srv.*", map[string]string{"priority": "20", "weight": "0", "port": "5060", "target": "backupbox.example.com."}),
				),
			},
			{
				PreConfig: func() { deleteSRVRecordSet(t) },
				Config:    testAccDnsSRVRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srv.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "srv.*", map[string]string{"priority": "10", "weight": "60", "port": "5060", "target": "bigbox.example.com."}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "srv.*", map[string]string{"priority": "20", "weight": "0", "port": "5060", "target": "backupbox.example.com."}),
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

func TestAccDnsSRVRecordSet_Basic_Upgrade(t *testing.T) {
	resourceName := "dns_srv_record_set.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDnsSRVRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsSRVRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srv.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "srv.*", map[string]string{"priority": "10", "weight": "60", "port": "5060", "target": "bigbox.example.com."}),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsSRVRecordSet_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsSRVRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srv.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "srv.*", map[string]string{"priority": "10", "weight": "60", "port": "5060", "target": "bigbox.example.com."}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "srv.*", map[string]string{"priority": "20", "weight": "0", "port": "5060", "target": "backupbox.example.com."}),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsSRVRecordSet_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				PreConfig:         func() { deleteSRVRecordSet(t) },
				Config:            testAccDnsSRVRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srv.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "srv.*", map[string]string{"priority": "10", "weight": "60", "port": "5060", "target": "bigbox.example.com."}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "srv.*", map[string]string{"priority": "20", "weight": "0", "port": "5060", "target": "backupbox.example.com."}),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsSRVRecordSet_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func deleteSRVRecordSet(t *testing.T) {
	name := "_sip._tcp"
	zone := "example.com."

	msg := new(dns.Msg)

	msg.SetUpdate(zone)

	fqdn := testResourceFQDN(name, zone)

	rrStr := fmt.Sprintf("%s 0 SRV", fqdn)

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

func testAccCheckDnsSRVRecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_srv_record_set", dns.TypeSRV)
}

var testAccDnsSRVRecordSet_basic = `
  resource "dns_srv_record_set" "foo" {
    zone = "example.com."
    name = "_sip._tcp"
    srv {
      priority = 10
      weight = 60
      port = 5060
      target = "bigbox.example.com."
    }
    ttl = 300
  }`

var testAccDnsSRVRecordSet_update = `
  resource "dns_srv_record_set" "foo" {
    zone = "example.com."
    name = "_sip._tcp"
    srv {
      priority = 10
      weight = 60
      port = 5060
      target = "bigbox.example.com."
    }
    srv {
      priority = 20
      weight = 0
      port = 5060
      target = "backupbox.example.com."
    }
    ttl = 300
  }`
