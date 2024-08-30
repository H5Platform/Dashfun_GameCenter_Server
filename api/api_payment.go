package api

import (
	"dashfun_gamecenter/paymentcenter"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type PaymentRequest struct {
	GameId  string `form:"game_id" binding:"required"`
	Title   string `form:"title" binding:"required"`
	Desc    string `form:"desc" binding:"required"`
	Payload string `form:"payload" binding:"required"`
	Price   int    `form:"price" binding:"required"`
}

type GetPaymentRequest struct {
	GameId    string `form:"game_id" binding:"required"`
	UserId    string `form:"user_id" binding:"required"`
	PaymentId string `form:"payment_id" binding:"required"`
}

// @Summary	telegram用户请求支付订单
// @Tags		Payment API
// @Produce	json
// @Param	game_id query string true "请求支付的游戏Id"
// @Param	title query string true "支付项目名称"
// @Param	desc query string true "支付项目描述"
// @Param	payload query string true "自定义数据"
// @Param	price query int true "支付金额"
// @Authorize "tma {token}"
// @Success	200		{object}	api.JSONResult{paymentId:string, invoiceLink:string}	"payment info"
// @Router		/api/v1/payment/request [get]
func apiUserRequestPayment(c *gin.Context) {
	req := &PaymentRequest{}
	err := c.ShouldBindQuery(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	tgAuthData := c.GetHeader("authorization")
	authParts := strings.Split(tgAuthData, " ")
	if len(authParts) < 2 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError("Unauthorized"))
		return
	}

	user, err := usercenter.Get().GetDashFunUserByTgAuthData(authParts[1])
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}

	payment, err := paymentcenter.Get().RequestTGPayment(user.Id, req.GameId, req.Title, req.Desc, req.Price)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(gin.H{
		"paymentId":   payment.Id,
		"invoiceLink": payment.ExtraData,
	}))
}

// @Summary	获取用户的支付订单
// @Tags		Payment API
// @Produce	json
// @Param	game_id query string true "游戏Id"
// @Param	user_id query string true "用户Id"
// @Param	payment_id query string true "订单ID"
// @Success	200		{object}	api.JSONResult{data=[]data.DashFunPaymentData}	"DashFunPaymentData"
// @Router		/api/v1/payment/get [get]
func apiGetPayment(c *gin.Context) {
	req := &GetPaymentRequest{}
	err := c.ShouldBindQuery(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	payment, err := paymentcenter.Get().FindPayment(req.PaymentId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	if payment.GameId != req.GameId || payment.UserId != req.UserId {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("gameId or userId not match"))
		return
	}
	c.JSON(http.StatusOK, RSuccess(payment))
}

func init() {
	web.GetService().RegisterApi(web.ApiModulePayment, web.GET, "request", apiUserRequestPayment)
	web.GetService().RegisterApi(web.ApiModulePayment, web.GET, "get", apiGetPayment)
}
