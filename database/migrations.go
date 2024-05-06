package database

import (
	"fmt"
	"jurrien/dnding-bot/models"
)

func Migrations(database *DB) error {
	if err := database.Connection.Migrator().DropColumn(&models.Character{}, "class"); err != nil {
		return fmt.Errorf("Could not drop column class for character")
	}
	return nil
}
