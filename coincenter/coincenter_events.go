package coincenter

import (
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"go.uber.org/zap"
)

func (c *CoinCenter) onUserLogin(user *data.OnlineUser) {
	//用户登录后读取用户coin列表数据
	_, err := c.loadUserCoins(user.User.Id)
	if err != nil {
		zap.S().Errorw("loadUserCoins failed", "user", user.User.Id, "err", err)
	}

	if config.IsTest() || config.IsDev() {
		//测试环境下，给用户赠送一些coin
		diamond := c.GetDashFunDiamond()
		userData := c.GetCoinUserData(user.User.Id, diamond.Id)
		if userData.Amount < 500 {
			c.AddUserCoinAmount(user.User.Id, diamond.Id, 1000)
		}
	}

}

func (c *CoinCenter) onUserOff(user *data.DashFunUser) {
	//用户登出后清理用户缓存
	c.users.RemoveCoinsUserData(user.Id)
}
