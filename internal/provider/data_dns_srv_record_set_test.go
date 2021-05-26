package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsSRVRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_srv_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_srv_record_set" "test" {
  service = "_http._tcp.mxtoolbox.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "_http._tcp.mxtoolbox.com"),
					resource.TestCheckResourceAttr(recordName, "srv.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(recordName, "srv.*", map[string]string{
						"port":     "80",
						"priority": "10",
						"target":   "mxtoolbox.com.",
						"weight":   "100",
					}),
				),
			},
		},
	})
}
