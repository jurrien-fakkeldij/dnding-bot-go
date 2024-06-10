package utils

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
