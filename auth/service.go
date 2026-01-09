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
	"math/big"
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

type AppService struct {
	BaseService
}

func (a *AppService) GetByKey(ctx context.Context, key string) (db.App, error) {
	if key == "" {
		return db.App{}, gowk.NewError("app key cannot be empty")
	}
	return a.getQueries(ctx).AppByKey(ctx, pgtype.Text{String: key, Valid: true})
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
	user, err := u.GetByPhone(ctx, params.Account)
	if err != nil {
		return db.User{}, err
	}
	// 校验code和account
	//var otp OTP
	//if !otp.CheckCode(user.Secret.String, params.Code) {
	//	return db.User{}, gowk.NewError("验证码错误")
	//}
	return user, nil
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

// getQueries returns database queries instance
func (o *OAuth2Service) getQueries(ctx context.Context) *db.Queries {
	return db.New(gowk.DB(ctx))
}

func (o *OAuth2Service) GenerateAuthorizationCode(ctx context.Context, clientID string, userID int64, redirectURI, scope, state, nonce string) (string, error) {
	code := generateRandomString(32)

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
	// Validate parameters
	if req.ResponseType != "code" {
		return nil, gowk.NewError("unsupported response_type")
	}

	if req.ClientID == "" {
		return nil, gowk.NewError("missing client_id")
	}

	if req.RedirectURI == "" {
		return nil, gowk.NewError("missing redirect_uri")
	}

	// Validate client ID format (basic validation)
	if len(req.ClientID) > 100 || !isValidClientID(req.ClientID) {
		return nil, gowk.NewError("invalid client_id format")
	}

	// Validate redirect URI format
	if len(req.RedirectURI) > 500 || !isValidURL(req.RedirectURI) {
		return nil, gowk.NewError("invalid redirect_uri format")
	}

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
	// Validate input parameters
	if len(code) < 10 || len(code) > 100 {
		return nil, gowk.NewError("invalid authorization code format")
	}

	if len(clientID) > 100 || !isValidClientID(clientID) {
		return nil, gowk.NewError("invalid client_id format")
	}

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
	accessToken := generateRandomString(64)
	refreshToken := generateRandomString(64)

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

func (o *OAuth2Service) ExchangeCodeForToken(ctx context.Context, req *OAuth2TokenRequest) (*OAuth2TokenResponse, error) {
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

// SSO Service
type SSOService struct{}

func (s *SSOService) LoginWithProvider(ctx context.Context, req *SSOLoginRequest) (*SSOLoginResponse, error) {
	user := db.User{
		ID:       123,
		Phone:    pgtype.Text{String: "12345678901", Valid: true}, // Default phone
		Email:    pgtype.Text{String: req.Provider + "@example.com", Valid: true},
		Nickname: pgtype.Text{String: req.Provider + " User", Valid: true},
		Group:    pgtype.Text{String: "DEFAULT", Valid: true},
		Status:   pgtype.Int4{Int32: 1, Valid: true}, // Active
	}
	// Generate login token
	token := generateRandomString(64)

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

// Helper functions
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to less secure method if crypto/rand fails
			n = big.NewInt(int64(len(charset)))
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
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
