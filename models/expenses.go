package models

type Expense struct {
	Characters *[]Character `gorm:"many2many:character_expenses;"`
	Name       string
	ID         uint `gorm:"primaryKey;autoIncrement;unique"`
}

type CharacterExpense struct {
	Character   Character
	Expense     Expense
	CharacterID uint `gorm:"primaryKey"`
	ExpenseID   uint `gorm:"primaryKey"`
	Amount      int
}
