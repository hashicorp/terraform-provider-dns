package dns

import (
	"fmt"
	"strconv"

	r "github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testCheckAttrIntArray(name, key string, value []int) r.TestCheckFunc {
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

	Next:
		for i := 0; i < gotCount; i++ {
			attrKey = fmt.Sprintf("%s.%d", key, i)
			got, ok := is.Attributes[attrKey]
			if !ok {
				return fmt.Errorf("Missing array item for %s", attrKey)
			}
			intGot, _ := strconv.Atoi(got)
			//for _, want := range value {
			//	fmt.Println(intGot, want)
			//	if intGot == want {
			//		continue Next
			//	}
			//}
			if intGot == value[i] {
				continue Next
			}
			return fmt.Errorf(
				"Unexpected array item for %s: expected %d, got %s",
				attrKey,
				value[i],
				got)
		}

		return nil
	}
}
