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

	// Create handlers
	var u UserHandler
	var o OAuth2Handler

	// User endpoints
	ro.POST("/login", u.Login)
	//ro.GET("/auth/smscode", u.Smscode)

	// OAuth2 endpoints
	ro.GET("/oauth2/auth", u.BasicAuthMiddleware, o.OAuth2Auth)
	ro.POST("/oauth2/token", o.OAuth2Token)

	// OIDC endpoints
	ro.GET("/.well-known/openid_configuration", o.OIDCDiscovery)
	ro.GET("/oidc/userinfo", OAuth2TokenMiddleware, o.OIDCUserInfo)
	ro.GET("/oidc/jwks", o.OIDCJwks)

	return ro
}
