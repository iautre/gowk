package auth

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

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
