/*
Package gss implements RFC 3645 GSS-TSIG functions. This permits sending
signed dynamic DNS update messages to Windows servers that have the zone
require "Secure only" updates.

Example client:

        import (
                "fmt"
                "time"

                "github.com/bodgit/tsig"
                "github.com/bodgit/tsig/gss"
                "github.com/miekg/dns"
        )

        func main() {
                dnsClient := new(dns.Client)
                dnsClient.Net = "tcp"

                gssClient, err := gss.NewClient(dnsClient)
                if err != nil {
                        panic(err)
                }
                defer gssClient.Close()

                host := "ns.example.com:53"

                // Negotiate a context with the chosen server using the
                // current user. See also
                // gssClient.NegotiateContextWithCredentials() and
                // gssClient.NegotiateContextWithKeytab() for alternatives
                keyname, _, err := gssClient.NegotiateContext(host)
                if err != nil {
                        panic(err)
                }

                dnsClient.TsigProvider = gssClient

                // Use the DNS client as normal

                msg := new(dns.Msg)
                msg.SetUpdate(dns.Fqdn("example.com"))

                insert, err := dns.NewRR("test.example.com. 300 A 192.0.2.1")
                if err != nil {
                        panic(err)
                }
                msg.Insert([]dns.RR{insert})

                msg.SetTsig(keyname, tsig.GSS, 300, time.Now().Unix())

                rr, _, err := dnsClient.Exchange(msg, host)
                if err != nil {
                        panic(err)
                }

                if rr.Rcode != dns.RcodeSuccess {
                        fmt.Printf("DNS error: %s (%d)\n", dns.RcodeToString[rr.Rcode], rr.Rcode)
                }

                // Cleanup the context
                err = gssClient.DeleteContext(keyname)
                if err != nil {
                        panic(err)
                }
        }

Under the hood, GSSAPI is used on platforms other than Windows whilst Windows
uses native SSPI which has a similar API.
*/
package gss

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/bodgit/tsig"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/miekg/dns"
)

var (
	errNotSupported = errors.New("not supported")
)

// gssNoVerify is a dns.TsigProvider that skips any GSS-TSIG verification.
//
// BIND doesn't sign TKEY responses but Windows does, using the key you're
// currently negotiating so it creates a chicken & egg problem. According
// to the RFC, verification isn't needed as the TKEY response should be
// cryptographically secure anyway.
type gssNoVerify struct{}

func (*gssNoVerify) Generate(_ []byte, t *dns.TSIG) ([]byte, error) {
	if dns.CanonicalName(t.Algorithm) != tsig.GSS {
		return nil, dns.ErrKeyAlg
	}
	return nil, dns.ErrSecret
}

func (*gssNoVerify) Verify(_ []byte, t *dns.TSIG) error {
	if dns.CanonicalName(t.Algorithm) != tsig.GSS {
		return dns.ErrKeyAlg
	}
	return nil
}

func generateTKEYName(host string) string {

	seed := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(seed)

	return dns.Fqdn(fmt.Sprintf("%d.sig-%s", rng.Int31(), host))
}

func generateSPN(host string) string {

	if dns.IsFqdn(host) {
		return fmt.Sprintf("DNS/%s", host[:len(host)-1])
	}

	return fmt.Sprintf("DNS/%s", host)
}

func (c *Client) close() error {

	c.m.RLock()
	keys := make([]string, 0, len(c.ctx))
	for k := range c.ctx {
		keys = append(keys, k)
	}
	c.m.RUnlock()

	var errs error
	for _, k := range keys {
		errs = multierror.Append(errs, c.DeleteContext(k))
	}

	return errs
}
