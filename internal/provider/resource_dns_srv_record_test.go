// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsSrvRecord_Basic(t *testing.T) {
	resourceName := "dns_srv_record.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsSrvRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsSrvRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "priority", "10"),
					resource.TestCheckResourceAttr(resourceName, "weight", "60"),
					resource.TestCheckResourceAttr(resourceName, "port", "5060"),
					resource.TestCheckResourceAttr(resourceName, "target", "bigbox.example.com."),
				),
			},
			{
				Config: testAccDnsSrvRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "priority", "20"),
					resource.TestCheckResourceAttr(resourceName, "weight", "0"),
					resource.TestCheckResourceAttr(resourceName, "port", "5070"),
					resource.TestCheckResourceAttr(resourceName, "target", "backupbox.example.com."),
				),
			},
			{
				PreConfig: func() { testRemoveRecord(t, "SRV", "foo._sip._tcp") },
				Config:    testAccDnsSrvRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "priority", "20"),
					resource.TestCheckResourceAttr(resourceName, "weight", "0"),
					resource.TestCheckResourceAttr(resourceName, "port", "5070"),
					resource.TestCheckResourceAttr(resourceName, "target", "backupbox.example.com."),
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

func TestAccDnsSrvRecord_Basic_Upgrade(t *testing.T) {
	resourceName := "dns_srv_record.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDnsSrvRecordDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsSrvRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "priority", "10"),
					resource.TestCheckResourceAttr(resourceName, "weight", "60"),
					resource.TestCheckResourceAttr(resourceName, "port", "5060"),
					resource.TestCheckResourceAttr(resourceName, "target", "bigbox.example.com."),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsSrvRecord_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsSrvRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "priority", "20"),
					resource.TestCheckResourceAttr(resourceName, "weight", "0"),
					resource.TestCheckResourceAttr(resourceName, "port", "5070"),
					resource.TestCheckResourceAttr(resourceName, "target", "backupbox.example.com."),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsSrvRecord_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				PreConfig:         func() { testRemoveRecord(t, "SRV", "foo._sip._tcp") },
				Config:            testAccDnsSrvRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "priority", "20"),
					resource.TestCheckResourceAttr(resourceName, "weight", "0"),
					resource.TestCheckResourceAttr(resourceName, "port", "5070"),
					resource.TestCheckResourceAttr(resourceName, "target", "backupbox.example.com."),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsSrvRecord_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckDnsSrvRecordDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_srv_record", dns.TypeSRV)
}

var testAccDnsSrvRecord_basic = `
  resource "dns_srv_record" "foo" {
    zone = "example.com."
    name = "foo"
    service = "_sip._tcp"
    priority = 10
    weight = 60
    port = 5060
    target = "bigbox.example.com."
    ttl = 300
  }`

var testAccDnsSrvRecord_update = `
  resource "dns_srv_record" "foo" {
    zone = "example.com."
    name = "foo"
    service = "_sip._tcp"
    priority = 20
    weight = 0
    port = 5070
    target = "backupbox.example.com."
    ttl = 300
  }`
