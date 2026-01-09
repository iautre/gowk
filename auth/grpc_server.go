package auth

import (
	"context"
	"net/url"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/iautre/gowk"
	authpb "github.com/iautre/gowk/auth/proto"
)

// AuthServer implements the AuthService gRPC interface
type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
}

// NewAuthServer creates a new AuthServer
func NewAuthServer() *AuthServer {
	return &AuthServer{}
}

// Login handles user login
func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	if req.Account == "" || req.Code == "" {
		return nil, status.Errorf(codes.InvalidArgument, "account and code are required")
	}

	userService := UserService{}
	user, err := userService.Login(ctx, &LoginParams{
		Account: req.Account,
		Code:    req.Code,
	})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	// Create token directly for gRPC context
	token := gowk.UUID()

	// TODO: Store token in database/cache for validation

	return &authpb.LoginResponse{
		Token:    token,
		UserId:   user.ID,
		Nickname: user.Nickname.String,
	}, nil
}

// Register handles user registration
func (s *AuthServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.LoginResponse, error) {
	// TODO: Implement user registration logic
	return nil, status.Errorf(codes.Unimplemented, "registration not implemented")
}

// UserInfo handles user info request
func (s *AuthServer) UserInfo(ctx context.Context, req *authpb.UserInfoRequest) (*authpb.UserInfoResponse, error) {
	if req.UserId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	userService := UserService{}
	user, err := userService.GetById(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &authpb.UserInfoResponse{
		Id:       user.ID,
		Phone:    user.Phone.String,
		Email:    user.Email.String,
		Nickname: user.Nickname.String,
		Group:    user.Group.String,
	}, nil
}

// OAuth2Auth handles OAuth2 authorization endpoint
func (s *AuthServer) OAuth2Auth(ctx context.Context, req *authpb.OAuth2AuthRequest) (*authpb.OAuth2AuthResponse, error) {
	if req.ClientId == "" || req.RedirectUri == "" || req.ResponseType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "client_id, redirect_uri, and response_type are required")
	}

	// Validate client using OAuth2Service
	oauth2Service := NewOAuth2Service(ctx)

	// Convert gRPC request to internal request format
	authReq := &OAuth2AuthRequest{
		ResponseType: req.ResponseType,
		ClientID:     req.ClientId,
		RedirectURI:  req.RedirectUri,
		Scope:        req.Scope,
		State:        req.State,
		Nonce:        req.Nonce,
	}

	_, err := oauth2Service.ValidateOAuth2AuthRequest(ctx, authReq)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "invalid client: %v", err)
	}

	// Generate authorization code
	authCode, err := oauth2Service.GenerateAuthorizationCode(ctx, req.ClientId, 0, req.RedirectUri, req.Scope, req.State, req.Nonce)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate authorization code: %v", err)
	}

	// Build redirect URL with authorization code
	redirectURL, err := url.Parse(req.RedirectUri)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid redirect URI: %v", err)
	}

	params := redirectURL.Query()
	params.Set("code", authCode)
	if req.State != "" {
		params.Set("state", req.State)
	}
	redirectURL.RawQuery = params.Encode()

	return &authpb.OAuth2AuthResponse{
		RedirectUrl: redirectURL.String(),
	}, nil
}

// OAuth2Token handles OAuth2 token endpoint
func (s *AuthServer) OAuth2Token(ctx context.Context, req *authpb.OAuth2TokenRequest) (*authpb.OAuth2TokenResponse, error) {
	if req.GrantType == "" || req.ClientId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "grant_type and client_id are required")
	}

	oauth2Service := NewOAuth2Service(ctx)

	switch req.GrantType {
	case "authorization_code":
		if req.Code == "" {
			return nil, status.Errorf(codes.InvalidArgument, "code is required for authorization_code grant")
		}

		// Convert gRPC request to internal format
		tokenReq := &OAuth2TokenRequest{
			GrantType:    req.GrantType,
			Code:         req.Code,
			RedirectURI:  req.RedirectUri,
			ClientID:     req.ClientId,
			ClientSecret: req.ClientSecret,
			Scope:        req.Scope,
		}

		// Exchange code for token
		response, err := oauth2Service.ExchangeCodeForToken(ctx, tokenReq)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "token exchange failed: %v", err)
		}

		return &authpb.OAuth2TokenResponse{
			AccessToken:  response.AccessToken,
			TokenType:    response.TokenType,
			ExpiresIn:    response.ExpiresIn,
			RefreshToken: response.RefreshToken,
			Scope:        response.Scope,
			IdToken:      response.IDToken,
		}, nil

	case "refresh_token":
		if req.RefreshToken == "" {
			return nil, status.Errorf(codes.InvalidArgument, "refresh_token is required for refresh_token grant")
		}

		// Refresh token
		response, err := oauth2Service.RefreshToken(ctx, req.RefreshToken)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "refresh token failed: %v", err)
		}

		return &authpb.OAuth2TokenResponse{
			AccessToken:  response.AccessToken,
			TokenType:    response.TokenType,
			ExpiresIn:    response.ExpiresIn,
			RefreshToken: response.RefreshToken,
			Scope:        response.Scope,
			IdToken:      response.IDToken,
		}, nil

	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported grant_type: %s", req.GrantType)
	}
}

// OIDCDiscovery handles OIDC discovery endpoint
func (s *AuthServer) OIDCDiscovery(ctx context.Context, req *emptypb.Empty) (*authpb.OIDCDiscoveryResponse, error) {
	baseURL := "http://localhost:8087" // TODO: Get from config

	return &authpb.OIDCDiscoveryResponse{
		Issuer:                           baseURL,
		AuthorizationEndpoint:            baseURL + "/auth/oauth2/authorize",
		TokenEndpoint:                    baseURL + "/auth/oauth2/token",
		UserinfoEndpoint:                 baseURL + "/auth/oidc/userinfo",
		JwksUri:                          baseURL + "/auth/oidc/jwks",
		ScopesSupported:                  []string{"openid", "profile", "email"},
		ResponseTypesSupported:           []string{"code", "id_token", "token id_token"},
		GrantTypesSupported:              []string{"authorization_code", "refresh_token"},
		SubjectTypesSupported:            []string{"public"},
		IdTokenSigningAlgValuesSupported: []string{"RS256"},
	}, nil
}

// OIDCUserInfo handles OIDC userinfo endpoint
func (s *AuthServer) OIDCUserInfo(ctx context.Context, req *authpb.OIDCUserInfoRequest) (*authpb.OIDCUserInfoResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "access_token is required")
	}

	// TODO: Validate access token and get user info
	// For now, return mock response
	return &authpb.OIDCUserInfoResponse{
		Sub:           "123456789",
		Name:          "John Doe",
		Email:         "john.doe@example.com",
		EmailVerified: true,
	}, nil
}

// OIDCJwks handles OIDC JWKS endpoint
func (s *AuthServer) OIDCJwks(ctx context.Context, req *emptypb.Empty) (*authpb.OIDCJwksResponse, error) {
	// TODO: Return actual public keys
	// For now, return mock JWKS
	mockJWK := &authpb.OIDCJwk{
		Kty: "RSA",
		Use: "sig",
		Kid: "1",
		N:   "mock_n_value",
		E:   "AQAB",
		Alg: "RS256",
	}

	return &authpb.OIDCJwksResponse{
		Keys: []*authpb.OIDCJwk{mockJWK},
	}, nil
}

// Helper functions
func isValidRedirectURI(validURIs []string, providedURI string) bool {
	for _, uri := range validURIs {
		if uri == providedURI {
			return true
		}
	}
	return false
}
