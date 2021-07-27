package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsNSRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_ns_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_ns_record_set" "test" {
  host = "terraform-provider-dns-ns.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "terraform-provider-dns-ns.hashicorptest.com"),
					resource.TestCheckResourceAttr(recordName, "nameservers.#", "2"),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "adaline.ns.cloudflare.com."),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "kenneth.ns.cloudflare.com."),
				),
			},
		},
	})
}
