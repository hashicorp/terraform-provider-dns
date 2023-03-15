package dnsvalidator

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/miekg/dns"
)

var _ validator.String = dnsRecordNameValidator{}

// dnsZoneValidator validates if the provided value is a fully qualified DNS zone name and contains no whitespace.
type dnsRecordNameValidator struct {
}

func (validator dnsRecordNameValidator) Description(ctx context.Context) string {
	return "value must be a fully qualified DNS record name"
}

func (validator dnsRecordNameValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator dnsRecordNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Only validate the attribute configuration value if it is known.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	if strings.TrimSpace(value) != value || len(value) == 0 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeTypeDiagnostic(
			req.Path,
			"DNS record name must not contain whitespace or be empty: ",
			req.ConfigValue.ValueString(),
		))
	}
	if dns.IsFqdn(value) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeTypeDiagnostic(
			req.Path,
			"DNS record name must not be fully qualified: ",
			req.ConfigValue.ValueString(),
		))
	}
	return
}

// IsRecordNameValid returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is a non-empty String.
//   - Contains no whitespace.
//   - Is NOT a fully qualified DNS zone name.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.

func IsRecordNameValid() validator.String {
	return dnsRecordNameValidator{}
}
