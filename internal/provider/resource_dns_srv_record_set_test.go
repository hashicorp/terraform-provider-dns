package provider

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/miekg/dns"

	"github.com/hashicorp/terraform-provider-dns/internal/hashcode"
)

func TestAccDnsSRVRecordSet_Basic(t *testing.T) {

	var name, zone string
	resourceName := "dns_srv_record_set.foo"

	deleteSRVRecordSet := func() {
		meta, err := initializeDNSClient(context.Background())
		if err != nil {
			t.Fatalf("Error creating DNS Client: %s", err.Error())
		}

		msg := new(dns.Msg)

		msg.SetUpdate(zone)

		fqdn := testResourceFQDN(name, zone)

		rrStr := fmt.Sprintf("%s 0 SRV", fqdn)

		rr_remove, err := dns.NewRR(rrStr)
		if err != nil {
			t.Fatalf("Error reading DNS record (%s): %s", rrStr, err)
		}

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

		meta := testAccProvider.Meta()

		msg := new(dns.Msg)
		msg.SetQuestion(fqdn, dns.TypeSRV)
		r, err := exchange(msg, false, meta)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error querying DNS record")
		}

		existing := schema.NewSet(resourceDnsSRVRecordSetHash, nil)
		expected := schema.NewSet(resourceDnsSRVRecordSetHash, srv)
		for _, record := range r.Answer {
			switch r := record.(type) {
			case *dns.SRV:
				s := map[string]interface{}{
					"priority": int(r.Priority),
					"weight":   int(r.Weight),
					"port":     int(r.Port),
					"target":   r.Target,
				}
				existing.Add(s)
			default:
				return fmt.Errorf("didn't get an SRV record")
			}
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

func resourceDnsSRVRecordSetHash(v interface{}) int {
	var buf bytes.Buffer
	//nolint:forcetypeassert
	m := v.(map[string]interface{})
	//nolint:forcetypeassert
	buf.WriteString(fmt.Sprintf("%d-", m["priority"].(int)))
	//nolint:forcetypeassert
	buf.WriteString(fmt.Sprintf("%d-", m["weight"].(int)))
	//nolint:forcetypeassert
	buf.WriteString(fmt.Sprintf("%d-", m["port"].(int)))
	//nolint:forcetypeassert
	buf.WriteString(fmt.Sprintf("%s-", m["target"].(string)))

	return hashcode.String(buf.String())
}
