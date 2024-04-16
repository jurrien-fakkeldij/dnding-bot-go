package commands

import "github.com/bwmarrin/discordgo"

type SessionModel interface {
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
}

type CommandFunction func(SessionModel, *discordgo.InteractionCreate) error
