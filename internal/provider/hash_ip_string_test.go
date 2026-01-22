// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestHashIPString(t *testing.T) {
	ipv4 := []string{"192.168.0.1", "192.168.000.001"}
	ipv6 := []string{"fdd5:e282::1234:5678:cafe:9012", "FDD5:E282:0000:0000:1234:5678:CAFE:9012"}
	invalid := "not.an.ip.address"

	if hashIPString(ipv4[0]) != hashIPString(ipv4[1]) {
		t.Errorf("IPv4 values %s and %s should hash to the same value", ipv4[0], ipv4[1])
	}

	if hashIPString(ipv6[0]) != hashIPString(ipv6[1]) {
		t.Errorf("IPv6 values %s and %s should hash to the same value", ipv6[0], ipv6[1])
	}

	if hashIPString(invalid) != schema.HashString(invalid) {
		t.Errorf("Invalid IP value %s should hash to the same result as HashString()", invalid)
	}
}

func TestStripLeadingZeros(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected string
	}{
		"single-zero": {
			input:    "192.168.0.1",
			expected: "192.168.0.1",
		},
		"double-zero": {
			input:    "192.168.00.1",
			expected: "192.168.0.1",
		},
		"triple-zero": {
			input:    "192.168.000.1",
			expected: "192.168.0.1",
		},
		"leading-zero": {
			input:    "192.168.010.1",
			expected: "192.168.10.1",
		},
	}

	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := stripLeadingZeros(testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
