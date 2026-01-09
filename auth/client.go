package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

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

	// 4. 验证JWT签名
	if err := verifyJWTSignature(idToken, publicKey); err != nil {
		return nil, fmt.Errorf("JWT signature verification failed: %v", err)
	}

	// 5. 解析payload获取用户信息并验证claims
	userInfo, err := c.parseAndValidateClaims(idToken)
	if err != nil {
		return nil, fmt.Errorf("JWT claims validation failed: %v", err)
	}

	return userInfo, nil
}

// parseRSAPublicKey 从JWK解析RSA公钥
func parseRSAPublicKey(jwk *authpb.OIDCJwk) (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}

	// 解码modulus (n)
	n, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %v", err)
	}

	// 解码exponent (e)
	e, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %v", err)
	}

	// 创建RSA公钥
	modulus := new(big.Int).SetBytes(n)
	exponent := new(big.Int).SetBytes(e)

	publicKey := &rsa.PublicKey{
		N: modulus,
		E: int(exponent.Int64()),
	}

	return publicKey, nil
}

// verifyJWTSignature 验证JWT签名
func verifyJWTSignature(idToken string, publicKey *rsa.PublicKey) error {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format")
	}

	// 获取签名部分
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("failed to decode signature: %v", err)
	}

	// 创建要验证的数据: header.payload
	data := parts[0] + "." + parts[1]

	// 根据算法选择哈希函数
	var hash crypto.Hash
	switch header := parseJWTHeader(parts[0]); header["alg"] {
	case "RS256":
		hash = crypto.SHA256
	default:
		return fmt.Errorf("unsupported algorithm: %v", header["alg"])
	}

	// 计算哈希
	hasher := hash.New()
	hasher.Write([]byte(data))
	hashed := hasher.Sum(nil)

	// 验证签名
	return rsa.VerifyPKCS1v15(publicKey, hash, hashed, signature)
}

// parseJWTHeader 解析JWT header
func parseJWTHeader(headerStr string) map[string]interface{} {
	headerData, _ := base64Decode(headerStr)
	var header map[string]interface{}
	json.Unmarshal(headerData, &header)
	return header
}

// parseAndValidateClaims 解析并验证JWT claims
func (c *AuthClient) parseAndValidateClaims(idToken string) (*authpb.OIDCUserInfoResponse, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// 解析payload
	payloadData, err := base64Decode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %v", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadData, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %v", err)
	}

	// 验证标准claims
	now := time.Now().Unix()

	// 验证过期时间 (exp)
	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < now {
			return nil, fmt.Errorf("token expired at %d", int64(exp))
		}
	}

	// 验证生效时间 (nbf)
	if nbf, ok := claims["nbf"].(float64); ok {
		if int64(nbf) > now {
			return nil, fmt.Errorf("token not valid until %d", int64(nbf))
		}
	}

	// 验证签发时间 (iat)
	if iat, ok := claims["iat"].(float64); ok {
		if int64(iat) > now {
			return nil, fmt.Errorf("token issued in the future at %d", int64(iat))
		}
	}

	// 验证必需的claims
	if _, ok := claims["sub"].(string); !ok {
		return nil, fmt.Errorf("missing required claim: sub")
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

// OIDCDiscovery 获取OIDC发现文档
func (c *AuthClient) OIDCDiscovery(ctx context.Context) (*authpb.OIDCDiscoveryResponse, error) {
	req := &emptypb.Empty{}

	resp, err := c.authClient.OIDCDiscovery(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get OIDC discovery failed: %v", err)
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
