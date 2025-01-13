package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

func IsSameDay(t1, t2 time.Time) bool {
	t1d := t1.YearDay()
	t1y := t1.Year()
	t2d := t2.YearDay()
	t2y := t2.Year()
	return t1y == t2y && t1d == t2d
}

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
