// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
)

func TestValidateKeyName(t *testing.T) {
	validNames := []string{
		"mykey",
		"update",
		"mykey.example.com.",
		"key.with.multiple.labels",
	}
	for _, v := range validNames {
		_, errors := validateKeyName(v, "key_name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid key name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"",
		" ",
	}
	for _, v := range invalidNames {
		_, errors := validateKeyName(v, "key_name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid key name", v)
		}
	}
}

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
