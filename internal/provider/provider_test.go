package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/miekg/dns"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

var testProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"dns": providerserver.NewProtocol5WithError(NewFrameworkProvider()),
}

func init() {
	testAccProvider = New()
	testAccProviders = map[string]*schema.Provider{
		"dns": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := New().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = New()
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

func initializeDNSClient(ctx context.Context) (*DNSClient, error) {
	var server, transport, timeout, keyname, keyalgo, keysecret, realm, username, password, keytab string
	var port, retries int
	var duration time.Duration
	var gssapi bool

	if len(os.Getenv("DNS_UPDATE_SERVER")) > 0 {
		server = os.Getenv("DNS_UPDATE_SERVER")
	}

	if len(os.Getenv("DNS_UPDATE_PORT")) > 0 {
		portStr := os.Getenv("DNS_UPDATE_PORT")
		envPort, err := strconv.Atoi(portStr)
		if err != nil {
			return &DNSClient{}, err
		}
		port = envPort
	} else {
		port = defaultPort
	}

	if len(os.Getenv("DNS_UPDATE_TRANSPORT")) > 0 {
		transport = os.Getenv("DNS_UPDATE_TRANSPORT")
	} else {
		transport = defaultTransport
	}

	if len(os.Getenv("DNS_UPDATE_TIMEOUT")) > 0 {
		timeout = os.Getenv("DNS_UPDATE_TIMEOUT")
	} else {
		timeout = defaultTimeout
	}

	// Try parsing as a duration
	var err error
	duration, err = time.ParseDuration(timeout)
	if err != nil {
		// Failing that, convert to an integer and treat as seconds
		var seconds int
		seconds, err = strconv.Atoi(timeout)
		if err != nil {
			return &DNSClient{}, err
		}
		duration = time.Duration(seconds) * time.Second
	}
	if duration < 0 {
		return &DNSClient{}, fmt.Errorf("timeout cannot be negative: %s", err.Error())
	}

	if len(os.Getenv("DNS_UPDATE_RETRIES")) > 0 {
		retriesStr := os.Getenv("DNS_UPDATE_RETRIES")

		var err error
		retries, err = strconv.Atoi(retriesStr)
		if err != nil {
			return &DNSClient{}, fmt.Errorf("invalid DNS_UPDATE_RETRIES environment variable: %s", err.Error())
		}
	} else {
		retries = defaultRetries
	}
	if len(os.Getenv("DNS_UPDATE_KEYNAME")) > 0 {
		keyname = os.Getenv("DNS_UPDATE_KEYNAME")
	}
	if len(os.Getenv("DNS_UPDATE_KEYALGORITHM")) > 0 {
		keyalgo = os.Getenv("DNS_UPDATE_KEYALGORITHM")
	}
	if len(os.Getenv("DNS_UPDATE_KEYSECRET")) > 0 {
		keysecret = os.Getenv("DNS_UPDATE_KEYSECRET")
	}

	if len(os.Getenv("DNS_UPDATE_REALM")) > 0 {
		realm = os.Getenv("DNS_UPDATE_REALM")
	}
	if len(os.Getenv("DNS_UPDATE_USERNAME")) > 0 {
		username = os.Getenv("DNS_UPDATE_USERNAME")
	}
	if len(os.Getenv("DNS_UPDATE_PASSWORD")) > 0 {
		password = os.Getenv("DNS_UPDATE_PASSWORD")
	}
	if len(os.Getenv("DNS_UPDATE_KEYTAB")) > 0 {
		keytab = os.Getenv("DNS_UPDATE_KEYTAB")
	}
	if realm != "" || username != "" || password != "" || keytab != "" {
		gssapi = true
	}

	config := Config{
		server:    server,
		port:      port,
		transport: transport,
		timeout:   duration,
		retries:   retries,
		keyname:   keyname,
		keyalgo:   keyalgo,
		keysecret: keysecret,
		gssapi:    gssapi,
		realm:     realm,
		username:  username,
		password:  password,
		keytab:    keytab,
	}

	var client, diags = config.Client(ctx)
	if diags.HasError() {
		return &DNSClient{}, fmt.Errorf("error creating DNS Client")
	}

	dnsClient, ok := client.(*DNSClient)

	if !ok {
		return &DNSClient{}, fmt.Errorf("error converting client to DNSClient type")
	}

	return dnsClient, nil
}
