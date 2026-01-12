// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsAAAARecordSet_basic(t *testing.T) {

	resourceName := "dns_aaaa_record_set.bar"
	resourceRoot := "dns_aaaa_record_set.root"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsAAAARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsAAAARecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:1234:5678:cafe:9012"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:cafe:9012:1234:5678"),
				),
			},
			{
				Config: testAccDnsAAAARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:5678:1234:9012:cafe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:9012:cafe:5678:1234"),
				),
			},
			{
				PreConfig: func() { testRemoveRecord(t, "AAAA", "bar") },
				Config:    testAccDnsAAAARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:5678:1234:9012:cafe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:9012:cafe:5678:1234"),
				),
			},
			{
				Config: testAccDnsAAAARecordSet_retry,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "14"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::5678:1234:9012:cafe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::9012:cafe:5678:1234"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::5678:9012:1234:cafe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::9012:5678:cafe:1234"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::cafe:5678:9012:1234"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::cafe:5678:1234:9012"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::cafe:9012:1234:5678"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::cafe:9012:5678:1234"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::1234:9012:cafe:5678"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::1234:9012:5678:cafe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::1234:cafe:9012:5678"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::1234:cafe:5678:9012"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::1234:5678:cafe:9012"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::1234:5678:9012:cafe"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsAAAARecordSet_root,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "addresses.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceRoot, "addresses.*", "fdd5:e282::5678:1234:9012:cafe"),
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

func TestAccDnsAAAARecordSet_basic_root_expanded(t *testing.T) {
	resourceRoot := "dns_aaaa_record_set.root"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsAAAARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsAAAARecordSet_root_expanded,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "addresses.#", "1"),
				),
			},
			{
				ResourceName:            resourceRoot,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"addresses.0"},
			},
		},
	})
}

func testAccCheckDnsAAAARecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_aaaa_record_set", dns.TypeAAAA)
}

var testAccDnsAAAARecordSet_basic = `
  resource "dns_aaaa_record_set" "bar" {
    zone = "example.com."
    name = "bar"
    addresses = ["fdd5:e282:0000:0000:1234:5678:cafe:9012", "fdd5:e282:0000:0000:cafe:9012:1234:5678"]
    ttl = 300
  }`

var testAccDnsAAAARecordSet_update = `
  resource "dns_aaaa_record_set" "bar" {
    zone = "example.com."
    name = "bar"
    addresses = ["fdd5:e282:0000:0000:5678:1234:9012:cafe", "fdd5:e282:0000:0000:9012:cafe:5678:1234"]
    ttl = 300
  }`

var testAccDnsAAAARecordSet_retry = `
  resource "dns_aaaa_record_set" "bar" {
    zone = "example.com."
    name = "bar"
    addresses = ["fdd5:e282::5678:1234:9012:cafe", "fdd5:e282::9012:cafe:5678:1234", "fdd5:e282::5678:9012:1234:cafe", "fdd5:e282::9012:5678:cafe:1234", "fdd5:e282::cafe:5678:9012:1234", "fdd5:e282::cafe:5678:1234:9012", "fdd5:e282::cafe:9012:1234:5678", "fdd5:e282::cafe:9012:5678:1234", "fdd5:e282::1234:9012:cafe:5678", "fdd5:e282::1234:9012:5678:cafe", "fdd5:e282::1234:cafe:9012:5678", "fdd5:e282::1234:cafe:5678:9012", "fdd5:e282::1234:5678:cafe:9012", "fdd5:e282::1234:5678:9012:cafe"]
    ttl = 300
  }`

var testAccDnsAAAARecordSet_root = `
  resource "dns_aaaa_record_set" "root" {
    zone = "example.com."
    addresses = ["fdd5:e282::5678:1234:9012:cafe"]
    ttl = 300
  }`

var testAccDnsAAAARecordSet_root_expanded = `
  resource "dns_aaaa_record_set" "root" {
    zone = "example.com."
    addresses = ["fdd5:e282:0000:0000:5678:1234:9012:cafe"]
    ttl = 300
  }`
