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
  host = "terraform-provider-dns-cname.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "cname", "example.com."),
					resource.TestCheckResourceAttr(recordName, "id", "terraform-provider-dns-cname.hashicorptest.com"),
					resource.TestCheckResourceAttr(recordName, "cname_tree.0", "terraform-provider-dns-cname.hashicorptest.com"),
				),
			},
		},
	})
}

func TestAccDataDnsCnameRecordSet_Advanced(t *testing.T) {
	recordName := "data.dns_cname_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_cname_record_set" "test" {
  host = "test2.tony.docusign.dev"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "cname", "stage.services.docusign.net."),
					resource.TestCheckResourceAttr(recordName, "id", "test2.tony.docusign.dev"),
					resource.TestCheckResourceAttr(recordName, "cname_tree.0", "test2.tony.docusign.dev"),
					resource.TestCheckResourceAttr(recordName, "cname_tree.1", "stage.services.docusign.net."),
				),
			},
		},
	})
}
