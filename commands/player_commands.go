package commands

import (
	"fmt"
	"jurrien/dnding-bot/database"
	"jurrien/dnding-bot/models"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

var (
	dmPermission   = false
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
					Required:    false,
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
				name = interaction.Member.User.Username
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

			result := database.Connection.Create(&player)
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

				return fmt.Errorf("Could not save player [%s - %s]: %v", name, discordID, result.Error)
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
	}
)

var registeredCommands = make([]*discordgo.ApplicationCommand, len(PlayerCommands))

func AddPlayerCommands(session *discordgo.Session, database *database.DB, logger *log.Logger) error {
	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		commandName := interaction.ApplicationCommandData().Name
		if h, ok := PlayerCommandHandlers[commandName]; ok {
			err := h(session, database, logger, interaction)
			if err != nil {
				logger.Error("Error in a player command", "command", commandName, "error", err)
			}
		}
	})

	for index, command := range PlayerCommands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, "", command)
		if err != nil {
			return fmt.Errorf("Cannot create '%v' command: %v", command.Name, err)
		}
		registeredCommands[index] = cmd
	}
	return nil
}

func RemovePlayerCommands(session *discordgo.Session) error {
	for _, command := range registeredCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, "", command.ID)
		if err != nil {
			return fmt.Errorf("Cannot delete '%v' command: %v", command.Name, err)
		}
	}
	return nil
}
