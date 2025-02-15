package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
	"strings"
)

func Middleware(ctx *gin.Context) {
	defaultAuth.Middleware(ctx)
}

func (a *Auth) Middleware(ctx *gin.Context) {
	if err := a.checkTokenType(ctx); err != nil {
		ctx.Error(err)
		return
	}
}
func (a *Auth) checkTokenType(ctx *gin.Context) error {
	if has := a.getAuthType(a.Type, Brear.toString()); has {
		authorization := ctx.Request.Header.Get(gowk.AUTHORIZATION)
		if authorization != "" {
			err := a.checkAuthTypeBear(ctx, authorization)
			if err != nil {
				return err
			}
		}
	}
	if has := a.getAuthType(a.Type, Api.toString()); has {
		key := ctx.Request.Header.Get(a.Config.ApiKeyName)
		if key == "" {
			key = ctx.Query(a.Config.ApiKeyName)
		}
		if key != "" {
			err := a.checkAuthTypeApi(ctx, key)
			if err != nil {
				return err
			}
		}
	}
	return gowk.ERR_AUTH
}
func (a *Auth) checkAuthTypeBear(ctx *gin.Context, authorization string) error {
	authorizations := strings.Split(authorization, " ")
	if len(authorizations) == 1 {
		return gowk.ERR_AUTH
	}
	token := strings.Join(authorizations[1:], " ")
	claim, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return a.Config.JwtSecret, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			// ValidationErrorMalformed是一个uint常量，表示token不可用
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return gowk.ERR_AUTH
				// ValidationErrorExpired表示Token过期
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return gowk.ERR_AUTH
				// ValidationErrorNotValidYet表示无效token
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return gowk.ERR_AUTH
			} else {
				return gowk.ERR_AUTH
			}
		}
		return gowk.ERR_AUTH
	}
	claims, ok := claim.Claims.(*jwt.StandardClaims)
	if !ok || !claim.Valid {
		return gowk.ERR_AUTH
	}
	ctx.Set(U_ID, claims.Id)
	ctx.Next()
	return nil
}
func (a *Auth) checkAuthTypeApi(ctx *gin.Context, key string) error {
	id, ok := a.HandlerFunc.CheckApiKey(key)
	if !ok {
		return gowk.ERR_AUTH
	}
	ctx.Set(A_ID, id)
	return nil
}

func (a *Auth) getAuthType(ats []AuthType, tokenType string) bool {
	for _, v := range ats {
		if tokenType == v.toString() {
			return true
		}
	}
	return false
}
