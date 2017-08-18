package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataDnsARecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		DataSourceName  string
		Expected        []string
		Host            string
	}{
		{
			`
			data "dns_NS_record_set" "foo" {
			  host = "test.nip.io"
			}
			`,
			"foo",
			[]string{
				"ns-test1.testdns.co.in",
			},
			"test.nip.io",
		},
		{
			`
			data "dns_NS_record_set" "ntp" {
			  host = "time-c.nist.gov"
			}
			`,
			"ntp",
			[]string{
				"ns-test2.testdns.co.uk",
			},
			"time-c.nist.gov",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_ns_record_set.%s", test.DataSourceName)

		resource.Test(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testCheckAttrStringArray(recordName, "nameservers", test.Expected),
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
