package models

type Player struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	DiscordID  string `gorm:"unique"`
	Name       string
	Characters *[]Character
}
