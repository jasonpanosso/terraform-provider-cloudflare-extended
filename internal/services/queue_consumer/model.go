package queue_consumer

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/apijson"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
)

type QueueConsumerResultEnvelope struct {
	Result QueueConsumerModel `json:"result"`
}

type QueueConsumerModel struct {
	AccountID       types.String                                         `tfsdk:"account_id" path:"account_id,required"`
	QueueID         types.String                                         `tfsdk:"queue_id" path:"queue_id,required"`
	ScriptName      types.String                                         `tfsdk:"script_name" json:"script_name,optional"`
	ConsumerID      types.String                                         `tfsdk:"consumer_id" json:"consumer_id,computed"`
	CreatedOn       types.String                                         `tfsdk:"created_on" json:"created_on,computed"`
	DeadLetterQueue types.String                                         `tfsdk:"dead_letter_queue" json:"dead_letter_queue,optional"`
	Environment     types.String                                         `tfsdk:"environment" json:"environment,computed_optional"`
	QueueName       types.String                                         `tfsdk:"queue_name" json:"queue_name,computed"`
	Settings        customfield.NestedObject[QueueConsumerSettingsModel] `tfsdk:"settings" json:"settings,computed_optional"`
	Type            types.String                                         `tfsdk:"type" json:"type,required"`
}

func (m QueueConsumerModel) MarshalJSON() (data []byte, err error) {
	return apijson.MarshalRoot(m)
}

func (m QueueConsumerModel) MarshalJSONForUpdate(state QueueConsumerModel) (data []byte, err error) {
	return apijson.MarshalForUpdate(m, state)
}

type QueueConsumerSettingsModel struct {
	BatchSize     types.Float64 `tfsdk:"batch_size" json:"batch_size,computed_optional"`
	MaxRetries    types.Float64 `tfsdk:"max_retries" json:"max_retries,computed_optional"`
	MaxWaitTimeMs types.Float64 `tfsdk:"max_wait_time_ms" json:"max_wait_time_ms,computed_optional"`
}
