package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataDnsPtrRecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		DataSourceName  string
		Expected        string
		IPAddress       string
	}{
		{
			`
			data "dns_ptr_record_set" "foo" {
			  ip_address = "8.8.8.8"
			}
			`,
			"foo",
			"dns.google.",
			"8.8.8.8",
		},
		{
			`
			data "dns_ptr_record_set" "non-existent" {
			  ip_address    = "255.255.255.255"
			  ignore_errors = true
			}
			`,
			"non-existent",
			"",
			"255.255.255.255",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_ptr_record_set.%s", test.DataSourceName)

		resource.UnitTest(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(recordName, "ptr", test.Expected),
					),
				},
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(recordName, "id", test.IPAddress),
					),
				},
			},
		})
	}
}
