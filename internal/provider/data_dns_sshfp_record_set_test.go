// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataDnsSSHFPRecordSet_Basic(t *testing.T) {
	t.Skip("SSHFP records not available on test DNS infrastructure")
	recordName := "data.dns_sshfp_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_sshfp_record_set" "test" {
  host = "_ssh.test.example.com."
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(recordName, "algorithm"),
					resource.TestCheckResourceAttrSet(recordName, "fingerprint_type"),
					resource.TestCheckResourceAttrSet(recordName, "fingerprint"),
				),
			},
		},
	})
}
