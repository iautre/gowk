package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/iautre/gowk"
)

type UserHandler struct {
}

var defaultUserHandler UserHandler

func Login() gin.HandlerFunc {
	return defaultUserHandler.Login
}
func UserInfo() gin.HandlerFunc {
	return defaultUserHandler.UserInfo
}

func (u *UserHandler) Login(ctx *gin.Context) {
	var params LoginParams
	err := ctx.ShouldBind(&params)
	if err != nil {
		ctx.Error(err)
		return
	}
	var userService UserService
	user := userService.Login(ctx, &params)
	token := gowk.Login(ctx, user.Id)
	gowk.Success(ctx, &LoginRes{
		Token:    token,
		UserId:   user.Id,
		Nickname: user.Nickname,
	})
}
func (u *UserHandler) UserInfo(ctx *gin.Context) {
	userId := gowk.LoginId[int64](ctx)
	var userService UserService
	user := userService.GetById(ctx, userId)
	gowk.Success(ctx, gowk.CopyByJson[User, UserRes](user))
}
func (u *UserHandler) Smscode(ctx *gin.Context) {
	params := &LoginParams{}
	err := ctx.ShouldBind(params)
	if err != nil {
		ctx.Error(err)
		return
	}
	var userService UserService
	user := userService.Login(ctx, params)
	token := gowk.Login(ctx, user.Id)
	gowk.Success(ctx, &LoginRes{
		Token:    token,
		UserId:   user.Id,
		Nickname: user.Nickname,
	})
}
