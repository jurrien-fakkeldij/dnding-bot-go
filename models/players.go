package models

type Player struct {
	Characters *[]Character
	DiscordID  string `gorm:"unique"`
	Name       string
	ID         uint `gorm:"primaryKey;autoIncrement"`
}
