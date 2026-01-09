package auth

type RegisterParams struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
type LoginParams struct {
	Account string `json:"account"`
	Code    string `json:"code"`
}

type LoginRes struct {
	Token    string `json:"token"`
	UserId   int64  `json:"userId"`
	Nickname string `json:"nickname"`
}
type UserRes struct {
	Id       int64  `json:"id"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Group    string `json:"group"`
}

// OAuth2 DTOs
type OAuth2AuthRequest struct {
	ResponseType string `json:"response_type" form:"response_type"`
	ClientID     string `json:"client_id" form:"client_id"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri"`
	Scope        string `json:"scope" form:"scope"`
	State        string `json:"state" form:"state"`
	Nonce        string `json:"nonce" form:"nonce"`
}

type OAuth2TokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type"`
	Code         string `json:"code" form:"code"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri"`
	ClientID     string `json:"client_id" form:"client_id"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
	Scope        string `json:"scope" form:"scope"`
	// OIDC specific fields
	CodeVerifier string `json:"code_verifier" form:"code_verifier"` // PKCE
}

type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	// OIDC specific fields
	IDToken string `json:"id_token,omitempty"` // OpenID Connect ID Token
}

type OAuth2Client struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Secret       string `json:"secret"`
	RedirectURIs string `json:"redirect_uris"`
	Scopes       string `json:"scopes"`
}

// SSO DTOs
type SSOLoginRequest struct {
	Provider    string `json:"provider" form:"provider"`
	Token       string `json:"token" form:"token"`
	RedirectURI string `json:"redirect_uri" form:"redirect_uri"`
	State       string `json:"state" form:"state"`
}

type SSOLoginResponse struct {
	Token    string `json:"token"`
	UserId   int64  `json:"userId"`
	Nickname string `json:"nickname"`
	Provider string `json:"provider"`
}

// OpenID Connect (OIDC) DTOs
type OIDCDiscoveryResponse struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	UserInfoEndpoint                 string   `json:"userinfo_endpoint"`
	JwksUri                          string   `json:"jwks_uri"`
	ScopesSupported                  []string `json:"scopes_supported"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	GrantTypesSupported              []string `json:"grant_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

type OIDCAddress struct {
	Formatted     string `json:"formatted,omitempty"`
	StreetAddress string `json:"street_address,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
}

type OIDCUserInfo struct {
	Sub               string       `json:"sub"`
	Name              string       `json:"name,omitempty"`
	GivenName         string       `json:"given_name,omitempty"`
	FamilyName        string       `json:"family_name,omitempty"`
	MiddleName        string       `json:"middle_name,omitempty"`
	Nickname          string       `json:"nickname,omitempty"`
	PreferredUsername string       `json:"preferred_username"`
	Profile           string       `json:"profile,omitempty"`
	Picture           string       `json:"picture,omitempty"`
	Website           string       `json:"website,omitempty"`
	Email             string       `json:"email,omitempty"`
	EmailVerified     bool         `json:"email_verified,omitempty"`
	Gender            string       `json:"gender,omitempty"`
	Birthdate         string       `json:"birthdate,omitempty"`
	ZoneInfo          string       `json:"zoneinfo,omitempty"`
	Locale            string       `json:"locale,omitempty"`
	PhoneNumber       string       `json:"phone_number,omitempty"`
	PhoneVerified     bool         `json:"phone_number_verified,omitempty"`
	Address           *OIDCAddress `json:"address,omitempty"`
	UpdatedAt         int64        `json:"updated_at,omitempty"`
}

type OIDCIDToken struct {
	Iss   string `json:"iss"`
	Sub   string `json:"sub"`
	Aud   string `json:"aud"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
	Nonce string `json:"nonce,omitempty"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type OIDCJwksResponse struct {
	Keys []OIDCJwk `json:"keys"`
}

type OIDCJwk struct {
	Kty string `json:"kty"` // Key Type (e.g., "RSA")
	Use string `json:"use"` // Public Key Use (e.g., "sig")
	Kid string `json:"kid"` // Key ID
	N   string `json:"n"`   // Modulus for RSA keys
	E   string `json:"e"`   // Exponent for RSA keys
	Alg string `json:"alg"` // Algorithm (e.g., "RS256")
}
