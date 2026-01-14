package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/pricepredictcenter"
	"dashfun_gamecenter/web"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary	用户投注
// @Tags	Price Predict API
// @Produce	json
// @Param	price		query	number	true	"Predicted Price"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=data.PricePredictData} "Prediction Record"
// @Router		/api/v1/price_predict/bet [post]
func apiUserBet(c *gin.Context, user *data.DashFunUser) {
	if !pricepredictcenter.Get().IsOpen() {
		c.JSON(http.StatusOK, RError("price prediction is closed"))
		return
	}
	// Simple parsing, or enhance with strconv
	// Since c.Query returns string, we need to parse it or binding
	// Let's use robust parsing if available, or just struct binding if we switch to POST body.
	// Query params fine for now.

	// A helper to parse float from query
	// For simplicity, let's use a struct binding or manual parsing
	type BetRequest struct {
		Price     float64 `json:"price"`
		BetAmount int64   `json:"bet_amount"`
	}
	var req BetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, RError("invalid parameters"))
		return
	}

	if req.Price <= 0 {
		c.JSON(http.StatusOK, RError("invalid price"))
		return
	}
	// Note: bet_amount validation is done in center logic, but we could do basic check here too
	if req.BetAmount <= 0 {
		c.JSON(http.StatusOK, RError("invalid bet amount"))
		return
	}

	result, err := pricepredictcenter.Get().CreateOrUpdateUserPredict(user.Id, req.Price, req.BetAmount)
	if err != nil {
		// Differentiate errors if possible
		c.JSON(http.StatusOK, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(result))
}

// @Summary	获取用户投注信息
// @Tags	Price Predict API
// @Produce	json
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=data.PricePredictData} "Prediction Record"
// @Router		/api/v1/price_predict/info [get]
func apiUserInfo(c *gin.Context, user *data.DashFunUser) {
	if !pricepredictcenter.Get().IsOpen() {
		c.JSON(http.StatusOK, RError("price prediction is closed"))
		return
	}
	record, err := pricepredictcenter.Get().GetUserPredict(user.Id)
	if err != nil {
		c.JSON(http.StatusOK, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(record))
}

// @Summary	用户领奖
// @Tags	Price Predict API
// @Produce	json
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=data.PricePredictData} "Updated Record"
// @Router		/api/v1/price_predict/claim [post]
func apiUserClaim(c *gin.Context, user *data.DashFunUser) {
	if !pricepredictcenter.Get().IsOpen() {
		c.JSON(http.StatusOK, RError("price prediction is closed"))
		return
	}
	record, err := pricepredictcenter.Get().ClaimReward(user.Id)
	if err != nil {
		// Ideally map errors to apperrors
		c.JSON(http.StatusOK, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(record))
}

// @Summary	获取价格预测配置信息
// @Tags	Price Predict API
// @Produce	json
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=map[string]interface{}} "Config Info"
// @Router		/api/v1/price_predict/config [get]
func apiConfigInfo(c *gin.Context, user *data.DashFunUser) {
	info, err := pricepredictcenter.Get().GetConfigInfo()
	if err != nil {
		c.JSON(http.StatusOK, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(info))
}

func init() {
	web.GetService().RegisterApi(web.ApiModulePricePredict, web.POST, "bet", userHandlerAuthWrapper(apiUserBet))
	web.GetService().RegisterApi(web.ApiModulePricePredict, web.GET, "info", userHandlerAuthWrapper(apiUserInfo))
	web.GetService().RegisterApi(web.ApiModulePricePredict, web.POST, "claim", userHandlerAuthWrapper(apiUserClaim))
	web.GetService().RegisterApi(web.ApiModulePricePredict, web.GET, "config", userHandlerAuthWrapper(apiConfigInfo))
}
