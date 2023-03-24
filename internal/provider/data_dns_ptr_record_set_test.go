package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsPtrRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_ptr_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_ptr_record_set" "test" {
  ip_address = "8.8.8.8"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "8.8.8.8"),
					resource.TestCheckResourceAttr(recordName, "ptr", "dns.google."),
				),
			},
		},
	})
}
