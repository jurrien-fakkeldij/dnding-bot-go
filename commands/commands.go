package commands

import (
	"cmp"
	"jurrien/dnding-bot/database"
	"maps"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

type SessionModel interface {
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
}

type CommandFunction func(SessionModel, *database.DB, *log.Logger, *discordgo.InteractionCreate) error

var (
	commandList = [][]*discordgo.ApplicationCommand{
		HelpCommands,
		PlayerCommands,
		CharacterCommands,
	}

	commandHandlers = []map[string]CommandFunction{
		PlayerCommandHandlers,
		HelpCommandHandlers,
		CharacterCommandHandlers,
	}

	AllCommands        = mergeCommandList()
	AllCommandHandlers = mergeHandlers()

	RegisteredCommands = make([]*discordgo.ApplicationCommand, len(AllCommands))

	DM_ROLE_NAME = "DM"
	dmPermission = false
)

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
