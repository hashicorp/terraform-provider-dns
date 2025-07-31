// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataDnsMXRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_mx_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_mx_record_set" "test" {
  domain = "mx.dns.tfacc.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "mx.dns.tfacc.hashicorptest.com"),
					resource.TestCheckResourceAttr(recordName, "mx.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(recordName, "mx.*", map[string]string{
						"exchange":   "example.com.",
						"preference": "1",
					}),
				),
			},
		},
	})
}
