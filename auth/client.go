package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/iautre/gowk/auth/proto"
)

// AuthClient OAuth2客户端
type AuthClient struct {
	conn         *grpc.ClientConn
	authClient   proto.AuthServiceClient
	clientID     string
	clientSecret string
	redirectURI  string
	baseURL      string
}

// NewOAuth2Client 创建新的OAuth2客户端
func NewOAuth2Client(serverAddr, clientID, clientSecret, redirectURI string) (*AuthClient, error) {
	// 建立gRPC连接
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth server: %w", err)
	}

	authClient := proto.NewAuthServiceClient(conn)

	return &AuthClient{
		conn:         conn,
		authClient:   authClient,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}, nil
}

// Close 关闭客户端连接
func (c *AuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Authorize 获取授权码
func (c *AuthClient) Authorize(ctx context.Context, scope, state string) (*AuthorizeResult, error) {
	// 生成PKCE参数
	codeVerifier, codeChallenge := generatePKCE()

	req := &proto.AuthorizeRequest{
		ClientId:            c.clientID,
		RedirectUri:         c.redirectURI,
		ResponseType:        "code",
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: "S256",
	}

	resp, err := c.authClient.Authorize(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("authorize failed: %w", err)
	}

	if resp.Error != "" {
		return &AuthorizeResult{
			Error:            resp.Error,
			ErrorDescription: resp.ErrorDescription,
		}, nil
	}

	return &AuthorizeResult{
		Code:         resp.Code,
		State:        resp.State,
		RedirectURI:  resp.RedirectUri,
		CodeVerifier: codeVerifier,
	}, nil
}

// GetToken 获取访问令牌
func (c *AuthClient) GetToken(ctx context.Context, code, codeVerifier string) (*TokenResult, error) {
	req := &proto.TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectUri:  c.redirectURI,
		ClientId:     c.clientID,
		ClientSecret: c.clientSecret,
		CodeVerifier: codeVerifier,
	}

	resp, err := c.authClient.GetToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	if resp.Error != "" {
		return &TokenResult{
			Error:            resp.Error,
			ErrorDescription: resp.ErrorDescription,
		}, nil
	}

	return &TokenResult{
		AccessToken:  resp.AccessToken,
		TokenType:    resp.TokenType,
		ExpiresIn:    resp.ExpiresIn,
		RefreshToken: resp.RefreshToken,
		Scope:        resp.Scope,
		IDToken:      resp.IdToken,
	}, nil
}

// RefreshToken 刷新访问令牌
func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error) {
	req := &proto.RefreshTokenRequest{
		RefreshToken: refreshToken,
		ClientId:     c.clientID,
		ClientSecret: c.clientSecret,
	}

	resp, err := c.authClient.RefreshToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("refresh token failed: %w", err)
	}

	if resp.Error != "" {
		return &TokenResult{
			Error:            resp.Error,
			ErrorDescription: resp.ErrorDescription,
		}, nil
	}

	return &TokenResult{
		AccessToken:  resp.AccessToken,
		TokenType:    resp.TokenType,
		ExpiresIn:    resp.ExpiresIn,
		RefreshToken: resp.RefreshToken,
		Scope:        resp.Scope,
		IDToken:      resp.IdToken,
	}, nil
}

// GetUserInfo 获取用户信息
func (c *AuthClient) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req := &proto.UserInfoRequest{
		AccessToken: accessToken,
	}

	resp, err := c.authClient.GetUserInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}

	if resp.Error != "" {
		return &UserInfo{
			Error:            resp.Error,
			ErrorDescription: resp.ErrorDescription,
		}, nil
	}

	return &UserInfo{
		Sub:               resp.Sub,
		Name:              resp.Name,
		Email:             resp.Email,
		Phone:             resp.Phone,
		Picture:           resp.Picture,
		Nickname:          resp.Nickname,
		PreferredUsername: resp.PreferredUsername,
		UpdatedAt:         resp.UpdatedAt,
	}, nil
}

// RevokeToken 撤销令牌
func (c *AuthClient) RevokeToken(ctx context.Context, token, tokenTypeHint string) (*RevokeResult, error) {
	req := &proto.RevokeTokenRequest{
		Token:         token,
		TokenTypeHint: tokenTypeHint,
		ClientId:      c.clientID,
		ClientSecret:  c.clientSecret,
	}

	resp, err := c.authClient.RevokeToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("revoke token failed: %w", err)
	}

	return &RevokeResult{
		Success:          resp.Success,
		Error:            resp.Error,
		ErrorDescription: resp.ErrorDescription,
	}, nil
}

// AuthorizeResult 授权结果
type AuthorizeResult struct {
	Code             string
	State            string
	RedirectURI      string
	CodeVerifier     string
	Error            string
	ErrorDescription string
}

// TokenResult 令牌结果
type TokenResult struct {
	AccessToken      string
	TokenType        string
	ExpiresIn        int64
	RefreshToken     string
	Scope            string
	IDToken          string
	Error            string
	ErrorDescription string
}

// UserInfo 用户信息
type UserInfo struct {
	Sub               string
	Name              string
	Email             string
	Phone             string
	Picture           string
	Nickname          string
	PreferredUsername string
	UpdatedAt         string
	Error             string
	ErrorDescription  string
}

// RevokeResult 撤销结果
type RevokeResult struct {
	Success          bool
	Error            string
	ErrorDescription string
}

// generatePKCE 生成PKCE参数
func generatePKCE() (string, string) {
	// 生成code verifier (43-128字符的随机字符串)
	codeVerifier := generateRandomStringForPKCE(64)

	// 计算code challenge (SHA256 + base64url)
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return codeVerifier, codeChallenge
}

// generateRandomStringForPKCE 生成PKCE用的随机字符串
func generateRandomStringForPKCE(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	b := make([]byte, length)
	for i := range b {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		b[i] = charset[int(randomByte[0])%len(charset)]
	}
	return string(b)
}
