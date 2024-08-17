package models

import (
	"fmt"
	"jurrien/dnding-bot/database"
)

func Migrations(database *database.DB) error {
	if err := database.Connection.Migrator().DropColumn(Character{}, "class"); err != nil {
		return fmt.Errorf("could not drop column class for character")
	}
	return nil
}
