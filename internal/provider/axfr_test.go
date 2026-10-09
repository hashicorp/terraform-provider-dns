// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
)

func TestAxfrQuery_NotAuthoritative(t *testing.T) {
	t.Skip("AXFR requires zone transfer permissions; cannot unit test without infrastructure")
}
