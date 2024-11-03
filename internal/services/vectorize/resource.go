package vectorize

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/cloudflare/cloudflare-go/v3/vectorize"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/logging"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.ResourceWithConfigure = (*VectorizeResource)(nil)
var _ resource.ResourceWithModifyPlan = (*VectorizeResource)(nil)

func NewResource() resource.Resource {
	return &VectorizeResource{}
}

// VectorizeResource defines the resource implementation.
type VectorizeResource struct {
	client *cloudflare.Client
}

func (r *VectorizeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vectorize_index"
}

func (r *VectorizeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cloudflare.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"unexpected resource configure type",
			fmt.Sprintf("Expected *cloudflare.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *VectorizeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *VectorizeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newIndex, err := r.client.Vectorize.Indexes.New(
		ctx,
		vectorize.IndexNewParams{
			AccountID:   cloudflare.F(data.AccountID.ValueString()),
			Name:        cloudflare.F(data.Name.ValueString()),
			Description: cloudflare.F(data.Description.ValueString()),
			Config: cloudflare.F(
				vectorize.IndexNewParamsConfigUnion(
					vectorize.IndexNewParamsConfig{
						Metric:     cloudflare.F(vectorize.IndexNewParamsConfigMetric(data.Metric.ValueString())),
						Dimensions: cloudflare.F(data.Dimensions.ValueInt64()),
					},
				),
			),
		},
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to create new vectorize index", err.Error())
		return
	}

	data.ID = basetypes.NewStringValue(newIndex.Name)
	data.Description = basetypes.NewStringValue(newIndex.Description)
	data.CreatedOn = basetypes.NewStringValue(newIndex.CreatedOn.String())
	data.ModifiedOn = basetypes.NewStringValue(newIndex.ModifiedOn.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	metadataIndexes := make(map[string]basetypes.StringValue)
	diags := data.MetadataIndexes.ElementsAs(ctx, &metadataIndexes, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for propertyName, indexType := range metadataIndexes {
		_, err := r.client.Vectorize.Indexes.MetadataIndex.New(
			ctx,
			data.Name.ValueString(),
			vectorize.IndexMetadataIndexNewParams{
				AccountID:    cloudflare.F(data.AccountID.ValueString()),
				PropertyName: cloudflare.F(propertyName),
				IndexType:    cloudflare.F(vectorize.IndexMetadataIndexNewParamsIndexType(indexType.ValueString())),
			},
			option.WithMiddleware(logging.Middleware(ctx)),
		)
		if err != nil {
			resp.Diagnostics.AddError("failed to create vectorize metadata index", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorizeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *VectorizeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	index, err := r.client.Vectorize.Indexes.Get(
		ctx,
		data.Name.ValueString(),
		vectorize.IndexGetParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to read current state of vectorize index", err.Error())
		return
	}

	data.ID = basetypes.NewStringValue(index.Name)
	data.Name = basetypes.NewStringValue(index.Name)
	data.Description = basetypes.NewStringValue(index.Description)
	data.Dimensions = basetypes.NewInt64Value(index.Config.Dimensions)
	data.Metric = basetypes.NewStringValue(string(index.Config.Metric))
	data.CreatedOn = basetypes.NewStringValue(index.CreatedOn.String())
	data.ModifiedOn = basetypes.NewStringValue(index.ModifiedOn.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorizeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *VectorizeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state *VectorizeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	curIndexes := make(map[string]basetypes.StringValue)
	diags := state.MetadataIndexes.ElementsAs(ctx, &curIndexes, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newIndexes := make(map[string]basetypes.StringValue)
	diags = data.MetadataIndexes.ElementsAs(ctx, &newIndexes, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for propName, oldType := range curIndexes {
		newType, exists := newIndexes[propName]

		if !exists || oldType != newType {
			_, err := r.client.Vectorize.Indexes.MetadataIndex.Delete(
				ctx,
				data.Name.ValueString(),
				vectorize.IndexMetadataIndexDeleteParams{
					AccountID:    cloudflare.F(data.AccountID.ValueString()),
					PropertyName: cloudflare.F(propName),
				},
				option.WithMiddleware(logging.Middleware(ctx)),
			)
			if err != nil {
				resp.Diagnostics.AddError("failed to update metadata indexes", err.Error())
				return
			}
		}

		if oldType != newType {
			_, err := r.client.Vectorize.Indexes.MetadataIndex.New(
				ctx,
				data.Name.ValueString(),
				vectorize.IndexMetadataIndexNewParams{
					AccountID:    cloudflare.F(data.AccountID.ValueString()),
					PropertyName: cloudflare.F(propName),
					IndexType:    cloudflare.F(vectorize.IndexMetadataIndexNewParamsIndexType(newType.ValueString())),
				},
				option.WithMiddleware(logging.Middleware(ctx)),
			)

			if err != nil {
				resp.Diagnostics.AddError("failed to update metadata indexes", err.Error())
				return
			}
		}
	}

	for propName, newType := range newIndexes {
		if _, exists := curIndexes[propName]; !exists {
			_, err := r.client.Vectorize.Indexes.MetadataIndex.New(
				ctx,
				data.Name.ValueString(),
				vectorize.IndexMetadataIndexNewParams{
					AccountID:    cloudflare.F(data.AccountID.ValueString()),
					PropertyName: cloudflare.F(propName),
					IndexType:    cloudflare.F(vectorize.IndexMetadataIndexNewParamsIndexType(newType.ValueString())),
				},
				option.WithMiddleware(logging.Middleware(ctx)),
			)
			if err != nil {
				resp.Diagnostics.AddError("failed to update metadata indexes", err.Error())
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorizeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *VectorizeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Vectorize.Indexes.Delete(
		ctx,
		data.ID.ValueString(),
		vectorize.IndexDeleteParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorizeResource) ModifyPlan(_ context.Context, _ resource.ModifyPlanRequest, _ *resource.ModifyPlanResponse) {

}
