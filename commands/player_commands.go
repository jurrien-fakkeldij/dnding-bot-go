package commands

import (
	"fmt"
	"jurrien/dnding-bot/database"
	"jurrien/dnding-bot/models"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

var dmPermission = false

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
					Required:    false,
				},
			},
		},
	}

	PlayerCommandHandlers = map[string]CommandFunction{
		"register_player": func(session SessionModel, database *database.DB, interaction *discordgo.InteractionCreate) error {
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

			player := models.Player{Name: name, DiscordID: interaction.Member.User.ID}

			result := database.Connection.Create(&player)
			if result.Error != nil {
				return fmt.Errorf("[register-player] database error: %v", result.Error)
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
			err := h(session, database, interaction)
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
