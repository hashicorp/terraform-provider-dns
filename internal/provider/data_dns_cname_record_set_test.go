package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsCnameRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_cname_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_cname_record_set" "test" {
  host = "www.hashicorp.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "cname", "cname.vercel-dns.com."),
					resource.TestCheckResourceAttr(recordName, "id", "www.hashicorp.com"),
				),
			},
		},
	})
}
