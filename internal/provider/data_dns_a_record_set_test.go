// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataDnsARecordSet_Basic(t *testing.T) {
	recordName := "data.dns_a_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_a_record_set" "test" {
  host = "a.dns.tfacc.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "addrs.#", "1"),
					resource.TestCheckTypeSetElemAttr(recordName, "addrs.*", "127.0.0.1"),
					resource.TestCheckResourceAttr(recordName, "id", "a.dns.tfacc.hashicorptest.com"),
				),
			},
		},
	})
}

func TestAccDataDnsARecordSet_BasicUpdateProvider(t *testing.T) {
	recordName := "data.dns_a_record_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_a_record_set" "test" {
  host = "terraform-provider-dns-a.hashicorptest.com"
  use_update_server = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "addrs.#", "1"),
					resource.TestCheckTypeSetElemAttr(recordName, "addrs.*", "127.0.0.1"),
					resource.TestCheckResourceAttr(recordName, "id", "terraform-provider-dns-a.hashicorptest.com"),
				),
			},
		},
	})
}
