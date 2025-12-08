// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package dnsvalidator

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/miekg/dns"
)

var _ validator.String = dnsZoneNameValidator{}

// dnsZoneValidator validates if the provided value is a fully qualified DNS zone name and contains no whitespace.
type dnsZoneNameValidator struct{}

func (validator dnsZoneNameValidator) Description(ctx context.Context) string {
	return "value must be a fully qualified DNS zone name"
}

func (validator dnsZoneNameValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator dnsZoneNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Only validate the attribute configuration value if it is known.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	if strings.TrimSpace(value) != value {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeTypeDiagnostic(
			req.Path,
			"DNS zone name must not contain whitespace",
			req.ConfigValue.ValueString(),
		))
	}
	if !dns.IsFqdn(value) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeTypeDiagnostic(
			req.Path,
			"DNS zone name must be fully qualified",
			req.ConfigValue.ValueString(),
		))
	}
}

// IsZoneNameValid returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is a non-empty String.
//   - Contains no whitespace.
//   - Is a fully qualified DNS zone name.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IsZoneNameValid() validator.String {
	return dnsZoneNameValidator{}
}
