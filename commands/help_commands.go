package commands

import (
	"fmt"
	"jurrien/dnding-bot/database"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/olekukonko/tablewriter"
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
			dm_role, err := HasMemberDMRole(session.(*discordgo.Session), interaction.Member, interaction.GuildID, logger)
			if err != nil {
				return err
			}
			tableString := &strings.Builder{}
			table := tablewriter.NewWriter(tableString)
			table.SetBorder(false)
			table.SetCenterSeparator("|")
			table.SetHeader([]string{"Command", "Description"})
			table.SetAutoWrapText(false)

			for _, command := range AllCommands {
				logger.Info("Command information", "command", command.Name, "description", command.Description)
				if (strings.Contains(command.Name, "[DM]") && dm_role) || !strings.Contains(command.Name, "[DM]") {
					table.Append([]string{command.Name, command.Description})
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
				return fmt.Errorf("error sending response for help command: %v", err)
			}

			return nil
		},
	}
)
