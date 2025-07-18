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
  host = "terraform-provider-dns-ns.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "terraform-provider-dns-ns.hashicorptest.com"),
					resource.TestCheckResourceAttr(recordName, "nameservers.#", "4"),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "ns-1407.awsdns-47.org."),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "ns-1816.awsdns-35.co.uk."),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "ns-307.awsdns-38.com."),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "ns-655.awsdns-17.net."),
				),
			},
		},
	})
}
