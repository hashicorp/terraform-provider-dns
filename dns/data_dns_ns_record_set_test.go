package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataDnsNSRecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		DataSourceName  string
		Expected        []string
		Host            string
	}{
		{
			`
			data "dns_ns_record_set" "foo" {
			  host = "terraform.io"
			}
			`,
			"foo",
			[]string{
				// These results may change if terraform.io moves to a new DNS host.
				// If you suspect the expected results have changed here, confirm
				// with e.g. dig terraform.io NS
				"sam.ns.cloudflare.com.",
				"zara.ns.cloudflare.com.",
			},
			"terraform.io",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_ns_record_set.%s", test.DataSourceName)

		resource.Test(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testCheckAttrStringArray(recordName, "nameservers", test.Expected),
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
