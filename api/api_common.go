package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func userHandlerAuthWrapper(handler func(ctx *gin.Context, user *data.DashFunUser)) func(*gin.Context) {
	return func(c *gin.Context) {
		auth, err := utils.CheckAuthorize(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
			return
		}

		user, err := usercenter.Get().GetDashFunUserByTgAuthData(auth.Token, false)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
			return
		}

		if handler != nil {
			handler(c, user)
		}
	}
}
