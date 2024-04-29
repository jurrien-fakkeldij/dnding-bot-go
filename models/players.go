package models

type Player struct {
	PlayerID   uint `gorm:"primaryKey;autoIncrement"`
	DiscordID  string
	Name       string
	Characters *[]Character
}
