package r2_event_notification

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/event_notifications"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/logging"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.ResourceWithConfigure = (*R2EventNotificationResource)(nil)
var _ resource.ResourceWithModifyPlan = (*R2EventNotificationResource)(nil)

func NewResource() resource.Resource {
	return &R2EventNotificationResource{}
}

// R2EventNotificationResource defines the resource implementation.
type R2EventNotificationResource struct {
	client *cloudflare.Client
}

func (r *R2EventNotificationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_r2_event_notification"
}

func (r *R2EventNotificationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *R2EventNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *R2EventNotificationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rules := make([]R2EventNotificationRuleModel, len(data.Rules.Elements()))
	diags := data.Rules.ElementsAs(ctx, &rules, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleParams := make([]event_notifications.R2ConfigurationQueueUpdateParamsRule, 0)
	for _, rule := range rules {
		actions := make([]event_notifications.R2ConfigurationQueueUpdateParamsRulesAction, 0)
		for _, action := range rule.Actions {
			actions = append(actions, event_notifications.R2ConfigurationQueueUpdateParamsRulesAction(action.ValueString()))
		}

		ruleParams = append(
			ruleParams,
			event_notifications.R2ConfigurationQueueUpdateParamsRule{
				Actions: cloudflare.F(actions),
				Prefix:  cloudflare.F(rule.Prefix.ValueString()),
				Suffix:  cloudflare.F(rule.Suffix.ValueString()),
			})
	}

	_, err := r.client.EventNotifications.R2.Configuration.Queues.Update(
		ctx,
		data.BucketName.ValueString(),
		data.QueueID.ValueString(),
		event_notifications.R2ConfigurationQueueUpdateParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
			Rules:     cloudflare.F(ruleParams),
		},
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to create new r2 event notification", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *R2EventNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *R2EventNotificationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eventNotifs, err := r.client.EventNotifications.R2.Configuration.Get(
		ctx,
		data.BucketName.ValueString(),
		event_notifications.R2ConfigurationGetParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
	)

	var queueConfig event_notifications.R2ConfigurationGetResponseQueue
	found := false
	for _, queue := range eventNotifs.Queues {
		if queue.QueueID == data.QueueID.ValueString() {
			queueConfig = queue
			found = true
			break
		}
	}
	if !found {
		resp.Diagnostics.AddError("failed to create new r2 event notification rule", err.Error())
		return
	}

	data.QueueName = types.StringValue(queueConfig.QueueName)
	resRules := make([]R2EventNotificationRuleModel, 0)
	for _, rule := range queueConfig.Rules {
		actions := make([]types.String, 0)
		for _, action := range rule.Actions {
			actions = append(actions, types.StringValue(string(action)))
		}

		resRules = append(
			resRules,
			R2EventNotificationRuleModel{
				RuleID:    types.StringValue(rule.RuleID),
				Suffix:    types.StringValue(rule.Suffix),
				Prefix:    types.StringValue(rule.Prefix),
				CreatedAt: types.StringValue(rule.CreatedAt),
				Actions:   actions,
			},
		)
	}

	set, diags := customfield.NewObjectSet(ctx, resRules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Rules = set

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *R2EventNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *R2EventNotificationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state *R2EventNotificationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *R2EventNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *R2EventNotificationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *R2EventNotificationResource) ModifyPlan(_ context.Context, _ resource.ModifyPlanRequest, _ *resource.ModifyPlanResponse) {

}
