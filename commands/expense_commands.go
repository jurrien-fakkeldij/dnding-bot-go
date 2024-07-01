package commands

import (
	"fmt"
	"jurrien/dnding-bot/database"
	"jurrien/dnding-bot/models"
	"jurrien/dnding-bot/utils"
	"math"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/olekukonko/tablewriter"
)

const INTEREST_RATE = float64(0.05)

var (
	newTabCharacters []models.Character
	wg               sync.WaitGroup
	ExpenseCommands  = []*discordgo.ApplicationCommand{
		{
			Name:         "add_expense_to_character",
			Description:  "[DM] Add a specific expense to a specific character.",
			DMPermission: &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionInteger,
					Name:         "character",
					Description:  "The character for whom you want to add an expense.",
					Required:     true,
					Autocomplete: true,
				}, {
					Type:         discordgo.ApplicationCommandOptionInteger,
					Name:         "expense",
					Description:  "The expense you want to add to a character.",
					Required:     true,
					Autocomplete: true,
				}, {
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "amount",
					Description: "The amount of the expense.",
					Required:    true,
				},
			},
		},
		{
			Name:         "list_expenses",
			Description:  "[DM] List the expenses of a certain character.",
			DMPermission: &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionInteger,
					Name:         "character",
					Description:  "The character for whom you want to list the expenses.",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		{
			Name:         "list_all_expenses",
			Description:  "[DM] List the expenses of all characters.",
			DMPermission: &dmPermission,
		},
		{
			Name:         "calculate_expenses",
			Description:  "[DM] Calculate the expenses for all the characters for this week.",
			DMPermission: &dmPermission,
		},
	}
	ExpenseCommandHandlers = map[string]CommandFunction{
		"add_expense_to_character": func(session SessionModel, db *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			dm_role, err := HasMemberDMRole(session.(*discordgo.Session), interaction.Member, interaction.GuildID, logger)
			if err != nil || !dm_role {
				logger.Warn("Error or could not find dm_role", "error", err, "dm_role", dm_role, "user", interaction.Member.Nick)
				_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You can not run this command, your are not recognized as a DM.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return fmt.Errorf("Could not find dm role or user doesn't have dm role")
			}

			switch interaction.Type {
			case discordgo.InteractionApplicationCommandAutocomplete:
				logger.Info("add_expense_to_character: Autocomplete")

				options := interaction.ApplicationCommandData().Options
				for _, option := range options {
					if option.Focused {
						if option.Name == "character" {
							name := option.Value.(string)

							var characters []models.Character
							err = db.GetConnection().Find(&characters).Error
							if err != nil {
								_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
									Type: discordgo.InteractionResponseChannelMessageWithSource,
									Data: &discordgo.InteractionResponseData{
										Content: "Something went wrong getting characters. Please contact the administrator.",
										Flags:   discordgo.MessageFlagsEphemeral,
									},
								})
								return fmt.Errorf("DB Error getting characters command: %s error: %v", "autocomplete: add_expense_to_character", err)
							}

							filteredCharacters := []*discordgo.ApplicationCommandOptionChoice{}
							for _, character := range characters {
								if strings.HasPrefix(*character.Name, name) || name == "" {
									filteredCharacters = append(filteredCharacters, &discordgo.ApplicationCommandOptionChoice{
										Name:  *character.Name,
										Value: character.ID,
									})
								}
							}

							err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionApplicationCommandAutocompleteResult,
								Data: &discordgo.InteractionResponseData{
									Choices: filteredCharacters,
								},
							})
							if err != nil {
								logger.Error("Error sending response for list_all_characters", "error", err)
								return fmt.Errorf("Error sending interaction: %v", err)
							}
						} else if option.Name == "expense" {
							name := option.Value.(string)
							var expenses []models.Expense
							err = db.GetConnection().Find(&expenses).Error
							if err != nil {
								_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
									Type: discordgo.InteractionResponseChannelMessageWithSource,
									Data: &discordgo.InteractionResponseData{
										Content: "Something went wrong getting expenses. Please contact the administrator.",
										Flags:   discordgo.MessageFlagsEphemeral,
									},
								})
								return fmt.Errorf("DB Error getting expenses command: %s error: %v", "autocomplete: add_expense_to_character", err)
							}
							filteredExpenses := []*discordgo.ApplicationCommandOptionChoice{}
							for _, expense := range expenses {
								if strings.HasPrefix(expense.Name, name) || name == "" {
									filteredExpenses = append(filteredExpenses, &discordgo.ApplicationCommandOptionChoice{
										Name:  expense.Name,
										Value: expense.ID,
									})
								}
							}

							err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionApplicationCommandAutocompleteResult,
								Data: &discordgo.InteractionResponseData{
									Choices: filteredExpenses,
								},
							})
							if err != nil {
								logger.Error("Error sending response for list_all_characters", "error", err)
								return fmt.Errorf("Error sending interaction: %v", err)
							}
						}
					}
				}
			case discordgo.InteractionApplicationCommand:
				logger.Info("add_expense_to_character: ApplicationCommand")

				options := interaction.ApplicationCommandData().Options
				optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
				for _, opt := range options {
					optionMap[opt.Name] = opt
				}
				character_id := -1
				expense_id := -1
				amount := 0
				option, ok := optionMap["character"]
				if ok {
					character_id = int(option.IntValue())
				}

				option, ok = optionMap["expense"]
				if ok {
					expense_id = int(option.IntValue())
				}

				option, ok = optionMap["amount"]
				if ok {
					amount = int(option.IntValue())
				}

				if character_id == -1 || expense_id == -1 {
					logger.Error("Trying to execute add_expense_to_character, wrong request", "character_id", character_id, "expense_id", expense_id, "amount", amount)

					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Something is wrong with the request, either the expense or character is not correct.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("add_expense_to_character went wrong - character: %d, expense:%d", character_id, expense_id)
				}
				expense := &models.Expense{ID: uint(expense_id)}
				character := &models.Character{ID: uint(character_id)}

				errExpense := db.GetConnection().Model(models.Expense{}).First(expense).Error
				errCharacter := db.GetConnection().Model(models.Character{}).Preload("Expenses").First(character).Error

				if errExpense != nil {
					logger.Error("Database problem getting expense", "expense_id", expense_id, "error", errExpense)
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Given expense is not found within the database or some error occured. Please contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("add_expense_to_character went wrong - character: %d, expense:%d, error:%v", character_id, expense_id, errExpense)
				}

				if errCharacter != nil {
					logger.Error("Database problem getting character", "character_id", character_id, "error", errCharacter)
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Given character is not found within the database or some error occured. Please contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("add_expense_to_character went wrong - character: %d, expense:%d, error:%v", character_id, expense_id, errCharacter)
				}

				var charExpenses []*models.CharacterExpense
				err = db.GetConnection().Where("character_id = ?", character_id).Find(&charExpenses).Error

				if err != nil {
					logger.Error("Database problem getting character expenses", "character_id", character_id, "error", err)
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Database error occured. Please contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("add_expense_to_character went wrong - character: %d, expense:%d, error:%v", character_id, expense_id, err)
				}

				expenseFound := false
				for _, charExpense := range charExpenses {
					if charExpense.ExpenseID == expense.ID {
						charExpense.Amount = amount
						expenseFound = true
					}
				}

				if !expenseFound {
					charExpense := &models.CharacterExpense{CharacterID: character.ID, ExpenseID: expense.ID, Amount: amount}
					charExpenses = append(charExpenses, charExpense)
				}

				err = db.GetConnection().Save(charExpenses).Error
				if err != nil {
					logger.Error("Database problem saving character expenses", "character_id", character_id, "error", err)
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Database error occured saving the expenses to the character. Please contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("add_expense_to_character went wrong - character: %d, expense:%d, error:%v", character_id, expense_id, err)
				}

				_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Saved expense %s to character %s with amount %d.", expense.Name, *character.Name, amount),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			}
			return nil
		},
		"list_expenses": func(session SessionModel, db *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			dm_role, err := HasMemberDMRole(session.(*discordgo.Session), interaction.Member, interaction.GuildID, logger)
			if err != nil || !dm_role {
				logger.Warn("Error or could not find dm_role", "error", err, "dm_role", dm_role, "user", interaction.Member.Nick)
				_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You can not run this command, your are not recognized as a DM.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return fmt.Errorf("Could not find dm role or user doesn't have dm role")
			}

			switch interaction.Type {
			case discordgo.InteractionApplicationCommandAutocomplete:
				logger.Info("list_expenses: Autocomplete")

				options := interaction.ApplicationCommandData().Options
				for _, option := range options {
					if option.Focused {
						if option.Name == "character" {
							name := option.Value.(string)

							var characters []models.Character
							err = db.GetConnection().Find(&characters).Error
							if err != nil {
								_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
									Type: discordgo.InteractionResponseChannelMessageWithSource,
									Data: &discordgo.InteractionResponseData{
										Content: "Something went wrong getting characters. Please contact the administrator.",
										Flags:   discordgo.MessageFlagsEphemeral,
									},
								})
								return fmt.Errorf("DB Error getting characters command: %s error: %v", "autocomplete: list_expenses", err)
							}

							filteredCharacters := []*discordgo.ApplicationCommandOptionChoice{}
							for _, character := range characters {
								if strings.HasPrefix(*character.Name, name) || name == "" {
									filteredCharacters = append(filteredCharacters, &discordgo.ApplicationCommandOptionChoice{
										Name:  *character.Name,
										Value: character.ID,
									})
								}
							}

							err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionApplicationCommandAutocompleteResult,
								Data: &discordgo.InteractionResponseData{
									Choices: filteredCharacters,
								},
							})
							if err != nil {
								logger.Error("Error sending response for list_expenses", "error", err)
								return fmt.Errorf("Error sending interaction: %v", err)
							}
						}
					}
				}
			case discordgo.InteractionApplicationCommand:

				options := interaction.ApplicationCommandData().Options
				optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
				for _, opt := range options {
					optionMap[opt.Name] = opt
				}
				character_id := -1
				option, ok := optionMap["character"]
				if ok {
					character_id = int(option.IntValue())
				}

				if character_id == -1 {
					logger.Error("Trying to execute add_expense_to_character, wrong request", "character_id", character_id)

					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Something is wrong with the request, the character is not correct.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("add_expense_to_character went wrong - character: %d", character_id)
				}

				var charExpenses []*models.CharacterExpense
				//				err = db.GetConnection().Joins("JOIN characters ON characters.id = character_expenses.character_id").Joins("JOIN expenses ON expenses.id = character_expenses.expense_id").Where("character_id = ?", character_id).Find(&charExpenses).Error
				err = db.GetConnection().Preload("Character").Preload("Expense").Where("character_id = ?", character_id).Find(&charExpenses).Error

				if err != nil {
					logger.Error("Database problem getting character expenses", "character_id", character_id, "error", err)
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Database error occured. Please contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("add_expense_to_character went wrong - character: %d, error:%v", character_id, err)
				}

				totalExpenses := 0
				expenseNames := utils.ExpenseType.GetArrayNames()
				expenseAmounts := []string{}

				tableString := &strings.Builder{}
				table := tablewriter.NewWriter(tableString)
				table.SetBorder(false)
				table.SetCenterSeparator("|")
				table.SetAutoWrapText(false)

				for _, expense := range expenseNames {
					currentExpenseAmount := 0
					for _, charExpense := range charExpenses {
						if strings.EqualFold(charExpense.Expense.Name, expense) {
							currentExpenseAmount = charExpense.Amount
						}
					}
					expenseAmounts = append(expenseAmounts, fmt.Sprintf("%d", currentExpenseAmount))
					totalExpenses += currentExpenseAmount
				}
				expenseNames = append(expenseNames, "TOTAL")
				expenseAmounts = append(expenseAmounts, fmt.Sprintf("%d", totalExpenses))

				table.SetHeader(expenseNames)
				table.Append(expenseAmounts)

				table.Render()

				err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("```%s```", tableString.String()),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					logger.Error("Error sending response for list_my_characters", "error", err)
					return fmt.Errorf("Error sending interaction: %v", err)
				}
			}
			return nil
		},
		"list_all_expenses": func(session SessionModel, db *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			dm_role, err := HasMemberDMRole(session.(*discordgo.Session), interaction.Member, interaction.GuildID, logger)
			if err != nil || !dm_role {
				logger.Warn("Error or could not find dm_role", "error", err, "dm_role", dm_role, "user", interaction.Member.Nick)
				_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You can not run this command, your are not recognized as a DM.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return fmt.Errorf("Could not find dm role or user doesn't have dm role")
			}

			var characters []models.Character

			err = db.GetConnection().Find(&characters).Error

			if err != nil {
				logger.Error("Database problem getting characters")
				_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Database error occured. Please contact an administrator.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return fmt.Errorf("list_all_expenses went wrong - error:%v", err)
			}
			expenseNames := []string{"NAME"}
			expenseNames = append(expenseNames, utils.ExpenseType.GetArrayNames()...)
			expenseNames = append(expenseNames, "TOTAL")

			tableString := &strings.Builder{}
			table := tablewriter.NewWriter(tableString)
			table.SetBorder(false)
			table.SetCenterSeparator("|")
			table.SetAutoWrapText(false)
			table.SetHeader(expenseNames)

			for _, character := range characters {
				expenseAmounts := []string{}
				character_id := character.ID

				var charExpenses []*models.CharacterExpense
				err = db.GetConnection().Preload("Character").Preload("Expense").Where("character_id = ?", character_id).Find(&charExpenses).Error

				if err != nil {
					logger.Error("Database problem getting character expenses", "character_id", character_id, "error", err)
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Database error occured. Please contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("list_all_expenses went wrong - character: %d, error:%v", character_id, err)
				}

				totalExpenses := 0
				expenseAmounts = append(expenseAmounts, *character.Name)

				for _, expense := range expenseNames {
					if strings.EqualFold(expense, "total") || strings.EqualFold(expense, "name") {
						continue
					}
					currentExpenseAmount := 0
					for _, charExpense := range charExpenses {
						if strings.EqualFold(charExpense.Expense.Name, expense) {
							currentExpenseAmount = charExpense.Amount
						}
					}
					expenseAmounts = append(expenseAmounts, fmt.Sprintf("%d", currentExpenseAmount))
					totalExpenses += currentExpenseAmount
				}
				expenseAmounts = append(expenseAmounts, fmt.Sprintf("%d", totalExpenses))

				table.Append(expenseAmounts)
			}
			table.Render()

			err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("```%s```", tableString.String()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.Error("Error sending response for list_all_expenses", "error", err)
				return fmt.Errorf("Error sending interaction: %v", err)
			}
			return nil
		},
		"calculate_expenses": func(session SessionModel, db *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			dm_role, err := HasMemberDMRole(session.(*discordgo.Session), interaction.Member, interaction.GuildID, logger)
			if err != nil || !dm_role {
				logger.Warn("Error or could not find dm_role", "error", err, "dm_role", dm_role, "user", interaction.Member.Nick)
				_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You can not run this command, your are not recognized as a DM.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return fmt.Errorf("Could not find dm role or user doesn't have dm role")
			}
			var characters []models.Character
			newTabCharacters = []models.Character{}

			err = db.GetConnection().Find(&characters).Error

			if err != nil {
				logger.Error("Database problem getting characters")
				_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Database error occured. Please contact an administrator.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return fmt.Errorf("calculate_expenses went wrong - error:%v", err)
			}

			headerNames := []string{"Name", "Current", "Interest", "Expenses", "New"}

			tableString := &strings.Builder{}
			table := tablewriter.NewWriter(tableString)
			table.SetBorder(false)
			table.SetCenterSeparator("|")
			table.SetAutoWrapText(false)
			table.SetHeader(headerNames)

			for _, character := range characters {
				character_id := character.ID

				var charExpenses []*models.CharacterExpense
				err = db.GetConnection().Preload("Character").Preload("Expense").Where("character_id = ?", character_id).Find(&charExpenses).Error

				if err != nil {
					logger.Error("Database problem getting character expenses", "character_id", character_id, "error", err)
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Database error occured. Please contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("calculate_expenses went wrong - character: %d, error:%v", character_id, err)
				}

				totalExpenses := 0
				for _, charExpense := range charExpenses {
					totalExpenses += charExpense.Amount
				}

				interest := math.Floor(math.Max(float64(*character.Tab)*INTEREST_RATE, 0))
				new_tab := *character.Tab + int(interest) + totalExpenses

				table.Append([]string{*character.Name, utils.ToDNDMoneyFormat(*character.Tab), utils.ToDNDMoneyFormat(int(interest)), utils.ToDNDMoneyFormat(totalExpenses), utils.ToDNDMoneyFormat(new_tab)})

				character.Tab = &new_tab

				newTabCharacters = append(newTabCharacters, character)
			}
			table.Render()

			confirmBtn := discordgo.Button{
				CustomID: "calculate_confirm",
				Label:    "Yes",
				Style:    discordgo.SuccessButton,
				Disabled: false,
			}

			cancelBtn := discordgo.Button{
				CustomID: "calculate_cancel",
				Label:    "No",
				Style:    discordgo.DangerButton,
				Disabled: false,
			}

			actionRow := discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{cancelBtn, confirmBtn},
			}

			err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content:    fmt.Sprintf("```%s``` Are you sure you want to execute above expenses calculation for this week?", tableString.String()),
					Flags:      discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsCrossPosted,
					Components: []discordgo.MessageComponent{actionRow},
				},
			})
			if err != nil {
				logger.Error("Error sending response for calculate_expenses", "error", err)
				return fmt.Errorf("Error sending interaction: %v", err)
			}

			return nil
		},
		"calculate_cancel": func(session SessionModel, db *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			logger.Warn("Canceling new tabs")

			_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content: "Stopped calculating, moving to the next patron.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return nil
		},
		"calculate_confirm": func(session SessionModel, db *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			logger.Info("Saving calculations")
			err := db.GetConnection().Save(newTabCharacters).Error
			if err != nil {
				logger.Error("Database problem saving character tabs")
				_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Content: "Database error occured. Please contact an administrator.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return fmt.Errorf("calculate_expenses save tabs went wrong - error:%v", err)
			}

			_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content: "Updated the ledger thank you for doing business at this establishment.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return nil
		},
	}
)
