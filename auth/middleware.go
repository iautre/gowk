package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
)

// OAuth2 token validation middleware
func OAuth2TokenMiddleware(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		gowk.Response(ctx, http.StatusUnauthorized, nil, gowk.NewError("Missing Authorization header"))
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		gowk.Response(ctx, http.StatusUnauthorized, nil, gowk.NewError("Invalid Authorization header format"))
		return
	}

	// Use OAuth2Service to validate token
	var oauth2Service OAuth2Service
	oauth2Token, err := oauth2Service.ValidateAccessToken(ctx, token)
	if err != nil {
		gowk.Response(ctx, http.StatusUnauthorized, nil, gowk.NewError("Invalid or expired access token"))
		return
	}

	// Store token info in context
	StoreContextOAuth2Token(ctx, oauth2Token)
	ctx.Next()
}

// AdminMiddleware checks if the user has admin privileges
func AdminMiddleware(ctx *gin.Context) {
	// Get user ID from context (should be set by OAuth2TokenMiddleware)
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		gowk.Response(ctx, http.StatusUnauthorized, nil, gowk.NewError("User not authenticated"))
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		gowk.Response(ctx, http.StatusUnauthorized, nil, gowk.NewError("Invalid user ID"))
		return
	}

	// Get user from database to check admin status
	var userService UserService
	user, err := userService.GetById(ctx, userID)
	if err != nil {
		gowk.Response(ctx, http.StatusUnauthorized, nil, gowk.NewError("User not found"))
		return
	}

	// Check if user is admin (assuming admin users have a specific group or status)
	// For now, we'll check if user belongs to "admin" group
	if user.Group.String != "admin" {
		gowk.Response(ctx, http.StatusForbidden, nil, gowk.NewError("Admin access required"))
		return
	}

	// Store admin status in context
	StoreContextAdmin(ctx, true)
	ctx.Next()
}
