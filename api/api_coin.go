package api

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GetCoinsResult struct {
	Coins    []*data.CoinData              `json:"coins"`
	UserData map[string]*data.CoinUserData `json:"user_data"`
}

type GetCoinsRequest struct {
	Ids    []string `json:"ids"`
	IdType string   `json:"id_type"` // id类型，id or gameId，如果是gameId则返回绑定了该gameId的coin
}

// @Summary	获取可用的coins以及用户coin相关数据
// @Tags		Coin API
// @Produce	json
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=api.GetCoinsResult}	"coins"
// @Router		/api/v1/coin/get [get]
func apiGetCoins(c *gin.Context, user *data.DashFunUser) {
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

// @Summary	获取指定id的coin数据
// @Tags		Coin API
// @Produce	json
// @Param		ids	body	string	true	"coin id array"
// @Param		id_type	body	string	true	"coin id type"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=data.CoinData}	"coin"
// @Router		/api/v1/coin/list [post]
func apiGetCoinsData(c *gin.Context, user *data.DashFunUser) {
	var req GetCoinsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	coinIds := req.Ids
	coins := make([]*data.CoinData, 0)

	f := coincenter.Get().GetCoinById
	if req.IdType == "gameId" {
		f = coincenter.Get().GetCoinByGame
	}

	for _, coinId := range coinIds {
		if req.IdType == "gameId" && (coinId == "" || coinId == "DashFun") {
			//获取DashFunPoint
			dashFunCoins := coincenter.Get().GetDashFunCoins()
			coins = append(coins, dashFunCoins...)
		} else if coin, ok := f(coinId); ok {
			coins = append(coins, coin)
		}
	}
	c.JSON(http.StatusOK, RSuccess(coins))
}

// @Summary	获取用户coin数据
// @Tags		Coin API
// @Produce	json
// @Param		coin_id	query	string	true	"coin id array"
// @Authorize "tma {token}"
// @Success	200	{object}	api.JSONResult{data=map[string]*data.CoinUserData}	"coin user data"
// @Router		/api/v1/coin/user_data [post]

func apiGetUserCoinsData(c *gin.Context, user *data.DashFunUser) {
	var req GetCoinsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	coinIds := req.Ids

	f := coincenter.Get().GetCoinById
	if req.IdType == "gameId" {
		f = coincenter.Get().GetCoinByGame
	}

	userData := make(map[string]*data.CoinUserData)
	for _, coinId := range coinIds {
		if req.IdType == "gameId" && (coinId == "" || coinId == "DashFun") {
			//获取DashFunPoint
			dashFunCoins := coincenter.Get().GetDashFunCoins()
			for _, coin := range dashFunCoins {
				ud := coincenter.Get().GetCoinUserData(user.Id, coin.Id)
				userData[coin.Id] = ud
			}
		} else if coin, ok := f(coinId); ok {
			ud := coincenter.Get().GetCoinUserData(user.Id, coin.Id)
			userData[coin.Id] = ud
		}
	}
	c.JSON(http.StatusOK, RSuccess(userData))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleCoin, web.GET, "get", userHandlerAuthWrapper(apiGetCoins))
	web.GetService().RegisterApi(web.ApiModuleCoin, web.POST, "list", userHandlerAuthWrapper(apiGetCoinsData))
	web.GetService().RegisterApi(web.ApiModuleCoin, web.POST, "user_data", userHandlerAuthWrapper(apiGetUserCoinsData))
}
