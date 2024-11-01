package vectorize

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/consts"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/utils"
)

func (r *VectorizeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: heredoc.Doc(`
			The [Vectorize](add link to doc) resource allows you to create and manage a Vectorize database.
		`),
		Version: 1,

		Attributes: map[string]schema.Attribute{
			consts.ZoneIDSchemaKey: schema.StringAttribute{
				MarkdownDescription: consts.ZoneIDSchemaDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"rules": schema.SetNestedBlock{
				MarkdownDescription: "List of Cloud Connector Rules",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"provider": schema.StringAttribute{
							Required: true,
							MarkdownDescription: fmt.Sprintf("Type of provider. %s", utils.RenderAvailableDocumentationValuesStringSlice([]string{
								"aws_s3",
								"cloudflare_r2",
								"azure_storage",
								"gcp_storage",
							})),
							Validators: []validator.String{
								stringvalidator.OneOf(
									"aws_s3",
									"cloudflare_r2",
									"azure_storage",
									"gcp_storage",
								),
							},
						},
						"description": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Brief summary of the Vectorize database and its intended use.",
						},
					},
					Blocks: map[string]schema.Block{
						"parameters": schema.SingleNestedBlock{
							MarkdownDescription: "Cloud Connector Rule Parameters",
							Attributes: map[string]schema.Attribute{
								"host": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Host parameter for cloud connector rule",
								},
							},
						},
					},
				},
			},
		},
	}
}
