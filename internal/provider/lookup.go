// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net"

	"github.com/miekg/dns"
)

func getSystemResolver() (string, error) {
	var addr string
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			addr = address
			d := net.Dialer{}
			return d.DialContext(ctx, network, address)
		},
	}
	// Trigger a DNS query to capture the system resolver address via the Dial hook.
	// "." (root zone) is used because /etc/hosts entries (e.g. "localhost") are
	// resolved without calling Dial. The query result itself is irrelevant.
	r.LookupHost(context.Background(), ".") //nolint:errcheck
	if addr == "" {
		return "", fmt.Errorf("could not determine system DNS resolver")
	}
	return addr, nil
}

func extractCNAME(answers []dns.RR, fqdn string) (string, error) {
	for _, ans := range answers {
		if cname, ok := ans.(*dns.CNAME); ok {
			if dns.Fqdn(cname.Hdr.Name) == fqdn {
				return cname.Target, nil
			}
		}
	}
	return "", fmt.Errorf("no CNAME record found for %s", fqdn)
}

func lookupCNAME(fqdn string) (string, error) {
	serverAddr, err := getSystemResolver()
	if err != nil {
		return "", err
	}

	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, dns.TypeCNAME)
	msg.RecursionDesired = true

	client := new(dns.Client)
	r, _, err := client.Exchange(msg, serverAddr)
	if err != nil {
		return "", fmt.Errorf("DNS query failed: %s", err)
	}
	if r.Rcode != dns.RcodeSuccess {
		return "", fmt.Errorf("DNS query returned %s", dns.RcodeToString[r.Rcode])
	}
	return extractCNAME(r.Answer, fqdn)
}

func lookupIP(host string) ([]string, []string, error) {
	records, err := net.LookupIP(host)
	if err != nil {
		return nil, nil, err
	}

	a := make([]string, 0)
	aaaa := make([]string, 0)
	for _, ip := range records {
		if ipv4 := ip.To4(); ipv4 != nil {
			a = append(a, ipv4.String())
		} else {
			aaaa = append(aaaa, ip.String())
		}
	}

	return a, aaaa, nil
}
