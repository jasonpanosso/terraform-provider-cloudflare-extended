package vectorize

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/cloudflare/cloudflare-go/v3/vectorize"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/apijson"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
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

	res := new(http.Response)
	env := VectorizeResultEnvelope{*data}
	_, err := r.client.Vectorize.Indexes.New(
		ctx,
		vectorize.IndexNewParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
			Name:      cloudflare.F(data.Name.ValueString()),
			Config: cloudflare.F(
				vectorize.IndexNewParamsConfigUnion(
					vectorize.IndexNewParamsConfig{
						Metric:     cloudflare.F(vectorize.IndexNewParamsConfigMetric(data.Metric.ValueString())),
						Dimensions: cloudflare.F(data.Dimensions.ValueInt64()),
					},
				),
			),
			Description: cloudflare.F(data.Description.ValueString()),
		},
		option.WithResponseBodyInto(&res),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	err = apijson.UnmarshalComputed(bytes, &env)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	data = &env.Result
	data.ID = data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorizeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *VectorizeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res := new(http.Response)
	env := VectorizeResultEnvelope{*data}
	_, err := r.client.Vectorize.Indexes.Get(
		ctx,
		data.Name.ValueString(),
		vectorize.IndexGetParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithResponseBodyInto(&res),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	err = apijson.UnmarshalComputed(bytes, &env)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	data = &env.Result

	var metadataIndexRes *vectorize.IndexMetadataIndexListResponse
	metadataIndexRes, err = r.client.Vectorize.Indexes.MetadataIndex.List(
		ctx,
		data.Name.ValueString(),
		vectorize.IndexMetadataIndexListParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to read metadata indexes", err.Error())
		return
	}

	var metadataIndexes []VectorizeMetadataIndexModel
	for _, mi := range metadataIndexRes.MetadataIndexes {
		metadataIndexes = append(metadataIndexes, VectorizeMetadataIndexModel{
			PropertyName: types.StringValue(mi.PropertyName),
			IndexType:    types.StringValue(string(mi.IndexType)),
		})

	}

	result, diags := customfield.NewObjectSet(ctx, metadataIndexes)
	resp.Diagnostics.Append(diags...)

	data.MetadataIndexes = result

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorizeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *VectorizeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddError("failed to update Vectorize index", "Not implemented")
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
