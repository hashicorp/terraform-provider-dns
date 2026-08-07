// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataDnsSOARecordSet_Basic(t *testing.T) {
	recordName := "data.dns_soa_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_soa_record_set" "test" {
  zone = "google.com."
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "google.com."),
					resource.TestCheckResourceAttr(recordName, "mname", "ns1.google.com."),
					resource.TestCheckResourceAttr(recordName, "rname", "dns-admin.google.com."),
					resource.TestCheckResourceAttrSet(recordName, "serial"),
					resource.TestCheckResourceAttrSet(recordName, "refresh"),
					resource.TestCheckResourceAttrSet(recordName, "retry"),
					resource.TestCheckResourceAttrSet(recordName, "expire"),
					resource.TestCheckResourceAttrSet(recordName, "ttl"),
				),
			},
		},
	})
}
