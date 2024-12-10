package config

import (
	"flag"
	"fmt"

	"github.com/g3ksa/warden_bot/internal/config"
	"github.com/spf13/viper"
)

type WardenBotConfig struct {
	CronSchedule    string
	HttpAddr        string
	ModelServiceURL string
	BotToken        string
	RunImmediate    bool
	Database        config.Database
}

func NewWardenBotConfig() (*WardenBotConfig, error) {
	v := viper.GetViper()

	configPath := "config/warden_bot.yaml"

	run := flag.Bool("run", false, "generate sitemaps immediately")
	flag.Parse()

	baseConfig, err := config.NewConfig(v, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %v", err)
	}

	return &WardenBotConfig{
		CronSchedule:    v.GetString("service.cron_schedule"),
		HttpAddr:        v.GetString("service.http_addr"),
		ModelServiceURL: v.GetString("service.model_service_url"),
		BotToken:        v.GetString("service.bot_token"),
		RunImmediate:    *run,
		Database:        *baseConfig.NewDatabase(),
	}, nil
}
