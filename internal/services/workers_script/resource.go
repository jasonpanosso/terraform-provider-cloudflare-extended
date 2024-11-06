package workers_script

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/cloudflare/cloudflare-go/v3/workers"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/apijson"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/importpath"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/logging"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.ResourceWithConfigure = (*WorkersScriptResource)(nil)
var _ resource.ResourceWithModifyPlan = (*WorkersScriptResource)(nil)
var _ resource.ResourceWithImportState = (*WorkersScriptResource)(nil)

func NewResource() resource.Resource {
	return &WorkersScriptResource{}
}

// WorkersScriptResource defines the resource implementation.
type WorkersScriptResource struct {
	client *cloudflare.Client
}

func (r *WorkersScriptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workers_script"
}

func (r *WorkersScriptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkersScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *WorkersScriptModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataBytes, contentType, err := data.MarshalMultipart()
	if err != nil {
		resp.Diagnostics.AddError("failed to serialize multipart http request", err.Error())
		return
	}
	created, err := r.client.Workers.Scripts.Update(
		ctx,
		data.ScriptName.ValueString(),
		workers.ScriptUpdateParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithRequestBody(contentType, dataBytes),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}

	updateModelFromResponse(ctx, data, created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkersScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *WorkersScriptModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var state *WorkersScriptModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	r.handleUpdate(ctx, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkersScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *WorkersScriptModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res := new(http.Response)
	path := fmt.Sprintf("accounts/%s/workers/scripts/%s/settings", data.AccountID.ValueString(), data.ID.ValueString())
	err := r.client.Execute(
		ctx,
		http.MethodGet,
		path,
		nil,
		&res,
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}

	bytes := make([]byte, 0)
	_, err = res.Body.Read(bytes)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}

	env := WorkersScriptSettingResponseEnvelope{}
	err = apijson.Unmarshal(bytes, &env)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}

	data.Logpush = env.Result.Logpush
	data.UsageModel = env.Result.UsageModel

	tcModel, diags := env.Result.TailConsumers.AsStructSliceT(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.TailConsumers = customfield.NewObjectSetMust(ctx, tcModel)

	placement := WorkersScriptMetadataPlacementModel{}
	diags = env.Result.Placement.As(ctx, &placement, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	resp.Diagnostics = append(resp.Diagnostics, diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.PlacementMode = placement.Mode

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkersScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *WorkersScriptModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Workers.Scripts.Delete(
		ctx,
		data.ScriptName.ValueString(),
		workers.ScriptDeleteParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	data.ID = data.ScriptName

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkersScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data *WorkersScriptModel

	path_account_id := ""
	path_script_name := ""
	diags := importpath.ParseImportID(
		req.ID,
		"<account_id>/<script_name>",
		&path_account_id,
		&path_script_name,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res := new(http.Response)
	_, err := r.client.Workers.Scripts.Get(
		ctx,
		path_script_name,
		workers.ScriptGetParams{
			AccountID: cloudflare.F(path_account_id),
		},
		option.WithResponseBodyInto(&res),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	err = apijson.UnmarshalComputed(bytes, &data)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	data.ID = data.ScriptName

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkersScriptResource) ModifyPlan(_ context.Context, _ resource.ModifyPlanRequest, _ *resource.ModifyPlanResponse) {

}

func (r *WorkersScriptResource) handleUpdate(ctx context.Context, data *WorkersScriptModel, diags *diag.Diagnostics) {
	dataBytes, contentType, err := data.MarshalMultipart()
	if err != nil {
		diags.AddError("failed to serialize multipart http request", err.Error())
		return
	}

	created, err := r.client.Workers.Scripts.Update(
		ctx,
		data.ScriptName.ValueString(),
		workers.ScriptUpdateParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithRequestBody(contentType, dataBytes),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		diags.AddError("failed to make http request", err.Error())
		return
	}

	updateModelFromResponse(ctx, data, created)
}

func updateModelFromResponse(ctx context.Context, model *WorkersScriptModel, res *workers.ScriptUpdateResponse) {
	model.Etag = types.StringValue(res.Etag)
	model.ID = types.StringValue(res.ID)
	model.Logpush = types.BoolValue(res.Logpush)
	model.UsageModel = types.StringValue(res.UsageModel)
	model.CreatedOn = timetypes.NewRFC3339TimeValue(res.CreatedOn)
	model.ModifiedOn = timetypes.NewRFC3339TimeValue(res.ModifiedOn)
	model.PlacementMode = types.StringValue(res.PlacementMode)
	model.StartupTimeMs = types.Int64Value(res.StartupTimeMs)

	var tailConsumerModels []WorkersScriptTailConsumersModel
	for _, tc := range res.TailConsumers {
		tailConsumerModels = append(
			tailConsumerModels,
			WorkersScriptTailConsumersModel{
				Service:     types.StringValue(tc.Service),
				Environment: types.StringValue(tc.Environment),
				Namespace:   types.StringValue(tc.Namespace),
			})
	}

	model.TailConsumers = customfield.NewObjectSetMust(ctx, tailConsumerModels)
}
