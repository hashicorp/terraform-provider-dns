// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataDnsDnameRecordSet_Basic(t *testing.T) {
	t.Skip("DNAME records are rare; requires test infrastructure")
	recordName := "data.dns_dname_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_dname_record_set" "test" {
  host = "_dname.test.example.com."
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(recordName, "target"),
					resource.TestCheckResourceAttrSet(recordName, "ttl"),
					resource.TestCheckResourceAttr(recordName, "id", "_dname.test.example.com."),
				),
			},
		},
	})
}
