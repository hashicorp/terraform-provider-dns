package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataDnsMXRecordSet_Basic(t *testing.T) {
	recordName := "data.dns_mx_record_set.test"

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "dns_mx_record_set" "test" {
  domain = "google.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(recordName, "id", "google.com"),
					resource.TestCheckResourceAttr(recordName, "mx.#", "5"),
					resource.TestCheckTypeSetElemNestedAttrs(recordName, "mx.*", map[string]string{
						"exchange":   "aspmx.l.google.com.",
						"preference": "10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(recordName, "mx.*", map[string]string{
						"exchange":   "alt1.aspmx.l.google.com.",
						"preference": "20",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(recordName, "mx.*", map[string]string{
						"exchange":   "alt2.aspmx.l.google.com.",
						"preference": "30",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(recordName, "mx.*", map[string]string{
						"exchange":   "alt3.aspmx.l.google.com.",
						"preference": "40",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(recordName, "mx.*", map[string]string{
						"exchange":   "alt4.aspmx.l.google.com.",
						"preference": "50",
					}),
				),
			},
		},
	})
}
