package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/exchangecenter"
	"dashfun_gamecenter/web"
	"net/http"

	"github.com/gin-gonic/gin"
)

// apiGetExchangeInfo
// @Summary 获取积分兑换活动信息
// @Description 获取当前活动的配置、状态、额度使用情况
// @Tags Exchange
// @Produce json
// @Authorize	"tma {token}"
// @Success 200 {object} api.JSONResult{data=exchangecenter.ExchangeActivityInfo}
// @Router /api/v1/exchange/info [get]
func apiGetExchangeInfo(c *gin.Context, user *data.DashFunUser) {
	info, err := exchangecenter.Get().GetActivityInfo(c.Request.Context(), user.Id)
	if err != nil {
		c.JSON(http.StatusOK, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(info))
}

// apiDoExchange
// @Summary 执行积分兑换
// @Description 消耗积分兑换Token
// @Tags Exchange
// @Produce json
// @Authorize	"tma {token}"
// @Param request body struct{Amount int64 `json:"amount"`; WalletAddress string `json:"wallet_address"`} true "Exchange Request"
// @Success 200 {object} api.JSONResult{data=float64} "Token Amount Received"
// @Router /api/v1/exchange/do [post]
func apiDoExchange(c *gin.Context, user *data.DashFunUser) {
	type ExchangeRequest struct {
		Amount        int64  `json:"amount"`         // Points to exchange
		WalletAddress string `json:"wallet_address"` // User wallet address
	}
	var req ExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, RError("invalid parameters"))
		return
	}

	tokenAmount, err := exchangecenter.Get().Exchange(c.Request.Context(), user.Id, req.Amount, req.WalletAddress, false)
	if err != nil {
		c.JSON(http.StatusOK, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(tokenAmount))
}

// apiGetExchangeHistory
// @Summary 获取积分兑换历史
// @Description 获取最近50条兑换记录
// @Tags Exchange
// @Produce json
// @Authorize	"tma {token}"
// @Success 200 {object} api.JSONResult{data=[]data.ExchangeLog}
// @Router /api/v1/exchange/history [get]
func apiGetExchangeHistory(c *gin.Context, user *data.DashFunUser) {
	history, err := exchangecenter.Get().GetHistory(c.Request.Context(), user.Id)
	if err != nil {
		c.JSON(http.StatusOK, RError(err.Error()))
		return
	}
	// Return empty slice instead of null if empty
	if history == nil {
		history = []*data.ExchangeLog{}
	}
	c.JSON(http.StatusOK, RSuccess(history))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleExchange, web.GET, "info", userHandlerAuthWrapper(apiGetExchangeInfo))
	web.GetService().RegisterApi(web.ApiModuleExchange, web.GET, "history", userHandlerAuthWrapper(apiGetExchangeHistory))
	web.GetService().RegisterApi(web.ApiModuleExchange, web.POST, "do", userHandlerAuthWrapper(apiDoExchange))
}
