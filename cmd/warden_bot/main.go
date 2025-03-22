package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	postrgesql "github.com/g3ksa/warden_bot/internal/tools/database/postgresql"
	"github.com/g3ksa/warden_bot/internal/warden_bot/service/storage"

	"github.com/g3ksa/warden_bot/internal/warden_bot/config"
	"github.com/g3ksa/warden_bot/internal/warden_bot/service"
	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func init() {

}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.NewWardenBotConfig()
	if err != nil {
		slog.Error(err.Error())
	}

	db, err := postrgesql.New(&postrgesql.Config{
		Host:       cfg.Database.Host,
		Port:       cfg.Database.Port,
		DBUser:     cfg.Database.DBUser,
		DBPassword: cfg.Database.DBPassword,
		DBName:     cfg.Database.DBName,
	})
	if err != nil {
		slog.Error(err.Error())
	}

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	slog.Info("Authorized on account:", slog.String("username", bot.Self.UserName))

	dbStorage := storage.NewDBStorage(db)
	wardenBotservice := service.NewWardenBotService(cfg.ModelServiceURL, bot, dbStorage)

	cron := gocron.NewScheduler(time.Local)
	scheduler := cron.Cron(cfg.CronSchedule)

	if cfg.RunImmediate {
		fmt.Println("run immediatly")
		_, err = scheduler.StartImmediately().Do(run, ctx, wardenBotservice)
	} else {
		_, err = scheduler.Do(run, ctx, wardenBotservice)
	}

	slog.Info(fmt.Sprintf("cron schedule: %v, task: %s", cfg.CronSchedule, "SitemapService.GenerateSitemaps"))
	if err != nil {
		panic(err)
	}

	cron.StartAsync()

	go func() {
		err := wardenBotservice.ProcessUpdatesFromBot(ctx)
		if err != nil {
			log.Panic(err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signalChan
	slog.Info("Received signal", sig)

	cancel()
	cron.Stop()
}

func run(ctx context.Context, service *service.WardenBotService) error {
	if err := service.ProcessMessages(ctx); err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}
