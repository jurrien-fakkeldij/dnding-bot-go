package models

import (
	"fmt"
	"jurrien/dnding-bot/database"
	"jurrien/dnding-bot/utils"
	"os"
	"reflect"
	"time"

	"github.com/charmbracelet/log"
)

var model_logger *log.Logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
	TimeFormat:      time.DateTime,
	Prefix:          "model_logger",
})

func InitializeModels(database *database.DB) error {
	model_logger.Info("Connected to database", "database", database.Connection.Migrator().CurrentDatabase())
	if err := database.Connection.Migrator().AutoMigrate(Player{}, Character{}, Expense{}, CharacterExpense{}); err != nil {
		return fmt.Errorf("failed to run auto migration on the database: %v", err)
	}

	intialiseExpenseTypes(database)

	return nil
}

func intialiseExpenseTypes(db *database.DB) {
	model_logger.Info("Initialise expenses")
	model_logger.Debug("Reading Current Expense Types")
	v := reflect.ValueOf(utils.ExpenseType)
	typeOfExpenseType := v.Type()

	var expenses []Expense

	err := db.GetConnection().Model(&Expense{}).Find(&expenses).Error
	if err != nil {
		model_logger.Fatal("Could not get expenses", "error", err)
	}

	for i := 0; i < v.NumField(); i++ {
		expenseType := typeOfExpenseType.Field(i).Name
		expenseValue := v.Field(i).Interface()

		model_logger.Info("Initialising ExpenseType", "type", expenseType)
		found_expense := false
		for _, expense := range expenses {
			if expense.Name == expenseValue {
				found_expense = true
			}
		}

		if found_expense {
			model_logger.Warn("Expense already exists in database. Doing nothing", "expense", expenseType)
		} else {
			expense := &Expense{
				Name: expenseValue.(string),
			}
			model_logger.Info("Saving Expense", "expense", expenseType)
			err := db.GetConnection().Save(expense).Error
			if err != nil {
				model_logger.Fatal("Saving Expense went wrong", "expense", expenseType, "error", err)
			}
		}
	}
}
