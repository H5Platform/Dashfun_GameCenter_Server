package coincenter

import (
	"dashfun_gamecenter/datasource/data"
	"go.uber.org/zap"
)

func (c *CoinCenter) onUserLogin(user *data.OnlineUser) {
	//用户登录后读取用户coin列表数据
	_, err := c.loadUserCoins(user.User.Id)
	if err != nil {
		zap.S().Errorw("loadUserCoins failed", "user", user.User.Id, "err", err)
	}
}

func (c *CoinCenter) onUserOff(user *data.DashFunUser) {
	//用户登出后清理用户缓存
	c.users.RemoveCoinsUserData(user.Id)
}
