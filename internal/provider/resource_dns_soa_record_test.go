// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDnsSoaRecord_Basic(t *testing.T) {
	resourceName := "dns_soa_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsSoaRecord_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "mname"),
					resource.TestCheckResourceAttrSet(resourceName, "rname"),
					resource.TestCheckResourceAttrSet(resourceName, "refresh"),
					resource.TestCheckResourceAttrSet(resourceName, "retry"),
					resource.TestCheckResourceAttrSet(resourceName, "expire"),
					resource.TestCheckResourceAttrSet(resourceName, "ttl"),
				),
			},
		},
	})
}

var testAccDnsSoaRecord_basic = `
  resource "dns_soa_record" "test" {
    zone = "example.com."
    mname = "ns1.example.com."
    rname = "admin.example.com."
    refresh = 3600
    retry = 900
    expire = 604800
    ttl = 300
  }`
