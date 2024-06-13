package models

type Expense struct {
	ID         uint `gorm:"primaryKey;autoIncrement;unique"`
	Name       string
	Characters *[]Character `gorm:"many2many:character_expenses;"`
}

type CharacterExpense struct {
	CharacterID uint `gorm:"primaryKey"`
	ExpenseID   uint `gorm:"primaryKey"`
	Amount      int
	Expense     Expense
	Character   Character
}
