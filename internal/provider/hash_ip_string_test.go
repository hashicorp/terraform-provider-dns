package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestHashIPString(t *testing.T) {
	ipv6 := []string{"fdd5:e282::dead:beef:cafe:babe", "FDD5:E282:0000:0000:DEAD:BEEF:CAFE:BABE"}
	invalid := "not.an.ip.address"

	if hashIPString(ipv6[0]) != hashIPString(ipv6[1]) {
		t.Errorf("IPv6 values %s and %s should hash to the same value", ipv6[0], ipv6[1])
	}

	if hashIPString(invalid) != schema.HashString(invalid) {
		t.Errorf("Invalid IP value %s should hash to the same result as HashString()", invalid)
	}
}
