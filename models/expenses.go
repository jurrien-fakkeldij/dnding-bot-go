package models

type Expense struct {
	ExpenseID  uint `gorm:"primaryKey;autoIncrement;unique"`
	Name       string
	Characters *[]Character `gorm:"many2many:character_expenses;"`
}
