package tgbot

import (
	"context"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/gamecenter"
	"encoding/base64"
	"fmt"
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
		b.RegisterHandler(bot.HandlerTypeMessageText, "/p-chat-id", bot.MatchTypePrefix, printChatIdHandler)

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

func (t *TGBot) GetUserPhotoFilePath(userId int64) string {
	tgbot := t.bot

	photos, err := tgbot.GetUserProfilePhotos(context.TODO(), &bot.GetUserProfilePhotosParams{
		UserID: userId,
		Limit:  1,
	})
	if err != nil || photos.TotalCount == 0 {
		//没有头像
		return ""
	} else {
		fileId := photos.Photos[0][0].FileID
		file, err := tgbot.GetFile(context.TODO(), &bot.GetFileParams{
			FileID: fileId,
		})

		if err != nil {
			//头像获取失败
			return ""
		} else {
			photoUrl := file.FilePath
			return photoUrl
		}
	}
}

func (t *TGBot) GetUserPhotoUrlByFile(file string) string {
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", config.GetConfig().TG.Token, file)
}

func (t *TGBot) GetUserPhotoUrl(userId int64) string {
	path := t.GetUserPhotoFilePath(userId)
	if path == "" {
		return ""
	} else {
		return t.GetUserPhotoUrlByFile(path)
	}
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
			//b.SendMessage(ctx, &bot.SendMessageParams{
			//	ChatID: update.Message.Chat.ID,
			//	Text:   "Say /start",
			//})
			startHandler(ctx, b, update)
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
	} else if update.CallbackQuery != nil {
		//game := update.CallbackQuery.GameShortName
		//game, err := dao.GetGameDao().GetGameByName("War Three Kingdoms")
		_, err := b.AnswerCallbackQuery(context.TODO(), &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            update.CallbackQuery.GameShortName,
			ShowAlert:       false,
			URL:             "https://entry-tma.3kweb3.com",
			CacheTime:       0,
		})
		if err != nil {
			return
		}
	}
}

func printChatIdHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	zap.S().Infow("print chat:", "chat", update.Message.Chat)
	return
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

	if !(strings.HasPrefix(msg[1], "http://") || strings.HasPrefix(msg[1], "https://")) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /test https://game_url\nthe game url must start with http:// or https://",
		})
		return
	}
	u, _ := url.Parse(botLink() + "/Games")
	//q := u.Query()
	//q.Set("startapp", "test-"+base64.StdEncoding.EncodeToString([]byte(msg[1])))

	//gzip url
	//var buf bytes.Buffer
	//gz := gzip.NewWriter(&buf)
	//_, err := gz.Write([]byte(msg[1]))
	//defer gz.Close()
	//if err != nil {
	//	zap.S().Errorw("gzip write error", "error", err)
	//	return
	//}

	u.RawQuery = "startapp=test-" + base64.RawURLEncoding.EncodeToString([]byte(msg[1]))
	//base64.StdEncoding.EncodeToString(buf.Bytes())
	//base64.StdEncoding.EncodeToString([]byte(msg[1]))
	gameLink := u.String()

	zap.S().Infow("Test", "gameLink", gameLink)

	buttons := [][]models.InlineKeyboardButton{
		{
			{
				Text: "Open Test Game",
				URL:  gameLink,
			},
		},
	}
	text := "Open"
	if config.IsDev() {
		text += " " + gameLink
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   text,
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

	//buttons := [][]models.KeyboardButton{
	//	{
	//		//{
	//		//	Text: "\U0001F44FGame Center123",
	//		//	WebApp: &models.WebAppInfo{
	//		//		URL: "https://t.me/DashFunBot/Center",
	//		//	},
	//		//}, {
	//		//	Text: "\U0001F3E6Wallet132",
	//		//	WebApp: &models.WebAppInfo{
	//		//		URL: "https://t.me/DashFunBot/Games",
	//		//	},
	//		//},
	//	},
	//}

	//msg := &bot.SendMessageParams{
	//	ChatID: update.Message.Chat.ID,
	//	Text:   "Welcome to DashFun Game Center!",
	//	ReplyMarkup: &models.ReplyKeyboardMarkup{
	//		Keyboard:       buttons,
	//		ResizeKeyboard: true,
	//	},
	//}
	//b.SendMessage(ctx, msg)
	//
	//buttons1 := [][]models.InlineKeyboardButton{
	//	{
	//		{
	//			Text: "Open Game Center",
	//			WebApp: &models.WebAppInfo{
	//				URL: appLink(),
	//			},
	//		},
	//	},
	//}
	//msg1 := &bot.SendPhotoParams{
	//	ChatID: update.Message.Chat.ID,
	//	Photo: &models.InputFileUpload{
	//		Filename: "dashfun.jpg",
	//		Data:     bytes.NewReader(logoData),
	//	},
	//	Caption: "Play lots of games in DashFun Game Center!\nEarn $TON & $NEXU rewards",
	//	ReplyMarkup: &models.InlineKeyboardMarkup{
	//		InlineKeyboard: buttons1,
	//	},
	//}
	//
	//b.SendPhoto(ctx, msg1)

	msgCenter := &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Welcome! I bet you’re here for the fun—head over to the Game Center and hey, don’t forget to stack some airdrop points while you’re at it!",
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text: "GameCenter",
						URL:  botLink() + "/Center",
					},
				},
			},
		},
	}

	b.SendMessage(ctx, msgCenter)

	game, err := gamecenter.Get().FindGameByName("War Three Kingdoms")
	if err != nil || game == nil {
		return
	}
	buttons2 := [][]models.InlineKeyboardButton{
		{
			{
				Text: "Open " + game.Name,
				URL:  gameLink(game.Id),
			},
		},
	}
	if err == nil {
		msg2 := &bot.SendPhotoParams{
			ChatID: update.Message.Chat.ID,
			Photo: &models.InputFileString{
				Data: game.MainPicUrl,
			},
			Caption: game.Desc,
			ReplyMarkup: &models.InlineKeyboardMarkup{
				InlineKeyboard: buttons2,
			},
		}

		_, err := b.SendPhoto(ctx, msg2)
		if err != nil {
			log.Printf("SendPhoto: %v  ", err)
		}
	}

	//b.SendGame(context.TODO(), &bot.SendGameParams{
	//	ChatID:       update.Message.Chat.ID,
	//	GameShorName: "threekweb3",
	//})

}

func appLink() string {
	if config.IsTest() {
		return "https://dashfun-test.nexgami.com"
	} else if config.IsDev() {
		return "https://tma-test.nexgami.com/"
	} else {
		return "https://tma.dashfun.games"
	}
}

func botLink() string {
	if config.IsTest() {
		return "https://t.me/DashFunTestBot"
	} else if config.IsDev() {
		return "https://t.me/LocalTestBot"
	} else {
		return "https://t.me/DashFunBot"
	}
}

func gameLink(gameId string) string {
	return botLink() + "/Games?startapp=" + gameId
}
