package commands

import (
	"jurrien/dnding-bot/database"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

type SessionModel interface {
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
}

type CommandFunction func(SessionModel, *database.DB, *log.Logger, *discordgo.InteractionCreate) error
