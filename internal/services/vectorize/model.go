package vectorize

import "github.com/hashicorp/terraform-plugin-framework/types"

type VectorizeModel struct {
	AccountID   types.String `tfsdk:"account_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	Dimensions  types.Int64  `tfsdk:"dimensions"`
	Metric      types.String `tfsdk:"metric"`
}
