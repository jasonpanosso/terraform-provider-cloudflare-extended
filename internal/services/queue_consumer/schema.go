package queue_consumer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
)

var _ resource.ResourceWithConfigValidators = (*QueueConsumerResource)(nil)

func ResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description:   "Identifier.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"queue_id": schema.StringAttribute{
				Description:   "Identifier.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"consumer_id": schema.StringAttribute{
				Description: "Identifier.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: `Type of queue consumer. One of "worker", or "http_pull"`,
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("worker", "http_pull")},
			},
			"script_name": schema.StringAttribute{
				Optional: true,
			},
			"created_on": schema.StringAttribute{
				Computed: true,
			},
			"dead_letter_queue": schema.StringAttribute{
				Optional: true,
			},
			"environment": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"queue_name": schema.StringAttribute{
				Computed: true,
			},
			"settings": schema.SingleNestedAttribute{
				Computed:   true,
				Optional:   true,
				CustomType: customfield.NewNestedObjectType[QueueConsumerSettingsModel](ctx),
				Attributes: map[string]schema.Attribute{
					"batch_size": schema.Float64Attribute{
						Computed: true,
						Optional: true,
					},
					"max_retries": schema.Float64Attribute{
						Description: "The maximum number of retries",
						Computed:    true,
						Optional:    true,
					},
					"max_wait_time_ms": schema.Float64Attribute{
						Computed: true,
						Optional: true,
					},
				},
			},
		},
	}
}

func (r *QueueConsumerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *QueueConsumerResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}
