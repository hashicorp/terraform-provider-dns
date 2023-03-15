package int64modifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// int64DefaultModifier is a plan modifier that sets a default value for a
// types.Int64Type attribute when it is not configured. The attribute must be
// marked as Optional and Computed. When setting the state during the resource
// Create, Read, or Update methods, this default value must also be included or
// the Terraform CLI will generate an error.
type int64DefaultModifier struct {
	Default int64
}

func (m int64DefaultModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %q", m.Default)
}

func (m int64DefaultModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %q", m.Default)
}

func (m int64DefaultModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// If the value is known, do not set default value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	resp.PlanValue = types.Int64Value(m.Default)
}

func Int64Default(defaultValue int64) planmodifier.Int64 {
	return int64DefaultModifier{
		Default: defaultValue,
	}
}
