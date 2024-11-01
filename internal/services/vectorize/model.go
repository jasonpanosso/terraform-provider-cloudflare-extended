package vectorize

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/apijson"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
)

type VectorizeResultEnvelope struct {
	Result VectorizeModel `json:"result"`
}

type VectorizeModel struct {
	ID              types.String                                             `tfsdk:"id"`
	AccountID       types.String                                             `tfsdk:"account_id" path:"account_id,required"`
	Name            types.String                                             `tfsdk:"name" path:"name,required"`
	Dimensions      types.Int64                                              `tfsdk:"dimensions" path:"dimensions,required"`
	Metric          types.String                                             `tfsdk:"metric" path:"metric,required"`
	Description     types.String                                             `tfsdk:"description"  path:"description,optional"`
	CreatedOn       types.String                                             `tfsdk:"created_on" json:"created_on,computed"`
	ModifiedOn      types.String                                             `tfsdk:"modified_on" json:"modified_on,computed"`
	MetadataIndexes customfield.NestedObjectSet[VectorizeMetadataIndexModel] `tfsdk:"metadata_indexes" path:"metadata_indexes,optional"`
}

type VectorizeMetadataIndexModel struct {
	IndexType    types.String `tfsdk:"index_type" path:"index_type,required"`
	PropertyName types.String `tfsdk:"property_name" path:"index_type,required"`
}

func (m VectorizeModel) MarshalJSON() (data []byte, err error) {
	return apijson.MarshalRoot(m)
}

func (m VectorizeModel) MarshalJSONForUpdate(state VectorizeModel) (data []byte, err error) {
	return apijson.MarshalForUpdate(m, state)
}
