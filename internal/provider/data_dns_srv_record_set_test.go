package provider

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataDnsSRVRecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		DataSourceName  string
		Expected        []map[string]string
		Service         string
	}{
		{
			`
			data "dns_srv_record_set" "srv" {
			  service = "_http._tcp.mxtoolbox.com"
			}
			`,
			"srv",
			[]map[string]string{
				{"priority": "10", "weight": "100", "port": "80", "target": "mxtoolbox.com."},
			},
			"_http._tcp.mxtoolbox.com",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_srv_record_set.%s", test.DataSourceName)

		resource.UnitTest(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testAccDataDnsSRVExpected(recordName, "srv", test.Expected),
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

func testAccDataDnsSRVExpected(name, key string, value []map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s", name)
		}

		attrKey := fmt.Sprintf("%s.#", key)
		count, ok := is.Attributes[attrKey]
		if !ok {
			return fmt.Errorf("Attributes not found for %s", attrKey)
		}

		gotCount, _ := strconv.Atoi(count)
		if gotCount != len(value) {
			return fmt.Errorf("Mismatch array count for %s: got %s, wanted %d", key, count, len(value))
		}

		for i := 0; i < gotCount; i++ {
			srv := make(map[string]string)

			for _, attr := range []string{"port", "priority", "target", "weight"} {
				attrKey = fmt.Sprintf("%s.%d.%s", key, i, attr)
				got, ok := is.Attributes[attrKey]
				if !ok {
					return fmt.Errorf("Missing attribute for %s", attrKey)
				}
				srv[attr] = got
			}

			if !reflect.DeepEqual(srv, value[i]) {
				return fmt.Errorf("Expected %v, got %v", value[i], srv)
			}
		}

		return nil
	}
}
