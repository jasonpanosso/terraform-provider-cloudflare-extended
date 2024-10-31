package provider

import (
	"context"
	"fmt"
	"math"
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
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/service/vectorize"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/utils"
)

// Ensure CloudflareExtendedProvider satisfies various provider interfaces.
var _ provider.Provider = &CloudflareExtendedProvider{}

// CloudflareExtendedProvider defines the provider implementation.
type CloudflareExtendedProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// CloudflareExtendedProviderModel describes the provider data model.
type CloudflareExtendedProviderModel struct {
	APIKey                  types.String `tfsdk:"api_key"`
	APIUserServiceKey       types.String `tfsdk:"api_user_service_key"`
	Email                   types.String `tfsdk:"email"`
	APIBasePath             types.String `tfsdk:"api_base_path"`
	APIToken                types.String `tfsdk:"api_token"`
	Retries                 types.Int64  `tfsdk:"retries"`
	APIClientLogging        types.Bool   `tfsdk:"api_client_logging"`
	APIHostname             types.String `tfsdk:"api_hostname"`
	UserAgentOperatorSuffix types.String `tfsdk:"user_agent_operator_suffix"`
}

func (p *CloudflareExtendedProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cloudflare-extended"
	resp.Version = p.version
}

func (p *CloudflareExtendedProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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

			consts.APIUserServiceKeySchemaKey: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("A special Cloudflare API key good for a restricted set of endpoints. Alternatively, can be configured using the `%s` environment variable. Must provide only one of `api_key`, `api_token`, `api_user_service_key`.", consts.APIUserServiceKeyEnvVarKey),
			},

			consts.RetriesSchemaKey: schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Maximum number of retries to perform when an API request fails. Alternatively, can be configured using the `%s` environment variable.", consts.RetriesEnvVarKey),
			},

			consts.APIClientLoggingSchemaKey: schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Whether to print logs from the API client (using the default log library logger). Alternatively, can be configured using the `%s` environment variable.", consts.APIClientLoggingEnvVarKey),
			},

			consts.APIHostnameSchemaKey: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Configure the hostname used by the API client. Alternatively, can be configured using the `%s` environment variable.", consts.APIHostnameEnvVarKey),
			},

			consts.APIBasePathSchemaKey: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Configure the base path used by the API client. Alternatively, can be configured using the `%s` environment variable.", consts.APIBasePathEnvVarKey),
			},

			consts.UserAgentOperatorSuffixSchemaKey: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("A value to append to the HTTP User Agent for all API calls. This value is not something most users need to modify however, if you are using a non-standard provider or operator configuration, this is recommended to assist in uniquely identifying your traffic. **Setting this value will remove the Terraform version from the HTTP User Agent string and may have unintended consequences**. Alternatively, can be configured using the `%s` environment variable.", consts.UserAgentOperatorSuffixEnvVarKey),
			},
		},
	}
}

func (p *CloudflareExtendedProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var (
		data CloudflareExtendedProviderModel

		email             string
		apiKey            string
		apiToken          string
		apiUserServiceKey string
		retries           int64
		baseHostname      string
		basePath          string
	)

	cloudflareOptions := make([]option.RequestOption, 0)

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.APIHostname.ValueString() != "" {
		baseHostname = data.APIHostname.ValueString()
	} else {
		baseHostname = utils.GetDefaultFromEnv(consts.APIHostnameEnvVarKey, consts.APIHostnameDefault)
	}

	if data.APIBasePath.ValueString() != "" {
		basePath = data.APIBasePath.ValueString()
	} else {
		basePath = utils.GetDefaultFromEnv(consts.APIBasePathEnvVarKey, consts.APIBasePathDefault)
	}
	cloudflareOptions = append(cloudflareOptions, option.WithBaseURL(fmt.Sprintf("https://%s%s", baseHostname, basePath)))

	if retries >= math.MaxInt32 {
		resp.Diagnostics.AddError(
			fmt.Sprintf("retries value of %d is too large, try a smaller value.", retries),
			fmt.Sprintf("retries value of %d is too large, try a smaller value.", retries),
		)
		return
	}

	cloudflareOptions = append(cloudflareOptions, option.WithMaxRetries(int(retries)))

	pluginVersion := utils.FindGoModuleVersion("github.com/hashicorp/terraform-plugin-framework")
	pluginType := "terraform-plugin-framework"
	userAgentParams := utils.UserAgentBuilderParams{
		ProviderVersion: &p.version,
		PluginType:      &pluginType,
		PluginVersion:   pluginVersion,
	}
	if !data.UserAgentOperatorSuffix.IsNull() {
		userAgentParams.OperatorSuffix = data.UserAgentOperatorSuffix.ValueStringPointer()
	} else {
		userAgentParams.TerraformVersion = &req.TerraformVersion
	}
	cloudflareOptions = append(cloudflareOptions, option.WithHeader("user-agent", userAgentParams.String()))

	if !data.APIToken.IsNull() {
		apiToken = data.APIToken.ValueString()
	} else {
		apiToken = utils.GetDefaultFromEnv(consts.APITokenEnvVarKey, "")
	}

	if apiToken != "" {
		cloudflareOptions = append(cloudflareOptions, option.WithAPIToken(apiToken))
	}

	if !data.APIKey.IsNull() {
		apiKey = data.APIKey.ValueString()
	} else {
		apiKey = utils.GetDefaultFromEnv(consts.APIKeyEnvVarKey, "")
	}

	if apiKey != "" {
		cloudflareOptions = append(cloudflareOptions, option.WithAPIKey(apiKey))

		if !data.Email.IsNull() {
			email = data.Email.ValueString()
		} else {
			email = utils.GetDefaultFromEnv(consts.EmailEnvVarKey, "")
		}

		if email == "" {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%q is not set correctly", consts.EmailSchemaKey),
				fmt.Sprintf("%q is required with %q and was not configured", consts.EmailSchemaKey, consts.APIKeySchemaKey))
			return
		}

		if email != "" {
			cloudflareOptions = append(cloudflareOptions, option.WithAPIEmail(email))
		}
	}

	if !data.APIUserServiceKey.IsNull() {
		apiUserServiceKey = data.APIUserServiceKey.ValueString()
	} else {
		apiUserServiceKey = utils.GetDefaultFromEnv(consts.APIUserServiceKeyEnvVarKey, "")
	}

	if apiUserServiceKey != "" {
		cloudflareOptions = append(cloudflareOptions, option.WithUserServiceKey(apiUserServiceKey))
	}

	if apiKey == "" && apiToken == "" && apiUserServiceKey == "" {
		resp.Diagnostics.AddError(
			fmt.Sprintf("must provide one of %q, %q or %q.", consts.APIKeySchemaKey, consts.APITokenSchemaKey, consts.APIUserServiceKeySchemaKey),
			fmt.Sprintf("must provide one of %q, %q or %q.", consts.APIKeySchemaKey, consts.APITokenSchemaKey, consts.APIUserServiceKeySchemaKey),
		)
		return
	}

	client := cloudflare.NewClient(cloudflareOptions...)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *CloudflareExtendedProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{vectorize.NewResource}
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
