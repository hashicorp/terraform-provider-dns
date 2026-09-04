// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	"github.com/miekg/dns"
)

func axfrQuery(client *DNSClient, zone string) ([]dns.RR, error) {
	m := new(dns.Msg)
	m.SetAxfr(zone)

	t := new(dns.Transfer)
	t.TsigProvider = client.c.TsigProvider
	t.DialTimeout = client.c.Timeout

	if len(client.srv_addrs) == 0 {
		return nil, fmt.Errorf("no DNS servers configured")
	}
	ch, err := t.In(m, client.srv_addrs[0])
	if err != nil {
		return nil, fmt.Errorf("error performing AXFR for zone %s: %s", zone, err)
	}

	var allRecords []dns.RR
	for env := range ch {
		if env.Error != nil {
			return nil, fmt.Errorf("error during AXFR for zone %s: %s", zone, env.Error)
		}
		allRecords = append(allRecords, env.RR...)
	}
	return allRecords, nil
}
