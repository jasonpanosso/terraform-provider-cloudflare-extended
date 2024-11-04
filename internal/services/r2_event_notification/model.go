package r2_event_notification

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
)

type R2EventNotificationModel struct {
	AccountID   types.String                                              `tfsdk:"account_id" path:"account_id,required"`
	BucketName  types.String                                              `tfsdk:"bucket_name" path:"bucket_name,required"`
	QueueID     types.String                                              `tfsdk:"queue_id" path:"queue_id,required"`
	QueueName   types.String                                              `tfsdk:"queue_name" path:"queue_name,computed"`
	Description types.String                                              `tfsdk:"description"  path:"description,optional"`
	Rules       customfield.NestedObjectSet[R2EventNotificationRuleModel] `tfsdk:"rules" path:"rules,required"`
	Timeouts    timeouts.Value                                            `tfsdk:"timeouts"`
}

type R2EventNotificationRuleModel struct {
	Actions   customfield.Set[types.String] `tfsdk:"actions" path:"actions,required"`
	RuleID    types.String                  `tfsdk:"rule_id" path:"rule_id,computed"`
	Prefix    types.String                  `tfsdk:"prefix" path:"prefix,computed_optional"`
	Suffix    types.String                  `tfsdk:"suffix" path:"suffix,computed_optional"`
	CreatedAt types.String                  `tfsdk:"created_at" path:"created_at,computed"`
}

type R2EventNotificationDeleteRequestBody struct {
	RuleIds []string `json:"ruleIds"`
}
