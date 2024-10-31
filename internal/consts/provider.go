package consts

const (
	// Schema key for the API token configuration.
	APITokenSchemaKey = "api_token"

	// Environment variable key for the API token configuration.
	APITokenEnvVarKey = "CLOUDFLARE_API_TOKEN"

	// Schema key for the API key configuration.
	APIKeySchemaKey = "api_key"

	// Environment variable key for the API key configuration.
	APIKeyEnvVarKey = "CLOUDFLARE_API_KEY"

	// Schema key for the email configuration.
	EmailSchemaKey = "email"

	// Environment variable key for the email configuration.
	EmailEnvVarKey = "CLOUDFLARE_EMAIL"

	// Schema key for the API user service key configuration.
	APIUserServiceKeySchemaKey = "api_user_service_key"

	// Environment variable key for the API user service key configuration.
	APIUserServiceKeyEnvVarKey = "CLOUDFLARE_API_USER_SERVICE_KEY"

	// Schema key for the API hostname configuration.
	APIHostnameSchemaKey = "api_hostname"

	// Environment variable key for the API hostname configuration.
	APIHostnameEnvVarKey = "CLOUDFLARE_API_HOSTNAME"

	// Default value for the API hostname.
	APIHostnameDefault = "api.cloudflare.com"

	// Schema key for the API base path configuration.
	APIBasePathSchemaKey = "api_base_path"

	// Environment variable key for the API base path configuration.
	APIBasePathEnvVarKey = "CLOUDFLARE_API_BASE_PATH"

	// Default value for the API base path.
	APIBasePathDefault = "/client/v4"

	// Schema key for the User Agent operator suffix.
	UserAgentOperatorSuffixSchemaKey = "user_agent_operator_suffix"

	// Environment variable key for the User Agent operator suffix.
	UserAgentOperatorSuffixEnvVarKey = "CLOUDFLARE_USER_AGENT_OPERATOR_SUFFIX"

	// Schema key for the retries configuration.
	RetriesSchemaKey = "retries"

	// Environment variable key for the retries configuration.
	RetriesEnvVarKey = "CLOUDFLARE_RETRIES"

	// Default value for the retries.
	RetriesDefault = "4"

	// Schema key for the API Client Logging Schema.
	APIClientLoggingSchemaKey = "api_client_logging"

	// Environment variable key for the API client logging configuration.
	APIClientLoggingEnvVarKey = "CLOUDFLARE_API_CLIENT_LOGGING"

	// Schema key for the zone ID configuration.
	ZoneIDSchemaKey = "zone_id"

	// Schema description for `zone_id` field.
	ZoneIDSchemaDescription = "The zone identifier to target for the resource."

	// Schema key for IDs.
	IDSchemaKey = "id"

	// Schema description for all ID fields.
	IDSchemaDescription = "The identifier of this resource."
)
