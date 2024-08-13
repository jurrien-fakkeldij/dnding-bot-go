package commands

import (
	"fmt"
	"jurrien/dnding-bot/database"
	"jurrien/dnding-bot/models"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

var (
	PlayerCommands = []*discordgo.ApplicationCommand{
		{
			Name:         "register_player",
			Description:  "Ability to register yourself as player",
			DMPermission: &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player_name",
					Description: "Name of the player",
					Required:    true,
				},
			},
		},
		///TODO: ADD CREATE PLAYER
		{
			Name:         "create_player",
			Description:  "[DM] Ability to register a player as a DM",
			DMPermission: &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player_name",
					Description: "Name of the player",
					Required:    true,
				},
			},
		},
	}

	PlayerCommandHandlers = map[string]CommandFunction{
		"register_player": func(session SessionModel, database *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			options := interaction.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			name := ""
			option, ok := optionMap["player_name"]
			if ok {
				name = option.StringValue()
			} else {
				name = interaction.Member.Nick
			}

			discordID := interaction.Member.User.ID

			player := models.Player{Name: name, DiscordID: discordID}
			var old_player []models.Player
			database.Connection.Where("discord_id = ?", discordID).Find(&old_player)
			if len(old_player) != 0 {
				player = old_player[0]
				logger.Warn("User is already registered.", "name", player.Name, "discordID", discordID, "new_name", name)
				err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("You already registered %v. If this is not correct please contact the DM or admin", player.Name),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					return fmt.Errorf("[register-player] response error: %v", err)
				}
				return nil
			}

			result := database.GetConnection().Save(&player)
			if result.Error != nil {
				err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Something went wrong, please try again or contact the server admin",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					return fmt.Errorf("[register-player] response error: %v -> %v", err, result.Error)
				}

				return fmt.Errorf("could not save player [%s - %s]: %v", name, discordID, result.Error)
			}
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("You have registered yourself with the name %s", name),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				return fmt.Errorf("[register-player] response error: %v", err)
			}

			return nil
		},
		"create_player": func(session SessionModel, database *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
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
				return fmt.Errorf("could not find dm role or user doesn't have dm role")
			}
			options := interaction.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			name := ""
			option, ok := optionMap["player_name"]
			if ok {
				name = option.StringValue()
			} else {
				name = interaction.Member.Nick
			}

			discordID := interaction.Member.User.ID

			player := models.Player{Name: name, DiscordID: discordID}
			var old_player []models.Player
			database.Connection.Where("discord_id = ?", discordID).Find(&old_player)
			if len(old_player) != 0 {
				player = old_player[0]
				logger.Warn("User is already registered.", "name", player.Name, "discordID", discordID, "new_name", name)
				err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("You already registered %v. If this is not correct please contact the DM or admin", player.Name),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					return fmt.Errorf("[create-player] response error: %v", err)
				}
				return nil
			}

			result := database.GetConnection().Save(&player)
			if result.Error != nil {
				err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Something went wrong, please try again or contact the server admin",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					return fmt.Errorf("[create-player] response error: %v -> %v", err, result.Error)
				}

				return fmt.Errorf("could not save player [%s - %s]: %v", name, discordID, result.Error)
			}
			err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("You have registered yourself with the name %s", name),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				return fmt.Errorf("[create-player] response error: %v", err)
			}

			return nil
		},
	}
)
