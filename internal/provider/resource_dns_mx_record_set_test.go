package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsMXRecordSet_Basic(t *testing.T) {

	var name, zone string
	resourceName := "dns_mx_record_set.foo"
	resourceRoot := "dns_mx_record_set.root"

	deleteMXRecordSet := func() {
		msg := new(dns.Msg)

		msg.SetUpdate(zone)

		fqdn := testResourceFQDN(name, zone)

		rrStr := fmt.Sprintf("%s 0 MX", fqdn)

		rr_remove, err := dns.NewRR(rrStr)
		if err != nil {
			t.Fatalf("Error reading DNS record (%s): %s", rrStr, err)
		}

		msg.RemoveRRset([]dns.RR{rr_remove})

		r, err := exchange(msg, true, dnsClient)
		if err != nil {
			t.Fatalf("Error deleting DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			t.Fatalf("Error deleting DNS record: %v", r.Rcode)
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDnsMXRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsMXRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "1"),
					testAccCheckDnsMXRecordSetExists(resourceName, []interface{}{map[string]interface{}{"preference": 10, "exchange": "smtp.example.org."}}, &name, &zone),
				),
			},
			{
				Config: testAccDnsMXRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "2"),
					testAccCheckDnsMXRecordSetExists(resourceName, []interface{}{map[string]interface{}{"preference": 10, "exchange": "smtp.example.org."}, map[string]interface{}{"preference": 20, "exchange": "backup.example.org."}}, &name, &zone),
				),
			},
			{
				PreConfig: deleteMXRecordSet,
				Config:    testAccDnsMXRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mx.#", "2"),
					testAccCheckDnsMXRecordSetExists(resourceName, []interface{}{map[string]interface{}{"preference": 10, "exchange": "smtp.example.org."}, map[string]interface{}{"preference": 20, "exchange": "backup.example.org."}}, &name, &zone),
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
					testAccCheckDnsMXRecordSetExists(resourceRoot, []interface{}{map[string]interface{}{"preference": 10, "exchange": "smtp.example.org."}}, &name, &zone),
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

func testAccCheckDnsMXRecordSetExists(n string, mx []interface{}, name, zone *string) resource.TestCheckFunc {
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

		msg := new(dns.Msg)
		msg.SetQuestion(fqdn, dns.TypeMX)
		r, err := exchange(msg, false, dnsClient)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error querying DNS record")
		}

		var answers []mxBlockConfig
		for _, record := range r.Answer {
			switch r := record.(type) {
			case *dns.MX:
				m := mxBlockConfig{
					Preference: types.Int64Value(int64(r.Preference)),
					Exchange:   types.StringValue(r.Mx),
				}
				answers = append(answers, m)
			default:
				return fmt.Errorf("didn't get an MX record")
			}
		}

		existing, diags := types.SetValueFrom(context.Background(), types.StringType, answers)
		if diags.HasError() {
			fmt.Errorf("couldn't create set from answers")
		}
		expected, diags := types.SetValueFrom(context.Background(), types.StringType, mx)
		if diags.HasError() {
			fmt.Errorf("couldn't create set from given mx param")
		}

		if !existing.Equal(expected) {
			return fmt.Errorf("DNS record differs: expected %v, found %v", expected, existing)
		}
		return nil
	}
}

var testAccDnsMXRecordSet_basic = `
  resource "dns_mx_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    mx {
      preference = 10
      exchange = "smtp.example.org."
    }
    ttl = 300
  }`

var testAccDnsMXRecordSet_update = `
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
  }`

var testAccDnsMXRecordSet_root = `
  resource "dns_mx_record_set" "root" {
    zone = "example.com."
    mx {
      preference = 10
      exchange = "smtp.example.org."
    }
    ttl = 300
  }`
