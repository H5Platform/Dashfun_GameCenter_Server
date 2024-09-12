package api

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GetCoinsResult struct {
	Coins    []*data.CoinData              `json:"coins"`
	UserData map[string]*data.CoinUserData `json:"user_data"`
}

// @Summary	获取可用的coins以及用户coin相关数据
// @Tags		Coin API
// @Produce	json
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=api.GetCoinsResult}	"coins"
// @Router		/api/v1/coin/get [get]
func apiGetCoins(c *gin.Context) {
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

	coins := coincenter.Get().GetAllCoins()
	userData := make(map[string]*data.CoinUserData)
	for _, coin := range coins {
		ud := coincenter.Get().GetCoinUserData(user.Id, coin.Id)
		userData[coin.Id] = ud
	}

	c.JSON(http.StatusOK, RSuccess(&GetCoinsResult{
		Coins:    coins,
		UserData: userData,
	}))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleCoin, web.GET, "get", apiGetCoins)
}
