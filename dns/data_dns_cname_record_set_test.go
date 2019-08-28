package dns

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataDnsCnameRecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		Expected        string
		Host            string
	}{
		{
			`
			data "dns_cname_record_set" "foo" {
			  host = "www.hashicorp.com"
			}
			`,
			"hashicorp.netlifyglobalcdn.com.",
			"www.hashicorp.com",
		},
	}

	for _, test := range tests {
		resource.UnitTest(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.dns_cname_record_set.foo", "cname", test.Expected),
					),
				},
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.dns_cname_record_set.foo", "id", test.Host),
					),
				},
			},
		})
	}
}
