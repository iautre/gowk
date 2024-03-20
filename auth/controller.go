package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/iautre/gowk"
)

type UserController struct {
}

var defaultUserController UserController

func Login() gin.HandlerFunc {
	return defaultUserController.Login
}
func UserInfo() gin.HandlerFunc {
	return defaultUserController.UserInfo
}

func (u *UserController) Login(ctx *gin.Context) {
	var params LoginParams
	err := ctx.ShouldBind(&params)
	if err != nil {
		gowk.Panic(gowk.ERR, err)
	}
	var userService UserService
	user := userService.Login(ctx, &params)
	token := gowk.Login(ctx, user.Id)
	userService.UpdateToken(ctx, user.Id, token)
	gowk.Success(ctx, &LoginRes{
		Token:    token,
		UserId:   user.Id,
		Nickname: user.Nickname,
	})
}
func (u *UserController) UserInfo(ctx *gin.Context) {
	userId := gowk.LoginId[int64](ctx)
	var userService UserService
	user := userService.GetById(ctx, userId)
	gowk.Success(ctx, gowk.CopyByJson[User, UserRes](user))
}
func (u *UserController) Smscode(ctx *gin.Context) {
	params := &LoginParams{}
	err := ctx.ShouldBind(params)
	if err != nil {
		gowk.Panic(gowk.ERR, err)
	}
	var userService UserService
	user := userService.Login(ctx, params)
	gowk.Success(ctx, &LoginRes{
		Token:    user.Token,
		UserId:   user.Id,
		Nickname: user.Nickname,
	})
}
