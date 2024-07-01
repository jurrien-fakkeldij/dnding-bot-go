package commands

import (
	"fmt"
	"jurrien/dnding-bot/database"
	"jurrien/dnding-bot/models"
	"jurrien/dnding-bot/utils"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/olekukonko/tablewriter"
)

var (
	CharacterCommands = []*discordgo.ApplicationCommand{
		{
			Name:         "register_character",
			Description:  "Register your character for your discord user",
			DMPermission: &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "character_name",
					Description: "Name of the character",
					Required:    true,
				},
			},
		}, {
			Name:         "list_my_characters",
			Description:  "Lists your characters",
			DMPermission: &dmPermission,
		}, {
			Name:         "list_all_characters",
			Description:  "[DM] Lists all characters",
			DMPermission: &dmPermission,
		}, {
			Name:         "set_character_tab",
			Description:  "[DM] Set the tab for a specific character",
			DMPermission: &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionInteger,
					Name:         "character",
					Description:  "The name of the character you want to set the tab for.",
					Autocomplete: true,
					Required:     true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "amount",
					Description: "The amount the tab should be",
					Required:    true,
				},
			},
		}, {
			Name:         "pay_my_tab",
			Description:  "Pay the tab of one of your characters",
			DMPermission: &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionInteger,
					Name:         "character",
					Description:  "The name of the character you want to pay the tab for.",
					Autocomplete: true,
					Required:     true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "amount",
					Description: "The amount of the tab, if left empty or 0 you pay all of it. Adding more can be used for later.",
					Required:    false,
				},
			},
		},

		//TODO: CREATE CHARACTER
	}

	CharacterCommandHandlers = map[string]CommandFunction{
		"register_character": func(session SessionModel, database *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			logger.Info("Registering character")
			options := interaction.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			characterName := ""
			option, ok := optionMap["character_name"]
			if ok {
				characterName = option.StringValue() //mandatory name so should be here
			}

			logger.Info("%v", interaction.Member.User)
			discordID := interaction.Member.User.ID

			var player models.Player
			result := database.GetConnection().Where("discord_id = ?", discordID).First(&player)
			if result.RowsAffected == 0 {
				logger.Warn("No player found for user", "discord_id", discordID)
				err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Please register as a player first, then you can register your character",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					logger.Error("Error sending response for register_character", "error", err)
					return fmt.Errorf("Error sending interaction: %v", err)
				}
				return fmt.Errorf("No player found for user: %s", discordID)
			}

			logger.Info("Registering character for user", "user", discordID, "character", characterName, "player", player)
			character := &models.Character{
				Name:     &characterName,
				PlayerID: player.ID,
			}

			logger.Info("Saving character to database", "character", character)

			result = database.GetConnection().Create(character)
			if result.Error != nil || result.RowsAffected == 0 {
				logger.Error("Error saving character to database", "character", character, "error", result.Error)

				err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Something went wrong saving your character, please try again or contact an admin.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					logger.Error("Error sending response for register_character", "error", err)
					return fmt.Errorf("Error sending interaction: %v", err)
				}
				return fmt.Errorf("Error saving character: %v", result.Error)
			}

			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("%s has been added for you.", *character.Name),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				logger.Error("Error sending response for register_character", "error", err)
				return fmt.Errorf("Error sending interaction: %v", err)
			}
			return nil
		},
		"list_my_characters": func(session SessionModel, db *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			discordId := interaction.Member.User.ID
			player := models.Player{
				DiscordID: discordId,
			}

			err := db.GetConnection().Model(&models.Player{}).Preload("Characters").First(&player).Error
			if err != nil {
				logger.Error("Error getting player", "discord_id", discordId)
				interactionErr := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Something went wrong getting your character list",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if interactionErr != nil {
					logger.Error("Error sending response for list_my_characters", "error", err)
					return fmt.Errorf("Error sending interaction: %v", err)
				}

				return err
			}

			tableString := &strings.Builder{}
			table := tablewriter.NewWriter(tableString)
			table.SetBorder(false)
			table.SetCenterSeparator("|")
			table.SetHeader([]string{"Character", "Tab"})
			table.SetAutoWrapText(false)

			for _, character := range *player.Characters {
				table.Append([]string{*character.Name, utils.ToDNDMoneyFormat(*character.Tab)})
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
				logger.Error("Error sending response for list_my_characters", "error", err)
				return fmt.Errorf("Error sending interaction: %v", err)
			}

			return nil
		},
		"list_all_characters": func(session SessionModel, db *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
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
			var players []models.Player
			err = db.GetConnection().Model(&models.Player{}).Preload("Characters").Find(&players).Error

			if err != nil {
				logger.Error("Error getting all characters")
				interactionErr := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Something went wrong getting the character list",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if interactionErr != nil {
					logger.Error("Error sending response for list_all_characters", "error", err)
					return fmt.Errorf("Error sending interaction: %v", err)
				}

				return err
			}

			tableString := &strings.Builder{}
			table := tablewriter.NewWriter(tableString)
			table.SetBorder(false)
			table.SetCenterSeparator("|")
			table.SetHeader([]string{"Player", "Character", "Tab"})
			table.SetAutoWrapText(false)
			for _, player := range players {
				for _, character := range *player.Characters {
					table.Append([]string{player.Name, *character.Name, utils.ToDNDMoneyFormat(*character.Tab)})
				}
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
				logger.Error("Error sending response for list_all_characters", "error", err)
				return fmt.Errorf("Error sending interaction: %v", err)
			}

			return nil
		},
		"set_character_tab": func(session SessionModel, db *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
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
				logger.Info("set_character_tab: Autocomplete")

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
					return fmt.Errorf("DB Error getting characters command: %s error: %v", "autocomplete: set_character_tab", err)
				}

				options := interaction.ApplicationCommandData().Options
				optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
				for _, opt := range options {
					optionMap[opt.Name] = opt
				}
				name := ""
				option, ok := optionMap["character"]
				if ok {
					name = option.Value.(string)
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

			case discordgo.InteractionApplicationCommand:
				logger.Info("set_character_tab: ApplicationCommand")

				options := interaction.ApplicationCommandData().Options
				optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
				for _, opt := range options {
					optionMap[opt.Name] = opt
				}
				characterId := int64(-1)
				amount := 0
				option, ok := optionMap["character"]
				if ok {
					characterId = option.IntValue()
				} else {
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Something went wrong with selecting the character. Please try again or contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("No character selected? character=%d", characterId)
				}

				option, ok = optionMap["amount"]
				if ok {
					amount = int(option.IntValue())
				}
				logger.Info("Saving character", "id", characterId, "amount", amount)

				character := &models.Character{ID: uint(characterId)}

				err := db.GetConnection().Model(&models.Character{}).First(&character).Error
				if err != nil {
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Could not find the correct character. Please contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					logger.Error("Character not found", "character", characterId, "error", err)
					return fmt.Errorf("No character saved, character not found! characterId=%d", characterId)
				}
				character.Tab = &amount
				err = db.GetConnection().Save(character).Error
				if err != nil {
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Something went wrong saving the character. Please try again or contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					logger.Error("Character not saved", "character", characterId, "amount", amount, "error", err)
					return fmt.Errorf("No character saved! characterId=%d amount=%d", characterId, amount)
				}

				return session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Setting tab of player %s to %s.", *character.Name, utils.ToDNDMoneyFormat(*character.Tab)),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			}
			return nil
		},
		"pay_my_tab": func(session SessionModel, db *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			switch interaction.Type {
			case discordgo.InteractionApplicationCommandAutocomplete:
				logger.Info("pay_my_tab: Autocomplete")

				discordId := interaction.Member.User.ID
				player := models.Player{
					DiscordID: discordId,
				}

				err := db.GetConnection().Model(&models.Player{}).Preload("Characters").First(&player).Error
				if err != nil {
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Something went wrong getting characters. Please contact the administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("DB Error getting characters command: %s error: %v", "autocomplete: set_character_tab", err)
				}

				options := interaction.ApplicationCommandData().Options
				optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
				for _, opt := range options {
					optionMap[opt.Name] = opt
				}
				name := ""
				option, ok := optionMap["character"]
				if ok {
					name = option.Value.(string)
				}

				filteredCharacters := []*discordgo.ApplicationCommandOptionChoice{}
				for _, character := range *player.Characters {
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
			case discordgo.InteractionApplicationCommand:
				logger.Info("pay_my_tab: ApplicationCommand")

				options := interaction.ApplicationCommandData().Options
				optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
				for _, opt := range options {
					optionMap[opt.Name] = opt
				}
				characterId := int64(-1)
				amount := 0
				option, ok := optionMap["character"]
				if ok {
					characterId = option.IntValue()
				} else {
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Something went wrong with selecting the character. Please try again or contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return fmt.Errorf("No character selected? character=%d", characterId)
				}

				option, ok = optionMap["amount"]
				if ok {
					amount = int(option.IntValue())
				}
				logger.Info("Saving character", "id", characterId, "amount", amount)
				discordId := interaction.Member.User.ID
				player := models.Player{
					DiscordID: discordId,
				}

				err := db.GetConnection().Model(&models.Player{}).Preload("Characters").First(&player).Error
				if err != nil {
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Player is not recognized in the system. Please register first.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					logger.Error("Player not found", "player", discordId, "error", err)
					return fmt.Errorf("Player not found player_id=%s", discordId)
				}

				found_character := models.Character{}
				characterFound := false
				for _, character := range *player.Characters {
					if character.ID == uint(characterId) {
						found_character = character
						characterFound = true
					}
				}

				if !characterFound {
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "The selected character is not your character.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					logger.Error("Character does not match with player", "id", characterId, "player", player.Name)
					return fmt.Errorf("Character does not match with requesting player")
				}

				if amount == 0 {
					amount = *found_character.Tab
				}

				new_tab := *found_character.Tab - amount
				found_character.Tab = &new_tab
				logger.Info("Paying characters tab", "character", found_character.Name, "id", found_character.ID, "amount", amount)
				err = db.GetConnection().Save(found_character).Error
				if err != nil {
					_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Something went wrong saving the character. Please try again or contact an administrator.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					logger.Error("Character not saved", "character", found_character.ID, "amount", amount, "error", err)
					return fmt.Errorf("No character saved! characterId=%d amount=%d", found_character.ID, amount)
				}

				return session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Payed tab of character %s with the amount %s.", *found_character.Name, utils.ToDNDMoneyFormat(amount)),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			}
			return nil
		},
	}
)
