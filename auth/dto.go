package auth

import "github.com/iautre/gowk/auth/db"

// User registration parameters
type RegisterParams struct {
	Phone  string `json:"phone" binding:"required,min=11,max=11"`
	Email  string `json:"email" binding:"required,email"`
	Name   string `json:"name" binding:"required,min=2,max=50"`
	Avatar string `json:"avatar,omitempty"`
}

// Login parameters
type LoginParams struct {
	Account string `json:"account" binding:"required"`    // Phone or Email
	Code    string `json:"code" binding:"required,len=6"` // OTP code
}

// Login response
type LoginRes struct {
	Token      string `json:"token"`
	UserId     int64  `json:"userId"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar,omitempty"`
	IsVerified bool   `json:"isVerified"`
}

// User response (excludes sensitive data)
type UserRes struct {
	Id          int64  `json:"id"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Nickname    string `json:"nickname"`
	Group       string `json:"group"`
	Avatar      string `json:"avatar,omitempty"`
	IsVerified  bool   `json:"isVerified"`
	Enabled     bool   `json:"enabled"`
	LastLoginAt string `json:"lastLoginAt,omitempty"`
	Created     string `json:"created"`
}

// User status update parameters
type UserStatusUpdateParams struct {
	UserId int64 `json:"userId" binding:"required"`
	Status int   `json:"status" binding:"required,oneof=0 1 2"` // disabled, active, suspended
}

// Password reset parameters
type PasswordResetParams struct {
	Account string `json:"account" binding:"required"`
	NewCode string `json:"newCode" binding:"required,len=6"`
}

// OAuth2 DTOs
type OAuth2AuthRequest struct {
	ResponseType string `json:"response_type" form:"response_type" binding:"required,oneof=code"`
	ClientID     string `json:"client_id" form:"client_id" binding:"required,min=1,max=100"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri" binding:"required,url"`
	Scope        string `json:"scope" form:"scope" binding:"required,min=1,max=500"`
	State        string `json:"state" form:"state" binding:"omitempty,max=100"`
	Nonce        string `json:"nonce" form:"nonce" binding:"omitempty,max=100"`
}

type OAuth2TokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type" binding:"required,oneof=authorization_code client_credentials refresh_token"`
	Code         string `json:"code" form:"code" binding:"omitempty,min=10,max=100"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri" binding:"omitempty,url"`
	ClientID     string `json:"client_id" form:"client_id" binding:"omitempty,min=1,max=100"`
	ClientSecret string `json:"client_secret" form:"client_secret" binding:"omitempty,min=16,max=128"`
	RefreshToken string `json:"refresh_token" form:"refresh_token" binding:"omitempty,min=10,max=500"`
	Scope        string `json:"scope" form:"scope" binding:"omitempty,min=1,max=500"`
	// OIDC specific fields
	CodeVerifier string `json:"code_verifier" form:"code_verifier" binding:"omitempty,min=43,max=128"` // PKCE
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

type OAuth2ClientResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Secret          string `json:"secret"`
	RedirectURIs    string `json:"redirect_uris"`
	Scopes          string `json:"scopes"`
	GrantTypes      string `json:"grant_types"`
	AccessTokenTTL  int64  `json:"access_token_ttl"`
	RefreshTokenTTL int64  `json:"refresh_token_ttl"`
	Created         string `json:"created"`
	Updated         string `json:"updated"`
	Enabled         bool   `json:"enabled"` // true = active, false = disabled
}

func BuildOAuth2ClientResponse(client db.Oauth2Client) *OAuth2ClientResponse {
	return &OAuth2ClientResponse{
		ID:              client.ID,
		Name:            client.Name,
		Secret:          client.Secret,
		RedirectURIs:    client.RedirectUris,
		Scopes:          client.Scopes,
		GrantTypes:      client.GrantTypes,
		AccessTokenTTL:  client.AccessTokenTtl,
		RefreshTokenTTL: client.RefreshTokenTtl,
		Enabled:         client.Enabled,
	}
}

// OAuth Client Management DTOs
type OAuth2ClientCreateParams struct {
	Name            string   `json:"name" binding:"required,min=2,max=100"`
	RedirectURIs    []string `json:"redirect_uris" binding:"required,min=1"`
	Scopes          []string `json:"scopes" binding:"required,min=1"`
	GrantTypes      []string `json:"grant_types" binding:"required,min=1"`
	AccessTokenTTL  int64    `json:"access_token_ttl" binding:"min=300"`   // min 5 minutes
	RefreshTokenTTL int64    `json:"refresh_token_ttl" binding:"min=3600"` // min 1 hour
}

type OAuth2ClientUpdateParams struct {
	ID              string   `json:"id" binding:"required"`
	Name            string   `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Secret          string   `json:"secret,omitempty" binding:"omitempty,min=16,max=128"`
	RedirectURIs    []string `json:"redirect_uris,omitempty" binding:"omitempty,min=1"`
	Scopes          []string `json:"scopes,omitempty" binding:"omitempty,min=1"`
	Enable          bool     `json:"enable,omitempty"`
	GrantTypes      []string `json:"grant_types,omitempty" binding:"omitempty,min=1"`
	AccessTokenTTL  int64    `json:"access_token_ttl,omitempty" binding:"omitempty,min=300"`
	RefreshTokenTTL int64    `json:"refresh_token_ttl,omitempty" binding:"omitempty,min=3600"`
}

// SSO DTOs
type SSOLoginRequest struct {
	Provider    string `json:"provider" form:"provider" binding:"required,oneof=google github wechat"`
	Token       string `json:"token" form:"token" binding:"required,min=10,max=1000"`
	RedirectURI string `json:"redirect_uri" form:"redirect_uri" binding:"required,url"`
	State       string `json:"state" form:"state" binding:"omitempty,max=100"`
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
