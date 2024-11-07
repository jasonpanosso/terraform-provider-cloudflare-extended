package vectorize

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
)

var _ resource.ResourceWithConfigValidators = (*VectorizeResource)(nil)

func ResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Vectorize database",
				Computed:    true,
			},
			"account_id": schema.StringAttribute{
				Description:   "Identifier.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Description:   "Name of the Vectorize Index.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^([a-z]+[a-z0-9_-]*[a-z0-9]+)$`),
						"Name must start with a lowercase letter, end with an alphanumeric character, and contain only lowercase letters, numbers, hyphens, and underscores.",
					),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Brief summary of the Vectorize database and its intended use.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
			},
			"dimensions": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Dimension of stored vectors",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"metric": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: `Distance metric to use for calculating vector similarity. One of "cosine", "dot-product", or "euclidean"`,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.OneOf("cosine", "dot-product", "euclidean"),
				},
			},
			"created_on": schema.StringAttribute{
				Computed: true,
			},
			"modified_on": schema.StringAttribute{
				Computed: true,
			},
			"metadata_indexes": schema.MapAttribute{
				Description: "Map of metadata index names to the attribute type",
				ElementType: types.StringType,
				CustomType:  customfield.NewMapType[basetypes.StringValue](ctx),
				Optional:    true,
				Validators: []validator.Map{
					mapvalidator.SizeAtMost(10),
					mapvalidator.ValueStringsAre(stringvalidator.OneOf([]string{"string", "number", "boolean"}...)),
				},
			},
		},
	}
}

func (r *VectorizeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *VectorizeResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}
