package main

import (
	_ "dashfun_gamecenter/api"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/taskcenter"
	"dashfun_gamecenter/tgbot"
	"dashfun_gamecenter/web"
	"go.uber.org/zap"
)

func main() {
	var logger *zap.Logger
	if config.IsProd() {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	zap.ReplaceGlobals(logger)
	tgbot.Get()
	taskcenter.Get()
	if err := web.GetService().Run(); err != nil {
		return
	}
}
