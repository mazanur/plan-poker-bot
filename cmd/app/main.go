package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-pkgz/lgr"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" //for db migration
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"gotestbot/internal/bot/bot_handler"
	"gotestbot/internal/bot/view"
	"gotestbot/internal/dao"
	"gotestbot/internal/service"
	"gotestbot/sdk/tgbot"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	InitConfig()

	if conf.Dry {
		lgr.Printf("[INFO] Started in dry mode ok\nnBye!")
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

	go func() {
		err = bot.StartLongPolling(application.Handle)
		if err != nil {
			lgr.Fatalf("[ERROR] unable to start app")
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}

func PgConnInit() *sqlx.DB {

	dsn := GetPgDsn()

	if err := MigrateDB(dsn); err != nil {
		lgr.Fatalf("[ERROR] Database migration failed: %s", err.Error())
	}
	lgr.Print("[INFO] Database migration succeeded")

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		lgr.Fatalf("Failed to connect to db. dsn='%s': %s", DsnMaskPass(dsn), err.Error())
	}
	db.SetMaxOpenConns(conf.Pg.MaxOpenConn)
	db.SetMaxIdleConns(conf.Pg.MaxIdleConn)
	db.SetConnMaxLifetime(conf.Pg.MaxLifeTime)
	db.SetConnMaxIdleTime(conf.Pg.MaxIdleTime)
	lgr.Print("[INFO] Connected to db")

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
	setupLog(conf.LogLevel == "debug", "")
}

func setupLog(dbg bool, lf string) {
	colorizer := lgr.Mapper{
		ErrorFunc:  func(s string) string { return color.New(color.FgHiRed).Sprint(s) },
		WarnFunc:   func(s string) string { return color.New(color.FgHiYellow).Sprint(s) },
		InfoFunc:   func(s string) string { return color.New(color.FgHiWhite).Sprint(s) },
		DebugFunc:  func(s string) string { return color.New(color.FgWhite).Sprint(s) },
		CallerFunc: func(s string) string { return color.New(color.FgBlue).Sprint(s) },
		TimeFunc:   func(s string) string { return color.New(color.FgCyan).Sprint(s) },
	}

	var stdout, stderr *os.File
	var err error
	if lf != "" {
		stdout, err = os.OpenFile(lf, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Printf("error opening log file: %v", err)
			os.Exit(2)
		}
		stderr = stdout
	} else {
		stdout = os.Stdout
		stderr = nil
	}
	if dbg {
		lgr.Setup(
			lgr.Debug,
			lgr.CallerFile,
			lgr.CallerFunc,
			lgr.Msec,
			lgr.LevelBraces,
			lgr.Out(stdout),
			lgr.Err(stderr),
			lgr.Map(colorizer),
		)
	}
	lgr.Setup(lgr.Out(stdout), lgr.Err(stderr), lgr.Map(colorizer), lgr.StackTraceOnError)
	lgr.Printf("INFO Logger successfully initialized")

}
