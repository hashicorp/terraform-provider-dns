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

func TestAccDnsDnameRecord_Basic(t *testing.T) {
	resourceName := "dns_dname_record.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsDnameRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsDnameRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target", "alias.example.com."),
				),
			},
			{
				Config: testAccDnsDnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target", "alias2.example.com."),
				),
			},
			{
				PreConfig: func() { testRemoveRecord(t, "DNAME", "dname-test") },
				Config:    testAccDnsDnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target", "alias2.example.com."),
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

func TestAccDnsDnameRecord_Basic_Upgrade(t *testing.T) {
	resourceName := "dns_dname_record.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDnsDnameRecordDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsDnameRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target", "alias.example.com."),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsDnameRecord_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsDnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target", "alias2.example.com."),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsDnameRecord_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				PreConfig:         func() { testRemoveRecord(t, "DNAME", "dname-test") },
				Config:            testAccDnsDnameRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target", "alias2.example.com."),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsDnameRecord_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckDnsDnameRecordDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_dname_record", dns.TypeDNAME)
}

var testAccDnsDnameRecord_basic = `
  resource "dns_dname_record" "foo" {
    zone = "example.com."
    name = "dname-test"
    target = "alias.example.com."
    ttl = 300
  }`

var testAccDnsDnameRecord_update = `
  resource "dns_dname_record" "foo" {
    zone = "example.com."
    name = "dname-test"
    target = "alias2.example.com."
    ttl = 300
  }`
