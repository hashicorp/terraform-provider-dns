// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataDnsRecursiveCnameRecordSet_Basic(t *testing.T) {
        recordName := "data.dns_recursive_cname_record_set.test"

        resource.UnitTest(t, resource.TestCase{
                ProtoV5ProviderFactories: testProtoV5ProviderFactories,
                Steps: []resource.TestStep{
                        {
                                Config: `
data "dns_recursive_cname_record_set" "test" {
  host = "terraform-provider-dns-cname.hashicorptest.com"
}
`,
                                Check: resource.ComposeAggregateTestCheckFunc(
                                        resource.TestCheckResourceAttr(recordName, "cnames.0", "example.com."),
                                        resource.TestCheckNoResourceAttr(recordName, "cnames.1"),
                                        resource.TestCheckResourceAttr(recordName, "last_cname", "example.com."),
                                        resource.TestCheckResourceAttr(recordName, "host", "terraform-provider-dns-cname.hashicorptest.com"),
                                        resource.TestCheckResourceAttr(recordName, "id", "terraform-provider-dns-cname.hashicorptest.com"),
                                ),
                        },
                },
        })
}

func TestAccDataDnsRecursiveCnameRecordSet_Complex(t *testing.T) {
	recordName := "data.dns_recursive_cname_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_recursive_cname_record_set" "test" {
  host = "test2.tony.docusign.dev"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "cnames.0", "stage.services.docusign.net."),
					resource.TestCheckResourceAttr(recordName, "cnames.1", "stage.services.docusign.net.akadns.net."),
					resource.TestCheckNoResourceAttr(recordName, "cnames.2"),
					resource.TestCheckResourceAttr(recordName, "last_cname", "stage.services.docusign.net.akadns.net."),
					resource.TestCheckResourceAttr(recordName, "host", "test2.tony.docusign.dev"),
					resource.TestCheckResourceAttr(recordName, "id", "test2.tony.docusign.dev"),
				),
			},
		},
	})
}
