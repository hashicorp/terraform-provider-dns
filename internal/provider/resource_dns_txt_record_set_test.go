package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

func TestAccDnsTXTRecordSet_Basic(t *testing.T) {

	var name, zone string
	resourceName := "dns_txt_record_set.foo"
	resourceRoot := "dns_txt_record_set.root"

	deleteTXTRecordSet := func() {
		msg := new(dns.Msg)

		msg.SetUpdate(zone)

		fqdn := testResourceFQDN(name, zone)

		rrStr := fmt.Sprintf("%s 0 TXT", fqdn)

		rr_remove, err := dns.NewRR(rrStr)
		if err != nil {
			t.Fatalf("Error reading DNS record: %s", err)
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
		CheckDestroy:             testAccCheckDnsARecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsTXTRecordSet_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "txt.#", "2"),
					testAccCheckDnsTXTRecordSetExists(resourceName, []interface{}{"foo", "bar"}, &name, &zone),
				),
			},
			{
				Config: testAccDnsTXTRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "txt.#", "3"),
					testAccCheckDnsTXTRecordSetExists(resourceName, []interface{}{"foo", "bar", "baz"}, &name, &zone),
				),
			},
			{
				PreConfig: deleteTXTRecordSet,
				Config:    testAccDnsTXTRecordSet_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "txt.#", "3"),
					testAccCheckDnsTXTRecordSetExists(resourceName, []interface{}{"foo", "bar", "baz"}, &name, &zone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDnsTXTRecordSet_root,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoot, "txt.#", "1"),
					testAccCheckDnsTXTRecordSetExists(resourceRoot, []interface{}{"foo"}, &name, &zone),
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

func testAccCheckDnsTXTRecordSetExists(n string, txt []interface{}, name, zone *string) resource.TestCheckFunc {
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
		msg.SetQuestion(fqdn, dns.TypeTXT)
		r, err := exchange(msg, false, dnsClient)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("Error querying DNS record")
		}

		var answers []string
		for _, record := range r.Answer {
			switch r := record.(type) {
			case *dns.TXT:
				answers = append(answers, strings.Join(r.Txt, ""))
			default:
				return fmt.Errorf("didn't get an TXT record")
			}
		}

		existing, diags := types.SetValueFrom(context.Background(), types.StringType, answers)
		if diags.HasError() {
			return fmt.Errorf("couldn't create set from answers")
		}
		expected, diags := types.SetValueFrom(context.Background(), types.StringType, txt)
		if diags.HasError() {
			return fmt.Errorf("couldn't create set from given txt param")
		}

		if !existing.Equal(expected) {
			return fmt.Errorf("DNS record differs: expected %v, found %v", expected, existing)
		}
		return nil
	}
}

var testAccDnsTXTRecordSet_basic = `
  resource "dns_txt_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    txt = ["foo", "bar"]
    ttl = 300
  }`

var testAccDnsTXTRecordSet_update = `
  resource "dns_txt_record_set" "foo" {
    zone = "example.com."
    name = "foo"
    txt = ["foo", "bar", "baz"]
    ttl = 300
  }`

var testAccDnsTXTRecordSet_root = `
  resource "dns_txt_record_set" "root" {
    zone = "example.com."
    txt = ["foo"]
    ttl = 300
  }`
