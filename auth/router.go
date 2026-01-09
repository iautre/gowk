package auth

import (
	"github.com/gin-gonic/gin"
)

func Router(r *gin.RouterGroup, relativePath ...string) *gin.RouterGroup {
	var ro *gin.RouterGroup
	if len(relativePath) > 0 {
		ro = r.Group(relativePath[0])
	} else {
		ro = r
	}
	var u UserHandler
	ro.POST("/login", u.Login)
	// ro.GET("/auth/token", a.Token)
	// ro.GET("/auth/qrcode", a.Qrcode)
	ro.GET("/auth/smscode", u.Smscode)

	// OAuth2 endpoints
	ro.GET("/oauth2/auth", u.BasicAuthMiddleware, u.OAuth2Auth)
	ro.POST("/oauth2/token", u.OAuth2Token)
	//
	//// SSO endpoints
	//ro.POST("/sso/login", SSOLogin())

	// OIDC endpoints
	ro.GET("/.well-known/openid_configuration", u.OIDCDiscovery)
	ro.GET("/oidc/userinfo", OAuth2TokenMiddleware, u.OIDCUserInfo)
	ro.GET("/oidc/jwks", u.OIDCJwks)

	return ro
}
