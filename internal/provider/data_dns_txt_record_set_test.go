package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				"_globalsign-domain-verification=O81xyb7YxpdGeHWkniit_VBT4vTXz9__NFrNMoTwFg",
				"google-site-verification=LQZvxDzrGE-ZLudDpkpj-gcXN-5yF7Z6C-4Rljs3I_Q",
				"google-site-verification=8d7FpfB8aOEYAIkoaVKxg7Ibj438CEypjZTH424Pews",
				"google-site-verification=y974ACvos30pN7_OBgEZb_byZV8qYtK0G6WZfE7OX8s",
				"google-site-verification=9D7erI6Bfd9EOHKSIXRe0XQaqAFAjToBtZmyYRzMm34",
				"keybase-site-verification=5HKqMvJnTWpe8W-Aa8r0y3wuy1bhQ6LwcjaxKE9BOQU",
			},
			"terraform.io",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_txt_record_set.%s", test.DataSourceName)
		resource.UnitTest(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testCheckAttrStringArray(recordName, "records", test.Expected),
					),
				},
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testCheckAttrStringArrayMember(recordName, "record", test.Expected),
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
