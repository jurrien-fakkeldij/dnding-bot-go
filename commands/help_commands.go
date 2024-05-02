package commands

import (
	"fmt"
	"jurrien/dnding-bot/database"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

var (
	HelpCommands = []*discordgo.ApplicationCommand{
		{
			Name:         "help",
			Description:  "Lists all the commands available for users",
			DMPermission: &dmPermission,
		},
	}

	HelpCommandHandlers = map[string]CommandFunction{
		"help": func(session SessionModel, database *database.DB, logger *log.Logger, interaction *discordgo.InteractionCreate) error {
			dm_role := false
			for role := range interaction.Member.Roles {
				logger.Info("Checking Role", "role", role, "dm_role", dm_role)
			}

			for _, command := range AllCommands {
				logger.Info("Command information", "command", command.Name, "description", command.Description)
			}

			return nil
		},
	}
)

var registeredHelpCommands = make([]*discordgo.ApplicationCommand, len(HelpCommands))

func AddHelpCommands(session *discordgo.Session, database *database.DB, logger *log.Logger) error {
	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		commandName := interaction.ApplicationCommandData().Name
		if h, ok := PlayerCommandHandlers[commandName]; ok {
			err := h(session, database, logger, interaction)
			if err != nil {
				logger.Error("Error in a player command", "command", commandName, "error", err)
			}
		}
	})

	for index, command := range HelpCommands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, "", command)
		if err != nil {
			return fmt.Errorf("Cannot create '%v' command: %v", command.Name, err)
		}
		registeredHelpCommands[index] = cmd
	}
	return nil
}

func RemoveHelpCommands(session *discordgo.Session) error {
	for _, command := range registeredHelpCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, "", command.ID)
		if err != nil {
			return fmt.Errorf("Cannot delete '%v' command: %v", command.Name, err)
		}
	}
	return nil
}
