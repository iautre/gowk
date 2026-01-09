package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	authpb "github.com/iautre/gowk/auth/proto"
)

// AuthClient OAuth2客户端
type AuthClient struct {
	conn       *grpc.ClientConn
	authClient authpb.AuthServiceClient
	clientID   string
	secret     string
}

// NewAuthClient 创建OAuth2客户端
func NewAuthClient(addr, clientID, secret string) (*AuthClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %v", err)
	}

	client := authpb.NewAuthServiceClient(conn)

	return &AuthClient{
		conn:       conn,
		authClient: client,
		clientID:   clientID,
		secret:     secret,
	}, nil
}

// OAuth2Token 交换访问令牌
func (c *AuthClient) OAuth2Token(ctx context.Context, grantType, code, redirectURI, clientID, clientSecret, refreshToken, scope, codeVerifier string) (*authpb.OAuth2TokenResponse, error) {
	req := &authpb.OAuth2TokenRequest{
		GrantType:    grantType,
		Code:         code,
		RedirectUri:  redirectURI,
		ClientId:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
		Scope:        scope,
		CodeVerifier: codeVerifier, // PKCE support
	}

	resp, err := c.authClient.OAuth2Token(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OAuth2 token exchange failed: %v", err)
	}

	return resp, nil
}

// OIDCUserInfo 获取OIDC用户信息
func (c *AuthClient) OIDCUserInfo(ctx context.Context, accessToken string) (*authpb.OIDCUserInfoResponse, error) {
	req := &authpb.OIDCUserInfoRequest{
		AccessToken: accessToken,
	}

	resp, err := c.authClient.OIDCUserInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get OIDC user info failed: %v", err)
	}

	return resp, nil
}

// OIDCJwks 获取JWKS公钥
func (c *AuthClient) OIDCJwks(ctx context.Context) (*authpb.OIDCJwksResponse, error) {
	req := &emptypb.Empty{}

	resp, err := c.authClient.OIDCJwks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get JWKS failed: %v", err)
	}

	return resp, nil
}

// Close 关闭客户端连接
func (c *AuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
