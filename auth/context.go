package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk/auth/db"
)

func StoreContextOAuth2Token(ctx *gin.Context, client *db.Oauth2Token) {
	ctx.Set(ContextOAuth2Token, client)
	ctx.Set(ContextUserID, client.UserID)
}

func LoadContextOAuth2Token(ctx context.Context) (*db.Oauth2Token, bool) {
	client, exists := ctx.Value(ContextOAuth2Token).(*db.Oauth2Token)
	return client, exists
}
func StoreContextAdmin(ctx *gin.Context, isAdmin bool) {
	ctx.Set(ContextIsAdmin, isAdmin)
}
func IsAdmin(ctx *gin.Context) bool {
	isAdmin, exists := ctx.Value(ContextIsAdmin).(bool)
	if exists && isAdmin {
		return true
	}
	return false
}
