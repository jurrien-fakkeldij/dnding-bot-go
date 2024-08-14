package commands

import (
	"cmp"
	"fmt"
	"jurrien/dnding-bot/database"
	"maps"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

type SessionModel interface {
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
	InteractionResponseEdit(interaction *discordgo.Interaction, params *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error)
	GuildMembers(guildID string, after string, limit int, options ...discordgo.RequestOption) (st []*discordgo.Member, err error)
}

type CommandFunction func(SessionModel, *database.DB, *log.Logger, *discordgo.InteractionCreate) error

var (
	commandList = [][]*discordgo.ApplicationCommand{
		HelpCommands,
		PlayerCommands,
		CharacterCommands,
		ExpenseCommands,
	}

	commandHandlers = []map[string]CommandFunction{
		PlayerCommandHandlers,
		HelpCommandHandlers,
		CharacterCommandHandlers,
		ExpenseCommandHandlers,
	}

	AllCommands        = mergeCommandList()
	AllCommandHandlers = mergeHandlers()

	RegisteredCommands = make([]*discordgo.ApplicationCommand, len(AllCommands))

	DM_ROLE_NAME = "DM"
	dmPermission = false
)

func HasMemberDMRole(session *discordgo.Session, member *discordgo.Member, guildID string, logger *log.Logger) (bool, error) {
	dm_role := false
	for _, roleID := range member.Roles {
		role, err := session.State.Role(guildID, roleID)
		if err != nil {
			return dm_role, fmt.Errorf("Something went wrong checking the role for the help interaction: %v", err)
		}
		if role.Name == DM_ROLE_NAME {
			dm_role = true
		}
		logger.Info("Checking Role", "role", role, "dm_role", dm_role)
	}

	return dm_role, nil
}

func cmpCommands(a, b *discordgo.ApplicationCommand) int {
	return cmp.Compare(a.Name, b.Name)
}

func mergeCommandList() []*discordgo.ApplicationCommand {
	mergedCommands := []*discordgo.ApplicationCommand{}
	for _, commandList := range commandList {
		mergedCommands = append(mergedCommands, commandList...)
	}
	slices.SortFunc(mergedCommands, cmpCommands)
	return mergedCommands
}

func mergeHandlers() map[string]CommandFunction {
	mergedHandlers := map[string]CommandFunction{}

	for _, elem := range commandHandlers {
		maps.Copy(mergedHandlers, elem)
	}
	return mergedHandlers
}
