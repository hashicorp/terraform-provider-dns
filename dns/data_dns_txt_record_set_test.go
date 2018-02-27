package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataDnsTxtRecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		DataSourceName  string
		Expected        []string
		Host            string
	}{
		{
			`
			data "dns_txt_record_set" "foo" {
			  host = "terraform.io"
			}
			`,
			"foo",
			[]string{
				"google-site-verification=LQZvxDzrGE-ZLudDpkpj-gcXN-5yF7Z6C-4Rljs3I_Q",
			},
			"terraform.io",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_txt_record_set.%s", test.DataSourceName)
		resource.UnitTest(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testCheckAttrStringArray(recordName, "records", test.Expected),
					),
				},
				resource.TestStep{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testCheckAttrStringArrayMember(recordName, "record", test.Expected),
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
