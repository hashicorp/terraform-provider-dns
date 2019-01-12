package dns

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/miekg/dns"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"dns": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	v := os.Getenv("DNS_UPDATE_SERVER")
	if v == "" {
		t.Fatal("DNS_UPDATE_SERVER must be set for acceptance tests")
	}
}

func testResourceFQDN(name, zone string) string {
	fqdn := zone
	if name != "" {
		fqdn = fmt.Sprintf("%s.%s", name, fqdn)
	}

	return fqdn
}

func testAccCheckDnsDestroy(s *terraform.State, resourceType string, rrType uint16) error {
	meta := testAccProvider.Meta()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceType {
			continue
		}

		fqdn := testResourceFQDN(rs.Primary.Attributes["name"], rs.Primary.Attributes["zone"])

		msg := new(dns.Msg)
		msg.SetQuestion(fqdn, rrType)
		r, err := exchange(msg, false, meta)
		if err != nil {
			return fmt.Errorf("Error querying DNS record: %s", err)
		}
		// Should either get an NXDOMAIN or a NOERROR and no answers
		// (and usually an authority record)
		if !(r.Rcode == dns.RcodeNameError || (r.Rcode == dns.RcodeSuccess && ((rrType == dns.TypeNS && len(r.Ns) == 0) || len(r.Answer) == 0))) {
			return fmt.Errorf("DNS record still exists: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode])
		}
	}

	return nil
}
