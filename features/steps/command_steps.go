package steps

import (
	"fmt"
	"jurrien/dnding-bot/commands"

	"github.com/bwmarrin/discordgo"
	"github.com/cucumber/godog"
)

type MockSession struct {
	Response *discordgo.InteractionResponse
}

func (s *MockSession) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	s.Response = resp
	return nil
}

type CommandSteps struct {
	Session     *discordgo.Session
	MockSession *MockSession
}

func (s *CommandSteps) InitializeSuite(suite *godog.TestSuiteContext) error {
	return nil
}

func (s *CommandSteps) anyUserSendsACommandWithName(commandName, name string) error {
	mockSession := &MockSession{}
	s.MockSession = mockSession

	var command *discordgo.ApplicationCommand
	for _, player_command := range commands.PlayerCommands {
		if player_command.Name == commandName {
			command = player_command
		}
	}
	if command == nil {
		return fmt.Errorf("No command: %s found", commandName)
	}

	return commands.PlayerCommandHandlers[command.Name](mockSession, &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:   "",
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "test",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Name: "player_name", Value: name, Type: discordgo.ApplicationCommandOptionString},
				},
			},
		},
	})
}

func (s *CommandSteps) aResponseShouldBeGiven() error {
	if s.MockSession.Response == nil {
		return fmt.Errorf("No response given")
	}
	return nil
}

func (s *CommandSteps) theResponseShouldBe(response string) error {
	if s.MockSession.Response.Data.Content != response {
		return fmt.Errorf("Response is not %s but %s", response, s.MockSession.Response.Data.Content)
	}
	return nil
}

func (s *CommandSteps) theResponseShouldBeEphimeral() error {
	if s.MockSession.Response.Data.Flags != discordgo.MessageFlagsEphemeral {
		return fmt.Errorf("Response is not ephemiral")
	}
	return nil
}

func (s *CommandSteps) InitializeScenario(ctx *godog.ScenarioContext) error {
	ctx.Step(`^any user sends a "([^"]*)" command with "([^"]*)" name$`, s.anyUserSendsACommandWithName)
	ctx.Step(`^a response should be given$`, s.aResponseShouldBeGiven)
	ctx.Step(`^the response should be "([^"]*)"$`, s.theResponseShouldBe)
	ctx.Step(`^the response should be ephimeral$`, s.theResponseShouldBeEphimeral)
	return nil
}
