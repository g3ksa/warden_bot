package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

type config struct {
	Viper *viper.Viper
}

type Database struct {
	Host       string
	Port       int
	DBUser     string
	DBPassword string
	DBName     string
}

func NewConfig(v *viper.Viper, configFile string) (*config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		slog.Error("failed to read a config file on <NewConfig>: %v", err)
		return nil, ErrConfig
	}

	return &config{Viper: v}, nil
}

func (c *config) NewDatabase() *Database {
	return &Database{
		Host:       c.Viper.GetString("database.host"),
		Port:       c.Viper.GetInt("database.port"),
		DBUser:     c.Viper.GetString("database.db_user"),
		DBPassword: c.Viper.GetString("database.db_password"),
		DBName:     c.Viper.GetString("database.db_name"),
	}
}
