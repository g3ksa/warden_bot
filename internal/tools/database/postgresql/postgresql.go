package postrgesql

import (
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host       string
	Port       int
	DBUser     string
	DBPassword string
	DBName     string
}

func New(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable",
		cfg.Host,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.Port,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, fmt.Errorf("failed connect to database: %v", err)
	}
	slog.Info("Database init")
	return db, nil
}
