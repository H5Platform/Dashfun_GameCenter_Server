package paymentcenter

import (
	"context"
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/tgbot"
	"dashfun_gamecenter/usercenter"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var instance *PaymentCenter

// PaymentCenter 支付中心
// 用于处理支付相关的逻辑，支付只能用DashFunDiamond，不足需要到充值中心充值
type PaymentCenter struct {
	idGen *snowflake.Worker
}

func Get() *PaymentCenter {
	once.Do(func() {
		instance = &PaymentCenter{}
		instance.init()
	})
	return instance
}

func (p *PaymentCenter) init() {
	p.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerPaymentId))
	events.TGPreCheckoutQueryEvents.On(p.onTGPreCheckoutQueryEvent)
	events.TGSuccessfulPaymentEvents.On(p.onTGPaymentSuccessEvent)
}

func (p *PaymentCenter) onTGPaymentSuccessEvent(message *models.Message) {
	paymentId := message.SuccessfulPayment.InvoicePayload

	//pi开头的才是支付订单
	if !strings.HasPrefix(paymentId, "pi") {
		return
	}

	payment, err := p.FindPayment(paymentId)
	if err != nil {
		zap.S().Errorw("SuccessfulPayment Received, payment data get error", "paymentId", paymentId, "error", err, "Message", message)
		return
	}

	payment.Status = data.DashFunPaymentStatus_Paid
	payment.Message = message.SuccessfulPayment.TelegramPaymentChargeID
	payment.PaidAt = time.Now().UnixMilli()
	_, err = dao.GetPaymentDao().SaveOrUpdate(payment)
	if err != nil {
		zap.S().Errorw("Save Payment Data Failed", "Payment", payment, "error", err, "Message", message)
		return
	}
	b := tgbot.Bot()
	b.SendMessage(context.TODO(), &bot.SendMessageParams{
		ChatID: message.Chat.ID,
		Text:   "Thank you for your purchase\n" + strconv.Itoa(message.SuccessfulPayment.TotalAmount) + " Stars"},
	)

	user, err := usercenter.Get().GetDashFunUser(payment.UserId)
	if err != nil {
		zap.S().Errorw("Get User Failed", "Payment", payment, "error", err, "Message", message)
		return
	}
	game, err := gamecenter.Get().FindGame(payment.GameId)
	if err != nil {
		zap.S().Errorw("Find Game Failed", "Payment", payment, "error", err, "Message", message)
		return
	}

	events.UserTGPaymentEvents.Emit(&events.EventUserPayment{
		User:    user,
		Game:    game,
		Payment: payment,
	})

}

func (p *PaymentCenter) onTGPreCheckoutQueryEvent(query *models.PreCheckoutQuery) {
	paymentId := query.InvoicePayload

	//pi开头的才是支付订单
	if !strings.HasPrefix(paymentId, "pi") {
		return
	}

	payment, err := p.FindPayment(paymentId)
	b := tgbot.Bot()
	if err != nil {
		zap.S().Errorw("get payment data error", "error", err, "paymentId", paymentId, "preCheckOutQuery", query)
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
		payment.Status = data.DashFunPaymentStatus_Pending
		dao.GetPaymentDao().SaveOrUpdate(payment)
	} else {
		zap.S().Errorw("Answer Pre Checkout Query Failed", "error", err, "paymentId", paymentId, "preCheckOutQuery", query)
		payment.Status = data.DashFunPaymentStatus_Failed
		payment.Message = err.Error()
		dao.GetPaymentDao().SaveOrUpdate(payment)
	}
}

func (p *PaymentCenter) newPaymentId() string {
	id := p.idGen.NextId()
	return "pi" + strconv.FormatInt(id, 36)
}

func (p *PaymentCenter) FindPayment(id string) (*data.DashFunPaymentData, error) {
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	//defer cancel()
	pd, err := dao.GetPaymentDao().FindPaymentById(id)
	if err != nil {
		return nil, err
	}
	return pd, nil
}

func (p *PaymentCenter) RequestDashFunPayment(userId, gameId, title, desc, payload string, price int, isTesting bool) (*data.DashFunPaymentData, error) {
	id := p.newPaymentId()
	currency := "DFD" //DashFunDiamond
	from := data.DashFunPaymentFrom_DashFun
	if isTesting {
		currency = "DFD_TEST"
		from = data.DashFunPaymentFrom_TEST
	}
	payment, err := dao.GetPaymentDao().CreatePayment(id, userId, gameId, "", title, desc, payload, currency, from, price, "")
	zap.S().Infow("RequestDashFunPayment", "Payment", payment, "error", err)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (p *PaymentCenter) ConfirmDashFunPayment(paymentId string, opUserId string) (*data.DashFunPaymentData, error) {
	payment, err := p.FindPayment(paymentId)
	if err != nil {
		return nil, err
	}

	if payment.Status > data.DashFunPaymentStatus_Pending {
		zap.S().Errorw("ConfirmDashFunPayment Failed", "Payment", payment, "error", "Status Not Pending")
		return payment, nil
	}

	if payment.UserId != opUserId {
		zap.S().Errorw("ConfirmDashFunPayment Failed", "Payment", payment, "User", opUserId, "error", "User Not Match")
		payment.Status = data.DashFunPaymentStatus_Failed
		payment.Message = "User Not Match"
		dao.GetPaymentDao().SaveOrUpdate(payment)
		return payment, nil
	}

	user, err := usercenter.Get().GetDashFunUser(payment.UserId)
	if err != nil {
		zap.S().Errorw("RequestTestPayment Get User Failed", "Payment", payment, "error", err)
		payment.Status = data.DashFunPaymentStatus_Failed
		payment.Message = err.Error()
		dao.GetPaymentDao().SaveOrUpdate(payment)
		return payment, err
	}

	game, err := gamecenter.Get().FindGame(payment.GameId)
	if err != nil {
		zap.S().Errorw("RequestTestPayment Find Game Failed", "Payment", payment, "error")
		payment.Status = data.DashFunPaymentStatus_Failed
		payment.Message = err.Error()
		dao.GetPaymentDao().SaveOrUpdate(payment)
		return payment, err
	}

	//扣费
	diamond := coincenter.Get().GetDashFunDiamond()
	_, err = coincenter.Get().DecUserCoinAmount(user.Id, diamond.Id, int32(payment.Price))

	if err != nil {
		zap.S().Errorw("Dec DashFunDiamond Failed", "Payment", payment.Id, "error", err)
		payment.Status = data.DashFunPaymentStatus_Failed
		payment.Message = err.Error()
		dao.GetPaymentDao().SaveOrUpdate(payment)
		return payment, err
	}

	payment.Status = data.DashFunPaymentStatus_Paid
	payment.PaidAt = time.Now().UnixMilli()

	_, err = dao.GetPaymentDao().SaveOrUpdate(payment)
	if err != nil {
		return nil, err
	}
	events.UserPaymentEvents.Emit(&events.EventUserPayment{
		User:    user,
		Game:    game,
		Payment: payment,
	})

	return payment, nil
}

// RequestTGPayment 请求使用tg支付
func (p *PaymentCenter) RequestTGPayment(userId, gameId, title, desc, payload string, price int, isTesting bool) (*data.DashFunPaymentData, error) {
	//向tg bot请求一个新的Invoice
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	id := p.newPaymentId()

	invoiceLink := ""
	var err error = nil

	if isTesting {
		invoiceLink = "test-" + id
	} else {
		invoiceLink, err = tgbot.Bot().CreateInvoiceLink(ctx, &bot.CreateInvoiceLinkParams{
			Title:         title,
			Description:   desc,
			Payload:       id,
			ProviderToken: "",
			Currency:      "XTR",
			Prices: []models.LabeledPrice{
				{
					Label:  title,
					Amount: price,
				},
			},
		})
	}

	if err != nil {
		return nil, err
	}

	paymentId, found := strings.CutPrefix(invoiceLink, "https://t.me/$")
	if !found {
		paymentId = invoiceLink
	}

	currency := "TG_XTR"
	from := data.DashFunPaymentFrom_TG
	if isTesting {
		currency = "TG_XTR_TEST"
		from = data.DashFunPaymentFrom_TEST
	}
	payment, err := dao.GetPaymentDao().CreatePayment(id, userId, gameId, paymentId, title, desc, payload, currency, from, price, invoiceLink)
	if err != nil {
		return nil, err
	}

	if isTesting {
		//测试状态下直接完成订单
		payment.Status = data.DashFunPaymentStatus_Paid
		payment.Message = "Test Payment"
		payment.PaidAt = time.Now().UnixMilli()
		_, err = dao.GetPaymentDao().SaveOrUpdate(payment)

		user, err := usercenter.Get().GetDashFunUser(payment.UserId)
		if err != nil {
			zap.S().Errorw("RequestTestPayment Get User Failed", "Payment", payment, "error", err)
			return payment, nil
		}
		game, err := gamecenter.Get().FindGame(payment.GameId)
		if err != nil {
			zap.S().Errorw("RequestTestPayment Find Game Failed", "Payment", payment, "error")
			return payment, nil
		}
		events.UserTGPaymentEvents.Emit(&events.EventUserPayment{
			User:    user,
			Game:    game,
			Payment: payment,
		})
	}
	return payment, nil
}
