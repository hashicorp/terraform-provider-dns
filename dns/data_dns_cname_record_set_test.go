package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataDnsCnameRecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		DataSourceName  string
		Expected        string
		Host            string
	}{
		{
			`
			data "dns_cname_record_set" "foo" {
			  host = "www.hashicorp.com"
			}
			`,
			"foo",
			"hashicorp.netlifyglobalcdn.com.",
			"www.hashicorp.com",
		},
		{
			`
			data "dns_cname_record_set" "non-existent" {
			  host          = "jolly.roger"
			  ignore_errors = true
			}
			`,
			"non-existent",
			"",
			"jolly.roger",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_cname_record_set.%s", test.DataSourceName)

		resource.UnitTest(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(recordName, "cname", test.Expected),
					),
				},
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(recordName, "id", test.Host),
					),
				},
			},
		})
	}
}
