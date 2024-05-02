package steps

import (
	"context"
	"fmt"
	"jurrien/dnding-bot/commands"
	"jurrien/dnding-bot/database"
	"jurrien/dnding-bot/models"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/cucumber/godog"
)

var logger *log.Logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
	TimeFormat:      time.DateTime,
	Level:           log.FatalLevel,
})

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
	Database    *database.DB
}

func (s *CommandSteps) InitializeScenario(scenario *godog.ScenarioContext) error {
	scenario.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.Interaction = nil
		var err error
		s.Database, err = database.SetupDB(ctx, &database.Config{DSN: ":memory:"})
		return ctx, err
	})
	scenario.Step(`^the user has a username "([^"]*)"$`, s.theUserHasAUsername)
	scenario.Step(`^the user has an ID "([^"]*)"$`, s.theUserHasAnID)
	scenario.Step(`^the user sends a "([^"]*)" command$`, s.theUserSendsACommand)
	scenario.Step(`^the user sends a "([^"]*)" command with "([^"]*)" name as a parameter$`, s.anyUserSendsACommandWithNameParameter)
	scenario.Step(`^the user sends a "([^"]*)" command without a name as a parameter$`, s.theUserSendsACommandWithoutANameAsAParameter)
	scenario.Step(`^the response "([^"]*)" is given$`, s.theResponseShouldBe)
	scenario.Step(`^the response is ephimeral$`, s.theResponseShouldBeEphimeral)
	scenario.Step(`^there is a player record in the database with "([^"]*)"$`, s.thereIsAPlayerRecordInTheDatabaseWith)
	scenario.Step(`^the user with ID "([^"]*)" is registered with the name "([^"]*)"$`, s.theUserWithIDIsRegisteredWithTheName)
	return nil
}

func (s *CommandSteps) InitializeSuite(suite *godog.TestSuiteContext) error {
	return nil
}

func (s *CommandSteps) theUserHasAUsername(username string) error {
	if s.Interaction == nil {
		s.Interaction = &discordgo.Interaction{}
	}

	if s.Interaction.Member == nil {
		s.Interaction.Member = &discordgo.Member{
			User: &discordgo.User{
				ID: "some_id",
			},
		}
	}

	s.Interaction.Member.User.Username = username

	return nil
}

func (s *CommandSteps) theUserHasAnID(id string) error {
	if s.Interaction == nil {
		s.Interaction = &discordgo.Interaction{}
	}

	if s.Interaction.Member == nil {
		s.Interaction.Member = &discordgo.Member{
			User: &discordgo.User{
				Username: "some_name",
			},
		}
	}

	s.Interaction.Member.User.ID = id

	return nil
}

func (s *CommandSteps) theUserSendsACommand(commandName string) error {
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
	for _, player_command := range commands.AllCommands {
		if player_command.Name == commandName {
			command = player_command
		}
	}
	if command == nil {
		return fmt.Errorf("No command: %s found", commandName)
	}

	return commands.AllCommandHandlers[command.Name](mockSession, s.Database, logger, &discordgo.InteractionCreate{
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

func (s *CommandSteps) thereIsAPlayerRecordInTheDatabaseWith(player_name string) error {
	var player models.Player

	result := s.Database.Connection.Where("name = ?", player_name).First(&player)
	if result.Error != nil {
		return fmt.Errorf("Error geting player with name %s: %v", player_name, result.Error)
	}

	if player.Name != player_name {
		return fmt.Errorf("Not sure what happened but found a different name")
	}

	return nil
}

func (s *CommandSteps) theUserWithIDIsRegisteredWithTheName(id string, name string) error {
	player := models.Player{Name: name, DiscordID: id}
	s.Database.Connection.Save(&player)
	return nil
}
