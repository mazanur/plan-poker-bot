package main

import (
	"github.com/go-pkgz/lgr"
	"gotestbot/internal/bot/bot_handler"
	"gotestbot/internal/bot/view"
	"gotestbot/internal/dao"
	"gotestbot/internal/service"
	"gotestbot/sdk/tgbot"
	"net/http"
	"os"
)

func Handler(rw http.ResponseWriter, req *http.Request) {

	InitConfig()

	if conf.Dry {
		lgr.Printf("[INFO] Started in dry mode ok\nBye!")
		os.Exit(0)
	}

	InitLogger()

	pgDb := PgConnInit()
	pgRepository := dao.NewRepository(pgDb)

	bot, err := tgbot.NewBot(conf.TgToken, pgRepository)
	if err != nil {
		lgr.Fatalf("[ERROR] unable to start app")
	}

	rateService := service.NewRateService(pgRepository)
	taskService := service.NewTaskService(pgRepository)
	viewSender := view.NewView(pgRepository, pgRepository, pgRepository, taskService, rateService, bot)
	application := bot_handler.NewBotApp(viewSender,
		service.NewRoomService(pgRepository),
		service.NewTaskService(pgRepository),
		rateService)
	update, err := bot.WrapRequest(req)
	if err != nil {
		lgr.Printf("[ERROR] unable read request %v", err)
		return
	}

	application.Handle(update)

	rw.WriteHeader(200)
}
