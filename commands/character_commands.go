package commands

import (
	"fmt"
	"jurrien/dnding-bot/database"
	"jurrien/dnding-bot/models"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
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
		},
	}

	CharacterCommandHandlers = map[string]CommandFunction{
		"register_character": func(session SessionModel, database *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			logger.Info("Registering player")
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

			discordID := interaction.User.ID

			var player models.Player
			result := database.Connection.Where("discord_id = ?", discordID).First(&player)
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
			}

			logger.Info("Registering character for user", "user", discordID, "character", characterName, "player", player)
			character := &models.Character{
				Name:     &characterName,
				PlayerID: player.PlayerID,
			}

			logger.Info("Saving character to database", "character", character)

			result = database.Connection.Create(character)
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
	}
)
