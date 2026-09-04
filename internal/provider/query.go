// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
)

func queryDNS(client *DNSClient, fqdn string, rrType uint16) (*dns.Msg, error) {
	if client == nil || len(client.queryNameservers) == 0 {
		return nil, nil
	}

	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, rrType)

	c := new(dns.Client)
	c.Net = client.queryTransport
	c.Timeout = client.queryTimeout
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}

	var lastErr error
	for _, ns := range client.queryNameservers {
		r, _, err := c.Exchange(msg, ns)
		if err != nil {
			lastErr = err
			continue
		}
		if r != nil {
			return r, nil
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("error querying DNS records from custom nameservers: %s", lastErr)
	}

	return nil, fmt.Errorf("no response from any custom nameserver")
}
