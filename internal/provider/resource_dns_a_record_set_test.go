// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsARecordSet_Basic(t *testing.T) {

	resourceName := "dns_a_record_set.foo"
	resourceRoot := "dns_a_record_set.root"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsARecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "192.168.000.002"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "192.168.000.001"),
				),
			},
			{
				Config: testAccDnsARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.1"),
				),
			},
			{
				PreConfig: func() { testRemoveRecord(t, "A", "foo") },
				Config:    testAccDnsARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "10.0.0.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsARecordSet_root,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "addresses.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceRoot, "addresses.*", "192.168.0.1"),
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

func testAccCheckDnsARecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_a_record_set", dns.TypeA)
}

var testAccDnsARecordSet_basic = `
  resource "dns_a_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    addresses = ["192.168.000.001", "192.168.000.002"]
    ttl = 300
  }`

var testAccDnsARecordSet_update = `
  resource "dns_a_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    addresses = ["10.0.0.1", "10.0.0.2", "10.0.0.3"]
    ttl = 300
  }`

var testAccDnsARecordSet_root = `
  resource "dns_a_record_set" "root" {
    zone = "example.com."
    addresses = ["192.168.0.1"]
    ttl = 300
  }`
