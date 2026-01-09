package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth/constant"
	"github.com/iautre/gowk/auth/db"
)

func CheckAppMiddleware(ctx *gin.Context) {
	key := ctx.Request.Header.Get(gowk.APPKEY)
	if key == "" {
		key = ctx.Query(gowk.APPKEY)
	}
	if key == "" {
		ctx.Error(gowk.ERR_AUTH)
		ctx.Abort()
		return
	}
	var appService AppService
	app, err := appService.GetByKey(ctx, key)
	if err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}
	ctx.Set(constant.CONTEXT_APP_KEY, app)
}

// OAuth2 token validation middleware
func OAuth2TokenMiddleware(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.Error(gowk.NewError("Missing Authorization header"))
		ctx.Abort()
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		ctx.Error(gowk.NewError("Invalid Authorization header format"))
		ctx.Abort()
		return
	}

	// Validate token from database
	queries := db.New(gowk.DB(ctx))
	oauth2Token, err := queries.GetOAuth2Token(ctx, token)
	if err != nil {
		ctx.Error(gowk.NewError("Invalid or expired access token"))
		ctx.Abort()
		return
	}

	// Store token info in context
	ctx.Set("oauth2_token", oauth2Token)
	ctx.Set("user_id", oauth2Token.UserID)
	ctx.Next()
}

// OAuth2 scope validation middleware
func OAuth2ScopeMiddleware(requiredScopes []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get token from context (must be set by OAuth2TokenMiddleware)
		tokenInterface, exists := ctx.Get("oauth2_token")
		if !exists {
			ctx.Error(gowk.NewError("OAuth2 token not found in context"))
			ctx.Abort()
			return
		}

		token := tokenInterface.(db.Oauth2Token)
		tokenScopes := strings.Fields(token.Scope.String)

		// Check if all required scopes are present
		for _, requiredScope := range requiredScopes {
			found := false
			for _, tokenScope := range tokenScopes {
				if tokenScope == requiredScope {
					found = true
					break
				}
			}
			if !found {
				ctx.Error(gowk.NewError("Insufficient scope: " + requiredScope + " required"))
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}
