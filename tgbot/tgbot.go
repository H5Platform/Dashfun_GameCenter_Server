package tgbot

import (
	"bytes"
	"context"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/events"
	"encoding/base64"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
)

var logoData []byte
var once sync.Once
var inst *TGBot

type TGBot struct {
	bot *bot.Bot
}

func Get() *TGBot {
	once.Do(func() {
		opts := []bot.Option{
			bot.WithDefaultHandler(defaultHandler),
		}

		d, err := os.ReadFile("./assets/dashfun.jpg")
		if err != nil {
			panic(err)
		}

		logoData = d

		botToken := config.GetConfig().TG.Token

		b, err := bot.New(botToken, opts...)
		if err != nil {
			panic(err)
		}

		b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, startHandler)
		b.RegisterHandler(bot.HandlerTypeMessageText, "/test", bot.MatchTypePrefix, testHandler)

		inst = &TGBot{
			bot: b,
		}

		go b.Start(context.Background())

	})
	return inst
}

func Bot() *bot.Bot {
	return Get().bot
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.MyChatMember != nil {
		if update.MyChatMember.NewChatMember.Administrator != nil {
			zap.S().Infow("DashFun Joined into ", "Chat", update.MyChatMember.Chat)
		}
	} else if update.Message != nil {
		if update.Message.SuccessfulPayment != nil {
			//log.Printf("Payment Received: %s", update.Message.SuccessfulPayment)
			//b.SendMessage(ctx, &bot.SendMessageParams{
			//	ChatID: update.Message.Chat.ID,
			//	Text:   "Thank you for your purchase\n" + strconv.Itoa(update.Message.SuccessfulPayment.TotalAmount) + " starts"},
			//)
			events.TGSuccessfulPaymentEvents.Emit(update.Message)
			return
		}
		if update.Message.Chat.Type == "private" {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Say /start",
			})
		}
	} else if update.PreCheckoutQuery != nil {
		//paymentId := update.PreCheckoutQuery.InvoicePayload
		//payment, err := paymentcenter.Get().FindPayment(paymentId)
		//_ = payment
		//if err != nil {
		//	zap.S().Errorw("get payment data error", "error", err, "paymentId", paymentId, "preCheckOutQuery", update.PreCheckoutQuery)
		//	b.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
		//		PreCheckoutQueryID: update.PreCheckoutQuery.ID,
		//		OK:                 false,
		//		ErrorMessage:       err.Error(),
		//	})
		//	return
		//}

		events.TGPreCheckoutQueryEvents.Emit(update.PreCheckoutQuery)

		//b.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
		//	PreCheckoutQueryID: update.PreCheckoutQuery.ID,
		//	OK:                 true,
		//	ErrorMessage:       "",
		//})
	}
}

func testHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := strings.Split(update.Message.Text, " ")

	if len(msg) < 2 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /test https://game_url",
		})
		return
	}

	if !strings.HasPrefix(msg[1], "https://") {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /test https://game_url\nthe game url must start with https://",
		})
		return
	}
	u, _ := url.Parse(botLink() + "/Games")
	q := u.Query()

	q.Set("startapp", "test-"+base64.StdEncoding.EncodeToString([]byte(msg[1])))
	u.RawQuery = q.Encode()
	gameLink := u.String()
	buttons := [][]models.InlineKeyboardButton{
		{
			{
				Text: "Open Test Game",
				URL:  gameLink,
			},
		},
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Open",
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	})
	if err != nil {
		log.Printf("SendMessage: %v  ", err)
	}
}

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Chat.Type != "private" {
		return
	}

	//member, err := b.GetChatMember(context.TODO(), &bot.GetChatMemberParams{
	//	ChatID: "-1002198592933",
	//	UserID: 1484579418,
	//})
	//
	//if err != nil {
	//	return
	//}

	//zap.S().Infow("member info", "", member)

	buttons := [][]models.KeyboardButton{
		{
			//{
			//	Text: "\U0001F44FGame Center123",
			//	WebApp: &models.WebAppInfo{
			//		URL: "https://t.me/DashFunBot/Center",
			//	},
			//}, {
			//	Text: "\U0001F3E6Wallet132",
			//	WebApp: &models.WebAppInfo{
			//		URL: "https://t.me/DashFunBot/Games",
			//	},
			//},
		},
	}

	msg := &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Welcome to DashFun Game Center!",
		ReplyMarkup: &models.ReplyKeyboardMarkup{
			Keyboard:       buttons,
			ResizeKeyboard: true,
		},
	}
	b.SendMessage(ctx, msg)

	buttons1 := [][]models.InlineKeyboardButton{
		{
			{
				Text: "Open Game Center",
				WebApp: &models.WebAppInfo{
					URL: appLink(),
				},
			},
		},
	}
	msg1 := &bot.SendPhotoParams{
		ChatID: update.Message.Chat.ID,
		Photo: &models.InputFileUpload{
			Filename: "dashfun.jpg",
			Data:     bytes.NewReader(logoData),
		},
		Caption: "Play lots of games in DashFun Game Center!\nEarn $TON & $NEXU rewards",
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: buttons1,
		},
	}

	b.SendPhoto(ctx, msg1)
}

func appLink() string {
	if config.IsTest() {
		return "https://dashfun-test.nexgami.com"
	} else if config.IsDev() {
		return "https://tma-test.nexgami.com/"
	} else {
		return "https://dashfun.nexgami.com"
	}
}

func botLink() string {
	if config.IsTest() {
		return "https://t.me/DashFunTestBot"
	} else if config.IsDev() {
		return "https://t.me/DashFunBot"
	} else {
		return "https://t.me/DashFunBot"
	}
}
