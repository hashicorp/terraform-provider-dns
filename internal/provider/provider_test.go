// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/miekg/dns"
)

var dnsClient *DNSClient
var testProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"dns": func() (tfprotov5.ProviderServer, error) {
		providers := []func() tfprotov5.ProviderServer{
			providerserver.NewProtocol5(NewFrameworkProvider()),
			New().GRPCProvider,
		}
		return tf5muxserver.NewMuxServer(context.Background(), providers...)
	},
}

func providerVersion324() map[string]resource.ExternalProvider {
	return map[string]resource.ExternalProvider{
		"dns": {
			VersionConstraint: "3.2.4",
			Source:            "hashicorp/dns",
		},
	}
}

func TestMain(m *testing.M) {
	var clientErr error
	dnsClient, clientErr = initializeDNSClient(context.Background())
	if clientErr != nil {
		os.Exit(1)
	}

	m.Run()
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
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceType {
			continue
		}

		fqdn := testResourceFQDN(rs.Primary.Attributes["name"], rs.Primary.Attributes["zone"])

		msg := new(dns.Msg)
		msg.SetQuestion(fqdn, rrType)
		r, err := exchange(msg, false, dnsClient)
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

func TestAccProvider_Update_Gssapi_Realm(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				provider "dns" {
					update {
						gssapi {
							realm = "EXAMPLE.COM"
						}

						server = "127.0.0.1"
					}
				}

				data "dns_a_record_set" "test" {
					# Same host as data source testing
					host = "terraform-provider-dns-a.hashicorptest.com"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dns_a_record_set.test", "addrs.#", "1"),
				),
			},
		},
	})
}

func TestAccProvider_Update_Server_Config(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				provider "dns" {
					update {
						server = "127.0.0.1"
					}
				}

				data "dns_a_record_set" "test" {
					# Same host as data source testing
					host = "terraform-provider-dns-a.hashicorptest.com"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dns_a_record_set.test", "addrs.#", "1"),
				),
			},
		},
	})
}

func TestAccProvider_Update_Server_Env(t *testing.T) {
	t.Setenv("DNS_UPDATE_SERVER", "example.com")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "dns_a_record_set" "test" {
					# Same host as data source testing
					host = "terraform-provider-dns-a.hashicorptest.com"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dns_a_record_set.test", "addrs.#", "1"),
				),
			},
		},
	})
}

func TestAccProvider_Update_Timeout_Config(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				provider "dns" {
					update {
						timeout = 5
					}
				}

				data "dns_a_record_set" "test" {
					# Same host as data source testing
					host = "terraform-provider-dns-a.hashicorptest.com"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dns_a_record_set.test", "addrs.#", "1"),
				),
			},
		},
	})
}

func TestAccProvider_Update_Timeout_Env(t *testing.T) {
	t.Setenv("DNS_UPDATE_TIMEOUT", "5")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "dns_a_record_set" "test" {
					# Same host as data source testing
					host = "terraform-provider-dns-a.hashicorptest.com"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dns_a_record_set.test", "addrs.#", "1"),
				),
			},
		},
	})
}

func TestAccProvider_InvalidClientConfig(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				provider "dns" {
					update {
						gssapi {
							realm = ""
						}

						server = "127.0.0.1"
					}
				}

				 resource "dns_ptr_record" "foo" {
					zone = "example.com."
					name = "r._dns-sd._udp"
					ptr = "bar.example.com."
					ttl = 300
  				}
				`,
				ExpectError: regexp.MustCompile(`.*Error configuring provider:`),
			},
		},
	})
}

func TestAccProvider_Validators(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				//multiple update blocks
				Config: `provider "dns" {
					update {
						server        = "192.168.0.1"
						key_algorithm = "hmac-md5"
						key_secret    = "3VwZXJzZWNyZXQ="
					}
					update {
						server        = "192.168.0.1"
						key_algorithm = "hmac-md5"
						key_secret    = "3VwZXJzZWNyZXQ="
					}
				}

				resource "dns_a_record_set" "foo" {}`,
				ExpectError: regexp.MustCompile(`.*Error: Invalid Attribute Value`),
			},
			{
				//multiple gssapi blocks
				Config: `provider "dns" {
					update {
						server        = "192.168.0.1"
						gssapi {
							realm    = "EXAMPLE.COM"
							username = "user"
							keytab   = "/path/to/keytab"
						}
						gssapi {
							realm    = "EXAMPLE.COM"
							username = "user"
							keytab   = "/path/to/keytab"
						}
					}
				}

				resource "dns_a_record_set" "foo" {}`,
				ExpectError: regexp.MustCompile(`.*Error: Invalid Attribute Value`),
			},
			{
				//missing key_name (required with key_algorithm and key_secret)
				Config: `provider "dns" {
					update {
						server        = "192.168.0.1"
						key_algorithm = "hmac-md5"
						key_secret    = "3VwZXJzZWNyZXQ="
					}
				}

				resource "dns_a_record_set" "foo" {}`,
				ExpectError: regexp.MustCompile(`.*Error: Invalid Attribute Combination`),
			},
			{
				//update block key_ arguments set with gssapi block
				Config: `provider "dns" {
					update {
						server        = "192.168.0.1"
						key_name      = "example.com."
						key_algorithm = "hmac-md5"
						key_secret    = "3VwZXJzZWNyZXQ="
						
						gssapi {
							realm    = "EXAMPLE.COM"
							username = "user"
							keytab   = "/path/to/keytab"
						}
					}
				}

				resource "dns_a_record_set" "foo" {}`,
				ExpectError: regexp.MustCompile(`.*Error: Invalid Attribute Combination`),
			},
			{
				//username not set (required with keytab)
				Config: `provider "dns" {
					update {
						server        = "192.168.0.1"
						gssapi {
							realm    = "EXAMPLE.COM"
							keytab   = "/path/to/keytab"
						}
					}
				}

				resource "dns_a_record_set" "foo" {}`,
				ExpectError: regexp.MustCompile(`.*Error: Invalid Attribute Combination`),
			},
			{
				//both password and keytab set
				Config: `provider "dns" {
					update {
						server        = "192.168.0.1"
						gssapi {
							realm    = "EXAMPLE.COM"
							username = "user"
							password = "password"
							keytab   = "/path/to/keytab"
						}
					}
				}

				resource "dns_a_record_set" "foo" {}`,
				ExpectError: regexp.MustCompile(`.*Error: Invalid Attribute Combination`),
			},
		},
	})
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

	client, configErr := config.Client(ctx)
	if configErr != nil {
		return &DNSClient{}, fmt.Errorf("error creating DNS Client")
	}

	dnsClient, ok := client.(*DNSClient)

	if !ok {
		return &DNSClient{}, fmt.Errorf("error converting client to DNSClient type")
	}

	return dnsClient, nil
}

func testRemoveRecord(t *testing.T, recordType string, recordName string) {
	msg := new(dns.Msg)
	recordZone := "example.com."

	msg.SetUpdate(recordZone)

	rrStr := fmt.Sprintf("%s.%s 0 %s", recordName, recordZone, recordType)

	rr, err := dns.NewRR(rrStr)

	if err != nil {
		t.Fatalf("Error generating DNS record (%s): %s", rrStr, err)
	}

	msg.RemoveRRset([]dns.RR{rr})

	resp, err := exchange(msg, true, dnsClient)

	if err != nil {
		t.Fatalf("Error deleting DNS record (%s): %s", rrStr, err)
	}

	if resp.Rcode != dns.RcodeSuccess {
		t.Fatalf("Error deleting DNS record (%s): %v", rrStr, resp.Rcode)
	}
}
