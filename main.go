package main

import (
	_ "dashfun_gamecenter/_initdata"
	_ "dashfun_gamecenter/api"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/taskcenter"
	"dashfun_gamecenter/tgbot"
	"dashfun_gamecenter/web"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"time"
)

func main() {
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

	if config.IsProd() {
		zapcfg = zap.NewProductionConfig()
		writer = rotator
	} else {
		zapcfg = zap.NewDevelopmentConfig()
		writer = os.Stdout
	}
	defer logger.Sync()
	zapcfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoder := zapcore.NewConsoleEncoder(zapcfg.EncoderConfig)
	core := zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel)
	logger = zap.New(core, zap.AddCaller())

	zap.ReplaceGlobals(logger)
	tgbot.Get()
	taskcenter.Get()

	logger.Info("dashfun gamecenter started")

	if err := web.GetService().Run(); err != nil {
		return
	}
}
