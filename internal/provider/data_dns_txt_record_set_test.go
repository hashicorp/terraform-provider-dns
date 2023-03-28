package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsTxtRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_txt_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_txt_record_set" "test" {
  host = "terraform-provider-dns-txt.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "terraform-provider-dns-txt.hashicorptest.com"),
					resource.TestCheckResourceAttr(recordName, "record", "v=spf1 -all"),
					resource.TestCheckResourceAttr(recordName, "records.#", "1"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "v=spf1 -all"),
				),
			},
		},
	})
}

func TestAccDataDnsTxtRecordSet_512Byte(t *testing.T) {
	recordName := "data.dns_txt_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_txt_record_set" "test" {
  host = "terraform.io"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "terraform.io"),
					resource.TestMatchResourceAttr(recordName, "record", regexp.MustCompile("^(_globalsign-domain-verification=|fastly-domain-delegation-|google-site-verification=|keybase-site-verification=|v=)")),
					resource.TestCheckResourceAttr(recordName, "records.#", "10"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "_globalsign-domain-verification=b1pmSjP4FyG8hkZunkD3Aoz8tK0FWCje80-YwtLeDU"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "_globalsign-domain-verification=O81xyb7YxpdGeHWkniit_VBT4vTXz9__NFrNMoTwFg"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "_globalsign-domain-verification=PO95qR5HARP3zbvcQ1WFBVyLGBUvgjmgc16ENjCtmy"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "fastly-domain-delegation-Yo7fwKrQkFhSCURj-430751-2021-09-07"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "google-site-verification=8d7FpfB8aOEYAIkoaVKxg7Ibj438CEypjZTH424Pews"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "google-site-verification=9D7erI6Bfd9EOHKSIXRe0XQaqAFAjToBtZmyYRzMm34"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "google-site-verification=LQZvxDzrGE-ZLudDpkpj-gcXN-5yF7Z6C-4Rljs3I_Q"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "google-site-verification=y974ACvos30pN7_OBgEZb_byZV8qYtK0G6WZfE7OX8s"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "keybase-site-verification=5HKqMvJnTWpe8W-Aa8r0y3wuy1bhQ6LwcjaxKE9BOQU"),
					resource.TestCheckTypeSetElemAttr(recordName, "records.*", "v=spf1 -all"),
				),
			},
		},
	})
}
