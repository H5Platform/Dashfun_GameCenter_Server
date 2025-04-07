package api

import (
	"dashfun_gamecenter/RechargeCenter"
	"dashfun_gamecenter/apperrors"
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.uber.org/zap"
	"net/http"
)

type CreateOrderRequest struct {
	Platform            string `json:"platform" required:"true"`              //充值平台，ios, android, tdesktop, browser
	RechargeOptionIndex int    `json:"recharge_option_index" required:"true"` //充值选项索引，对应RechargeCfg中的索引
	Channel             string `json:"channel" required:"true"`               //当前充值渠道
	GameId              string `json:"game_id"`                               //当前所在的游戏，空串表示是DashFun平台
}

type RechargeOrderDetail struct {
	UserId      string                    `json:"user_id"`
	Photo       string                    `json:"photo"`
	DisplayName string                    `json:"display_name"`
	Balance     int                       `json:"balance"`
	Order       *data.DashFunRechargeData `json:"order"`
}

func apiGetRechargePlatformOptions(c *gin.Context, user *data.DashFunUser) {
	platform := c.Query("platform")
	if platform == "" {
		platform = "browser"
	}

	options := RechargeCenter.Get().GetRechargePlatformOptions(platform)
	c.JSON(http.StatusOK, RSuccess(options))
}

func apiRequestRechargeOrder(c *gin.Context, user *data.DashFunUser) {
	req := &CreateOrderRequest{}
	err := c.ShouldBindBodyWithJSON(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	idx := req.RechargeOptionIndex
	if idx < 0 || idx > len(config.GetConfig().RechargeCfg.Options) {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("invalid recharge option index"))
		return
	}

	if req.GameId != "" {
		//验证游戏
		_, err := gamecenter.Get().FindGame(req.GameId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
			return
		}
	}

	option := config.GetConfig().RechargeCfg.Options[idx]

	from := data.DashFunRechargeFrom_TG
	if req.Channel == "test" && !config.IsProd() {
		from = data.DashFunRechargeFrom_TEST
	}
	order, err := RechargeCenter.Get().CreateRechargeOrder(user.Id, option, req.Platform, req.GameId, from)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(order))
}

func apiCancelRechargeOrder(c *gin.Context, user *data.DashFunUser) {
	orderId := c.PostForm("id")
	if orderId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("order id is required"))
		return
	}

	_, err := RechargeCenter.Get().CancelRechargeOrder(orderId, user.Id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(true))
}

func apiGetRechargeOrder(c *gin.Context) {
	orderId := c.Param("id")
	order, err := RechargeCenter.Get().GetRechargeOrder(orderId)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(order))
}

func apiGetRechargeOrderDetails(c *gin.Context) {
	orderId := c.Param("id")
	order, err := RechargeCenter.Get().GetRechargeOrder(orderId)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	user, err := usercenter.Get().GetDashFunUser(order.UserId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	coin := coincenter.Get().GetDashFunDiamond()
	userCoinData := coincenter.Get().GetCoinUserData(order.UserId, coin.Id)

	detail := &RechargeOrderDetail{
		UserId:      user.Id,
		Photo:       user.AvatarUrl,
		DisplayName: user.DisplayName,
		Balance:     int(userCoinData.Amount),
		Order:       order,
	}

	c.JSON(http.StatusOK, RSuccess(detail))
}

func apiStripeCreateCheckoutSession(c *gin.Context) {
	orderId := c.PostForm("order_id")
	if orderId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("order id is required"))
		return
	}

	order, err := RechargeCenter.Get().GetRechargeOrder(orderId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	if order.PriceType != data.RechargePlatformOptionPriceTypeUSD || (order.Status != data.DashFunRechargeStatus_Created && order.Status != data.DashFunRechargeStatus_Pending) {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(apperrors.ErrRechargeOrderStatus.Error()))
		return
	}

	domain := config.GetConfig().StripeCfg.ReturnHost
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String(string(stripe.CurrencyUSD)),
					UnitAmount: stripe.Int64(int64(order.Price)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(fmt.Sprintf("%d Diamonds", order.Diamond)),
						Images:      []*string{stripe.String("https://res.dashfun.games/icons/dashfun-diamond.png")},
						Description: stripe.String("Use them on all DashFun Games!"),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"order_id": order.Id,
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(domain + "recharge/result/success"),
		CancelURL:  stripe.String(domain + "recharge/result/cancel"),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: map[string]string{
			"order_id": order.Id,
		},
	}
	s, err := session.New(params)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	RechargeCenter.Get().PendingRechargeOrder(orderId, "stripe")

	c.Redirect(http.StatusSeeOther, s.URL)
}

func apiStripeWebhook(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("failed to read request body"))
		return
	}
	endpointSecret := config.GetConfig().StripeCfg.WebhookKey
	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)
	zap.S().Infow("StripeWebhook", "type", event.Type, "err", err)
	if event.Type == stripe.EventTypePaymentIntentSucceeded {
		piId := event.Data.Object["id"].(string)
		metadata := event.Data.Object["metadata"]
		if metadata == nil {
			zap.S().Errorw("StripeWebhook", "EventTypePaymentIntentSucceeded", event.ID, "error", "metadata is nil")
			return
		}
		orderId := metadata.(map[string]interface{})["order_id"].(string)
		zap.S().Infow("StripeWebhook Event", "Event", "EventTypePaymentIntentSucceeded", "PID", piId, "order_id", orderId)
		RechargeCenter.Get().ConfirmRechargeOrder(orderId, piId)
	}
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleRecharge, web.GET, "options", userHandlerAuthWrapper(apiGetRechargePlatformOptions))
	web.GetService().RegisterApi(web.ApiModuleRecharge, web.POST, "order/create", userHandlerAuthWrapper(apiRequestRechargeOrder))
	web.GetService().RegisterApi(web.ApiModuleRecharge, web.POST, "order/cancel", userHandlerAuthWrapper(apiCancelRechargeOrder))

	web.GetService().RegisterApi(web.ApiModuleRecharge, web.GET, "order/:id", apiGetRechargeOrder)
	web.GetService().RegisterApi(web.ApiModuleRecharge, web.GET, "detail/:id", apiGetRechargeOrderDetails)
	web.GetService().RegisterApi(web.ApiModuleRecharge, web.POST, "stripe/checkout", apiStripeCreateCheckoutSession)
	web.GetService().RegisterApi(web.ApiModuleRecharge, web.POST, "stripe/webhook", apiStripeWebhook)
}
