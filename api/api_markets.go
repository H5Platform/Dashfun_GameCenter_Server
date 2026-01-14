package api

import (
	"dashfun_gamecenter/coingecko"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/openai_api"
	"dashfun_gamecenter/pricepredictcenter"
	"dashfun_gamecenter/web"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func apiUserGetTokenMarkets(c *gin.Context, user *data.DashFunUser) {
	ids := c.QueryArray("ids[]")
	tokenMarkets := make([]*coingecko.TokenMarketInfo, 0)
	for _, id := range ids {
		markets, err := coingecko.Get().GetMarketInfo(id)
		if err == nil && markets != nil {
			markets.Brief = openai_api.GetMarketSummarize().GetSummarize(id)
			tokenMarkets = append(tokenMarkets, markets)
		}
	}
	c.JSON(http.StatusOK, RSuccess(tokenMarkets))
}

// @Summary	获取代币聚合价格
// @Tags	Markets API
// @Produce	json
// @Param	symbol	path	string	true	"代币符号, 例如: BNB"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=number}	"聚合价格"
// @Router		/api/v1/markets/price/{symbol} [get]
func apiGetTokenPrice(c *gin.Context, user *data.DashFunUser) {
	symbol := c.Params.ByName("symbol")
	price, err := pricepredictcenter.Get().GetTokenPrice(symbol)
	if err != nil {
		c.JSON(http.StatusOK, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(price))
}

// @Summary	获取某个token的未来走势预测
// @Tags	Markets API
// @Produce	json
// @Param	symbol	path	string	true	"symbol , e.g., BTCUSDT"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=string}	"Forecast string"
// @Router		/api/v1/markets/forecast/{symbol} [get]
func apiUserGetTokenForecast(c *gin.Context, user *data.DashFunUser) {
	symbol := c.Params.ByName("symbol")
	marketUrl := config.GetConfig().ForecastConfig.Url
	if marketUrl == "" {
		marketUrl = "http://localhost:8080/"
	}
	marketUrl = marketUrl + "forecast/" + symbol + "/recent"

	resp, err := http.Get(marketUrl)
	if err != nil || resp.StatusCode != 200 {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError("failed to get forecast"))
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError("failed to read response body"))
		return
	}
	c.Data(http.StatusOK, "application/json", body)
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleMarkets, web.GET, "get", userHandlerAuthWrapper(apiUserGetTokenMarkets))
	web.GetService().RegisterApi(web.ApiModuleMarkets, web.GET, "forecast/:symbol", userHandlerAuthWrapper(apiUserGetTokenForecast))
	web.GetService().RegisterApi(web.ApiModuleMarkets, web.GET, "price/:symbol", userHandlerAuthWrapper(apiGetTokenPrice))

}
