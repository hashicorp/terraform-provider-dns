// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/miekg/dns"
)

func TestExtractCNAME_SingleRecord(t *testing.T) {
	fqdn := "example.com."
	answers := []dns.RR{
		&dns.CNAME{
			Hdr:    dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeCNAME, Class: dns.ClassINET},
			Target: "target.example.com.",
		},
	}

	got, err := extractCNAME(answers, fqdn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "target.example.com." {
		t.Errorf("got %q, want %q", got, "target.example.com.")
	}
}

func TestExtractCNAME_Chain(t *testing.T) {
	fqdn := "a.example.com."
	// Simulate a CNAME chain: a -> b -> c
	// A recursive resolver may return all chain entries in the answer section.
	answers := []dns.RR{
		&dns.CNAME{
			Hdr:    dns.RR_Header{Name: "a.example.com.", Rrtype: dns.TypeCNAME, Class: dns.ClassINET},
			Target: "b.example.com.",
		},
		&dns.CNAME{
			Hdr:    dns.RR_Header{Name: "b.example.com.", Rrtype: dns.TypeCNAME, Class: dns.ClassINET},
			Target: "c.example.com.",
		},
		&dns.CNAME{
			Hdr:    dns.RR_Header{Name: "c.example.com.", Rrtype: dns.TypeCNAME, Class: dns.ClassINET},
			Target: "d.example.com.",
		},
	}

	got, err := extractCNAME(answers, fqdn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return only the direct target for the queried FQDN, not the final chain target.
	if got != "b.example.com." {
		t.Errorf("got %q, want %q", got, "b.example.com.")
	}
}

func TestExtractCNAME_NoMatch(t *testing.T) {
	fqdn := "notfound.example.com."
	answers := []dns.RR{
		&dns.CNAME{
			Hdr:    dns.RR_Header{Name: "other.example.com.", Rrtype: dns.TypeCNAME, Class: dns.ClassINET},
			Target: "target.example.com.",
		},
	}

	_, err := extractCNAME(answers, fqdn)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractCNAME_EmptyAnswers(t *testing.T) {
	fqdn := "example.com."
	answers := []dns.RR{}

	_, err := extractCNAME(answers, fqdn)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
