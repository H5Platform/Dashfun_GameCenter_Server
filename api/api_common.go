package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var locker = utils.NewKeyRWMutexManager[string]("user_locker", 5*time.Minute)

func userHandlerAuthWrapper(handler func(ctx *gin.Context, user *data.DashFunUser)) func(*gin.Context) {
	return func(c *gin.Context) {
		auth, err := utils.CheckAuthorize(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
			return
		}

		// 同步锁用户token
		guard := locker.Lock(auth.Token)
		defer guard.Unlock()

		user, err := usercenter.Get().GetDashFunUserByAuthData(auth, false)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
			return
		}

		if handler != nil {
			handler(c, user)
		}
	}
}
