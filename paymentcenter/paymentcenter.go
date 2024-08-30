package paymentcenter

import (
	"context"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/tgbot"
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
	events.TGPreCheckoutQueryEvents.On(p.onPreCheckoutQueryEvent)
	events.TGSuccessfulPaymentEvents.On(p.onPaymentSuccessEvent)
}

func (p *PaymentCenter) onPaymentSuccessEvent(message *models.Message) {
	paymentId := message.SuccessfulPayment.InvoicePayload
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
}

func (p *PaymentCenter) onPreCheckoutQueryEvent(query *models.PreCheckoutQuery) {
	paymentId := query.InvoicePayload
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
	return strconv.FormatInt(id, 36)
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

// RequestTGPayment 请求使用tg支付
func (p *PaymentCenter) RequestTGPayment(userId, gameId, title, desc string, price int) (*data.DashFunPaymentData, error) {
	//向tg bot请求一个新的Invoice
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	id := p.newPaymentId()

	invoiceLink, err := tgbot.Bot().CreateInvoiceLink(ctx, &bot.CreateInvoiceLinkParams{
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
	if err != nil {
		return nil, err
	}

	paymentId, found := strings.CutPrefix(invoiceLink, "https://t.me/$")
	if !found {
		paymentId = invoiceLink
	}

	payment, err := dao.GetPaymentDao().CreatePayment(id, userId, gameId, paymentId, title, desc, "", "XTR", data.DashFunPaymentFrom_TG, price, invoiceLink)
	if err != nil {
		return nil, err
	}

	return payment, nil
}
