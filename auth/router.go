package auth

import (
	"github.com/gin-gonic/gin"
)

func Router(r *gin.RouterGroup, relativePath ...string) {
	var ro *gin.RouterGroup
	if len(relativePath) > 0 {
		ro = r.Group(relativePath[0])
	} else {
		ro = r
	}
	var u UserHandler
	// ro.GET("/auth/token", a.Token)
	// ro.GET("/auth/qrcode", a.Qrcode)
	ro.GET("/auth/smscode", u.Smscode)
}
