package steps

import (
	"context"
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
	Interaction *discordgo.Interaction
	MockSession *MockSession
}

func (s *CommandSteps) InitializeScenario(scenario *godog.ScenarioContext) error {
	scenario.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.Interaction = nil
		return ctx, nil
	})
	scenario.Step(`^the user has a username "([^"]*)" on the server$`, s.theUserHasAUsernameOnTheServer)
	scenario.Step(`^the user sends a "([^"]*)" command with "([^"]*)" name as a parameter$`, s.anyUserSendsACommandWithNameParameter)
	scenario.Step(`^the user sends a "([^"]*)" command without a name as a parameter$`, s.theUserSendsACommandWithoutANameAsAParameter)
	scenario.Step(`^the response "([^"]*)" is given$`, s.theResponseShouldBe)
	scenario.Step(`^the response is ephimeral$`, s.theResponseShouldBeEphimeral)
	return nil
}

func (s *CommandSteps) InitializeSuite(suite *godog.TestSuiteContext) error {
	return nil
}

func (s *CommandSteps) theUserHasAUsernameOnTheServer(username string) error {
	if s.Interaction == nil {
		s.Interaction = &discordgo.Interaction{}
	}

	s.Interaction.Member = &discordgo.Member{
		User: &discordgo.User{
			Username: username,
		},
	}

	return nil
}

func (s *CommandSteps) theUserSendsACommandWithoutANameAsAParameter(commandName string) error {
	mockSession := &MockSession{}
	s.MockSession = mockSession

	if s.Interaction == nil {
		s.Interaction = &discordgo.Interaction{}
	}
	s.Interaction.ID = ""
	s.Interaction.Type = discordgo.InteractionApplicationCommand
	s.Interaction.Data = discordgo.ApplicationCommandInteractionData{}
	return s.sendCommand(commandName)
}

func (s *CommandSteps) anyUserSendsACommandWithNameParameter(commandName, name string) error {
	if s.Interaction == nil {
		s.Interaction = &discordgo.Interaction{}
	}
	s.Interaction.ID = ""
	s.Interaction.Type = discordgo.InteractionApplicationCommand
	s.Interaction.Data = discordgo.ApplicationCommandInteractionData{
		Name: "test",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{Name: "player_name", Value: name, Type: discordgo.ApplicationCommandOptionString},
		},
	}
	return s.sendCommand(commandName)
}

func (s *CommandSteps) sendCommand(commandName string) error {
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
		Interaction: s.Interaction,
	})
}

func (s *CommandSteps) theResponseShouldBe(response string) error {
	if s.MockSession.Response == nil || s.MockSession.Response.Data == nil {
		return fmt.Errorf("No response given")
	}

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
