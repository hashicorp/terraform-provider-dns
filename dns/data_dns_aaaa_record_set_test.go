package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataDnsAAAARecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		DataSourceName  string
		Expected        []string
		Host            string
	}{
		{
			`
			data "dns_aaaa_record_set" "ntp" {
			  host = "example.com"
			}
			`,
			"ntp",
			[]string{
				"2606:2800:220:1:248:1893:25c8:1946",
			},
			"example.com",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_aaaa_record_set.%s", test.DataSourceName)

		resource.UnitTest(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testCheckAttrStringArray(recordName, "addrs", test.Expected),
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
