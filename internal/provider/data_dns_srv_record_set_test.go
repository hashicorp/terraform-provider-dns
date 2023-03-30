package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataDnsSRVRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_srv_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_srv_record_set" "test" {
  service = "_sip._tls.hashicorptest.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "_sip._tls.hashicorptest.com"),
					resource.TestCheckResourceAttr(recordName, "srv.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(recordName, "srv.*", map[string]string{
						"port":     "443",
						"priority": "0",
						"target":   "example.com.",
						"weight":   "0",
					}),
				),
			},
		},
	})
}
