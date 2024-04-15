package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

var dmPermission = false

var Player_commands = []*discordgo.ApplicationCommand{
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

var Player_command_handlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
	"register_player": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
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

		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("You have registered yourself with the name %s", name),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Println("[register-player] response error: ", err)
		}
	},
}

var registeredCommands = make([]*discordgo.ApplicationCommand, len(Player_commands))

func AddPlayerCommands(session *discordgo.Session) error {
	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if h, ok := Player_command_handlers[interaction.ApplicationCommandData().Name]; ok {
			h(session, interaction)
		}
	})

	for index, command := range Player_commands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, "", command)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", command.Name, err)
			return err
		}
		registeredCommands[index] = cmd
	}
	return nil
}

func RemovePlayerCommands(session *discordgo.Session) error {
	for _, command := range registeredCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, "", command.ID)
		if err != nil {
			log.Printf("Cannot delete '%v' command: %v", command.Name, err)
			return err
		}
	}
	return nil
}
