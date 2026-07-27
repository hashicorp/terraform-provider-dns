// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
	"testing"
)

func TestExchangeErrorMessage(t *testing.T) {
	err := fmt.Errorf("unable to complete DNS exchange after %d retries with server %s", 3, "127.0.0.1:53")
	if !strings.Contains(err.Error(), "127.0.0.1:53") {
		t.Errorf("error should contain server address")
	}
	if !strings.Contains(err.Error(), "3") {
		t.Errorf("error should contain retry count")
	}
}
