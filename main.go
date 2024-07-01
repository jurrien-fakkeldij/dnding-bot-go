package main

import (
	"context"
	"fmt"
	"jurrien/dnding-bot/commands"
	"jurrien/dnding-bot/database"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

var logger *log.Logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
	TimeFormat:      time.DateTime,
})

var guilds = []*discordgo.Guild{}

func main() {
	logger.Info("Starting server!")
	token := os.Getenv("DISCORD_TOKEN")
	databaseDSN := os.Getenv("DB_DSN")
	ctx := context.Background()
	StartServer(ctx, token, databaseDSN)
}

func StartServer(ctx context.Context, token string, dbDSN string) {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := &database.Config{
		DSN: dbDSN,
	}

	logger.Info("Setting up database", "dsn", dbDSN)
	db, err := database.SetupDB(ctx, config)

	if err != nil {
		logger.Fatal("failed to setup database", "err", err)
	}

	logger.Info("Setup discord bot", "token", token)
	session, err := SetupDiscordBot(token)

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		logger.Info("Logged in as", "user", s.State.User.Username, "#", s.State.User.Discriminator)
	})

	AddingInteractionCreateHandler(session, db, logger)

	session.AddHandler(func(s *discordgo.Session, gc *discordgo.GuildCreate) {
		logger.Info("Guild create", "guild", gc.Guild.Name, "id", gc.Guild.ID)
		if err = AddApplicationCommands(session, gc.Guild); err != nil {
			logger.Fatal("Something went wrong adding commands for the discord bot", "err", err)
		}
		guilds = append(guilds, gc.Guild)
	})

	session.AddHandler(func(s *discordgo.Session, gd *discordgo.GuildDelete) {
		logger.Info("Guild delete")
	})

	if err != nil {
		logger.Fatal("Something went wrong with setting up the discord bot", "err", err)
	}

	if err = StartDiscordBot(session); err != nil {
		logger.Fatal("Something went wrong opening the discord session", "err", err)
	}

	logger.Print("Press Ctrl+C to exit")
	<-ctx.Done()

	_, cancel = context.WithTimeout(context.Background(), 29*time.Second)
	defer cancel()

	RemoveApplicationCommands(session)

	if err = StopDiscordBot(session); err != nil {
		logger.Fatal("Could not stop discord session gracefully: %v", err)
	}

	log.Info("Gracefully shutting down.")
}

func SetupDiscordBot(token string) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Error("Invalid bot parameters", "err", err)
		return nil, err
	}
	logger.Info("Started the go discord bot!")
	return session, nil
}

func AddingInteractionCreateHandler(session *discordgo.Session, database *database.DB, logger *log.Logger) {
	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		commandName := ""

		if interaction.Type == discordgo.InteractionApplicationCommand || interaction.Type == discordgo.InteractionApplicationCommandAutocomplete {
			commandName = interaction.ApplicationCommandData().Name
		} else if interaction.Type == discordgo.InteractionMessageComponent {
			commandName = interaction.MessageComponentData().CustomID
		} else {
			logger.Error("Unknown command type found", "commandType", interaction.Type.String(), "interaction", interaction)
			return
		}

		logger.Info("Executing for command", "command", commandName)
		if h, ok := commands.AllCommandHandlers[commandName]; ok {
			err := h(session, database, logger, interaction)
			if err != nil {
				logger.Error("Error in a command", "command", commandName, "error", err)
			}
		}
	})
}

func AddApplicationCommands(session *discordgo.Session, guild *discordgo.Guild) error {
	for index, command := range commands.AllCommands {
		logger.Info("Adding command", "command", command.Name, "guild", guild.Name)
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, guild.ID, command)
		if err != nil {
			return fmt.Errorf("Cannot create '%v' command: %v", command.Name, err)
		}
		commands.RegisteredCommands[index] = cmd
	}

	return nil
}

func RemoveGuildApplicationCommands(session *discordgo.Session, guild *discordgo.Guild) error {
	for _, command := range commands.RegisteredCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, guild.ID, command.ID)
		if err != nil {
			return fmt.Errorf("Cannot delete '%v' command: %v", command.Name, err)
		}
	}
	return nil
}

func RemoveApplicationCommands(session *discordgo.Session) {
	for _, guild := range guilds {
		err := RemoveGuildApplicationCommands(session, guild)
		if err != nil {
			logger.Error("Error removing commands from guild", "guild", guild.Name, "error", err)
		}
	}
}

func StartDiscordBot(session *discordgo.Session) error {
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages
	return session.Open()
}

func StopDiscordBot(session *discordgo.Session) error {
	return session.Close()
}
