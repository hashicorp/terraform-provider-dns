// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"net"

	"github.com/hashicorp/terraform-provider-dns/internal/hashcode"
)

func hashIPString(v interface{}) int {
	//nolint:forcetypeassert
	addr := v.(string)
	ip := net.ParseIP(addr)
	if ip != nil {
		return hashcode.String(ip.String())
	}
	return hashcode.String(addr)
}
