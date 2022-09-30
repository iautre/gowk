package auth

import (
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
)

func Middleware(ctx *gin.Context) {
	defaultAuth.Middleware(ctx)
}

func (a *Auth) Middleware(ctx *gin.Context) {
	authorization := ctx.Request.Header.Get(gowk.AUTHORIZATION)
	if authorization != "" {
		authorizations := strings.Split(authorization, " ")
		tokenType := strings.ToLower(authorizations[0])
		if has := a.checkAuthType(a.Type, tokenType); has {
			token := strings.Join(authorizations[1:], " ")
			switch tokenType {
			// case Basic.toString(): //基础认证
			// 	username, password, ok := ctx.Request.BasicAuth()
			// 	if !ok {
			// 		gowk.Panic(gowk.ERR_AUTH, "获取认证信息异常")
			// 	}
			// 	id, ok := a.HandlerFunc.CheckByUsernameAndPassword(username, password)
			// 	if !ok {
			// 		gowk.Panic(gowk.ERR_AUTH, "校验失败")
			// 	}
			// 	//校验成功
			// 	ctx.Set(U_ID, id)
			// 	ctx.Next()
			// 	return
			case Brear.toString():
				claim, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
					return a.Config.JwtSecret, nil
				})
				if err != nil {
					if ve, ok := err.(*jwt.ValidationError); ok {
						// ValidationErrorMalformed是一个uint常量，表示token不可用
						if ve.Errors&jwt.ValidationErrorMalformed != 0 {
							gowk.Panic(gowk.ERR_AUTH, "token不可用")
							// ValidationErrorExpired表示Token过期
						} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
							gowk.Panic(gowk.ERR_AUTH, "token过期")
							// ValidationErrorNotValidYet表示无效token
						} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
							gowk.Panic(gowk.ERR_AUTH, "无效的token")
						} else {
							gowk.Panic(gowk.ERR_AUTH, "token不可用")
						}
					}
					gowk.Panic(gowk.ERR_AUTH, "token失败")
				}
				claims, ok := claim.Claims.(*jwt.StandardClaims)
				if !ok || !claim.Valid {
					gowk.Panic(gowk.ERR_AUTH, "校验失败了")
				}
				//校验成功了
				ctx.Set(U_ID, claims.Id)
				ctx.Next()
			default:
			}
		}
	}
	if has := a.checkAuthType(a.Type, Api.toString()); has {
		key := ctx.Request.Header.Get(a.Config.ApiKeyName)
		if key == "" {
			key = ctx.Query(a.Config.ApiKeyName)
		}
		id, ok := a.HandlerFunc.CheckApiKey(key)
		if !ok {
			gowk.Panic(gowk.ERR_AUTH, "校验失败")
		}
		//校验成功
		ctx.Set(A_ID, id)
		ctx.Next()
		return
	}
	gowk.Panic(gowk.ERR_AUTH, "校验失败")
}

func (a *Auth) checkAuthType(ats []AuthType, tokenType string) bool {
	for _, v := range ats {
		if tokenType == v.toString() {
			return true
		}
	}
	return false
}
