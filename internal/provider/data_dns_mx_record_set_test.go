package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsMXRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_mx_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_mx_record_set" "test" {
  domain = "terraform-provider-dns-mx.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "terraform-provider-dns-mx.hashicorptest.com"),
					resource.TestCheckResourceAttr(recordName, "mx.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(recordName, "mx.*", map[string]string{
						"exchange":   "example.com.",
						"preference": "10",
					}),
				),
			},
		},
	})
}
