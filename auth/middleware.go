package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth/respository"
)

func CheckAppMiddleware(ctx *gin.Context) {
	key := ctx.Request.Header.Get(gowk.APPKEY)
	if key == "" {
		key = ctx.Query(gowk.APPKEY)
	}
	if key == "" {
		ctx.Error(gowk.ERR_AUTH)
		return
	}
	repository := respository.NewAppRepository()
	app, err := repository.GetByKey(ctx, key)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.Set(CONTEXT_APP_KEY, app)
}
