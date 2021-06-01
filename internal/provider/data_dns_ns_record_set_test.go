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
  host = "terraform.io"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "terraform.io"),
					resource.TestCheckResourceAttr(recordName, "nameservers.#", "2"),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "sam.ns.cloudflare.com."),
					resource.TestCheckTypeSetElemAttr(recordName, "nameservers.*", "zara.ns.cloudflare.com."),
				),
			},
		},
	})
}
