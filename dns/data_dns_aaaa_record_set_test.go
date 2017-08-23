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
			  host = "nist-time-server.eoni.com"
			}
			`,
			"ntp",
			[]string{
				"2607:f248::45",
			},
			"nist-time-server.eoni.com",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_aaaa_record_set.%s", test.DataSourceName)

		resource.Test(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testCheckAttrStringArray(recordName, "addrs", test.Expected),
					),
				},
				resource.TestStep{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(recordName, "id", test.Host),
					),
				},
			},
		})
	}

}
