// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

func validateZone(v interface{}, k string) (ws []string, errors []error) {
	//nolint:forcetypeassert
	value := v.(string)
	if strings.TrimSpace(value) != value {
		errors = append(errors, fmt.Errorf("DNS zone name %q must not contain whitespace: %q", k, value))
	}
	if !dns.IsFqdn(value) {
		errors = append(errors, fmt.Errorf("DNS zone name %q must be fully qualified: %q", k, value))
	}
	return
}

func validateName(v interface{}, k string) (ws []string, errors []error) {
	//nolint:forcetypeassert
	value := v.(string)
	if strings.TrimSpace(value) != value || len(value) == 0 {
		errors = append(errors, fmt.Errorf("DNS record name %q must not contain whitespace or be empty: %q", k, value))
	}
	if dns.IsFqdn(value) {
		errors = append(errors, fmt.Errorf("DNS record name %q must not be fully qualified: %q", k, value))
	}
	return
}
