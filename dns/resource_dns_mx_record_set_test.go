package dns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsMXRecordSet_Basic(t *testing.T) {

	var name, zone string
	resourceName := "dns_mx_record_set.foo"
	resourceRoot := "dns_mx_record_set.root"

	deleteMXRecordSet := func() {
		meta := testAccProvider.Meta()

		msg := new(dns.Msg)

		msg.SetUpdate(zone)

		fqdn := testResourceFQDN(name, zone)

		rr_remove, _ := dns.NewRR(fmt.Sprintf("%s 0 MX", fqdn))
		msg.RemoveRRset([]dns.RR{rr_remove})

		r, err := exchange(msg, true, meta)
		if err != nil {
			t.Fatalf("Error deleting DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			t.Fatalf("Error deleting DNS record: %v", r.Rcode)
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDnsMXRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsMXRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "1"),
					testAccCheckDnsMXRecordSetExists(t, resourceName, []interface{}{map[string]interface{}{"preference": 10, "exchange": "smtp.example.org."}}, &name, &zone),
				),
			},
			{
				Config: testAccDnsMXRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "2"),
					testAccCheckDnsMXRecordSetExists(t, resourceName, []interface{}{map[string]interface{}{"preference": 10, "exchange": "smtp.example.org."}, map[string]interface{}{"preference": 20, "exchange": "backup.example.org."}}, &name, &zone),
				),
			},
			{
				PreConfig: deleteMXRecordSet,
				Config:    testAccDnsMXRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "2"),
					testAccCheckDnsMXRecordSetExists(t, resourceName, []interface{}{map[string]interface{}{"preference": 10, "exchange": "smtp.example.org."}, map[string]interface{}{"preference": 20, "exchange": "backup.example.org."}}, &name, &zone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsMXRecordSet_root,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "mx.#", "1"),
					testAccCheckDnsMXRecordSetExists(t, resourceRoot, []interface{}{map[string]interface{}{"preference": 10, "exchange": "smtp.example.org."}}, &name, &zone),
				),
			},
			{
				ResourceName:      resourceRoot,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDnsMXRecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_mx_record_set", dns.TypeMX)
}

func testAccCheckDnsMXRecordSetExists(t *testing.T, n string, mx []interface{}, name, zone *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		*name = rs.Primary.Attributes["name"]
		*zone = rs.Primary.Attributes["zone"]

		fqdn := testResourceFQDN(*name, *zone)

		meta := testAccProvider.Meta()

		msg := new(dns.Msg)
		msg.SetQuestion(fqdn, dns.TypeMX)
		r, err := exchange(msg, false, meta)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error querying DNS record")
		}

		existing := schema.NewSet(resourceDnsMXRecordSetHash, nil)
		expected := schema.NewSet(resourceDnsMXRecordSetHash, mx)
		for _, record := range r.Answer {
			switch r := record.(type) {
			case *dns.MX:
				m := map[string]interface{}{
					"preference": int(r.Preference),
					"exchange":   r.Mx,
				}
				existing.Add(m)
			default:
				return fmt.Errorf("didn't get an MX record")
			}
		}
		if !existing.Equal(expected) {
			return fmt.Errorf("DNS record differs: expected %v, found %v", expected, existing)
		}
		return nil
	}
}

var testAccDnsMXRecordSet_basic = fmt.Sprintf(`
  resource "dns_mx_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    mx {
      preference = 10
      exchange = "smtp.example.org."
    }
    ttl = 300
  }`)

var testAccDnsMXRecordSet_update = fmt.Sprintf(`
  resource "dns_mx_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    mx {
      preference = 10
      exchange = "smtp.example.org."
    }
    mx {
      preference = 20
      exchange = "backup.example.org."
    }
    ttl = 300
  }`)

var testAccDnsMXRecordSet_root = fmt.Sprintf(`
  resource "dns_mx_record_set" "root" {
    zone = "example.com."
    mx {
      preference = 10
      exchange = "smtp.example.org."
    }
    ttl = 300
  }`)
