package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk/auth/model"

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
	user, err := userService.Login(ctx, &params)
	if err != nil {
		ctx.Error(err)
		return
	}
	token, err := gowk.Login(ctx, user.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	gowk.Success(ctx, &LoginRes{
		Token:    token,
		UserId:   user.ID,
		Nickname: user.Nickname,
	})
}
func (u *UserHandler) Register(ctx *gin.Context) {
	var params RegisterParams
	err := ctx.ShouldBind(&params)
	if err != nil {
		ctx.Error(err)
		return
	}
}
func (u *UserHandler) UserInfo(ctx *gin.Context) {
	userId := gowk.LoginId(ctx)
	var userService UserService
	user, err := userService.GetById(ctx, userId)
	if err != nil {
		ctx.Error(err)
		return
	}
	gowk.Success(ctx, gowk.CopyByJson[model.User, UserRes](user))
}
func (u *UserHandler) Smscode(ctx *gin.Context) {
	params := &LoginParams{}
	err := ctx.ShouldBind(params)
	if err != nil {
		ctx.Error(err)
		return
	}
	var userService UserService
	user, err := userService.Login(ctx, params)
	if err != nil {
		ctx.Error(err)
		return
	}
	token, err := gowk.Login(ctx, user.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	gowk.Success(ctx, &LoginRes{
		Token:    token,
		UserId:   user.ID,
		Nickname: user.Nickname,
	})
}
