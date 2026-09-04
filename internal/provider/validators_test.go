// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"net"
	"testing"
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

func TestParseIPv6Address(t *testing.T) {
	invalidIPs := []string{"", "foo", "not-an-ip"}
	for _, ip := range invalidIPs {
		parsed := net.ParseIP(ip)
		if parsed != nil {
			t.Errorf("expected %q to be invalid", ip)
		}
	}

	validIPs := []string{"::1", "fdd5:e282::1234:5678:cafe:9012"}
	for _, ip := range validIPs {
		parsed := net.ParseIP(ip)
		if parsed == nil {
			t.Errorf("expected %q to be valid", ip)
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
	}
	for _, v := range invalidNames {
		_, errors := validateName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid DNS record", v)
		}
	}
}
