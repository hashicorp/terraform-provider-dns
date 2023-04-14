package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsMXRecordSet_Basic(t *testing.T) {
	resourceName := "dns_mx_record_set.foo"
	resourceRoot := "dns_mx_record_set.root"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsMXRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsMXRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "mx.*", map[string]string{"preference": "10", "exchange": "smtp.example.org."}),
				),
			},
			{
				Config: testAccDnsMXRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "mx.*", map[string]string{"preference": "10", "exchange": "smtp.example.org."}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "mx.*", map[string]string{"preference": "20", "exchange": "backup.example.org."}),
				),
			},
			{
				PreConfig: func() { deleteMXRecordSet(t) },
				Config:    testAccDnsMXRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "mx.*", map[string]string{"preference": "10", "exchange": "smtp.example.org."}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "mx.*", map[string]string{"preference": "20", "exchange": "backup.example.org."}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsMXRecordSet_root,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "mx.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceRoot, "mx.*", map[string]string{"preference": "10", "exchange": "smtp.example.org."}),
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

func TestAccDnsMXRecordSet_Basic_Upgrade(t *testing.T) {
	resourceName := "dns_mx_record_set.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDnsMXRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsMXRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "mx.*", map[string]string{"preference": "10", "exchange": "smtp.example.org."}),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsMXRecordSet_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsMXRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "mx.*", map[string]string{"preference": "10", "exchange": "smtp.example.org."}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "mx.*", map[string]string{"preference": "20", "exchange": "backup.example.org."}),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsMXRecordSet_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				PreConfig:         func() { deleteMXRecordSet(t) },
				Config:            testAccDnsMXRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "mx.*", map[string]string{"preference": "10", "exchange": "smtp.example.org."}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "mx.*", map[string]string{"preference": "20", "exchange": "backup.example.org."}),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsMXRecordSet_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func deleteMXRecordSet(t *testing.T) {
	name := "foo"
	zone := "example.com."

	msg := new(dns.Msg)

	msg.SetUpdate(zone)

	fqdn := testResourceFQDN(name, zone)

	rrStr := fmt.Sprintf("%s 0 MX", fqdn)

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

func testAccCheckDnsMXRecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_mx_record_set", dns.TypeMX)
}

var testAccDnsMXRecordSet_basic = `
  resource "dns_mx_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    mx {
      preference = 10
      exchange = "smtp.example.org."
    }
    ttl = 300
  }`

var testAccDnsMXRecordSet_update = `
  resource "dns_mx_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    mx {
      preference = 10
      exchange = "smtp.example.org."
    }
    mx {
      preference = 20
      exchange = "backup.example.org."
    }
    ttl = 300
  }`

var testAccDnsMXRecordSet_root = `
  resource "dns_mx_record_set" "root" {
    zone = "example.com."
    mx {
      preference = 10
      exchange = "smtp.example.org."
    }
    ttl = 300
  }`
