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
func NewAuthServer(ctx context.Context) *AuthServer {
	return &AuthServer{}
}

// OAuth2Token handles OAuth2 token endpoint - 使用现有OAuth2Service
func (s *AuthServer) OAuth2Token(ctx context.Context, req *authpb.OAuth2TokenRequest) (*authpb.OAuth2TokenResponse, error) {
	// Convert gRPC request to internal format
	tokenReq := &OAuth2TokenRequest{
		GrantType:    req.GrantType,
		Code:         req.Code,
		RedirectURI:  req.RedirectUri,
		ClientID:     req.ClientId,
		ClientSecret: req.ClientSecret,
		RefreshToken: req.RefreshToken,
		Scope:        req.Scope,
		CodeVerifier: req.CodeVerifier,
	}

	// Use existing OAuth2Service
	var oauth2Service OAuth2Service
	response, err := oauth2Service.ExchangeCodeForToken(ctx, tokenReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "token exchange failed: %v", err)
	}

	// Convert to gRPC response format
	return &authpb.OAuth2TokenResponse{
		AccessToken:  response.AccessToken,
		TokenType:    response.TokenType,
		ExpiresIn:    response.ExpiresIn,
		RefreshToken: response.RefreshToken,
		Scope:        response.Scope,
		IdToken:      response.IDToken,
	}, nil
}

// OIDCUserInfo handles OIDC userinfo endpoint - 使用现有OIDCService
func (s *AuthServer) OIDCUserInfo(ctx context.Context, req *authpb.OIDCUserInfoRequest) (*authpb.OIDCUserInfoResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "access_token is required")
	}

	// Use OAuth2Service to validate access token
	var oauth2Service OAuth2Service
	oauth2Token, err := oauth2Service.ValidateAccessToken(ctx, req.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid or expired access token: %v", err)
	}

	// Use existing OIDCService to get user info
	var oidcService OIDCService
	userInfo, err := oidcService.GetUserInfo(ctx, oauth2Token.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user info: %v", err)
	}

	// Convert to gRPC response format
	return &authpb.OIDCUserInfoResponse{
		Sub:                 userInfo.Sub,
		Name:                userInfo.Name,
		Email:               userInfo.Email,
		EmailVerified:       userInfo.EmailVerified,
		GivenName:           userInfo.GivenName,
		FamilyName:          userInfo.FamilyName,
		MiddleName:          userInfo.MiddleName,
		Nickname:            userInfo.Nickname,
		PreferredUsername:   userInfo.PreferredUsername,
		Picture:             userInfo.Picture,
		PhoneNumber:         userInfo.PhoneNumber,
		PhoneNumberVerified: userInfo.PhoneVerified,
		Locale:              userInfo.Locale,
		UpdatedAt:           userInfo.UpdatedAt,
	}, nil
}

// OIDCDiscovery handles OIDC discovery endpoint - 使用现有OIDCService
func (s *AuthServer) OIDCDiscovery(ctx context.Context, req *emptypb.Empty) (*authpb.OIDCDiscoveryResponse, error) {
	// Use existing OIDCService
	var oidcService OIDCService
	discovery := oidcService.GetDiscoveryDocument()

	// Convert to gRPC response format
	return &authpb.OIDCDiscoveryResponse{
		Issuer:                           discovery.Issuer,
		AuthorizationEndpoint:            discovery.AuthorizationEndpoint,
		TokenEndpoint:                    discovery.TokenEndpoint,
		UserinfoEndpoint:                 discovery.UserInfoEndpoint,
		JwksUri:                          discovery.JwksUri,
		ScopesSupported:                  discovery.ScopesSupported,
		ResponseTypesSupported:           discovery.ResponseTypesSupported,
		GrantTypesSupported:              discovery.GrantTypesSupported,
		SubjectTypesSupported:            discovery.SubjectTypesSupported,
		IdTokenSigningAlgValuesSupported: discovery.IDTokenSigningAlgValuesSupported,
	}, nil
}

// OIDCJwks handles OIDC JWKS endpoint - 使用现有OIDCService
func (s *AuthServer) OIDCJwks(ctx context.Context, req *emptypb.Empty) (*authpb.OIDCJwksResponse, error) {
	// Use existing OIDCService
	var oidcService OIDCService
	jwks := oidcService.GetJwks()

	// Convert to gRPC response format
	var keys []*authpb.OIDCJwk
	for _, key := range jwks.Keys {
		keys = append(keys, &authpb.OIDCJwk{
			Kty: key.Kty,
			Use: key.Use,
			Kid: key.Kid,
			N:   key.N,
			E:   key.E,
			Alg: key.Alg,
		})
	}

	return &authpb.OIDCJwksResponse{
		Keys: keys,
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
