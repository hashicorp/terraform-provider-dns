package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataDnsMXRecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		DataSourceName  string
		Expected        []string
		Host            string
	}{
		{
			`
			data "dns_mx_record_set" "foo" {
			  host = "hashicorp.net"
			}
			`,
			"foo",
			[]string{
				// These results may change if hashicorp.net changes MX hosts or providers.
				// If you suspect the expected results have changed here, confirm
				// with e.g. dig hashicorp.net MX +short
				"eforward1.registrar-servers.com.",
				"eforward2.registrar-servers.com.",
				"eforward3.registrar-servers.com.",
				"eforward4.registrar-servers.com.",
				"eforward5.registrar-servers.com.",
			},
			"hashicorp.net",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_mx_record_set.%s", test.DataSourceName)

		resource.Test(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testCheckAttrStringArray(recordName, "mxservers", test.Expected),
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
