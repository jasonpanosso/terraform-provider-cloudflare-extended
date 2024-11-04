package r2_event_notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/event_notifications"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

	createTimeout, diags := data.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	rules := make([]R2EventNotificationRuleModel, len(data.Rules.Elements()))
	resp.Diagnostics.Append(data.Rules.ElementsAs(ctx, &rules, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateEventNotification(ctx, data, rules, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.verifyConfigurationUpdatedAndSetRuleIDs(ctx, data, &resp.Diagnostics, &resp.State)
}

func (r *R2EventNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *R2EventNotificationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queue, err := r.getQueue(ctx, data, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("could not read r2 event notification", err.Error())
		return
	}

	apiRules, diags := convertToRuleModels(ctx, queue.Rules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	set, diags := customfield.NewObjectSet(ctx, apiRules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Rules = set
	data.QueueName = types.StringValue(queue.QueueName)

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

	updateTimeout, diags := data.Timeouts.Update(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	var stateRules []R2EventNotificationRuleModel
	resp.Diagnostics.Append(state.Rules.ElementsAs(ctx, &stateRules, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newRules []R2EventNotificationRuleModel
	resp.Diagnostics.Append(data.Rules.ElementsAs(ctx, &newRules, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleIDsToDelete, rulesToAdd := findRulesToDeleteAndAdd(stateRules, newRules)
	if len(ruleIDsToDelete) > 0 {
		reqBody := R2EventNotificationDeleteRequestBody{
			RuleIds: ruleIDsToDelete,
		}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			resp.Diagnostics.AddError("Error serializing delete request body", err.Error())
			return
		}

		_, err = r.client.EventNotifications.R2.Configuration.Queues.Delete(
			ctx,
			data.BucketName.ValueString(),
			data.QueueID.ValueString(),
			event_notifications.R2ConfigurationQueueDeleteParams{
				AccountID: cloudflare.F(data.AccountID.ValueString()),
			},
			option.WithMiddleware(logging.Middleware(ctx)),
			option.WithRequestBody("application/json", jsonData),
		)
		if err != nil {
			resp.Diagnostics.AddError("error deleting rules", err.Error())
			return
		}
	}

	if len(rulesToAdd) > 0 {
		r.updateEventNotification(ctx, data, rulesToAdd, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	r.verifyConfigurationUpdatedAndSetRuleIDs(ctx, data, &resp.Diagnostics, &resp.State)
}

func (r *R2EventNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *R2EventNotificationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.EventNotifications.R2.Configuration.Queues.Delete(
		ctx,
		data.BucketName.ValueString(),
		data.QueueID.ValueString(),
		event_notifications.R2ConfigurationQueueDeleteParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("error deleting r2 event notification", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *R2EventNotificationResource) ModifyPlan(_ context.Context, _ resource.ModifyPlanRequest, _ *resource.ModifyPlanResponse) {

}

func (r *R2EventNotificationResource) updateEventNotification(
	ctx context.Context,
	data *R2EventNotificationModel,
	rules []R2EventNotificationRuleModel,
	diagnostics *diag.Diagnostics,
) {
	updateParamsRules, diags := convertToUpdateParamsRules(ctx, rules)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	_, err := r.client.EventNotifications.R2.Configuration.Queues.Update(
		ctx,
		data.BucketName.ValueString(),
		data.QueueID.ValueString(),
		event_notifications.R2ConfigurationQueueUpdateParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
			Rules:     cloudflare.F(updateParamsRules),
		},
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		diagnostics.AddError("failed update r2 event notifications", err.Error())
		return
	}
}

func (r *R2EventNotificationResource) verifyConfigurationUpdatedAndSetRuleIDs(
	ctx context.Context,
	data *R2EventNotificationModel,
	diagnostics *diag.Diagnostics,
	state *tfsdk.State,
) {
	interval := 10 * time.Second
	for {
		select {
		case <-ctx.Done():
			diagnostics.AddError(
				"Timed out waiting for R2 event notification to become available",
				"The resource did not become available within the allotted time",
			)
			return
		default:
			queue, err := r.getQueue(ctx, data, diagnostics)
			if diagnostics.HasError() {
				return
			}

			if err != nil {
				time.Sleep(interval)
				continue
			}

			rules, diags := convertToRuleModels(ctx, queue.Rules)
			diagnostics.Append(diags...)
			if diagnostics.HasError() {
				return
			}

			var dataRules []R2EventNotificationRuleModel
			diagnostics.Append(data.Rules.ElementsAs(ctx, &dataRules, false)...)
			if diagnostics.HasError() {
				return
			}

			if rulesSlicesEqual(rules, dataRules) {
				// to set rule IDs from API
				set, diags := customfield.NewObjectSet(ctx, rules)
				diagnostics.Append(diags...)
				if diagnostics.HasError() {
					return
				}
				data.Rules = set
				data.QueueName = types.StringValue(queue.QueueName)

				diagnostics.Append(state.Set(ctx, &data)...)
				return
			}

			time.Sleep(interval)
		}
	}
}

func (r *R2EventNotificationResource) getQueue(
	ctx context.Context,
	data *R2EventNotificationModel,
	diagnostics *diag.Diagnostics,
) (*event_notifications.R2ConfigurationGetResponseQueue, error) {
	eventNotifs, err := r.client.EventNotifications.R2.Configuration.Get(
		ctx,
		data.BucketName.ValueString(),
		event_notifications.R2ConfigurationGetParams{
			AccountID: cloudflare.F(data.AccountID.ValueString()),
		},
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		diagnostics.AddError("failed to get event notifications", err.Error())
		return nil, err
	}

	for _, q := range eventNotifs.Queues {
		// remove hyphens from QueueID to match the format used in data.QueueID
		q.QueueID = strings.ReplaceAll(q.QueueID, "-", "")
		if q.QueueID == strings.ReplaceAll(data.QueueID.ValueString(), "-", "") {
			return &q, nil
		}
	}

	return nil, errors.New("could not find queue associated with event notification")
}

func convertToUpdateParamsRules(
	ctx context.Context,
	ruleModels []R2EventNotificationRuleModel,
) ([]event_notifications.R2ConfigurationQueueUpdateParamsRule, diag.Diagnostics) {
	var paramsRules []event_notifications.R2ConfigurationQueueUpdateParamsRule
	var diags diag.Diagnostics
	for _, rule := range ruleModels {
		updateActions := make([]event_notifications.R2ConfigurationQueueUpdateParamsRulesAction, len(rule.Actions.Elements()))

		actions := make([]string, len(rule.Actions.Elements()))
		diags = append(diags, rule.Actions.ElementsAs(ctx, &actions, false)...)
		if diags.HasError() {
			return nil, diags
		}

		for i, action := range actions {
			updateActions[i] = event_notifications.R2ConfigurationQueueUpdateParamsRulesAction(action)
		}

		paramsRules = append(
			paramsRules,
			event_notifications.R2ConfigurationQueueUpdateParamsRule{
				Actions: cloudflare.F(updateActions),
				Prefix:  cloudflare.F(rule.Prefix.ValueString()),
				Suffix:  cloudflare.F(rule.Suffix.ValueString()),
			},
		)
	}

	return paramsRules, diags
}

func convertToRuleModels(
	ctx context.Context,
	apiRules []event_notifications.R2ConfigurationGetResponseQueuesRule,
) ([]R2EventNotificationRuleModel, diag.Diagnostics) {
	var models []R2EventNotificationRuleModel
	var allDiags diag.Diagnostics
	for _, rule := range apiRules {
		actions := make([]types.String, len(rule.Actions))
		for i, action := range rule.Actions {
			actions[i] = types.StringValue(string(action))
		}

		actionSet, diags := customfield.NewSet[types.String](ctx, actions)
		allDiags = append(allDiags, diags...)

		models = append(
			models,
			R2EventNotificationRuleModel{
				RuleID:    types.StringValue(rule.RuleID),
				Suffix:    types.StringValue(rule.Suffix),
				Prefix:    types.StringValue(rule.Prefix),
				CreatedAt: types.StringValue(rule.CreatedAt),
				Actions:   actionSet,
			},
		)
	}

	return models, allDiags
}

func findRulesToDeleteAndAdd(stateRules, newRules []R2EventNotificationRuleModel) ([]string, []R2EventNotificationRuleModel) {
	stateRulesByID := make(map[string]R2EventNotificationRuleModel)
	for _, rule := range stateRules {
		stateRulesByID[rule.RuleID.ValueString()] = rule
	}

	newRulesByID := make(map[string]R2EventNotificationRuleModel)
	var dataRulesWithoutID []R2EventNotificationRuleModel
	for _, rule := range newRules {
		if !rule.RuleID.IsNull() && !rule.RuleID.IsUnknown() && rule.RuleID.ValueString() != "" {
			newRulesByID[rule.RuleID.ValueString()] = rule
		} else {
			dataRulesWithoutID = append(dataRulesWithoutID, rule)
		}
	}

	var ruleIDsToDelete []string
	for ruleID := range stateRulesByID {
		if _, exists := newRulesByID[ruleID]; !exists {
			// rule exists in state but not in data, mark for deletion
			ruleIDsToDelete = append(ruleIDsToDelete, ruleID)
		}
	}

	var rulesToUpdate []R2EventNotificationRuleModel
	for ruleID, newRule := range newRulesByID {
		if oldRule, exists := stateRulesByID[ruleID]; exists {
			if !rulesEqual(oldRule, newRule) {
				// content has changed, need to delete & re-add
				ruleIDsToDelete = append(ruleIDsToDelete, ruleID)
				rulesToUpdate = append(rulesToUpdate, newRule)
			}
		} else {
			// RuleID exists in data but not in state, treat as new rule
			rulesToUpdate = append(rulesToUpdate, newRule)
		}
	}

	rulesToAdd := append(dataRulesWithoutID, rulesToUpdate...)

	return ruleIDsToDelete, rulesToAdd
}

func rulesSlicesEqual(a, b []R2EventNotificationRuleModel) bool {
	if len(a) != len(b) {
		return false
	}

	for _, ruleA := range a {
		found := false
		for _, ruleB := range b {
			if rulesEqual(ruleA, ruleB) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func rulesEqual(a, b R2EventNotificationRuleModel) bool {
	return a.Prefix.ValueString() == b.Prefix.ValueString() &&
		a.Suffix.ValueString() == b.Suffix.ValueString() &&
		a.Actions.Equal(b.Actions)
}
