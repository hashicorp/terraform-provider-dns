package dns

import (
	"fmt"

	"github.com/miekg/dns"
)

func validateZone(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !dns.IsFqdn(value) {
		errors = append(errors, fmt.Errorf("DNS zone name %q must be fully qualified: %q", k, value))
	}
	return
}

func validateName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if dns.IsFqdn(value) {
		errors = append(errors, fmt.Errorf("DNS record name %q must not be fully qualified: %q", k, value))
	}
	return
}
