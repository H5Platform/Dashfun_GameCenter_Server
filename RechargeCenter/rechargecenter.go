package RechargeCenter

import (
	"context"
	"dashfun_gamecenter/apperrors"
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/tgbot"
	"dashfun_gamecenter/usercenter"
	"errors"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/stripe/stripe-go/v81"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var instance *RechargeCenter

// RechargeCenter 充值中心
type RechargeCenter struct {
	idGen               *snowflake.Worker
	processingOrdersChn chan *data.DashFunRechargeData //已经支付完毕，还没有发放钻石的订单队列
}

var telegramPlatforms = []string{"ios", "android", "tdesktop"}

func isTelegram(platform string) bool {
	for _, p := range telegramPlatforms {
		if strings.EqualFold(platform, p) {
			return true
		}
	}
	return false
}

func Get() *RechargeCenter {
	once.Do(func() {
		instance = &RechargeCenter{}
		instance.init()
	})
	return instance
}

func (r *RechargeCenter) init() {
	r.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerRechargeId))
	//init stripe
	stripe.Key = config.GetConfig().StripeCfg.SecretKey
	r.processingOrdersChn = make(chan *data.DashFunRechargeData, 100)
	go r.processPaidOrders()

	paidOrders, err := dao.GetRechargeDao().GetOrdersByStatus(data.DashFunRechargeStatus_Paid)
	if err != nil {
		zap.S().Errorw("Get Paid Orders Failed", "error", err)
	} else {
		for _, order := range paidOrders {
			r.processingOrdersChn <- order
		}
	}

	events.TGPreCheckoutQueryEvents.On(r.onTGPreCheckoutQueryEvent)
	events.TGSuccessfulPaymentEvents.On(r.onTGPaymentSuccessEvent)
}

func (r *RechargeCenter) onTGPreCheckoutQueryEvent(query *models.PreCheckoutQuery) {
	rechargeId := query.InvoicePayload

	//rc开头的才是充值订单
	if !strings.HasPrefix(rechargeId, "rc") {
		return
	}

	rechargeOrder, err := r.GetRechargeOrder(rechargeId)
	b := tgbot.Bot()
	if err != nil {
		zap.S().Errorw("get recharge data error", "error", err, "rechargeId", rechargeId, "preCheckOutQuery", query)
		b.AnswerPreCheckoutQuery(context.TODO(), &bot.AnswerPreCheckoutQueryParams{
			PreCheckoutQueryID: query.ID,
			OK:                 false,
			ErrorMessage:       err.Error(),
		})
		return
	}

	_, err = b.AnswerPreCheckoutQuery(context.TODO(), &bot.AnswerPreCheckoutQueryParams{
		PreCheckoutQueryID: query.ID,
		OK:                 true,
		ErrorMessage:       "",
	})

	if err == nil {
		rechargeOrder.Status = data.DashFunRechargeStatus_Pending
		dao.GetRechargeDao().SaveOrUpdate(rechargeOrder)
	} else {
		zap.S().Errorw("Answer Pre Checkout Query Failed", "error", err, "rechargeId", rechargeId, "preCheckOutQuery", query)
		rechargeOrder.Status = data.DashFunRechargeStatus_Failed
		rechargeOrder.Message = err.Error()
		dao.GetRechargeDao().SaveOrUpdate(rechargeOrder)
	}
}

func (r *RechargeCenter) onTGPaymentSuccessEvent(message *models.Message) {
	paymentId := message.SuccessfulPayment.InvoicePayload

	//rc开头的才是充值订单
	if !strings.HasPrefix(paymentId, "rc") {
		return
	}
	payment, err := r.GetRechargeOrder(paymentId)
	if err != nil {
		zap.S().Errorw("SuccessfulPayment Received, recharge data get error", "rechargeId", paymentId, "error", err, "Message", message)
		return
	}

	if payment.Status != data.DashFunRechargeStatus_Pending {
		zap.S().Errorw("SuccessfulPayment Received, recharge data status error", "rechargeId", paymentId, "error", "status incorrect", "Message", message)
		return
	}

	payment.Status = data.DashFunRechargeStatus_Paid
	payment.Message = message.SuccessfulPayment.TelegramPaymentChargeID
	payment.PaidAt = time.Now().UnixMilli()
	_, err = dao.GetRechargeDao().SaveOrUpdate(payment)
	if err != nil {
		zap.S().Errorw("Save Recharge Data Failed", "Recharge", payment, "error", err, "Message", message)
		return
	}
	b := tgbot.Bot()
	b.SendMessage(context.TODO(), &bot.SendMessageParams{
		ChatID: message.Chat.ID,
		Text:   "Thank you for your purchase\n" + strconv.Itoa(message.SuccessfulPayment.TotalAmount) + " Stars"},
	)
	//发送订单到处理队列
	r.processingOrdersChn <- payment

}
func (r *RechargeCenter) processPaidOrders() {
	diamond := coincenter.Get().GetDashFunDiamond()
	for {
		order := <-r.processingOrdersChn
		//发放钻石
		_, err := coincenter.Get().AddUserCoinAmount(order.UserId, diamond.Id, int32(order.Diamond), "TopUp", order.Id)
		if err != nil {
			zap.S().Errorw("AddUserCoinAmount failed", "order", order.Id, "err", err)
		}
		//更新order
		order.Status = data.DashFunRechargeStatus_Completed
		dao.GetRechargeDao().SaveOrUpdate(order)

		user, _ := usercenter.Get().GetDashFunUser(order.UserId)
		var game *data.DashFunGame = nil

		if order.GameId != "" {
			game, _ = dao.GetGameDao().GetGameById(order.GameId)
		}

		zap.S().Infow("recharge order completed", "user", order.UserId, "order", order.Id, "price", order.Price, "diamond", order.Diamond)
		events.UserRechargeEvents.Emit(&events.EventUserRecharge{
			User:     user,
			Game:     game,
			Recharge: order,
		})
	}
}

func (r *RechargeCenter) newRechargeOrderId() string {
	id := r.idGen.NextId()
	return "rc" + strconv.FormatInt(id, 36)
}

func (r *RechargeCenter) GetRechargePlatformOptions(platform string) RechargePlatformOptions {
	var options []RechargePlatformOption
	priceType := data.RechargePlatformOptionPriceTypeUSD

	if isTelegram(platform) && config.GetConfig().RechargeCfg.EnableStar {
		priceType = data.RechargePlatformOptionPriceTypeTGStar //目前ios和android都是tg的miniapp用户，使用星星支付
	}

	for _, option := range config.GetConfig().RechargeCfg.Options {
		price := option.Price
		if isTelegram(platform) && config.GetConfig().RechargeCfg.EnableStar { //目前ios和android都是tg的miniapp用户，使用星星支付
			price = option.TGStar
		}
		options = append(options, RechargePlatformOption{
			Price:    price,
			Diamond:  option.Diamond,
			PriceOff: option.PriceOff,
		})
	}

	ret := RechargePlatformOptions{
		PriceType: priceType,
		Options:   options,
	}

	return ret
}

func (r *RechargeCenter) CreateRechargeOrder(userId string, rechargeOption config.RechargeOption, platform, gameId string, from data.RechargeFrom) (*data.DashFunRechargeData, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()

	d := dao.GetRechargeDao()
	//user, err := usercenter.Get().GetDashFunUser(userId)
	//
	//if err != nil {
	//	zap.S().Errorw("CreateRechargeOrder Get User Failed", "UserId", userId, "error", err)
	//	return nil, err
	//}
	finalPrice, priceType := r.GetRechargePrice(rechargeOption, platform)
	recharge, err := d.CreateRecharge(r.newRechargeOrderId(), userId, from, gameId, finalPrice, priceType, rechargeOption.Diamond, "", "", time.Now().Unix())
	if err != nil {
		zap.S().Infow("CreateRechargeOrder", "recharge", recharge)
		return nil, err
	}
	if recharge == nil {
		return nil, apperrors.ErrRechargeOrderCreateFailed
	}

	if priceType == data.RechargePlatformOptionPriceTypeTGStar {
		title := fmt.Sprintf("%d Diamonds", rechargeOption.Diamond)
		//tgstar支付，直接向bot请求payment
		invoiceLink, err := tgbot.Bot().CreateInvoiceLink(ctx, &bot.CreateInvoiceLinkParams{
			Title:         title,
			Description:   fmt.Sprintf("%d Diamonds", rechargeOption.Diamond),
			Payload:       recharge.Id,
			ProviderToken: "",
			Currency:      "XTR",
			Prices: []models.LabeledPrice{
				{
					Label:  title,
					Amount: finalPrice,
				},
			},
		})

		if err != nil {
			zap.S().Errorw("CreateRechargeOrder Create Invoice Failed", "UserId", userId, "error", err)
			return nil, err
		}

		recharge.ChannelPayId = invoiceLink
		d.SaveOrUpdate(recharge)
	}

	return recharge, err
}

func (r *RechargeCenter) GetRechargePrice(rechargeOption config.RechargeOption, platform string) (int, data.RechargePlatformOptionPriceType) {
	price := rechargeOption.Price
	priceType := data.RechargePlatformOptionPriceTypeUSD
	if isTelegram(platform) && config.GetConfig().RechargeCfg.EnableStar {
		price = rechargeOption.TGStar
		priceType = data.RechargePlatformOptionPriceTypeTGStar
	}
	finalPrice := price
	if rechargeOption.PriceOff > 0 {
		finalPrice = price * (1000 - rechargeOption.PriceOff) / 1000
	}
	return finalPrice, priceType
}

func (r *RechargeCenter) GetRechargeOrder(id string) (*data.DashFunRechargeData, error) {
	order, err := dao.GetRechargeDao().FindRechargeById(id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrRechargeOrderNotFound
		} else {
			return nil, err
		}
	}
	return order, nil
}

func (r *RechargeCenter) PendingRechargeOrder(orderId string, payFrom string, payOrderId string) (*data.DashFunRechargeData, error) {
	order, err := r.GetRechargeOrder(orderId)
	if err != nil {
		return nil, err
	}

	if order.Status != data.DashFunRechargeStatus_Created {
		return nil, apperrors.ErrRechargeOrderStatus
	}

	order.Status = data.DashFunRechargeStatus_Pending
	order.PayFrom = payFrom

	order.ChannelPayId = payOrderId

	zap.S().Infow("PendingRechargeOrder", "orderId", orderId, "payFrom", payFrom, "payOrderId", payOrderId)

	return dao.GetRechargeDao().SaveOrUpdate(order)
}

func (r *RechargeCenter) ConfirmRechargeOrder(orderId string, channelOrderId string) (*data.DashFunRechargeData, error) {
	order, err := r.GetRechargeOrder(orderId)
	if err != nil {
		return nil, err
	}

	if order.Status != data.DashFunRechargeStatus_Pending {
		return nil, apperrors.ErrRechargeOrderStatus
	}

	order.ChannelPayId = channelOrderId
	order.Status = data.DashFunRechargeStatus_Paid
	order.PaidAt = time.Now().Unix()

	order, err = dao.GetRechargeDao().SaveOrUpdate(order)
	if err != nil {
		return nil, err
	}

	//发送订单到处理队列
	r.processingOrdersChn <- order
	return order, nil
}

func (r *RechargeCenter) CancelRechargeOrder(orderId, userId string) (*data.DashFunRechargeData, error) {
	order, err := r.GetRechargeOrder(orderId)
	if err != nil {
		return nil, err
	}

	if order.Status > data.DashFunRechargeStatus_Pending {
		return nil, apperrors.ErrRechargeOrderStatus
	}

	if order.UserId != userId {
		return nil, apperrors.ErrRechargeOrderCantCancel
	}

	order.Status = data.DashFunRechargeStatus_Canceled
	order.Message = "Canceled Manually"
	order.PaidAt = time.Now().Unix()

	return dao.GetRechargeDao().SaveOrUpdate(order)
}
