// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestAccStripLeadingZeroesFunction_ipv4(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
output "test" {
  value = provider::dns::strip_leading_zeroes(["127.000.000.001", "10.000.000.001"])
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.SetExact(
						[]knownvalue.Check{
							knownvalue.StringExact("127.0.0.1"),
							knownvalue.StringExact("10.0.0.1"),
						},
					)),
				},
			},
		},
	})
}

func TestAccStripLeadingZeroesFunction_ipv6(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
output "test" {
  value = provider::dns::strip_leading_zeroes(["FF01:0000:0000:0000:0000:0000:0000:0101", "2001:0DB8:0000:0000:0008:0800:200C:417A"])
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.SetExact(
						[]knownvalue.Check{
							knownvalue.StringExact("FF01:0:0:0:0:0:0:101"),
							knownvalue.StringExact("2001:DB8:0:0:8:800:200C:417A"),
						},
					)),
				},
			},
		},
	})
}
