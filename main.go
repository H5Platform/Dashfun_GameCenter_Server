package main

import (
	"dashfun_gamecenter/RechargeCenter"
	_ "dashfun_gamecenter/_initdata"
	_ "dashfun_gamecenter/api"
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/coingecko"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/fishingverse/fishingleaderboard"
	"dashfun_gamecenter/invitecenter"
	"dashfun_gamecenter/leaderboardcenter"
	"dashfun_gamecenter/nolandev"
	nolandevleaderboard "dashfun_gamecenter/nolandev/leaderboard"
	"dashfun_gamecenter/openai_api"
	"dashfun_gamecenter/paymentcenter"
	"dashfun_gamecenter/pricepredictcenter"
	"dashfun_gamecenter/taskcenter"
	"dashfun_gamecenter/tgbot"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"dashfun_gamecenter/web3center"
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	//将本地user数据迁移到DBUserCenter里，专门给usercenter服务使用
	//usercenter.MoveUserData()

	logPath := config.GetConfig().Log.Path
	if logPath == "" {
		logPath = "app.log"
	}

	rotator, err := rotatelogs.New(
		logPath+".%Y-%m-%d",
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithLinkName(logPath),
	)

	if err != nil {
		panic(err)
	}

	var logger *zap.Logger
	var zapcfg zap.Config
	var writer io.Writer

	if config.IsDev() {
		zapcfg = zap.NewDevelopmentConfig()
		writer = os.Stdout
	} else {
		zapcfg = zap.NewProductionConfig()
		writer = rotator
	}
	defer logger.Sync()
	zapcfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoder := zapcore.NewConsoleEncoder(zapcfg.EncoderConfig)
	core := zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel)
	logger = zap.New(core, zap.AddCaller())

	zap.ReplaceGlobals(logger)
	tgbot.Get()
	coincenter.Get()
	taskcenter.Get()
	invitecenter.Get()
	leaderboardcenter.Get()
	RechargeCenter.Get()
	paymentcenter.Get()
	usercenter.Get()

	switch config.RunningMode() {
	case config.ModeNolanDev:
		nolandev.Get()
		nolandevleaderboard.Get()
		openai_api.GetMarketSummarize()
	case config.ModeFishVerse:
		fishingleaderboard.Get()
	case config.ModeHowardAI:
		nolandev.Get()
		nolandevleaderboard.Get()
	}

	web3center.Get()
	pricepredictcenter.Get()

	coingecko.Get()
	//openai_api.GetMarketSummarize()
	//prompt, err := coingecko.GetTokenPricePrompt("ethereum", "usd")
	//if err != nil {
	//	return
	//}
	//log.Info(prompt)
	//resp, err := openai_api.SummarizeWithOpenAI(context.Background(), prompt)
	//if err != nil {
	//	return
	//}

	//log.Info(resp)

	logger.Info("dashfun gamecenter started")
	//ton.Get()
	if err := web.GetService().Run(); err != nil {
		return
	}
}
