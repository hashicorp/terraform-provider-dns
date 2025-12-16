// Copyright IBM Corp. 2016, 2025
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
