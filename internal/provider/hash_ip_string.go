package provider

import (
	"net"

	"github.com/hashicorp/terraform-provider-dns/internal/hashcode"
)

func hashIPString(v interface{}) int {
	addr := v.(string)
	ip := net.ParseIP(addr)
	if ip == nil {
		return hashcode.String(addr)
	}
	return hashcode.String(ip.String())
}
