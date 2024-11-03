package vectorize

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
)

type VectorizeModel struct {
	ID              types.String                           `tfsdk:"id"`
	AccountID       types.String                           `tfsdk:"account_id" path:"account_id,required"`
	Name            types.String                           `tfsdk:"name" path:"name,required"`
	Dimensions      types.Int64                            `tfsdk:"dimensions" path:"dimensions,required"`
	Metric          types.String                           `tfsdk:"metric" path:"metric,required"`
	Description     types.String                           `tfsdk:"description"  path:"description,computed_optional"`
	CreatedOn       types.String                           `tfsdk:"created_on" json:"created_on,computed"`
	ModifiedOn      types.String                           `tfsdk:"modified_on" json:"modified_on,computed"`
	MetadataIndexes customfield.Map[basetypes.StringValue] `tfsdk:"metadata_indexes" path:"metadata_indexes,optional"`
}
