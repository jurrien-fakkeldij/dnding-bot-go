package models

type Character struct {
	CharacterID uint `gorm:"primaryKey;autoIncrement;unique"`
	PlayerID    uint
	Name        *string
	Xp          *uint
	Tab         *int
	Class       *string
	Expenses    *[]Expense `gorm:"many2many:character_expenses;"`
}
