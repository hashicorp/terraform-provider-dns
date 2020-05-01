package dns

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsPtrRecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		Expected        string
		IPAddress       string
	}{
		{
			`
			data "dns_ptr_record_set" "foo" {
			  ip_address = "8.8.8.8"
			}
			`,
			"dns.google.",
			"8.8.8.8",
		},
	}

	for _, test := range tests {
		resource.UnitTest(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.dns_ptr_record_set.foo", "ptr", test.Expected),
					),
				},
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.dns_ptr_record_set.foo", "id", test.IPAddress),
					),
				},
			},
		})
	}
}
