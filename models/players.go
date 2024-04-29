package models

type Player struct {
	PlayerID   uint   `gorm:"primaryKey;autoIncrement"`
	DiscordID  string `gorm:"unique"`
	Name       string
	Characters *[]Character
}
