package dnstypes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satifies the expected interfaces
var _ basetypes.StringValuable = IPAddressValue{}
var _ basetypes.StringValuableWithSemanticEquals = IPAddressValue{}

type IPAddressValue struct {
	basetypes.StringValue
}

func (v IPAddressValue) StringSemanticEquals(ctx context.Context, valuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newIpAddress, ok := valuable.(IPAddressValue)

	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newIpAddress),
		)

		return false, diags
	}

	priorIpAddressStripped := stripLeadingZeros(v.StringValue.ValueString())

	newIpAddressStripped := stripLeadingZeros(newIpAddress.String())

	// If the times are equivalent, keep the prior value
	return priorIpAddressStripped == newIpAddressStripped, nil
}

func (v IPAddressValue) Equal(o attr.Value) bool {
	other, ok := o.(IPAddressValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v IPAddressValue) Type(ctx context.Context) attr.Type {
	return IPAddressType{}
}

func stripLeadingZeros(input string) string {
	if strings.Contains(input, ".") {
		classes := strings.Split(input, ".")
		if len(classes) != 4 {
			return input
		}
		for classIndex, class := range classes {
			if len(class) <= 1 {
				continue
			}
			classes[classIndex] = strings.TrimLeft(class, "0")
			if classes[classIndex] == "" {
				classes[classIndex] = "0"
			}
		}
		return strings.Join(classes, ".")
	}

	return input
}
