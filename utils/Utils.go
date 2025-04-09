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

// NeedReset 判断传入的时间是否需要重置
// 重置时间为当日UTC0点
func NeedReset(t time.Time) bool {
	//Truncate到天，去掉时分秒
	return t.UTC().Before(time.Now().UTC().Truncate(24 * time.Hour))
}

// GetNextResetTime 获取下次重置时间
func GetNextResetTime() time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)
}

// DaysDifference 计算输入的毫秒数与当前时间相差的天数
// 只要不是同一天，相差就是1，以此类推，使用UTC时间
func DaysDifference(milliseconds int64) int {
	inputTime := time.UnixMilli(milliseconds).UTC()
	currentTime := time.Now().UTC()

	// 计算输入时间和当前时间的UTC日期
	inputDate := inputTime.Truncate(24 * time.Hour)
	currentDate := currentTime.Truncate(24 * time.Hour)

	// 计算日期差异
	daysDiff := int(currentDate.Sub(inputDate).Hours() / 24)
	if daysDiff < 0 {
		daysDiff = -daysDiff
	}
	return daysDiff
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
