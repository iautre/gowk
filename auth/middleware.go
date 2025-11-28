package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth/constant"
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
