package vectorize

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/vectorize"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TODO/FIXME: metadata indexes

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VectorizeResource{}
var _ resource.ResourceWithImportState = &VectorizeResource{}

func NewResource() resource.Resource {
	return &VectorizeResource{}
}

// VectorizeResource defines the resource implementation.
type VectorizeResource struct {
	client *cloudflare.Client
}

func (r *VectorizeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vectorize_database"
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

	metric, err := validateNewMetric(data.Metric.String())
	if err != nil {
		resp.Diagnostics.AddError("failed to create vectorize database", err.Error())
		return
	}

	database, err := r.client.Vectorize.Indexes.New(ctx, vectorize.IndexNewParams{
		AccountID: cloudflare.String(data.AccountID.String()),
		Name:      cloudflare.String(data.Name.String()),
		Config: cloudflare.F(vectorize.IndexNewParamsConfigUnion(vectorize.IndexNewParamsConfig{
			Metric:     cloudflare.F(metric),
			Dimensions: cloudflare.Int(data.Dimensions.ValueInt64()),
		})),
		Description: cloudflare.F(data.Description.String()),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create vectorize database", err.Error())
		return
	}

	data.ID = types.StringValue(database.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorizeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *VectorizeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	database, err := r.client.Vectorize.Indexes.Get(ctx, data.Name.String(), vectorize.IndexGetParams{AccountID: cloudflare.F(data.AccountID.String())})
	if err != nil {
		resp.Diagnostics.AddError("failed reading vectorize database", err.Error())
		return
	}

	data.ID = types.StringValue(database.Name)
	data.Name = types.StringValue(database.Name)
	data.Metric = types.StringValue(string(database.Config.Metric))
	data.Dimensions = types.Int64Value(database.Config.Dimensions)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VectorizeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *VectorizeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddError("failed to update Vectorize database", "Not implemented")
}

func (r *VectorizeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *VectorizeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Vectorize.Indexes.Delete(ctx, data.ID.String(),
		vectorize.IndexDeleteParams{AccountID: cloudflare.F(data.AccountID.String())})

	if err != nil {
		resp.Diagnostics.AddError("failed to delete Vectorize database", err.Error())
		return
	}
}

func (r *VectorizeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idparts := strings.Split(req.ID, "/")
	if len(idparts) != 2 {
		resp.Diagnostics.AddError("error importing Vectorize database", "invalid ID specified. Please specify the ID as \"account_id/name\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("account_id"), idparts[0],
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("id"), idparts[1],
	)...)
}

func validateMetric(input string) (vectorize.IndexDimensionConfigurationMetric, error) {
	var metric vectorize.IndexDimensionConfigurationMetric
	if input == string(vectorize.IndexDimensionConfigurationMetricCosine) {
		metric = vectorize.IndexDimensionConfigurationMetricCosine
	} else if input == string(vectorize.IndexDimensionConfigurationMetricEuclidean) {
		metric = vectorize.IndexDimensionConfigurationMetricEuclidean
	} else if input == string(vectorize.IndexDimensionConfigurationMetricDOTProduct) {
		metric = vectorize.IndexDimensionConfigurationMetricDOTProduct
	} else {
		return "", errors.New("variable Metric must be 'cosine', 'euclidean', or 'dot-product'")
	}

	return metric, nil
}

func validateNewMetric(input string) (vectorize.IndexNewParamsConfigMetric, error) {
	var metric vectorize.IndexNewParamsConfigMetric
	if input == string(vectorize.IndexNewParamsConfigMetricCosine) {
		metric = vectorize.IndexNewParamsConfigMetricCosine
	} else if input == string(vectorize.IndexNewParamsConfigMetricEuclidean) {
		metric = vectorize.IndexNewParamsConfigMetricEuclidean
	} else if input == string(vectorize.IndexNewParamsConfigMetricDOTProduct) {
		metric = vectorize.IndexNewParamsConfigMetricDOTProduct
	} else {
		return "", errors.New("variable Metric must be 'cosine', 'euclidean', or 'dot-product'")
	}

	return metric, nil
}
