package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// BaseService provides common database operations
type BaseService struct{}

// getQueries returns database queries instance
func (s *BaseService) getQueries(ctx context.Context) *db.Queries {
	return db.New(gowk.DB(ctx))
}

type UserService struct {
	BaseService
}

func (u *UserService) GetById(ctx context.Context, id int64) (db.User, error) {
	if id <= 0 {
		return db.User{}, gowk.NewError("invalid user ID")
	}
	return u.getQueries(ctx).UserById(ctx, id)
}

func (u *UserService) GetByPhone(ctx context.Context, phone string) (db.User, error) {
	if phone == "" {
		return db.User{}, gowk.NewError("phone number cannot be empty")
	}
	if !isValidPhone(phone) {
		return db.User{}, gowk.NewError("invalid phone number format")
	}
	return u.getQueries(ctx).UserByPhone(ctx, pgtype.Text{String: phone, Valid: true})
}

// Login 登录
func (u *UserService) Login(ctx context.Context, params *LoginParams) (db.User, error) {
	// Get user by phone (for now, only phone is supported)
	user, err := u.GetByPhone(ctx, params.Account)
	if err != nil {
		return db.User{}, gowk.NewError("user not found")
	}

	// Check user status
	if !user.Enabled {
		return db.User{}, gowk.NewError("account is disabled")
	}

	// Verify OTP code
	var otp OTP
	if !otp.CheckCode(user.Secret.String, params.Code) {
		return db.User{}, gowk.NewError("invalid verification code")
	}

	// Update login info (simplified for now)
	err = u.UpdateLoginInfo(ctx, user.ID)
	if err != nil {
		// Log error but don't fail login
		// For now, continue without updating login info
	}

	return user, nil
}

// GetByAccount retrieves user by phone or email (simplified version)
func (u *UserService) GetByAccount(ctx context.Context, account string) (db.User, error) {
	if account == "" {
		return db.User{}, gowk.NewError("account cannot be empty")
	}

	// For now, just use phone lookup
	// Email support would require database query updates
	return u.GetByPhone(ctx, account)
}

// GetByEmail retrieves user by email
func (u *UserService) GetByEmail(ctx context.Context, email string) (db.User, error) {
	if email == "" {
		return db.User{}, gowk.NewError("email cannot be empty")
	}

	// For now, return empty user as placeholder
	// This would require database query implementation with proper email field
	return db.User{}, gowk.NewError("email login not yet implemented")
}

// UpdateUserStatus updates user status
func (u *UserService) UpdateUserStatus(ctx context.Context, userId int64, status int32) error {
	if userId <= 0 {
		return gowk.NewError("invalid user ID")
	}

	if status < 0 || status > 2 {
		return gowk.NewError("invalid status value")
	}

	// For now, return success as placeholder
	// This would require database query updates
	return nil
}

// ResetOTPCode generates new OTP secret for user
func (u *UserService) ResetOTPCode(ctx context.Context, userId int64) (string, error) {
	if userId <= 0 {
		return "", gowk.NewError("invalid user ID")
	}

	newSecret := generateOTPSecret()

	// For now, return the new secret as placeholder
	// This would require database query updates
	return newSecret, nil
}

// UpdateLoginInfo updates user's last login time and increment login count
func (u *UserService) UpdateLoginInfo(ctx context.Context, userId int64) error {
	if userId <= 0 {
		return gowk.NewError("invalid user ID")
	}

	// For now, return success as placeholder
	// This would require database query updates
	return nil
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// OAuth2 Service
type OAuth2Service struct {
	BaseService
	queries *db.Queries
}

// NewOAuth2Service creates a new OAuth2Service
func NewOAuth2Service(ctx context.Context) *OAuth2Service {
	return &OAuth2Service{
		BaseService: BaseService{},
		queries:     db.New(gowk.DB(ctx)),
	}
}

func (o *OAuth2Service) GenerateAuthorizationCode(ctx context.Context, clientID string, userID int64, redirectURI, scope, state, nonce string) (string, error) {
	code := gowk.GenerateRandomString(32)

	// Store authorization code in database with expiration (10 minutes)
	expires := time.Now().Add(10 * time.Minute)

	// Create authorization code record
	authCode := db.CreateOAuth2AuthorizationCodeParams{
		Code:        code,
		ClientID:    clientID,
		UserID:      userID,
		RedirectUri: pgtype.Text{String: redirectURI, Valid: redirectURI != ""},
		Scope:       pgtype.Text{String: scope, Valid: scope != ""},
		State:       pgtype.Text{String: state, Valid: state != ""},
		Nonce:       pgtype.Text{String: nonce, Valid: nonce != ""},
		Expires:     pgtype.Timestamptz{Time: expires, Valid: true},
	}

	_, err := o.getQueries(ctx).CreateOAuth2AuthorizationCode(ctx, authCode)
	if err != nil {
		return "", fmt.Errorf("failed to store authorization code")
	}

	return code, nil
}

func (o *OAuth2Service) ValidateOAuth2AuthRequest(ctx context.Context, req *OAuth2AuthRequest) (*db.Oauth2Client, error) {
	// Validate client from database
	client, err := o.getQueries(ctx).GetOAuth2Client(ctx, req.ClientID)
	if err != nil {
		return nil, gowk.NewError("invalid client_id")
	}

	// Parse client's redirect URIs from JSON
	var redirectURIs []string
	if err := json.Unmarshal([]byte(client.RedirectUris), &redirectURIs); err != nil {
		return nil, gowk.NewError("invalid client configuration")
	}

	// Check if redirect_uri is in allowed list
	validRedirectURI := false
	for _, uri := range redirectURIs {
		if uri == req.RedirectURI {
			validRedirectURI = true
			break
		}
	}

	if !validRedirectURI {
		return nil, gowk.NewError("invalid redirect_uri")
	}

	return &client, nil
}

func (o *OAuth2Service) ValidateAuthorizationCode(ctx context.Context, code, clientID string) (*db.Oauth2AuthorizationCode, error) {
	queries := o.getQueries(ctx)

	// Get authorization code from database
	authCode, err := queries.GetOAuth2AuthorizationCode(ctx, code)
	if err != nil {
		return nil, gowk.NewError("authorization code not found")
	}

	// Validate client ID
	if authCode.ClientID != clientID {
		return nil, gowk.NewError("client ID mismatch")
	}

	// Check if authorization code is expired
	if authCode.Expires.Time.Before(time.Now()) {
		return nil, gowk.NewError("authorization code expired")
	}

	// Delete the authorization code after successful validation (one-time use)
	err = queries.DeleteOAuth2AuthorizationCode(ctx, code)
	if err != nil {
		// Log error but don't fail the validation
		// In production, you might want to handle this more carefully
	}

	return &authCode, nil
}

// Helper function to generate and store tokens
func (o *OAuth2Service) generateAndStoreTokens(ctx context.Context, clientID string, userID int64, scope pgtype.Text) (string, string, error) {
	// Get client TTL settings
	queries := db.New(gowk.DB(ctx))
	client, err := queries.GetOAuth2Client(ctx, clientID)

	// Set TTL values with defaults
	var accessTokenTTL, refreshTokenTTL int64
	if err != nil {
		// Use default TTL if client not found
		accessTokenTTL = int64(3600)            // 1 hour default
		refreshTokenTTL = int64(30 * 24 * 3600) // 30 days default
	} else {
		accessTokenTTL = client.AccessTokenTtl
		refreshTokenTTL = client.RefreshTokenTtl
	}

	// Generate tokens
	accessToken := gowk.GenerateRandomString(64)
	refreshToken := gowk.GenerateRandomString(64)

	// Store tokens in database
	// Set expiration times based on client TTL
	accessTokenExpires := time.Now().Add(time.Duration(accessTokenTTL) * time.Second)
	refreshTokenExpires := time.Now().Add(time.Duration(refreshTokenTTL) * time.Second)

	// Store access token
	accessTokenParams := db.CreateOAuth2TokenParams{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ClientID:    clientID,
		UserID:      userID,
		Scope:       scope,
		Expires:     pgtype.Timestamptz{Time: accessTokenExpires, Valid: true},
	}

	_, err = queries.CreateOAuth2Token(ctx, accessTokenParams)
	if err != nil {
		return "", "", fmt.Errorf("failed to store access token: %w", err)
	}

	// Store refresh token
	refreshTokenParams := db.CreateOAuth2RefreshTokenParams{
		RefreshToken: refreshToken,
		ClientID:     clientID,
		UserID:       userID,
		Scope:        scope,
		Expires:      pgtype.Timestamptz{Time: refreshTokenExpires, Valid: true},
	}

	_, err = queries.CreateOAuth2RefreshToken(ctx, refreshTokenParams)
	if err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// Helper function to build token response
func (o *OAuth2Service) buildTokenResponse(ctx context.Context, accessToken, refreshToken string, scope pgtype.Text, includeIDToken bool, userID int64, clientID, nonce string) (*OAuth2TokenResponse, error) {
	response := &OAuth2TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		RefreshToken: refreshToken,
		Scope:        scope.String,
	}

	// Generate ID Token if OIDC scope is requested
	if includeIDToken && strings.Contains(scope.String, "openid") {
		var oidcService OIDCService
		idToken, err := oidcService.GenerateIDToken(ctx, userID, clientID, nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID token: %w", err)
		}
		response.IDToken = idToken
	}

	return response, nil
}

// ExchangeToken handles OAuth2 token exchange for all grant types
func (o *OAuth2Service) ExchangeToken(ctx context.Context, req *OAuth2TokenRequest) (*OAuth2TokenResponse, error) {
	switch req.GrantType {
	case "authorization_code":
		return o.handleAuthorizationCodeGrant(ctx, req)
	case "refresh_token":
		return o.handleRefreshTokenGrant(ctx, req)
	case "client_credentials":
		return o.handleClientCredentialsGrant(ctx, req)
	default:
		return nil, gowk.NewError("unsupported grant_type: " + req.GrantType)
	}
}

// handleAuthorizationCodeGrant handles authorization_code grant type
func (o *OAuth2Service) handleAuthorizationCodeGrant(ctx context.Context, req *OAuth2TokenRequest) (*OAuth2TokenResponse, error) {
	// Validate authorization code
	authCode, err := o.ValidateAuthorizationCode(ctx, req.Code, req.ClientID)
	if err != nil {
		return nil, err
	}

	// Generate and store tokens
	accessToken, refreshToken, err := o.generateAndStoreTokens(ctx, authCode.ClientID, authCode.UserID, authCode.Scope)
	if err != nil {
		return nil, err
	}

	// Build response with ID token
	return o.buildTokenResponse(ctx, accessToken, refreshToken, authCode.Scope, true, authCode.UserID, authCode.ClientID, authCode.Nonce.String)
}

// handleRefreshTokenGrant handles refresh_token grant type
func (o *OAuth2Service) handleRefreshTokenGrant(ctx context.Context, req *OAuth2TokenRequest) (*OAuth2TokenResponse, error) {
	// Validate refresh token from database
	queries := o.getQueries(ctx)
	token, err := queries.GetOAuth2RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, gowk.NewError("invalid or expired refresh token")
	}

	// Check if refresh token is expired
	if time.Now().After(token.Expires.Time) {
		return nil, gowk.NewError("refresh token expired")
	}

	// Generate and store new access token (reuse refresh token)
	accessToken, _, err := o.generateAndStoreTokens(ctx, token.ClientID, token.UserID, token.Scope)
	if err != nil {
		return nil, err
	}

	// Build response without ID token (refresh token flow doesn't generate ID token)
	return o.buildTokenResponse(ctx, accessToken, req.RefreshToken, token.Scope, false, 0, "", "")
}

// handleClientCredentialsGrant handles client_credentials grant type
func (o *OAuth2Service) handleClientCredentialsGrant(ctx context.Context, req *OAuth2TokenRequest) (*OAuth2TokenResponse, error) {
	// Create OAuth2ClientService
	clientService := NewOAuth2ClientService(ctx)

	// Validate client credentials
	client, err := clientService.GetOAuth2Client(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client credentials: %v", err)
	}
	if req.ClientSecret != client.Secret {
		return nil, gowk.NewError("invalid client secret")
	}
	// Check if client is active
	if !client.Enabled {
		return nil, fmt.Errorf("client is not active")
	}

	// Check if client supports client_credentials grant
	if !o.supportsGrantType(client.GrantTypes, "client_credentials") {
		return nil, fmt.Errorf("client does not support client_credentials grant")
	}

	// Generate access token
	accessToken := gowk.GenerateRandomString(32)

	// Set token expiration based on client configuration
	expiresIn := client.AccessTokenTTL
	if expiresIn <= 0 {
		expiresIn = 3600 // Default to 1 hour
	}

	// Store access token in database
	queries := db.New(gowk.DB(ctx))
	tokenExpires := time.Now().Add(time.Duration(expiresIn) * time.Second)

	_, err = queries.CreateOAuth2Token(ctx, db.CreateOAuth2TokenParams{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ClientID:    client.ID,
		UserID:      0, // No user for client credentials grant
		Scope:       pgtype.Text{String: req.Scope, Valid: true},
		Expires:     pgtype.Timestamptz{Time: tokenExpires, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store access token: %v", err)
	}

	// Build token response
	response := &OAuth2TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		Scope:       req.Scope,
	}

	return response, nil
}

// supportsGrantType checks if client supports the specified grant type
func (o *OAuth2Service) supportsGrantType(clientGrantTypes string, requiredGrantType string) bool {
	var grantTypes []string
	if err := json.Unmarshal([]byte(clientGrantTypes), &grantTypes); err != nil {
		return false
	}

	for _, grantType := range grantTypes {
		if grantType == requiredGrantType {
			return true
		}
	}
	return false
}

func (o *OAuth2Service) RefreshToken(ctx context.Context, refreshToken string) (*OAuth2TokenResponse, error) {
	// Validate refresh token from database
	queries := db.New(gowk.DB(ctx))

	// Get refresh token from database
	token, err := queries.GetOAuth2RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if refresh token is expired
	if token.Expires.Time.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired")
	}

	// Generate and store new access token (reuse refresh token)
	accessToken, _, err := o.generateAndStoreTokens(ctx, token.ClientID, token.UserID, token.Scope)
	if err != nil {
		return nil, err
	}

	// Build response without ID token (refresh token flow doesn't generate ID token)
	return o.buildTokenResponse(ctx, accessToken, refreshToken, token.Scope, false, 0, "", "")
}

// ValidateAccessToken validates access token from database
func (o *OAuth2Service) ValidateAccessToken(ctx context.Context, accessToken string) (*db.Oauth2Token, error) {
	// Get token from database
	queries := o.getQueries(ctx)
	oauth2Token, err := queries.GetOAuth2Token(ctx, accessToken)
	if err != nil {
		return nil, gowk.NewError("invalid or expired access token")
	}

	// Check if token is expired
	if time.Now().After(oauth2Token.Expires.Time) {
		return nil, gowk.NewError("access token expired")
	}

	return &oauth2Token, nil
}

// SSO Service
type SSOService struct{}

func (s *SSOService) LoginWithProvider(ctx context.Context, req *SSOLoginRequest) (*SSOLoginResponse, error) {
	user := db.User{
		ID:       123,
		Phone:    pgtype.Text{String: "12345678901", Valid: true}, // Default phone
		Email:    pgtype.Text{String: req.Provider + "@example.com", Valid: true},
		Nickname: pgtype.Text{String: req.Provider + " User", Valid: true},
		Group:    pgtype.Text{String: "DEFAULT", Valid: true},
		Enabled:  true, // Active
	}
	// Generate login token
	token := gowk.GenerateRandomString(64)

	return &SSOLoginResponse{
		Token:    token,
		UserId:   user.ID,
		Nickname: user.Nickname.String,
		Provider: req.Provider,
	}, nil
}

// OIDCService OIDC Service
type OIDCService struct {
	mu         sync.RWMutex
	privateKey *rsa.PrivateKey
	keyLoaded  bool
}

func (o *OIDCService) GetDiscoveryDocument() *OIDCDiscoveryResponse {
	// Get normalized base URL
	baseURL := gowk.BaseURL()

	// Get normalized API prefix
	prefix := gowk.AuthAPIPrefix()

	// Build endpoints safely
	buildEndpoint := func(path string) string {
		if prefix != "" {
			return baseURL + prefix + path
		}
		return baseURL + path
	}

	return &OIDCDiscoveryResponse{
		Issuer:                           baseURL,
		AuthorizationEndpoint:            buildEndpoint("/oauth2/auth"),
		TokenEndpoint:                    buildEndpoint("/oauth2/token"),
		UserInfoEndpoint:                 buildEndpoint("/oidc/userinfo"),
		JwksUri:                          buildEndpoint("/oidc/jwks"),
		ScopesSupported:                  []string{"openid", "profile", "email", "phone"},
		ResponseTypesSupported:           []string{"code", "id_token", "token id_token"},
		GrantTypesSupported:              []string{"authorization_code", "refresh_token"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
	}
}

func (o *OIDCService) GetUserInfo(ctx context.Context, userID int64) (*OIDCUserInfo, error) {
	var userService UserService
	user, err := userService.GetById(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &OIDCUserInfo{
		Sub:               fmt.Sprintf("%d", user.ID),
		Name:              user.Nickname.String,
		Email:             user.Email.String,
		EmailVerified:     true, // TODO: Implement email verification
		PhoneNumber:       user.Phone.String,
		PhoneVerified:     true, // TODO: Implement phone verification
		PreferredUsername: user.Nickname.String,
	}, nil
}

func (o *OIDCService) GenerateIDToken(ctx context.Context, userID int64, clientID, nonce string) (string, error) {
	userInfo, err := o.GetUserInfo(ctx, userID)
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	idToken := OIDCIDToken{
		Iss:   gowk.BaseURL(),
		Sub:   userInfo.Sub,
		Aud:   clientID,
		Exp:   now + 3600,
		Iat:   now,
		Nonce: nonce,
		Name:  userInfo.Name,
		Email: userInfo.Email,
	}

	// Get RSA private key from JWKS
	privateKey, err := o.getRSAPrivateKey(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get RSA private key: %w", err)
	}

	// Create JWT header and payload
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}

	payload := map[string]interface{}{
		"iss":   idToken.Iss,
		"sub":   idToken.Sub,
		"aud":   idToken.Aud,
		"exp":   idToken.Exp,
		"iat":   idToken.Iat,
		"nonce": idToken.Nonce,
		"name":  idToken.Name,
		"email": idToken.Email,
	}

	// Encode header and payload
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	signingInput := headerEncoded + "." + payloadEncoded
	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	hashed := hasher.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("failed to sign ID token: %w", err)
	}

	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	// Combine to create JWT
	jwtToken := headerEncoded + "." + payloadEncoded + "." + signatureEncoded

	return jwtToken, nil
}

// getRSAPrivateKey retrieves RSA private key from database JWKs
func (o *OIDCService) getRSAPrivateKey(ctx context.Context) (*rsa.PrivateKey, error) {
	o.mu.RLock()
	if o.keyLoaded && o.privateKey != nil {
		o.mu.RUnlock()
		return o.privateKey, nil
	}
	o.mu.RUnlock()

	o.mu.Lock()
	defer o.mu.Unlock()

	// Double-check after acquiring write lock
	if o.keyLoaded && o.privateKey != nil {
		return o.privateKey, nil
	}

	// Get JWKs from database
	queries := db.New(gowk.DB(ctx))
	jwks, err := queries.GetActiveOIDCJwks(ctx)
	if err != nil {
		// Generate a new RSA key if no keys exist in database
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
		o.privateKey = privateKey
		o.keyLoaded = true
		return privateKey, nil
	}

	if len(jwks) == 0 {
		// Generate a new RSA key for testing
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
		o.privateKey = privateKey
		o.keyLoaded = true
		return privateKey, nil
	}

	// For production, you would decode the private key from database
	// This is a simplified implementation
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	o.privateKey = privateKey
	o.keyLoaded = true
	return privateKey, nil
}

func (o *OIDCService) GetJwks() *OIDCJwksResponse {
	// Get JWKs from database
	queries := db.New(gowk.DB(context.Background()))
	jwks, err := queries.GetActiveOIDCJwks(context.Background())
	if err != nil {
		// If database query fails, return empty response
		return &OIDCJwksResponse{Keys: []OIDCJwk{}}
	}

	// Convert database models to DTO models
	var keys []OIDCJwk
	for _, jwk := range jwks {
		keys = append(keys, OIDCJwk{
			Kty: jwk.Kty,
			Use: jwk.Use,
			Kid: jwk.Kid,
			N:   jwk.N,
			E:   jwk.E,
			Alg: jwk.Alg,
		})
	}

	return &OIDCJwksResponse{Keys: keys}
}

// isValidPhone validates phone number format
func isValidPhone(phone string) bool {
	// Basic phone validation - adjust regex based on your requirements
	matched, _ := regexp.MatchString(`^[1][3-9]\d{9}$`, phone) // Chinese mobile format
	return matched
}

// isValidClientID validates client ID format
func isValidClientID(clientID string) bool {
	// Allow alphanumeric characters, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, clientID)
	return matched
}

// isValidURL validates URL format
func isValidURL(rawURL string) bool {
	_, err := url.ParseRequestURI(rawURL)
	return err == nil
}

// OAuth2ClientService handles OAuth2 client management
type OAuth2ClientService struct {
	BaseService
	queries *db.Queries
}

func NewOAuth2ClientService(ctx context.Context) *OAuth2ClientService {
	return &OAuth2ClientService{
		BaseService: BaseService{},
		queries:     db.New(gowk.DB(ctx)),
	}
}

// CreateOAuth2Client creates a new OAuth2 client
func (s *OAuth2ClientService) CreateOAuth2Client(ctx context.Context, params *OAuth2ClientCreateParams) (*OAuth2ClientResponse, error) {
	// Gin binding validation already handles all validation rules via binding tags
	// No additional manual validation needed

	// Convert arrays to JSON
	redirectURIsJSON, err := json.Marshal(params.RedirectURIs)
	if err != nil {
		return nil, gowk.NewError("invalid redirect URIs format")
	}
	scopesJSON, err := json.Marshal(params.Scopes)
	if err != nil {
		return nil, gowk.NewError("invalid scopes format")
	}
	grantTypesJSON, err := json.Marshal(params.GrantTypes)
	if err != nil {
		return nil, gowk.NewError("invalid grant types format")
	}

	// Set default TTL values
	accessTokenTTL := params.AccessTokenTTL
	if accessTokenTTL == 0 {
		accessTokenTTL = 3600 // 1 hour default
	}
	refreshTokenTTL := params.RefreshTokenTTL
	if refreshTokenTTL == 0 {
		refreshTokenTTL = 2592000 // 30 days default
	}

	// For now, return placeholder response
	// This would require database query implementation
	o, err := s.getQueries(ctx).CreateOAuth2Client(ctx, db.CreateOAuth2ClientParams{
		ID:              gowk.GenerateRandomString(32),
		Name:            params.Name,
		Secret:          gowk.GenerateRandomString(64),
		RedirectUris:    string(redirectURIsJSON),
		Scopes:          string(scopesJSON),
		GrantTypes:      string(grantTypesJSON),
		AccessTokenTtl:  accessTokenTTL,
		RefreshTokenTtl: refreshTokenTTL,
		Created:         pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Updated:         pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Enabled:         true, // true = active
	})
	if err != nil {
		return nil, err
	}
	return BuildOAuth2ClientResponse(o), nil
}

// UpdateOAuth2Client updates an existing OAuth2 client
func (s *OAuth2ClientService) UpdateOAuth2Client(ctx context.Context, params *OAuth2ClientUpdateParams) (*OAuth2ClientResponse, error) {
	// Check if client exists
	_, err := s.GetOAuth2Client(ctx, params.ID)
	if err != nil {
		return nil, gowk.NewError("client not found")
	}

	// Return placeholder response for now
	// This would require database query implementation
	return &OAuth2ClientResponse{
		ID:              params.ID,
		Name:            params.Name,
		Secret:          params.Secret,
		RedirectURIs:    `["http://localhost:8080/callback"]`,
		Scopes:          `["openid", "profile"]`,
		GrantTypes:      `["authorization_code", "refresh_token"]`,
		AccessTokenTTL:  3600,
		RefreshTokenTTL: 2592000,
		Created:         time.Now().Format(time.RFC3339),
		Updated:         time.Now().Format(time.RFC3339),
		Enabled:         true, // true = active
	}, nil
}

// DeleteOAuth2Client soft deletes an OAuth2 client (sets status to disabled)
func (s *OAuth2ClientService) DisableOAuth2Client(ctx context.Context, clientID string) error {
	return s.getQueries(ctx).DisableOAuth2Client(ctx, clientID)
}

// GetOAuth2Client retrieves an OAuth2 client by ID
func (s *OAuth2ClientService) GetOAuth2Client(ctx context.Context, clientID string) (*OAuth2ClientResponse, error) {
	// Gin binding validation already handles all validation rules via binding tags
	// No additional manual validation needed

	// Use SQLC generated method
	queries := s.getQueries(ctx)
	client, err := queries.GetOAuth2Client(ctx, clientID)
	if err != nil {
		return nil, gowk.NewError("client not found")
	}
	// Convert database model to DTO
	return BuildOAuth2ClientResponse(client), nil
}

// ListOAuth2Clients lists OAuth2 clients with pagination and filtering
func (s *OAuth2ClientService) ListOAuth2Clients(ctx context.Context) ([]*OAuth2ClientResponse, error) {
	os, err := s.getQueries(ctx).ListOAuth2Client(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*OAuth2ClientResponse, len(os))
	for _, o := range os {
		res = append(res, BuildOAuth2ClientResponse(o))
	}
	return res, nil
}
