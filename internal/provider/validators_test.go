// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/miekg/dns"
)

func TestValidateZone(t *testing.T) {
	validNames := []string{
		"example.com.",
	}
	for _, v := range validNames {
		_, errors := validateZone(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid DNS zone: %q", v, errors)
		}
	}

	invalidNames := []string{
		"example.com",
		" example.com.",
		" ",
		"",
	}
	for _, v := range invalidNames {
		_, errors := validateZone(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid DNS zone", v)
		}
	}
}

func TestValidateName(t *testing.T) {
	validNames := []string{
		"test",
	}
	for _, v := range validNames {
		_, errors := validateName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid DNS record: %q", v, errors)
		}
	}

	invalidNames := []string{
		"test.",
		" test. ",
		" ",
		"",
		"test name",
		"test  name",
	}
	for _, v := range invalidNames {
		_, errors := validateName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid DNS record", v)
		}
	}
}

func TestGetPtrValNoTruncate(t *testing.T) {
	target := "host.example.com."
	ptr := &dns.PTR{
		Ptr: target,
		Hdr: dns.RR_Header{
			Ttl: 3600,
		},
	}
	result, ttl, err := getPtrVal(ptr)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if result != target {
		t.Fatalf("expected %q, got %q", target, result)
	}
	if ttl != 3600 {
		t.Fatalf("expected TTL 3600, got %d", ttl)
	}
}
