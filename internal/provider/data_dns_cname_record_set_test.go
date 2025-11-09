// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataDnsCnameRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_cname_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_cname_record_set" "test" {
  host = "cname.dns.tfacc.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "cname", "example.com."),
					resource.TestCheckResourceAttr(recordName, "id", "cname.dns.tfacc.hashicorptest.com"),
				),
			},
		},
	})
}
