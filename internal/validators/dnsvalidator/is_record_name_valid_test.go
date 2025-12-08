// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package dnsvalidator

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIsRecordNameValid(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
	}

	tests := map[string]testCase{
		"string unknown": {
			val:         types.StringUnknown(),
			expectError: false,
		},
		"string null": {
			val:         types.StringNull(),
			expectError: false,
		},
		"string empty": {
			val:         types.StringValue(""),
			expectError: true,
		},
		"is a fully qualified DNS name": {
			val:         types.StringValue("example.com."),
			expectError: true,
		},
		"string contains whitespace": {
			val:         types.StringValue(" example"),
			expectError: true,
		},
		"string only whitespace": {
			val:         types.StringValue(" "),
			expectError: true,
		},
		"success scenario": {
			val:         types.StringValue("test"),
			expectError: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}

			response := validator.StringResponse{}
			IsRecordNameValid().ValidateString(context.TODO(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
