package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataDnsSRVRecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		DataSourceName  string
		ExpectedTarget  string
		Service         string
	}{
		{
			`
			data "dns_srv_record_set" "srv" {
			  service = "_http._tcp.mxtoolbox.com"
			}
			`,
			"srv",
			"mxtoolbox.com.",
			"_http._tcp.mxtoolbox.com",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_srv_record_set.%s", test.DataSourceName)

		resource.Test(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(recordName, "srv.0.target", test.ExpectedTarget),
					),
				},
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(recordName, "id", test.Service),
					),
				},
			},
		})
	}

}
