package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsAAAARecordSet_Basic(t *testing.T) {
	recordName := "data.dns_aaaa_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_aaaa_record_set" "test" {
  host = "example.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "addrs.#", "1"),
					resource.TestCheckTypeSetElemAttr(recordName, "addrs.*", "2606:2800:220:1:248:1893:25c8:1946"),
					resource.TestCheckResourceAttr(recordName, "id", "example.com"),
				),
			},
		},
	})
}
