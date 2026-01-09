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

	switch params.GrantType {
	case "authorization_code":
		response, err := o.oauth2Service.ExchangeCodeForToken(ctx, &params)
		if err != nil {
			gowk.Response(ctx, http.StatusBadRequest, nil, err)
			return
		}
		ctx.JSON(http.StatusOK, response)
		ctx.Abort()
	case "refresh_token":
		response, err := o.oauth2Service.RefreshToken(ctx, params.RefreshToken)
		if err != nil {
			gowk.Response(ctx, http.StatusBadRequest, nil, err)
			return
		}
		ctx.JSON(http.StatusOK, response)
		ctx.Abort()
	default:
		gowk.Response(ctx, http.StatusBadRequest, nil, gowk.NewError("Unsupported grant_type"))
	}
}

func (o *OAuth2Handler) OIDCDiscovery(ctx *gin.Context) {
	discovery := o.oidcService.GetDiscoveryDocument()
	ctx.JSON(200, discovery)
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

	ctx.JSON(200, userInfo)
	ctx.Abort()
}

func (o *OAuth2Handler) OIDCJwks(ctx *gin.Context) {
	jwks := o.oidcService.GetJwks()
	ctx.JSON(200, jwks)
}
