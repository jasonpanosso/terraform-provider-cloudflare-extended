package queue_consumer

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/cloudflare/cloudflare-go/v3/queues"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/apijson"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/logging"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.ResourceWithConfigure = (*QueueConsumerResource)(nil)
var _ resource.ResourceWithModifyPlan = (*QueueConsumerResource)(nil)

func NewResource() resource.Resource {
	return &QueueConsumerResource{}
}

// QueueConsumerResource defines the resource implementation.
type QueueConsumerResource struct {
	client *cloudflare.Client
}

func (r *QueueConsumerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_queue_consumer"
}

func (r *QueueConsumerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *QueueConsumerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *QueueConsumerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dataBytes, err := data.MarshalJSON()
	if err != nil {
		resp.Diagnostics.AddError("failed to serialize http request", err.Error())
		return
	}
	res := new(http.Response)
	env := QueueConsumerResultEnvelope{*data}
	_, err = r.client.Queues.Consumers.New(
		ctx,
		data.QueueID.ValueString(),
		queues.ConsumerNewParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithRequestBody("application/json", dataBytes),
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
	data.ConsumerID = env.Result.ConsumerID
	data.Type = env.Result.Type
	data.Settings = env.Result.Settings
	data.QueueName = env.Result.QueueName
	data.CreatedOn = env.Result.CreatedOn
	data.Environment = env.Result.Environment

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *QueueConsumerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *QueueConsumerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state *QueueConsumerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataBytes, err := data.MarshalJSONForUpdate(*state)
	if err != nil {
		resp.Diagnostics.AddError("failed to serialize http request", err.Error())
		return
	}
	res := new(http.Response)
	env := QueueConsumerResultEnvelope{*data}
	_, err = r.client.Queues.Consumers.Update(
		ctx,
		data.QueueID.ValueString(),
		data.ConsumerID.ValueString(),
		queues.ConsumerUpdateParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithRequestBody("application/json", dataBytes),
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *QueueConsumerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *QueueConsumerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	consumers, err := r.client.Queues.Consumers.Get(
		ctx,
		data.QueueID.ValueString(),
		queues.ConsumerGetParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	if consumers == nil {
		req.State.RemoveResource(ctx)
		return
	}

	found := false
	for _, consumer := range *consumers {
		if consumer.Service == data.ScriptName.ValueString() {
			found = true
			data.CreatedOn = basetypes.NewStringValue(consumer.CreatedOn)
			data.Environment = basetypes.NewStringValue(consumer.Environment)
			data.QueueName = basetypes.NewStringValue(consumer.QueueName)
			data.Settings = customfield.NewObjectMust(
				ctx,
				&QueueConsumerSettingsModel{
					BatchSize:     basetypes.NewFloat64Value(consumer.Settings.BatchSize),
					MaxRetries:    basetypes.NewFloat64Value(consumer.Settings.MaxRetries),
					MaxWaitTimeMs: basetypes.NewFloat64Value(consumer.Settings.MaxWaitTimeMs),
				})
		}
	}

	if !found {
		req.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *QueueConsumerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *QueueConsumerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Queues.Consumers.Delete(
		ctx,
		data.QueueID.ValueString(),
		data.ConsumerID.ValueString(),
		queues.ConsumerDeleteParams{
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

func (r *QueueConsumerResource) ModifyPlan(_ context.Context, _ resource.ModifyPlanRequest, _ *resource.ModifyPlanResponse) {

}
