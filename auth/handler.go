package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth/db"
	authpb "github.com/iautre/gowk/auth/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserHandler struct {
}

func NewUserHandler(ctx context.Context) *UserHandler {
	return &UserHandler{}
}

func (u *UserHandler) Login(ctx *gin.Context) {
	var params LoginParams
	err := ctx.ShouldBind(&params)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}
	var userService UserService
	user, err := userService.Login(ctx, &params)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}
	token, err := gowk.Login(ctx, user.ID)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}
	gowk.Response(ctx, http.StatusOK, token, nil)
}

func (u *UserHandler) BasicAuthMiddleware(ctx *gin.Context) {
	err := u.validateBasicAuth(ctx)
	if err != nil {
		u.requireBasicAuth(ctx)
		return
	}
	// 验证通过（实际场景）
	ctx.Next()
}

func (u *UserHandler) requireBasicAuth(ctx *gin.Context) {
	// 关键1：设置WWW-Authenticate响应头（必须）
	ctx.Header("WWW-Authenticate", `Basic realm="请输入系统账号密码"`)
	// 关键2：返回401 Unauthorized状态码（必须）
	gowk.Response(ctx, http.StatusUnauthorized, nil, gowk.NewError("请输入系统账号密码"))
}

func (u *UserHandler) validateBasicAuth(ctx *gin.Context) error {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		return gowk.NewError("no Authorization")
	}
	username, pwd, ok := u.parseBasicAuth(ctx)
	if !ok {
		return gowk.NewError("no Authorization")
	}
	params := &LoginParams{
		Account: username,
		Code:    pwd,
	}
	var userService UserService
	user, err := userService.Login(ctx, params)
	if err != nil {
		return err
	}
	_, err = gowk.Login(ctx, user.ID)
	if err != nil {
		return err
	}
	return nil
}

// ParseBasicAuth 从Gin上下文解析Basic Auth的用户名和密码
// 返回值：username-用户名，password-密码，ok-是否解析成功，err-错误信息
func (u *UserHandler) parseBasicAuth(c *gin.Context) (username, password string, ok bool) {
	// 1. 提取Authorization请求头
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", "", false
	}

	// 2. 拆分"Basic "和base64字符串（必须是两部分）
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	authType := strings.ToLower(parts[0])
	base64Str := parts[1]

	// 3. 校验认证类型是否为Basic
	if authType != "basic" {
		return "", "", false
	}

	// 4. Base64解码
	decodedBytes, decodeErr := base64.StdEncoding.DecodeString(base64Str)
	if decodeErr != nil {
		return "", "", false
	}
	decodedStr := string(decodedBytes)

	// 5. 按":"拆分用户名和密码（必须拆分出两部分）
	credParts := strings.SplitN(decodedStr, ":", 2)
	if len(credParts) != 2 {
		return "", "", false
	}
	username = credParts[0]
	password = credParts[1]

	// 6. 校验用户名/密码是否为空
	if username == "" || password == "" {
		return "", "", false
	}

	// 解析成功
	return username, password, true
}

func (u *UserHandler) Register(ctx *gin.Context) {
	var params RegisterParams
	err := ctx.ShouldBind(&params)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}
}

func (u *UserHandler) UserInfo(ctx *gin.Context) {
	userId := gowk.LoginId(ctx)
	var userService UserService
	user, err := userService.GetById(ctx, userId)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}
	ctx.JSON(200, gowk.CopyByJson[db.User, UserRes](user))
	ctx.Abort()
}

func (u *UserHandler) Smscode(ctx *gin.Context) {
	params := &LoginParams{}
	err := ctx.ShouldBind(params)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}
	var userService UserService
	user, err := userService.Login(ctx, params)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}
	token, err := gowk.Login(ctx, user.ID)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}
	ctx.JSON(200, &LoginRes{
		Token:    token,
		UserId:   user.ID,
		Nickname: user.Nickname.String,
	})
	ctx.Abort()
}

// SSO Login endpoint
func (u *UserHandler) SSOLogin(ctx *gin.Context) {
	var params SSOLoginRequest
	err := ctx.ShouldBind(&params)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}

	var ssoService SSOService
	response, err := ssoService.LoginWithProvider(ctx, &params)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}

	ctx.JSON(200, response)
	ctx.Abort()
}

type OAuth2Handler struct {
	oauth2Service *OAuth2Service
	oidcService   *OIDCService
}

func NewOAuth2Handler(ctx context.Context) *OAuth2Handler {
	return &OAuth2Handler{
		oauth2Service: NewOAuth2Service(ctx),
		oidcService:   &OIDCService{},
	}
}

func (o *OAuth2Handler) OAuth2Auth(ctx *gin.Context) {
	// Parse request parameters
	var params OAuth2AuthRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}

	// Get user ID from session or token
	userID := gowk.LoginId(ctx)

	// Validate request using existing service layer
	_, err := o.oauth2Service.ValidateOAuth2AuthRequest(ctx, &params)
	if err != nil {
		gowk.Response(ctx, http.StatusUnauthorized, nil, err)
		return
	}

	// Generate authorization code using existing service
	authCode, err := o.oauth2Service.GenerateAuthorizationCode(ctx, params.ClientID, userID, params.RedirectURI, params.Scope, params.State, params.Nonce)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}

	// Build redirect URL
	redirectURL, err := url.Parse(params.RedirectURI)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}

	queryParams := redirectURL.Query()
	queryParams.Set("code", authCode)
	if params.State != "" {
		queryParams.Set("state", params.State)
	}
	redirectURL.RawQuery = queryParams.Encode()

	ctx.Redirect(http.StatusFound, redirectURL.String())
}

func (o *OAuth2Handler) OAuth2Token(ctx *gin.Context) {
	var params OAuth2TokenRequest
	err := ctx.ShouldBind(&params)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}

	// Use unified ExchangeToken method
	response, err := o.oauth2Service.ExchangeToken(ctx, &params)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}
	gowk.Response(ctx, http.StatusOK, response, nil)
}

func (o *OAuth2Handler) OIDCDiscovery(ctx *gin.Context) {
	discovery := o.oidcService.GetDiscoveryDocument()
	gowk.Response(ctx, http.StatusOK, discovery, nil)
}

func (o *OAuth2Handler) OIDCUserInfo(ctx *gin.Context) {
	// Get user ID from OAuth2TokenMiddleware
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		gowk.Response(ctx, http.StatusUnauthorized, nil, gowk.NewError("User ID not found in context"))
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		gowk.Response(ctx, http.StatusUnauthorized, nil, gowk.NewError("Invalid user ID"))
		return
	}

	userInfo, err := o.oidcService.GetUserInfo(ctx, userID)
	if err != nil {
		gowk.Response(ctx, http.StatusBadRequest, nil, err)
		return
	}
	gowk.Response(ctx, http.StatusOK, userInfo, nil)
}

func (o *OAuth2Handler) OIDCJwks(ctx *gin.Context) {
	jwks := o.oidcService.GetJwks()
	gowk.Response(ctx, http.StatusOK, jwks, nil)
}

// GrpcHandler 处理gRPC相关的请求
type GrpcHandler struct {
	oauth2Service *OAuth2Service
	oidcService   *OIDCService
}

func NewGrpcHandler(ctx context.Context) *GrpcHandler {
	return &GrpcHandler{
		oauth2Service: NewOAuth2Service(ctx),
		oidcService:   &OIDCService{},
	}
}

// OAuth2Token handles OAuth2 token endpoint - gRPC version
func (g *GrpcHandler) OAuth2Token(ctx context.Context, req *authpb.OAuth2TokenRequest) (*authpb.OAuth2TokenResponse, error) {
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

	// Use unified ExchangeToken method
	response, err := g.oauth2Service.ExchangeToken(ctx, tokenReq)
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

// OIDCUserInfo handles OIDC userinfo endpoint - gRPC version
func (g *GrpcHandler) OIDCUserInfo(ctx context.Context, req *authpb.OIDCUserInfoRequest) (*authpb.OIDCUserInfoResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "access_token is required")
	}

	// Use OAuth2Service to validate access token
	oauth2Token, err := g.oauth2Service.ValidateAccessToken(ctx, req.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid or expired access token: %v", err)
	}

	// Use existing OIDCService to get user info
	userInfo, err := g.oidcService.GetUserInfo(ctx, oauth2Token.UserID)
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

// OIDCDiscovery handles OIDC discovery endpoint - gRPC version
func (g *GrpcHandler) OIDCDiscovery(ctx context.Context, req *emptypb.Empty) (*authpb.OIDCDiscoveryResponse, error) {
	// Use existing OIDCService
	discovery := g.oidcService.GetDiscoveryDocument()

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

// OIDCJwks handles OIDC JWKS endpoint - gRPC version
func (g *GrpcHandler) OIDCJwks(ctx context.Context, req *emptypb.Empty) (*authpb.OIDCJwksResponse, error) {
	// Use existing OIDCService
	jwks := g.oidcService.GetJwks()

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
