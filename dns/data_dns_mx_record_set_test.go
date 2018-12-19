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
		Priority        []int
		Zone            string
	}{
		{
			`
			data "dns_mx_record_set" "foo" {
			zone = "hashicorptest.com"
			}
			`,
			"foo",
			[]string{
				// These results may change if hashicorptest.com changes MX hosts or providers.
				// If you suspect the expected results have changed here, confirm
				// with e.g. dig hashicorptest.com MX +short
				"alt1.aspmx.l.google.com.",
				"alt2.aspmx.l.google.com.",
				"alt3.aspmx.l.google.com.",
				"alt4.aspmx.l.google.com.",
				"aspmx.l.google.com.",
			},
			[]int{
				// These results may change if hashicorptest.com changes MX host priorities, hosts, or providers.
				// If you suspect the expected results have changed here, confirm
				// with e.g. dig hashicorptest.net MX +short
				5,
				5,
				10,
				10,
				1,
			},
			"hashicorptest.com",
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
						testCheckAttrIntArray(recordName, "priorities", test.Priority),
					),
				},
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(recordName, "id", test.Zone),
					),
				},
			},
		})
	}

}
