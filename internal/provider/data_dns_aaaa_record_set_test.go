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
  host = "terraform-provider-dns-aaaa.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "addrs.#", "1"),
					resource.TestCheckTypeSetElemAttr(recordName, "addrs.*", "::1"),
					resource.TestCheckResourceAttr(recordName, "id", "terraform-provider-dns-aaaa.hashicorptest.com"),
				),
			},
		},
	})
}
