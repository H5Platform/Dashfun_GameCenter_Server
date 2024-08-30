package api

import (
	"errors"
	"github.com/gin-gonic/gin"
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
