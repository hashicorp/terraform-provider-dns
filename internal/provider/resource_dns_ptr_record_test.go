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

func TestAccDnsPtrRecord_Basic(t *testing.T) {
	resourceName := "dns_ptr_record.foo"
	resourceRoot := "dns_ptr_record.root"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsPtrRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsPtrRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ptr", "bar.example.com."),
				),
			},
			{
				Config: testAccDnsPtrRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ptr", "baz.example.com."),
				),
			},
			{
				PreConfig: func() { testRemoveRecord(t, "PTR", "r._dns-sd._udp") },
				Config:    testAccDnsPtrRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ptr", "baz.example.com."),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsPtrRecord_root,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "ptr", "baz.example.com."),
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

func TestAccDnsPtrRecord_Basic_Upgrade(t *testing.T) {
	resourceName := "dns_ptr_record.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDnsPtrRecordDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsPtrRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ptr", "bar.example.com."),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsPtrRecord_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsPtrRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ptr", "baz.example.com."),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsPtrRecord_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				PreConfig:         func() { testRemoveRecord(t, "PTR", "r._dns-sd._udp") },
				Config:            testAccDnsPtrRecord_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ptr", "baz.example.com."),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsPtrRecord_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckDnsPtrRecordDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_ptr_record", dns.TypePTR)
}

var testAccDnsPtrRecord_basic = `
  resource "dns_ptr_record" "foo" {
    zone = "example.com."
    name = "r._dns-sd._udp"
    ptr = "bar.example.com."
    ttl = 300
  }`

var testAccDnsPtrRecord_update = `
  resource "dns_ptr_record" "foo" {
    zone = "example.com."
    name = "r._dns-sd._udp"
    ptr = "baz.example.com."
    ttl = 300
  }`

var testAccDnsPtrRecord_root = `
  resource "dns_ptr_record" "root" {
    zone = "example.com."
    ptr = "baz.example.com."
    ttl = 300
  }`
