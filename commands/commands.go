package commands

import (
	"jurrien/dnding-bot/database"
	"maps"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

type SessionModel interface {
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
}

type CommandFunction func(SessionModel, *database.DB, *log.Logger, *discordgo.InteractionCreate) error

var allCommandHandlers = map[string]CommandFunction{}

var AllCommands = append(HelpCommands, PlayerCommands...)
var AllCommandHandlers = appendMaps(allCommandHandlers, PlayerCommandHandlers, HelpCommandHandlers)

var RegisteredCommands = make([]*discordgo.ApplicationCommand, len(AllCommands))

func appendMaps(dst map[string]CommandFunction, commandFunctions ...map[string]CommandFunction) map[string]CommandFunction {
	for _, elem := range commandFunctions {
		maps.Copy(dst, elem)
	}
	return dst
}
