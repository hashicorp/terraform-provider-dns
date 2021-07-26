package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsARecordSet_Basic(t *testing.T) {
	recordName := "data.dns_a_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_a_record_set" "test" {
  host = "terraform-provider-dns-a.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "addrs.#", "1"),
					resource.TestCheckTypeSetElemAttr(recordName, "addrs.*", "127.0.0.1"),
					resource.TestCheckResourceAttr(recordName, "id", "terraform-provider-dns-a.hashicorptest.com"),
				),
			},
		},
	})
}
