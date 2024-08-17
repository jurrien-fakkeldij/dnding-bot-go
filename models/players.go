package models

import (
	"jurrien/dnding-bot/database"

	"github.com/charmbracelet/log"
)

type Player struct {
	Characters *[]Character
	DiscordID  string `gorm:"unique"`
	Name       string
	ID         uint `gorm:"primaryKey;autoIncrement"`
}

func GetAllPlayers(database *database.DB, logger *log.Logger) ([]Player, error) {
	var players []Player
	err := database.GetConnection().Model(&Player{}).Preload("Characters").Find(&players).Error
	if err != nil {
		logger.Error("Error getting all characters")
		return nil, err
	}

	return players, nil
}
