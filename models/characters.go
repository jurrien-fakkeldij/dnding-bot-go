package models

type Character struct {
	Expenses *[]Expense `gorm:"many2many:character_expenses;"`
	Name     *string
	Tab      *int  `gorm:"default:0"`
	Xp       *uint `gorm:"default:0"`
	ID       uint  `gorm:"primaryKey;autoIncrement;unique"`
	PlayerID uint
}
