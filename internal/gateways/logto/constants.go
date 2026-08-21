package logto

// OAuth2 form field names sent to Logto's token and device endpoints.
const (
	fieldClientID     = "client_id"
	fieldGrantType    = "grant_type"
	fieldRefreshToken = "refresh_token"
)

// OAuth2 grant_type values used by the M2M client-credentials grant and the
// device authorization flow (RFC 8628).
const (
	grantTypeClientCredentials = "client_credentials"
	grantTypeDeviceCode        = "urn:ietf:params:oauth:grant-type:device_code"
	grantTypeRefreshToken      = "refresh_token"
)
