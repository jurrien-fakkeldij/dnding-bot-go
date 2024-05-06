package models

type Character struct {
	CharacterID uint `gorm:"primaryKey;autoIncrement;unique"`
	PlayerID    uint
	Name        *string
	Xp          *uint      `gorm:"default:0"`
	Tab         *int       `gorm:"default:0"`
	Expenses    *[]Expense `gorm:"many2many:character_expenses;"`
}
