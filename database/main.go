package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"jurrien/dnding-bot/models"

	"github.com/charmbracelet/log"
	_ "github.com/tursodatabase/go-libsql"
	sqlite "github.com/ytsruh/gorm-libsql"
	"gorm.io/gorm"
)

type Config struct {
	DSN string `default:":memory:"`
}

type DB struct {
	Connection *gorm.DB
}

var db_logger *log.Logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
	TimeFormat:      time.DateTime,
	Prefix:          "db_logger",
})

func SetupDB(ctx context.Context, config *Config) (*DB, error) {
	database := &DB{}
	var err error

	// Pass a DSN or Turso DB url to Gorm. The url must also include an authToken as a query parameter
	if database.Connection, err = gorm.Open(sqlite.New(sqlite.Config{
		DSN: config.DSN,
	}), &gorm.Config{}); err != nil {
		return nil, fmt.Errorf("Failed to connect and open database: %w", err)
	}

	db_logger.Info("Connected to database", "database", database.Connection.Migrator().CurrentDatabase())
	if err = database.Connection.Migrator().AutoMigrate(&models.Player{}, &models.Character{}, &models.Expense{}); err != nil {
		return nil, fmt.Errorf("Failed to run auto migration on the database: %v", err)
	}

	go func() {
		<-ctx.Done()
		db_logger.Info("Shutting down database connection")
		sqlDriver, err := database.Connection.DB()
		if err != nil {
			db_logger.Error("Failed closing database", "error", err)
		}
		sqlDriver.Close()
		db_logger.Info("Gracefully shut down database connection")
	}()
	return database, nil
}
