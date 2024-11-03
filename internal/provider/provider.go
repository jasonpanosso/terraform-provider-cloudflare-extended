package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/consts"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/services/queue_consumer"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/services/r2_event_notification"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/services/vectorize"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/services/workers_script"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/utils"
)

var _ provider.ProviderWithConfigValidators = &CloudflareExtendedProvider{}

// CloudflareExtendedProvider defines the provider implementation.
type CloudflareExtendedProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// CloudflareExtendedProviderModel describes the provider data model.
type CloudflareExtendedProviderModel struct {
	APIKey                  types.String `tfsdk:"api_key" json:"api_key"`
	APIUserServiceKey       types.String `tfsdk:"api_user_service_key" json:"api_user_service_key"`
	Email                   types.String `tfsdk:"email" json:"email"`
	APIToken                types.String `tfsdk:"api_token" json:"api_token"`
	UserAgentOperatorSuffix types.String `tfsdk:"user_agent_operator_suffix" json:"user_agent_operator_suffix"`
	BaseURL                 types.String `tfsdk:"base_url" json:"base_url"`
}

func (p *CloudflareExtendedProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cloudflare-extended"
	resp.Version = p.version
}

func (p *CloudflareExtendedProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			consts.EmailSchemaKey: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("A registered Cloudflare email address. Alternatively, can be configured using the `%s` environment variable. Required when using `api_key`. Conflicts with `api_token`.", consts.EmailEnvVarKey),
				Validators:          []validator.String{},
			},

			consts.APIKeySchemaKey: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("The API key for operations. Alternatively, can be configured using the `%s` environment variable. API keys are [now considered legacy by Cloudflare](https://developers.cloudflare.com/fundamentals/api/get-started/keys/#limitations), API tokens should be used instead. Must provide only one of `api_key`, `api_token`, `api_user_service_key`.", consts.APIKeyEnvVarKey),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`[0-9a-f]{37}`),
						"API key must be 37 characters long and only contain characters 0-9 and a-f (all lowercased)",
					),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot(consts.EmailSchemaKey),
					}...),
				},
			},

			consts.APITokenSchemaKey: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("The API Token for operations. Alternatively, can be configured using the `%s` environment variable. Must provide only one of `api_key`, `api_token`, `api_user_service_key`.", consts.APITokenEnvVarKey),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`[A-Za-z0-9-_]{40}`),
						"API tokens must be 40 characters long and only contain characters a-z, A-Z, 0-9, hyphens and underscores",
					),
				},
			},

			consts.APIUserServiceKeySchemaKey: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("A special Cloudflare API key good for a restricted set of endpoints. Alternatively, can be configured using the `%s` environment variable. Must provide only one of `api_key`, `api_token`, `api_user_service_key`.", consts.APIUserServiceKeyEnvVarKey),
			},

			consts.UserAgentOperatorSuffixSchemaKey: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("A value to append to the HTTP User Agent for all API calls. This value is not something most users need to modify however, if you are using a non-standard provider or operator configuration, this is recommended to assist in uniquely identifying your traffic. **Setting this value will remove the Terraform version from the HTTP User Agent string and may have unintended consequences**. Alternatively, can be configured using the `%s` environment variable.", consts.UserAgentOperatorSuffixEnvVarKey),
			},

			consts.BaseURLSchemaKey: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Value to override the default HTTP client base URL. Alternatively, can be configured using the `%s` environment variable.", consts.BaseURLSchemaKey),
			},
		},
	}
}

func (p *CloudflareExtendedProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data CloudflareExtendedProviderModel
	opts := []option.RequestOption{}

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if !data.BaseURL.IsNull() {
		opts = append(opts, option.WithBaseURL(data.BaseURL.ValueString()))
	}
	if !data.APIToken.IsNull() {
		opts = append(opts, option.WithAPIToken(data.APIToken.ValueString()))
	}
	if !data.APIKey.IsNull() {
		opts = append(opts, option.WithAPIKey(data.APIKey.ValueString()))
	}
	if !data.Email.IsNull() {
		opts = append(opts, option.WithAPIEmail(data.Email.ValueString()))
	}
	if !data.APIUserServiceKey.IsNull() {
		opts = append(opts, option.WithUserServiceKey(data.APIUserServiceKey.ValueString()))
	}

	pluginVersion := utils.FindGoModuleVersion("github.com/hashicorp/terraform-plugin-framework")
	if pluginVersion != nil {
		opts = append(opts, option.WithHeader("x-stainless-runtime-version", *pluginVersion))
	}

	framework := "terraform-plugin-framework"
	userAgentParams := utils.UserAgentBuilderParams{
		ProviderVersion: &p.version,
		PluginType:      &framework,
		PluginVersion:   pluginVersion,
	}

	if !data.UserAgentOperatorSuffix.IsNull() {
		operatorSuffix := data.UserAgentOperatorSuffix.String()
		userAgentParams.OperatorSuffix = &operatorSuffix
	} else {
		userAgentParams.TerraformVersion = &req.TerraformVersion
	}

	opts = append(opts, option.WithHeader("user-agent", userAgentParams.String()))

	client := cloudflare.NewClient(
		opts...,
	)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *CloudflareExtendedProvider) ConfigValidators(_ context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{}
}

func (p *CloudflareExtendedProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		vectorize.NewResource,
		workers_script.NewResource,
		queue_consumer.NewResource,
		r2_event_notification.NewResource,
	}
}

func (p *CloudflareExtendedProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *CloudflareExtendedProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CloudflareExtendedProvider{
			version: version,
		}
	}
}
