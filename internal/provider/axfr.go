// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	"github.com/miekg/dns"
)

// axfrQuery performs a zone transfer (AXFR) and returns all records
func axfrQuery(client *DNSClient, zone string) ([]dns.RR, error) {
	m := new(dns.Msg)
	m.SetAxfr(zone)

	t := new(dns.Transfer)
	t.TsigProvider = client.c.TsigProvider
	t.DialTimeout = client.c.Timeout

	ch, err := t.In(m, client.srv_addr)
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
