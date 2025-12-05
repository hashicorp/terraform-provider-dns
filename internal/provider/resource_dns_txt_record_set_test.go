// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
				PreConfig: func() { testRemoveRecord(t, "TXT", "foo") },
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
				PreConfig:         func() { testRemoveRecord(t, "TXT", "foo") },
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
