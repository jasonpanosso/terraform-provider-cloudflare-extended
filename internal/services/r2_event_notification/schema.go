package r2_event_notification

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
)

var _ resource.ResourceWithConfigValidators = (*R2EventNotificationResource)(nil)

func ResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description:   "Identifier.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"bucket_name": schema.StringAttribute{
				Description:   "Name of the R2 Bucket for the event notification",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"queue_id": schema.StringAttribute{
				Description:   "Queue ID",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"queue_name": schema.StringAttribute{
				Description: "Name of the queue.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Brief summary of the event notifications and their intended use.",
			},
			"rules": schema.SetNestedAttribute{
				Description: "List of r2 event notification rules",
				CustomType:  customfield.NewNestedObjectSetType[R2EventNotificationRuleModel](ctx),
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"rule_id": schema.StringAttribute{
							Description: "Identifier.",
							Computed:    true,
						},
						"prefix": schema.StringAttribute{
							Description: "Notifications will be sent only for objects with this prefix.",
							Optional:    true,
							Computed:    true,
						},
						"suffix": schema.StringAttribute{
							Description: "Notifications will be sent only for objects with this suffix.",
							Optional:    true,
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the rule was created.",
							Computed:    true,
						},
						"actions": schema.SetAttribute{
							ElementType: types.StringType,
							CustomType:  customfield.NewSetType[types.String](ctx),
							Description: "Set of R2 object actions that will trigger notifications",
							Required:    true,
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(
									stringvalidator.OneOf([]string{"PutObject", "CopyObject", "DeleteObject", "CompleteMultipartUpload", "LifecycleDeletion"}...),
								),
							},
						},
					},
				},
			},
		},
	}
}

func (r *R2EventNotificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *R2EventNotificationResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}
