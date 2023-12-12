// Copyright (c) HashiCorp, Inc.
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
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:dead:beef:cafe:babe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:cafe:babe:dead:beef"),
				),
			},
			{
				Config: testAccDnsAAAARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:beef:dead:babe:cafe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:babe:cafe:beef:dead"),
				),
			},
			{
				PreConfig: func() { testRemoveRecord(t, "AAAA", "bar") },
				Config:    testAccDnsAAAARecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:beef:dead:babe:cafe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282:0000:0000:babe:cafe:beef:dead"),
				),
			},
			{
				Config: testAccDnsAAAARecordSet_retry,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "14"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::beef:dead:babe:cafe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::babe:cafe:beef:dead"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::beef:babe:dead:cafe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::babe:beef:cafe:dead"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::cafe:beef:babe:dead"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::cafe:beef:dead:babe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::cafe:babe:dead:beef"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::cafe:babe:beef:dead"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::dead:babe:cafe:beef"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::dead:babe:beef:cafe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::dead:cafe:babe:beef"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::dead:cafe:beef:babe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::dead:beef:cafe:babe"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "fdd5:e282::dead:beef:babe:cafe"),
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
					resource.TestCheckTypeSetElemAttr(resourceRoot, "addresses.*", "fdd5:e282::beef:dead:babe:cafe"),
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
    addresses = ["fdd5:e282:0000:0000:dead:beef:cafe:babe", "fdd5:e282:0000:0000:cafe:babe:dead:beef"]
    ttl = 300
  }`

var testAccDnsAAAARecordSet_update = `
  resource "dns_aaaa_record_set" "bar" {
    zone = "example.com."
    name = "bar"
    addresses = ["fdd5:e282:0000:0000:beef:dead:babe:cafe", "fdd5:e282:0000:0000:babe:cafe:beef:dead"]
    ttl = 300
  }`

var testAccDnsAAAARecordSet_retry = `
  resource "dns_aaaa_record_set" "bar" {
    zone = "example.com."
    name = "bar"
    addresses = ["fdd5:e282::beef:dead:babe:cafe", "fdd5:e282::babe:cafe:beef:dead", "fdd5:e282::beef:babe:dead:cafe", "fdd5:e282::babe:beef:cafe:dead", "fdd5:e282::cafe:beef:babe:dead", "fdd5:e282::cafe:beef:dead:babe", "fdd5:e282::cafe:babe:dead:beef", "fdd5:e282::cafe:babe:beef:dead", "fdd5:e282::dead:babe:cafe:beef", "fdd5:e282::dead:babe:beef:cafe", "fdd5:e282::dead:cafe:babe:beef", "fdd5:e282::dead:cafe:beef:babe", "fdd5:e282::dead:beef:cafe:babe", "fdd5:e282::dead:beef:babe:cafe"]
    ttl = 300
  }`

var testAccDnsAAAARecordSet_root = `
  resource "dns_aaaa_record_set" "root" {
    zone = "example.com."
    addresses = ["fdd5:e282::beef:dead:babe:cafe"]
    ttl = 300
  }`

var testAccDnsAAAARecordSet_root_expanded = `
  resource "dns_aaaa_record_set" "root" {
    zone = "example.com."
    addresses = ["fdd5:e282:0000:0000:beef:dead:babe:cafe"]
    ttl = 300
  }`
