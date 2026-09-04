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

func TestAccDnsCAARecordSet_Basic(t *testing.T) {
	resourceName := "dns_caa_record_set.foo"
  resourceRoot := "dns_caa_record_set.root"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsCAARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsCAARecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "caa.#", "1"),
          resource.TestCheckTypeSetElemNestedAttrs(resourceName, "caa.*", map[string]string{"flags": "0", "tag": "issue", "value": ";"}),
				),
			},
			{
				Config: testAccDnsCAARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "caa.#", "2"),
          resource.TestCheckTypeSetElemNestedAttrs(resourceName, "caa.*", map[string]string{"flags": "0", "tag": "issue", "value": "example.com;"}),
				),
			},
			{
				PreConfig: func() { testRemoveRecord(t, "CAA", "foo") },
				Config:    testAccDnsCAARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "caa.#", "2"),
          resource.TestCheckTypeSetElemNestedAttrs(resourceName, "caa.*", map[string]string{"flags": "0", "tag": "issue", "value": "example.com;"}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
      {
        Config: testAccDNSCAARecordSet_root,
        Check: resource.ComposeTestCheckFunc(
          resource.TestCheckResourceAttr(resourceRoot, "caa.#", "1"),
          resource.TestCheckTypeSetElemNestedAttrs(resourceName, "caa.*", map[string]string{"flags": "0", "tag": "issue", "value": ";"}),
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

func TestAccDnsCAARecordSet_Basic_Upgrade(t *testing.T) {
	resourceName := "dns_caa_record_set.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDnsCAARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsCAARecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "caa.#", "1"),
          resource.TestCheckTypeSetElemNestedAttrs(resourceName, "caa.*", map[string]string{"flags": "0", "tag": "issue", "value": ";"}),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsCAARecordSet_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				Config:            testAccDnsCAARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "caa.#", "2"),
          resource.TestCheckTypeSetElemNestedAttrs(resourceName, "caa.*", map[string]string{"flags": "0", "tag": "issue", "value": "example.com;"}),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsCAARecordSet_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ExternalProviders: providerVersion324(),
				PreConfig:         func() { testRemoveRecord(t, "CAA", "foo") },
				Config:            testAccDnsCAARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "caa.#", "2"),
          resource.TestCheckTypeSetElemNestedAttrs(resourceName, "caa.*", map[string]string{"flags": "0", "tag": "issue", "value": "example.com;"}),
				),
			},
			{
				ProtoV5ProviderFactories: testProtoV5ProviderFactories,
				Config:                   testAccDnsCAARecordSet_update,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckDnsCAARecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_caa_record_set", dns.TypeCAA)
}

var testAccDnsCAARecordSet_basic = `
  resource "dns_caa_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    caa {
      flags = 0
      tag = "issue"
      value = ";"
    }
    ttl = 300
  }`

var testAccDnsCAARecordSet_update = `
  resource "dns_caa_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    caa {
      flags = 0
      tag = "issue"
      value = "example.com;"
    }
    ttl = 300
  }`

var testAccDNSCAARecordSet_root = `
    zone = "example.com."
    caa {
      flags = 0
      tag = "issue"
      value = ";"
    }
    ttl = 300
  }`
