package database

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"jurrien/dnding-bot/models"
	"jurrien/dnding-bot/utils"

	"github.com/charmbracelet/log"
	_ "github.com/tursodatabase/go-libsql"
	sqlite "github.com/ytsruh/gorm-libsql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	DSN string `default:":memory:"`
}

type DB struct {
	Connection *gorm.DB
	Config     *Config
}

var db_logger *log.Logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
	TimeFormat:      time.DateTime,
	Prefix:          "db_logger",
})

func (db *DB) GetConnection() *gorm.DB {
	var err error
	if db.Connection, err = gorm.Open(sqlite.New(sqlite.Config{
		DSN: db.Config.DSN,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), //TODO: for now silent need to figure this out later
	}); err != nil {
		db_logger.Error("Failed to connect and open database", "error", err)
		return nil
	}

	return db.Connection
}

func SetupDB(ctx context.Context, config *Config) (*DB, error) {
	database := &DB{}
	var err error
	database.Config = config
	// Pass a DSN or Turso DB url to Gorm. The url must also include an authToken as a query parameter
	if database.Connection, err = gorm.Open(sqlite.New(sqlite.Config{
		DSN: config.DSN,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), //TODO: for now silent need to figure this out later
	}); err != nil {
		return nil, fmt.Errorf("Failed to connect and open database: %w", err)
	}

	db_logger.Info("Connected to database", "database", database.Connection.Migrator().CurrentDatabase())
	if err = database.Connection.Migrator().AutoMigrate(&models.Player{}, &models.Character{}, &models.Expense{}); err != nil {
		return nil, fmt.Errorf("Failed to run auto migration on the database: %v", err)
	}

	if err = Migrations(database); err != nil {
		return nil, fmt.Errorf("Could not do migrations: %v", err)
	}

	intialiseExpenseTypes(database)

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

func intialiseExpenseTypes(db *DB) {
	db_logger.Info("Initialise expenses")
	db_logger.Debug("Reading Current Expense Types")
	v := reflect.ValueOf(utils.ExpenseType)
	typeOfExpenseType := v.Type()

	var expenses []models.Expense

	err := db.GetConnection().Model(&models.Expense{}).Find(&expenses).Error
	if err != nil {
		db_logger.Fatal("Could not get expenses", "error", err)
	}

	for i := 0; i < v.NumField(); i++ {
		expenseType := typeOfExpenseType.Field(i).Name
		expenseValue := v.Field(i).Interface()

		db_logger.Info("Initialising ExpenseType", "type", expenseType)
		found_expense := false
		for _, expense := range expenses {
			if expense.Name == expenseValue {
				found_expense = true
			}
		}

		if found_expense {
			db_logger.Warn("Expense already exists in database. Doing nothing", "expense", expenseType)
		} else {
			expense := &models.Expense{
				Name: expenseValue.(string),
			}
			db_logger.Info("Saving Expense", "expense", expenseType)
			err := db.GetConnection().Save(expense).Error
			if err != nil {
				db_logger.Fatal("Saving Expense went wrong", "expense", expenseType, "error", err)
			}
		}
	}
}
