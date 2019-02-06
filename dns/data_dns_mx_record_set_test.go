package dns

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataDnsMXRecordSet_Basic(t *testing.T) {
	tests := []struct {
		DataSourceBlock string
		DataSourceName  string
		Expected        []map[string]string
		Domain          string
	}{
		{
			`
			data "dns_mx_record_set" "mx" {
			  domain = "google.com"
			}
			`,
			"mx",
			[]map[string]string{
				{"preference": "10", "exchange": "aspmx.l.google.com."},
				{"preference": "20", "exchange": "alt1.aspmx.l.google.com."},
				{"preference": "30", "exchange": "alt2.aspmx.l.google.com."},
				{"preference": "40", "exchange": "alt3.aspmx.l.google.com."},
				{"preference": "50", "exchange": "alt4.aspmx.l.google.com."},
			},
			"google.com",
		},
		{
			`
			data "dns_mx_record_set" "non-existent" {
			  domain        = "jolly.roger"
			  ignore_errors = true
			}
			`,
			"non-existent",
			[]map[string]string{},
			"jolly.roger",
		},
	}

	for _, test := range tests {
		recordName := fmt.Sprintf("data.dns_mx_record_set.%s", test.DataSourceName)

		resource.Test(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						testAccDataDnsMXExpected(recordName, "mx", test.Expected),
					),
				},
				{
					Config: test.DataSourceBlock,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(recordName, "id", test.Domain),
					),
				},
			},
		})
	}
}

func testAccDataDnsMXExpected(name, key string, value []map[string]string) resource.TestCheckFunc {
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
			mx := make(map[string]string)

			for _, attr := range []string{"exchange", "preference"} {
				attrKey = fmt.Sprintf("%s.%d.%s", key, i, attr)
				got, ok := is.Attributes[attrKey]
				if !ok {
					return fmt.Errorf("Missing attribute for %s", attrKey)
				}
				mx[attr] = got
			}

			if !reflect.DeepEqual(mx, value[i]) {
				return fmt.Errorf("Expected %v, got %v", value[i], mx)
			}
		}

		return nil
	}
}
