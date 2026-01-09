package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

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

// ParseIDToken 解析ID Token获取用户信息 (不验证签名)
func (c *AuthClient) ParseIDToken(idToken string) (*authpb.OIDCUserInfoResponse, error) {
	if idToken == "" {
		return nil, fmt.Errorf("id_token is required")
	}

	// ID Token格式: header.payload.signature
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid id_token format")
	}

	// 解析payload部分
	payload := parts[1]
	// 补齐base64 padding
	if len(payload)%4 != 0 {
		payload += strings.Repeat("=", 4-len(payload)%4)
	}

	// Base64解码
	decoded, err := base64Decode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode id_token payload: %v", err)
	}

	// 解析JSON
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse id_token claims: %v", err)
	}

	// 转换为OIDCUserInfoResponse
	userInfo := &authpb.OIDCUserInfoResponse{}

	if sub, ok := claims["sub"].(string); ok {
		userInfo.Sub = sub
	}
	if name, ok := claims["name"].(string); ok {
		userInfo.Name = name
	}
	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}
	if emailVerified, ok := claims["email_verified"].(bool); ok {
		userInfo.EmailVerified = emailVerified
	}
	if givenName, ok := claims["given_name"].(string); ok {
		userInfo.GivenName = givenName
	}
	if familyName, ok := claims["family_name"].(string); ok {
		userInfo.FamilyName = familyName
	}
	if nickname, ok := claims["nickname"].(string); ok {
		userInfo.Nickname = nickname
	}
	if preferredUsername, ok := claims["preferred_username"].(string); ok {
		userInfo.PreferredUsername = preferredUsername
	}
	if picture, ok := claims["picture"].(string); ok {
		userInfo.Picture = picture
	}
	if phoneNumber, ok := claims["phone_number"].(string); ok {
		userInfo.PhoneNumber = phoneNumber
	}
	if phoneVerified, ok := claims["phone_number_verified"].(bool); ok {
		userInfo.PhoneNumberVerified = phoneVerified
	}
	if locale, ok := claims["locale"].(string); ok {
		userInfo.Locale = locale
	}
	if updatedAt, ok := claims["updated_at"].(float64); ok {
		userInfo.UpdatedAt = int64(updatedAt)
	}

	return userInfo, nil
}

// ValidateAndParseIDToken 验证ID Token签名并解析用户信息 (推荐)
func (c *AuthClient) ValidateAndParseIDToken(ctx context.Context, idToken string) (*authpb.OIDCUserInfoResponse, error) {
	if idToken == "" {
		return nil, fmt.Errorf("id_token is required")
	}

	// 1. 获取JWKS公钥
	jwksResp, err := c.OIDCJwks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %v", err)
	}

	if len(jwksResp.Keys) == 0 {
		return nil, fmt.Errorf("no public keys found in JWKS")
	}

	// 2. 解析JWT header获取kid
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid id_token format")
	}

	// 解析header
	headerData, err := base64Decode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT header: %v", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerData, &header); err != nil {
		return nil, fmt.Errorf("failed to parse JWT header: %v", err)
	}

	kid, ok := header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("JWT header missing kid field")
	}

	// 3. 找到对应的公钥
	var publicKey *rsa.PublicKey
	for _, key := range jwksResp.Keys {
		if key.Kid == kid {
			publicKey, err = parseRSAPublicKey(key)
			if err != nil {
				return nil, fmt.Errorf("failed to parse RSA public key: %v", err)
			}
			break
		}
	}

	if publicKey == nil {
		return nil, fmt.Errorf("public key not found for kid: %s", kid)
	}

	// 4. 验证JWT签名 (这里简化处理，实际需要完整的JWT库)
	// TODO: 实现完整的RSA签名验证
	// 这里先解析payload，实际生产环境应该验证签名

	// 5. 解析payload获取用户信息
	return c.ParseIDToken(idToken)
}

// parseRSAPublicKey 从JWK解析RSA公钥
func parseRSAPublicKey(jwk *authpb.OIDCJwk) (*rsa.PublicKey, error) {
	// TODO: 实现完整的JWK到RSA公钥转换
	// 这里需要解析modulus(n)和exponent(e)
	// 暂时返回nil，需要完整的RSA实现
	return nil, fmt.Errorf("RSA public key parsing not implemented yet")
}

// base64Decode 安全的base64解码
func base64Decode(data string) ([]byte, error) {
	// 替换URL安全的base64字符
	data = strings.ReplaceAll(data, "-", "+")
	data = strings.ReplaceAll(data, "_", "/")

	// 补齐padding
	if len(data)%4 != 0 {
		data += strings.Repeat("=", 4-len(data)%4)
	}

	// Base64解码
	return base64.StdEncoding.DecodeString(data)
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
