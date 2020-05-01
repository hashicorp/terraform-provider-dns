package dns

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestHashIPString(t *testing.T) {
	ipv4 := []string{"192.168.0.1", "192.168.000.001"}
	ipv6 := []string{"fdd5:e282::dead:beef:cafe:babe", "FDD5:E282:0000:0000:DEAD:BEEF:CAFE:BABE"}
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
