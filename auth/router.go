package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
)

func Router(r *gin.RouterGroup, relativePath ...string) *gin.RouterGroup {
	var ro *gin.RouterGroup
	if len(relativePath) > 0 {
		ro = r.Group(relativePath[0])
	} else {
		ro = r
	}

	// Create handlers with context
	ctx := context.Background()
	u := NewUserHandler(ctx)
	o := NewOAuth2Handler(ctx)
	oc := NewOAuth2ClientHandler(ctx)

	// User endpoints
	ro.POST("/login", u.Login)
	ro.GET("/user/info", gowk.CheckLogin, u.UserInfo)
	ro.POST("/sso/login", u.SSOLogin)

	// Admin endpoints (require admin middleware)
	ro.POST("/user/:userId/reset-otp", gowk.CheckLogin, AdminMiddleware, u.ResetOTPCode)

	// OAuth2 endpoints
	ro.GET("/oauth2/auth", u.BasicAuthMiddleware, o.OAuth2Auth)
	ro.POST("/oauth2/token", o.OAuth2Token)

	// OIDC endpoints
	ro.GET("/.well-known/openid_configuration", o.OIDCDiscovery)
	ro.GET("/oidc/userinfo", OAuth2TokenMiddleware, o.OIDCUserInfo)
	ro.GET("/oidc/jwks", o.OIDCJwks)

	// OAuth2 Client Management endpoints (admin only)
	ro.POST("/oauth2/clients", OAuth2TokenMiddleware, AdminMiddleware, oc.CreateOAuth2Client)
	ro.GET("/oauth2/clients", OAuth2TokenMiddleware, AdminMiddleware, oc.ListOAuth2Clients)
	ro.GET("/oauth2/clients/:id", OAuth2TokenMiddleware, AdminMiddleware, oc.GetOAuth2Client)
	ro.PUT("/oauth2/clients/:id", OAuth2TokenMiddleware, AdminMiddleware, oc.UpdateOAuth2Client)
	ro.DELETE("/oauth2/clients/:id/disable", OAuth2TokenMiddleware, AdminMiddleware, oc.DisableOAuth2Client)
	ro.POST("/oauth2/clients/:id/regenerate-secret", OAuth2TokenMiddleware, AdminMiddleware, oc.RegenerateClientSecret)

	return ro
}
