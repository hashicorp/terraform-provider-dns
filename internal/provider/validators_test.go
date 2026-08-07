// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
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

func TestValidateName(t *testing.T) {
	validNames := []string{
		"test",
		"alias.subdomain",
		"host.subdomain.example",
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

func TestValidateNameSubdomain(t *testing.T) {
	validNames := []string{
		"alias.subdomain",
		"host.subdomain.example",
		"a.b.c.d",
		"my-host.subdomain",
		"www",
		"sub1.sub2.sub3.sub4",
	}
	for _, v := range validNames {
		_, errors := validateName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid DNS record name with subdomains: %q", v, errors)
		}
	}

	invalidNames := []string{
		"alias.subdomain.",
		" host.subdomain",
		"",
		" ",
	}
	for _, v := range invalidNames {
		_, errors := validateName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid DNS record name", v)
		}
	}
}
