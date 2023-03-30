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

func TestAccDnsSRVRecordSet_Basic(t *testing.T) {

	var name, zone string
	resourceName := "dns_srv_record_set.foo"

	deleteSRVRecordSet := func() {
		msg := new(dns.Msg)

		msg.SetUpdate(zone)

		fqdn := testResourceFQDN(name, zone)

		rrStr := fmt.Sprintf("%s 0 SRV", fqdn)

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
		CheckDestroy:             testAccCheckDnsSRVRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsSRVRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srv.#", "1"),
					testAccCheckDnsSRVRecordSetExists(resourceName, []interface{}{map[string]interface{}{"priority": 10, "weight": 60, "port": 5060, "target": "bigbox.example.com."}}, &name, &zone),
				),
			},
			{
				Config: testAccDnsSRVRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srv.#", "2"),
					testAccCheckDnsSRVRecordSetExists(resourceName, []interface{}{map[string]interface{}{"priority": 10, "weight": 60, "port": 5060, "target": "bigbox.example.com."}, map[string]interface{}{"priority": 20, "weight": 0, "port": 5060, "target": "backupbox.example.com."}}, &name, &zone),
				),
			},
			{
				PreConfig: deleteSRVRecordSet,
				Config:    testAccDnsSRVRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srv.#", "2"),
					testAccCheckDnsSRVRecordSetExists(resourceName, []interface{}{map[string]interface{}{"priority": 10, "weight": 60, "port": 5060, "target": "bigbox.example.com."}, map[string]interface{}{"priority": 20, "weight": 0, "port": 5060, "target": "backupbox.example.com."}}, &name, &zone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDnsSRVRecordSetDestroy(s *terraform.State) error {
	return testAccCheckDnsDestroy(s, "dns_srv_record_set", dns.TypeSRV)
}

func testAccCheckDnsSRVRecordSetExists(n string, srv []interface{}, name, zone *string) resource.TestCheckFunc {
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
		msg.SetQuestion(fqdn, dns.TypeSRV)
		r, err := exchange(msg, false, dnsClient)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error querying DNS record")
		}

		var answers []srvBlockConfig
		for _, record := range r.Answer {
			switch r := record.(type) {
			case *dns.SRV:
				s := srvBlockConfig{
					Priority: types.Int64Value(int64(r.Priority)),
					Weight:   types.Int64Value(int64(r.Weight)),
					Port:     types.Int64Value(int64(r.Port)),
					Target:   types.StringValue(r.Target),
				}
				answers = append(answers, s)
			default:
				return fmt.Errorf("didn't get an SRV record")
			}
		}

		existing, diags := types.SetValueFrom(context.Background(), types.StringType, answers)
		if diags.HasError() {
			fmt.Errorf("couldn't create set from answers")
		}
		expected, diags := types.SetValueFrom(context.Background(), types.StringType, srv)
		if diags.HasError() {
			fmt.Errorf("couldn't create set from given srv param")
		}

		if !existing.Equal(expected) {
			return fmt.Errorf("DNS record differs: expected %v, found %v", expected, existing)
		}
		return nil
	}
}

var testAccDnsSRVRecordSet_basic = `
  resource "dns_srv_record_set" "foo" {
    zone = "example.com."
    name = "_sip._tcp"
    srv {
      priority = 10
      weight = 60
      port = 5060
      target = "bigbox.example.com."
    }
    ttl = 300
  }`

var testAccDnsSRVRecordSet_update = `
  resource "dns_srv_record_set" "foo" {
    zone = "example.com."
    name = "_sip._tcp"
    srv {
      priority = 10
      weight = 60
      port = 5060
      target = "bigbox.example.com."
    }
    srv {
      priority = 20
      weight = 0
      port = 5060
      target = "backupbox.example.com."
    }
    ttl = 300
  }`
