// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
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
				PreConfig: func() { testRemoveRecord(t, "CNAME", "baz") },
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
				PreConfig:         func() { testRemoveRecord(t, "CNAME", "baz") },
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
