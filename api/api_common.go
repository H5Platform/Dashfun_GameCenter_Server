package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/usercenter"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type AuthData struct {
	Method string
	Token  string
}

// CheckAuthorize 检查请求中的authorization header，返回验证信息
func CheckAuthorize(c *gin.Context) (*AuthData, error) {
	tgAuthData := c.GetHeader("authorization")
	authParts := strings.Split(tgAuthData, " ")
	if len(authParts) < 2 {
		return nil, errors.New("unauthorized")
	}
	return &AuthData{
		Method: authParts[0],
		Token:  authParts[1],
	}, nil
}

func userHandlerAuthWrapper(handler func(ctx *gin.Context, user *data.DashFunUser)) func(*gin.Context) {
	return func(c *gin.Context) {
		auth, err := CheckAuthorize(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
			return
		}

		user, err := usercenter.Get().GetDashFunUserByTgAuthData(auth.Token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
			return
		}

		if handler != nil {
			handler(c, user)
		}
	}
}
