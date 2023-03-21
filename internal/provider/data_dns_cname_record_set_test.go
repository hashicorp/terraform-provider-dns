package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsCnameRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_cname_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_cname_record_set" "test" {
  host = "terraform-provider-dns-cname.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "cname", "example.com."),
					resource.TestCheckResourceAttr(recordName, "id", "terraform-provider-dns-cname.hashicorptest.com"),
				),
			},
		},
	})
}
