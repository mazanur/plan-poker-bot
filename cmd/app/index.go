package main

import (
	"github.com/rs/zerolog/log"
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
		log.Info().Msg("Started in dry mode ok\nBye!")
		os.Exit(0)
	}

	InitLogger()

	pgDb := PgConnInit()
	pgRepository := dao.NewRepository(pgDb)

	bot, err := tgbot.NewBot(conf.TgToken, pgRepository)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to start app")
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
		log.Error().Err(err).Msg("unable read request")
		return
	}

	application.Handle(update)

	rw.WriteHeader(200)
}
