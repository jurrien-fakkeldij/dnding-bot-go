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
	"github.com/google/go-cmp/cmp"
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

func (s *MockSession) GuildMembers(guildID string, after string, limit int, options ...discordgo.RequestOption) (st []*discordgo.Member, err error) {
	///TODO: implement function when used later
	return nil, nil
}

func (s *MockSession) InteractionResponseEdit(interaction *discordgo.Interaction, params *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	///TODO: implement functon when used later
	return nil, nil
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
	scenario.Step(`^the response is$`, s.theResponseIs)
	scenario.Step(`^the response is ephemeral$`, s.theResponseShouldBeEphemeral)
	scenario.Step(`^there is a player record in the database with "([^"]*)"$`, s.thereIsAPlayerRecordInTheDatabaseWith)
	scenario.Step(`^the user with ID "([^"]*)" is registered with the name "([^"]*)"$`, s.theUserWithIDIsRegisteredWithTheName)
	scenario.Step(`^there is no player record in the database with "([^"]*)"$`, s.thereIsNoPlayerRecordInTheDatabaseWith)
	scenario.Step(`^there is a character record in the database for "([^"]*)" with the name "([^"]*)"$`, s.thereIsACharacterRecordInTheDatabaseForWithTheName)
	scenario.Step(`^the user with ID "([^"]*)" has a character with the name "([^"]*)" and tab amount (\d+) registered$`, s.theUserWithIDHasACharacterWithTheNameAndTabAmountRegistered)
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

	if s.Interaction.User == nil {
		s.Interaction.User = &discordgo.User{
			Username: "some_name",
		}
	}

	s.Interaction.User.ID = id

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
	switch commandName {
	case "register_player":
		s.Interaction.Data = discordgo.ApplicationCommandInteractionData{
			Name: "test",
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Name: "player_name", Value: name, Type: discordgo.ApplicationCommandOptionString},
			},
		}
	case "register_character":
		s.Interaction.Data = discordgo.ApplicationCommandInteractionData{
			Name: "test",
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Name: "character_name", Value: name, Type: discordgo.ApplicationCommandOptionString},
			},
		}
	}
	return s.sendCommand(commandName)
}

func (s *CommandSteps) sendCommand(commandName string) error {
	mockSession := &MockSession{}
	s.MockSession = mockSession

	var command *discordgo.ApplicationCommand
	for _, _command := range commands.AllCommands {
		if _command.Name == commandName {
			command = _command
		}
	}

	if command == nil {
		return fmt.Errorf("no command: %s found", commandName)
	}

	fmt.Printf("Command %s interaction %v\n", command.Name, s.Interaction)
	fmt.Printf("Command %s interaction %v\n", command.Name, s.Interaction.Member.User.ID)

	err := commands.AllCommandHandlers[command.Name](mockSession, s.Database, logger, &discordgo.InteractionCreate{
		Interaction: s.Interaction,
	})

	fmt.Printf("ERROR: %v", err)

	return err
}

func (s *CommandSteps) theResponseShouldBe(response string) error {
	if s.MockSession.Response == nil || s.MockSession.Response.Data == nil {
		return fmt.Errorf("no response given")
	}
	if s.MockSession.Response.Data.Content != response {
		if diff := cmp.Diff(response, s.MockSession.Response.Data.Content); diff != "" {
			return fmt.Errorf("response mismatch (-want +got):\n%s", diff)
		}
		return fmt.Errorf("response is not \n%s but \n%s", response, s.MockSession.Response.Data.Content)
	}
	return nil
}

func (s *CommandSteps) theResponseIs(response *godog.DocString) error {
	if s.MockSession.Response == nil || s.MockSession.Response.Data == nil {
		return fmt.Errorf("no response given")
	}

	if s.MockSession.Response.Data.Content != response.Content {
		if diff := cmp.Diff(response.Content, s.MockSession.Response.Data.Content); diff != "" {
			return fmt.Errorf("response mismatch (-want +got):\n%s", diff)
		}
		return fmt.Errorf("response is not \n%s but \n%s", response.Content, s.MockSession.Response.Data.Content)
	}
	return nil
}

func (s *CommandSteps) theResponseShouldBeEphemeral() error {
	if s.MockSession.Response.Data.Flags != discordgo.MessageFlagsEphemeral {
		return fmt.Errorf("response is not ephemiral")
	}
	return nil
}

func (s *CommandSteps) thereIsAPlayerRecordInTheDatabaseWith(player_name string) error {
	var player models.Player

	result := s.Database.Connection.Where("name = ?", player_name).First(&player)
	if result.Error != nil {
		return fmt.Errorf("error geting player with name %s: %v", player_name, result.Error)
	}

	if player.Name != player_name {
		return fmt.Errorf("not sure what happened but found a different name")
	}

	return nil
}

func (s *CommandSteps) thereIsNoPlayerRecordInTheDatabaseWith(player_name string) error {
	var player models.Player

	result := s.Database.Connection.Where("name = ?", player_name).First(&player)
	if result.RowsAffected != 0 {
		return fmt.Errorf("found player record for %s: %v", player_name, player)
	}
	return nil
}

func (s *CommandSteps) theUserWithIDIsRegisteredWithTheName(id string, name string) error {
	player := models.Player{Name: name, DiscordID: id}
	s.Database.Connection.Save(&player)
	return nil
}

func (s *CommandSteps) thereIsACharacterRecordInTheDatabaseForWithTheName(player_name, character_name string) error {
	var player models.Player
	var character models.Character

	result := s.Database.Connection.Where("name = ?", player_name).First(&player)
	if result.RowsAffected == 0 {
		return fmt.Errorf("have not found player record for %s", player_name)
	}

	result = s.Database.Connection.Where("name = ?", character_name).First(&character)

	if result.RowsAffected == 0 {
		return fmt.Errorf("have not found character record with name %s for player %s", character_name, player.Name)
	}

	return nil
}

func (s *CommandSteps) theUserWithIDHasACharacterWithTheNameAndTabAmountRegistered(discordId, characterName string, tabAmount int) error {
	var player models.Player
	var character models.Character

	player = models.Player{Name: "test_name", DiscordID: discordId}
	result := s.Database.Connection.Save(&player)
	if result.RowsAffected == 0 {
		return fmt.Errorf("have not found player record for %s", "test_name")
	}

	character = models.Character{Name: &characterName, Tab: &tabAmount, PlayerID: player.ID}
	result = s.Database.Connection.Save(&character)

	if result.RowsAffected == 0 {
		return fmt.Errorf("could not create character %s with tab %d", *character.Name, *character.Tab)
	}

	return nil
}
