package utils

import "reflect"

type ExpenseName = string

type expense struct {
	ROOM    ExpenseName
	FOOD    ExpenseName
	TAXES   ExpenseName
	ELVES   ExpenseName
	MEAT    ExpenseName
	INCOME  ExpenseName
	GENERAL ExpenseName
	OTHER   ExpenseName
}

var ExpenseType = expense{
	ROOM:    "room",
	FOOD:    "food",
	TAXES:   "taxes",
	ELVES:   "elves",
	MEAT:    "meat",
	INCOME:  "income",
	GENERAL: "general",
	OTHER:   "other",
}

func (expenseType *expense) GetArrayNames() []string {
	expenseTypeArray := []string{}
	v := reflect.ValueOf(ExpenseType)
	typeOfExpenseType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		expenseTypeArray = append(expenseTypeArray, typeOfExpenseType.Field(i).Name)
	}

	return expenseTypeArray
}
