package main

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" //for db migration
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gotestbot/internal/bot/bot_handler"
	"gotestbot/internal/bot/view"
	"gotestbot/internal/dao"
	"gotestbot/internal/service"
	"gotestbot/sdk/tgbot"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {

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

	go func() {
		err = bot.StartLongPolling(application.Handle)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to start app")
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}

func PgConnInit() *sqlx.DB {

	dsn := GetPgDsn()

	if err := MigrateDB(dsn); err != nil {
		log.Fatal().Msgf("Database migration failed: %s", err.Error())
	}
	log.Info().Msg("Database migration succeeded")

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatal().Msgf("Failed to connect to db. dsn='%s': %s", DsnMaskPass(dsn), err.Error())
	}
	db.SetMaxOpenConns(conf.Pg.MaxOpenConn)
	db.SetMaxIdleConns(conf.Pg.MaxIdleConn)
	db.SetConnMaxLifetime(conf.Pg.MaxLifeTime)
	db.SetConnMaxIdleTime(conf.Pg.MaxIdleTime)
	log.Info().Msg("Connected to db")
	return db
}

func MigrateDB(dsn string) error {
	m, err := migrate.New("file://db/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func InitLogger() {
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z"
	zerolog.TimestampFieldName = "@timestamp"

	logLvl, err := zerolog.ParseLevel(strings.ToLower(conf.LogLevel))
	if err != nil {
		log.Fatal().Msgf("Failed to parse log level '%s': %s", conf.LogLevel, err.Error())
	}

	zerolog.SetGlobalLevel(logLvl)

	switch conf.LogFormat {
	case "plain":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	case "logstash":
		// do nothing
	default:
		log.Fatal().Msgf("Unknown log format '%s'", conf.LogFormat)
	}

	log.Info().Msg("Logger successfully initialized")
}
