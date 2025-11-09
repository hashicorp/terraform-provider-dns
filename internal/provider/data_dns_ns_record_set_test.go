// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataDnsNSRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_ns_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_ns_record_set" "test" {
  host = "ns.dns.tfacc.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "ns.dns.tfacc.hashicorptest.com"),
					resource.TestCheckResourceAttr(recordName, "nameservers.#", "4"),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "ns-1172.awsdns-18.org."),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "ns-1747.awsdns-26.co.uk."),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "ns-67.awsdns-08.com."),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "ns-555.awsdns-05.net."),
				),
			},
		},
	})
}
